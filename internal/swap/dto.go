package swap

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
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
	CreatedAt            string             `json:"createdAt"`
	ApprovedAt           *string            `json:"approvedAt,omitempty"`
}
