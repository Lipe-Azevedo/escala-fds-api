package swap

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	CreateSwap(swap entity.Swap, requesterID uint) (*entity.Swap, *ierr.RestErr)
	ApproveOrRejectSwap(swapID, approverID uint, newStatus entity.SwapStatus) (*entity.Swap, *ierr.RestErr)
	FindSwapByID(id uint) (*entity.Swap, *ierr.RestErr)
	FindSwapsForUser(userID uint) ([]entity.Swap, *ierr.RestErr)
	FindAllSwaps() ([]entity.Swap, *ierr.RestErr)
	DeleteSwap(id, requesterID uint, requesterType entity.UserType) *ierr.RestErr
}

type service struct {
	swapRepo Repository
	userRepo user.Repository
}

func NewService(swapRepo Repository, userRepo user.Repository) Service {
	return &service{swapRepo: swapRepo, userRepo: userRepo}
}

func (s *service) CreateSwap(swap entity.Swap, requesterID uint) (*entity.Swap, *ierr.RestErr) {
	swap.RequesterID = requesterID
	swap.Status = entity.StatusPending

	if err := s.validateSwap(swap); err != nil {
		return nil, err
	}

	if err := s.swapRepo.CreateSwap(&swap); err != nil {
		return nil, ierr.NewInternalServerError("error creating swap request")
	}

	// ALTERAÇÃO CRÍTICA AQUI:
	// Em vez de retornar o objeto em memória, buscamos o objeto recém-criado
	// para que o GORM carregue as relações (como o Requester).
	newSwap, err := s.swapRepo.FindSwapByID(swap.ID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching newly created swap")
	}

	return newSwap, nil
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
	if newStatus == entity.StatusApproved {
		now := time.Now()
		swap.ApprovedAt = &now
		swap.ApprovedByID = &approverID
	} else {
		swap.ApprovedAt = nil
		swap.ApprovedByID = nil
	}

	if err := s.swapRepo.UpdateSwap(swap); err != nil {
		return nil, ierr.NewInternalServerError("error updating swap status")
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

func (s *service) FindSwapsForUser(userID uint) ([]entity.Swap, *ierr.RestErr) {
	swaps, err := s.swapRepo.FindSwapsByUserID(userID)
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

	if err := s.swapRepo.DeleteSwap(id); err != nil {
		return ierr.NewInternalServerError("error deleting swap request")
	}
	return nil
}

func (s *service) validateSwap(swap entity.Swap) *ierr.RestErr {
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

	return nil
}
