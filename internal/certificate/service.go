package certificate

import (
	"escala-fds-api/internal/constants"
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	CreateCertificate(certificate entity.Certificate) (*CertificateResponse, *ierr.RestErr)
	ApproveOrReject(id, approverID uint, status entity.CertificateStatus) (*CertificateResponse, *ierr.RestErr)
	FindAll() ([]CertificateResponse, *ierr.RestErr)
	FindByCollaborator(collaboratorID uint) ([]CertificateResponse, *ierr.RestErr)
}

type service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) CreateCertificate(certificate entity.Certificate) (*CertificateResponse, *ierr.RestErr) {
	certificate.Status = entity.CertificateStatusPending

	if err := s.repo.Create(&certificate); err != nil {
		return nil, ierr.NewInternalServerError("error creating certificate")
	}

	return s.buildSingleResponse(certificate.ID)
}

func (s *service) ApproveOrReject(id, approverID uint, status entity.CertificateStatus) (*CertificateResponse, *ierr.RestErr) {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("certificate not found")
		}
		return nil, ierr.NewInternalServerError("error finding certificate")
	}

	now := time.Now().UTC()
	cert.Status = status
	cert.ApprovedByID = &approverID
	cert.ApprovedAt = &now

	if err := s.repo.Update(cert); err != nil {
		return nil, ierr.NewInternalServerError("error updating certificate status")
	}

	return s.buildSingleResponse(id)
}

func (s *service) FindAll() ([]CertificateResponse, *ierr.RestErr) {
	certificates, err := s.repo.FindAll()
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding certificates")
	}
	return s.buildResponseList(certificates)
}

func (s *service) FindByCollaborator(collaboratorID uint) ([]CertificateResponse, *ierr.RestErr) {
	certificates, err := s.repo.FindByCollaboratorID(collaboratorID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding certificates for collaborator")
	}
	return s.buildResponseList(certificates)
}

func (s *service) buildSingleResponse(id uint) (*CertificateResponse, *ierr.RestErr) {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching certificate")
	}

	list, restErr := s.buildResponseList([]entity.Certificate{*cert})
	if restErr != nil {
		return nil, restErr
	}
	if len(list) == 0 {
		return nil, ierr.NewNotFoundError("certificate response could not be built")
	}
	return &list[0], nil
}

func (s *service) buildResponseList(certificates []entity.Certificate) ([]CertificateResponse, *ierr.RestErr) {
	var userIDs []uint
	userIDsSet := make(map[uint]bool)

	for _, cert := range certificates {
		if !userIDsSet[cert.CollaboratorID] {
			userIDs = append(userIDs, cert.CollaboratorID)
			userIDsSet[cert.CollaboratorID] = true
		}
		if cert.ApprovedByID != nil && !userIDsSet[*cert.ApprovedByID] {
			userIDs = append(userIDs, *cert.ApprovedByID)
			userIDsSet[*cert.ApprovedByID] = true
		}
	}

	users, err := s.userRepo.FindUsersByIDs(userIDs)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching user data for certificates")
	}

	userMap := make(map[uint]*entity.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	var responses []CertificateResponse
	for _, cert := range certificates {
		collaborator := userMap[cert.CollaboratorID]
		var approver *entity.User
		if cert.ApprovedByID != nil {
			approver = userMap[*cert.ApprovedByID]
		}
		if collaborator != nil {
			responses = append(responses, s.toResponse(&cert, collaborator, approver))
		}
	}
	return responses, nil
}

func (s *service) toResponse(cert *entity.Certificate, collaborator *entity.User, approvedBy *entity.User) CertificateResponse {
	var approvedByResponse *user.UserResponse
	if approvedBy != nil {
		res := user.ToUserResponse(approvedBy)
		approvedByResponse = &res
	}

	var approvedAt *string
	if cert.ApprovedAt != nil {
		formatted := cert.ApprovedAt.Format(constants.ApiTimestampLayout)
		approvedAt = &formatted
	}

	return CertificateResponse{
		ID:           cert.ID,
		Collaborator: user.ToUserResponse(collaborator),
		StartDate:    cert.StartDate.Format(constants.ApiDateLayout),
		EndDate:      cert.EndDate.Format(constants.ApiDateLayout),
		Reason:       cert.Reason,
		Status:       cert.Status,
		ApprovedBy:   approvedByResponse,
		CreatedAt:    cert.CreatedAt.Format(constants.ApiTimestampLayout),
		ApprovedAt:   approvedAt,
	}
}
