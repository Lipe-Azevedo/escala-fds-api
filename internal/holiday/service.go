package holiday

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"

	"gorm.io/gorm"
)

type Service interface {
	CreateHoliday(holiday entity.Holiday) (*entity.Holiday, *ierr.RestErr)
	FindHolidayByID(id uint) (*entity.Holiday, *ierr.RestErr)
	FindAllHolidays() ([]entity.Holiday, *ierr.RestErr)
	UpdateHoliday(id uint, holidayData entity.Holiday) (*entity.Holiday, *ierr.RestErr)
	DeleteHoliday(id uint) *ierr.RestErr
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateHoliday(holiday entity.Holiday) (*entity.Holiday, *ierr.RestErr) {
	if err := s.repo.CreateHoliday(&holiday); err != nil {
		return nil, ierr.NewInternalServerError("error creating holiday")
	}
	return &holiday, nil
}

func (s *service) FindHolidayByID(id uint) (*entity.Holiday, *ierr.RestErr) {
	holiday, err := s.repo.FindHolidayByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("holiday not found")
		}
		return nil, ierr.NewInternalServerError("error finding holiday")
	}
	return holiday, nil
}

func (s *service) FindAllHolidays() ([]entity.Holiday, *ierr.RestErr) {
	holidays, err := s.repo.FindAllHolidays()
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding holidays")
	}
	return holidays, nil
}

func (s *service) UpdateHoliday(id uint, holidayData entity.Holiday) (*entity.Holiday, *ierr.RestErr) {
	holiday, restErr := s.FindHolidayByID(id)
	if restErr != nil {
		return nil, restErr
	}

	holiday.Name = holidayData.Name
	holiday.Date = holidayData.Date
	holiday.Type = holidayData.Type

	if err := s.repo.UpdateHoliday(holiday); err != nil {
		return nil, ierr.NewInternalServerError("error updating holiday")
	}
	return holiday, nil
}

func (s *service) DeleteHoliday(id uint) *ierr.RestErr {
	if _, err := s.FindHolidayByID(id); err != nil {
		return err
	}
	if err := s.repo.DeleteHoliday(id); err != nil {
		return ierr.NewInternalServerError("error deleting holiday")
	}
	return nil
}
