package swap

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"time"
)

type CreateSwapRequest struct {
	InvolvedCollaboratorID *uint            `json:"involvedCollaboratorId"`
	OriginalDate           string           `json:"originalDate" binding:"required"`
	NewDate                string           `json:"newDate" binding:"required"`
	OriginalShift          entity.ShiftName `json:"originalShift" binding:"required"`
	NewShift               entity.ShiftName `json:"newShift" binding:"required"`
	Reason                 string           `json:"reason"`
}

type UpdateSwapStatusRequest struct {
	Status entity.SwapStatus `json:"status" binding:"required,oneof=approved rejected"`
}

type SwapResponse struct {
	ID                   uint               `json:"id"`
	Requester            user.UserResponse  `json:"requester"`
	InvolvedCollaborator *user.UserResponse `json:"involvedCollaborator,omitempty"`
	OriginalDate         string             `json:"originalDate"`
	NewDate              string             `json:"newDate"`
	OriginalShift        entity.ShiftName   `json:"originalShift"`
	NewShift             entity.ShiftName   `json:"newShift"`
	Reason               string             `json:"reason"`
	Status               entity.SwapStatus  `json:"status"`
	ApprovedBy           *user.UserResponse `json:"approvedBy,omitempty"`
	CreatedAt            time.Time          `json:"createdAt"`
	ApprovedAt           *time.Time         `json:"approvedAt,omitempty"`
}

func ToSwapResponse(swap *entity.Swap) SwapResponse {
	var involved *user.UserResponse
	if swap.InvolvedCollaborator != nil {
		res := user.ToUserResponse(swap.InvolvedCollaborator)
		involved = &res
	}

	var approvedBy *user.UserResponse
	if swap.ApprovedBy != nil {
		res := user.ToUserResponse(swap.ApprovedBy)
		approvedBy = &res
	}

	return SwapResponse{
		ID:                   swap.ID,
		Requester:            user.ToUserResponse(&swap.Requester),
		InvolvedCollaborator: involved,
		OriginalDate:         swap.OriginalDate.Format("2006-01-02"),
		NewDate:              swap.NewDate.Format("2006-01-02"),
		OriginalShift:        swap.OriginalShift,
		NewShift:             swap.NewShift,
		Reason:               swap.Reason,
		Status:               swap.Status,
		ApprovedBy:           approvedBy,
		CreatedAt:            swap.CreatedAt,
		ApprovedAt:           swap.ApprovedAt,
	}
}
