package swap

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/holiday"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	CreateSwap(swap entity.Swap, requesterID uint, requesterType entity.UserType) (*entity.Swap, *ierr.RestErr)
	ApproveOrRejectSwap(swapID, approverID uint, newStatus entity.SwapStatus) (*entity.Swap, *ierr.RestErr)
	FindSwapByID(id uint) (*entity.Swap, *ierr.RestErr)
	FindSwapsForUser(userID uint, statusFilter string) ([]entity.Swap, *ierr.RestErr)
	FindAllSwaps() ([]entity.Swap, *ierr.RestErr)
	DeleteSwap(id, requesterID uint, requesterType entity.UserType) *ierr.RestErr
}

type service struct {
	swapRepo    Repository
	userRepo    user.Repository
	holidayRepo holiday.Repository
}

func NewService(swapRepo Repository, userRepo user.Repository, holidayRepo holiday.Repository) Service {
	return &service{
		swapRepo:    swapRepo,
		userRepo:    userRepo,
		holidayRepo: holidayRepo,
	}
}

var shiftTimings = map[entity.ShiftName]struct {
	start time.Duration
	end   time.Duration
}{
	entity.ShiftMorning:   {start: 6 * time.Hour, end: 14 * time.Hour},
	entity.ShiftAfternoon: {start: 14 * time.Hour, end: 22 * time.Hour},
	entity.ShiftNight:     {start: 22 * time.Hour, end: 30 * time.Hour},
}

func (s *service) CreateSwap(swap entity.Swap, requesterID uint, requesterType entity.UserType) (*entity.Swap, *ierr.RestErr) {
	swap.RequesterID = requesterID

	if requesterType == entity.UserTypeMaster {
		swap.Status = entity.StatusApproved
		now := time.Now().UTC()
		swap.ApprovedAt = &now
		swap.ApprovedByID = &requesterID
	} else {
		swap.Status = entity.StatusPending
	}

	if err := s.validateSwap(&swap); err != nil {
		return nil, err
	}
	if err := s.swapRepo.CreateSwap(&swap); err != nil {
		return nil, ierr.NewInternalServerError("error creating swap request")
	}
	newSwap, err := s.swapRepo.FindSwapByID(swap.ID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching newly created swap")
	}
	return newSwap, nil
}

func (s *service) validateSwap(swap *entity.Swap) *ierr.RestErr {
	requester, err := s.userRepo.FindUserByID(swap.RequesterID)
	if err != nil {
		return ierr.NewBadRequestError("requester not found")
	}

	if swap.InvolvedCollaboratorID != nil {
		involved, err := s.userRepo.FindUserByID(*swap.InvolvedCollaboratorID)
		if err != nil {
			return ierr.NewBadRequestError("involved collaborator not found")
		}
		if requester.Team != involved.Team {
			return ierr.NewBadRequestError("swaps can only occur between members of the same team")
		}
	}

	originalIsWeekend := isWeekend(swap.OriginalDate)
	newIsWeekend := isWeekend(swap.NewDate)
	if originalIsWeekend != newIsWeekend {
		return ierr.NewBadRequestError("weekday off can only be swapped for another weekday, and weekend off for another weekend day")
	}

	if err := s.checkRestInterval(requester, swap); err != nil {
		return err
	}

	return nil
}

func (s *service) checkRestInterval(user *entity.User, swap *entity.Swap) *ierr.RestErr {
	dayBefore := swap.NewDate.AddDate(0, 0, -1)
	dayAfter := swap.NewDate.AddDate(0, 0, 1)

	shiftBefore, isWorkDayBefore, err := s.getShiftForDay(user, dayBefore)
	if err != nil {
		return ierr.NewInternalServerError("could not determine schedule for previous day")
	}
	if isWorkDayBefore {
		endOfShiftBefore := dayBefore.Add(shiftTimings[shiftBefore].end)
		startOfNewShift := swap.NewDate.Add(shiftTimings[swap.NewShift].start)
		if startOfNewShift.Sub(endOfShiftBefore) < 11*time.Hour {
			return ierr.NewBadRequestError("the proposed swap violates the minimum 11-hour rest interval with the previous day's shift")
		}
	}

	shiftAfter, isWorkDayAfter, err := s.getShiftForDay(user, dayAfter)
	if err != nil {
		return ierr.NewInternalServerError("could not determine schedule for next day")
	}
	if isWorkDayAfter {
		endOfNewShift := swap.NewDate.Add(shiftTimings[swap.NewShift].end)
		startOfShiftAfter := dayAfter.Add(shiftTimings[shiftAfter].start)
		if startOfShiftAfter.Sub(endOfNewShift) < 11*time.Hour {
			return ierr.NewBadRequestError("the proposed swap violates the minimum 11-hour rest interval with the next day's shift")
		}
	}
	return nil
}

