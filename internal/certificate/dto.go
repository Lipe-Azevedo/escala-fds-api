package certificate

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
)

type CreateCertificateRequest struct {
	StartDate string `json:"startDate" binding:"required"`
	EndDate   string `json:"endDate" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

type UpdateStatusRequest struct {
	Status entity.CertificateStatus `json:"status" binding:"required,oneof=approved rejected"`
}

type CertificateResponse struct {
	ID           uint                     `json:"id"`
	Collaborator user.UserResponse        `json:"collaborator"`
	StartDate    string                   `json:"startDate"`
	EndDate      string                   `json:"endDate"`
	Reason       string                   `json:"reason"`
	Status       entity.CertificateStatus `json:"status"`
	ApprovedBy   *user.UserResponse       `json:"approvedBy,omitempty"`
	CreatedAt    string                   `json:"createdAt"`
	ApprovedAt   *string                  `json:"approvedAt,omitempty"`
}