func (s *service) getShiftForDay(u *entity.User, date time.Time) (entity.ShiftName, bool, error) {
	swaps, err := s.swapRepo.FindApprovedSwapsForDateRange(u.ID, date, date)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", false, err
	}
	if len(swaps) > 0 {
		for _, swap := range swaps {
			isSameNewDate := swap.NewDate.Year() == date.Year() && swap.NewDate.YearDay() == date.YearDay()
			isSameOriginalDate := swap.OriginalDate.Year() == date.Year() && swap.OriginalDate.YearDay() == date.YearDay()
			isRequester := swap.RequesterID == u.ID
			isInvolved := swap.InvolvedCollaboratorID != nil && *swap.InvolvedCollaboratorID == u.ID

			if isSameNewDate && isRequester {
				return "", false, nil
			}
			if isSameOriginalDate && isRequester {
				return swap.NewShift, true, nil
			}
			if isSameNewDate && isInvolved {
				return swap.NewShift, true, nil
			}
			if isSameOriginalDate && isInvolved {
				return "", false, nil
			}
		}
	}

	isHoliday, err := s.holidayRepo.IsHoliday(date)
	if err != nil {
		return "", false, err
	}
	if isHoliday {
		return "", false, nil
	}

	if isRegularDayOff(date, u) {
		return "", false, nil
	}

	return u.Shift, true, nil
}

func isWeekend(date time.Time) bool {
	wd := date.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

func isRegularDayOff(date time.Time, user *entity.User) bool {
	weekdayMap := map[time.Weekday]entity.WeekdayName{
		time.Monday:    entity.WeekdayMonday,
		time.Tuesday:   entity.WeekdayTuesday,
		time.Wednesday: entity.WeekdayWednesday,
		time.Thursday:  entity.WeekdayThursday,
		time.Friday:    entity.WeekdayFriday,
	}
	if user.WeekdayOff == weekdayMap[date.Weekday()] {
		return true
	}

	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		if user.InitialWeekendOff == "" {
			return false
		}
		firstWeekendOffDay := time.Sunday
		if user.InitialWeekendOff == entity.WeekendSaturday {
			firstWeekendOffDay = time.Saturday
		}
		firstOccurrence := user.CreatedAt
		for firstOccurrence.Weekday() != firstWeekendOffDay {
			firstOccurrence = firstOccurrence.AddDate(0, 0, 1)
		}
		firstOccurrence = time.Date(firstOccurrence.Year(), firstOccurrence.Month(), firstOccurrence.Day(), 0, 0, 0, 0, time.UTC)
		currentDayOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if currentDayOnly.Before(firstOccurrence) {
			return false
		}
		daysDiff := currentDayOnly.Sub(firstOccurrence).Hours() / 24
		weekDiff := int(daysDiff / 7)
		currentWeekendOffDay := firstWeekendOffDay
		if weekDiff%2 != 0 {
			if firstWeekendOffDay == time.Saturday {
				currentWeekendOffDay = time.Sunday
			} else {
				currentWeekendOffDay = time.Saturday
			}
		}
		if date.Weekday() == currentWeekendOffDay {
			return true
		}
	}
	return false
}

func (s *service) ApproveOrRejectSwap(swapID, approverID uint, newStatus entity.SwapStatus) (*entity.Swap, *ierr.RestErr) {
	swap, restErr := s.FindSwapByID(swapID)
	if restErr != nil {
		return nil, restErr
	}
	approver, err := s.userRepo.FindUserByID(approverID)
	if err != nil {
		return nil, ierr.NewInternalServerError("approver not found")
	}
	requester, err := s.userRepo.FindUserByID(swap.RequesterID)
	if err != nil {
		return nil, ierr.NewInternalServerError("requester not found")
	}
	if approver.UserType != entity.UserTypeMaster && (requester.SuperiorID == nil || *requester.SuperiorID != approverID) {
		return nil, ierr.NewForbiddenError("you do not have permission to approve this request")
	}
	swap.Status = newStatus
	now := time.Now().UTC()
	if newStatus == entity.StatusApproved {
		swap.ApprovedAt = &now
		swap.ApprovedByID = &approverID
	} else {
		swap.ApprovedAt = nil
		swap.ApprovedByID = nil
	}
	if err := s.swapRepo.UpdateSwap(swap); err != nil {
		return nil, ierr.NewInternalServerError(fmt.Sprintf("error updating swap status: %v", err))
	}
	return swap, nil
}

func (s *service) FindSwapByID(id uint) (*entity.Swap, *ierr.RestErr) {
	swap, err := s.swapRepo.FindSwapByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("swap not found")
		}
		return nil, ierr.NewInternalServerError("error finding swap")
	}
	return swap, nil
}

func (s *service) FindSwapsForUser(userID uint, statusFilter string) ([]entity.Swap, *ierr.RestErr) {
	swaps, err := s.swapRepo.FindSwapsByUserID(userID, statusFilter)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching swaps for user")
	}
	return swaps, nil
}

func (s *service) FindAllSwaps() ([]entity.Swap, *ierr.RestErr) {
	swaps, err := s.swapRepo.FindAllSwaps()
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching all swaps")
	}
	return swaps, nil
}

func (s *service) DeleteSwap(id, requesterID uint, requesterType entity.UserType) *ierr.RestErr {
	swap, err := s.FindSwapByID(id)
	if err != nil {
		return err
	}
	if requesterType != entity.UserTypeMaster && swap.RequesterID != requesterID {
		return ierr.NewForbiddenError("you can only delete your own swap requests")
	}

	if swap.Status == entity.StatusApproved {
		return ierr.NewForbiddenError("cannot delete an approved swap request")
	}

	if err := s.swapRepo.DeleteSwap(id); err != nil {
		return ierr.NewInternalServerError("error deleting swap request")
	}
	return nil
}
