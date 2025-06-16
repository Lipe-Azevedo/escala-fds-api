package certificate

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	CreateCertificate(certificate entity.Certificate) (*entity.Certificate, *ierr.RestErr)
	ApproveOrReject(id, approverID uint, status entity.CertificateStatus) (*entity.Certificate, *ierr.RestErr)
	FindAll() ([]entity.Certificate, *ierr.RestErr)
	FindByCollaborator(collaboratorID uint) ([]entity.Certificate, *ierr.RestErr)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateCertificate(certificate entity.Certificate) (*entity.Certificate, *ierr.RestErr) {
	certificate.Status = entity.CertificateStatusPending

	if err := s.repo.Create(&certificate); err != nil {
		return nil, ierr.NewInternalServerError("error creating certificate")
	}

	createdCert, err := s.repo.FindByID(certificate.ID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching newly created certificate")
	}

	return createdCert, nil
}

func (s *service) ApproveOrReject(id, approverID uint, status entity.CertificateStatus) (*entity.Certificate, *ierr.RestErr) {
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

	updatedCert, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching updated certificate")
	}

	return updatedCert, nil
}

func (s *service) FindAll() ([]entity.Certificate, *ierr.RestErr) {
	certificates, err := s.repo.FindAll()
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding certificates")
	}
	return certificates, nil
}

func (s *service) FindByCollaborator(collaboratorID uint) ([]entity.Certificate, *ierr.RestErr) {
	certificates, err := s.repo.FindByCollaboratorID(collaboratorID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding certificates for collaborator")
	}
	return certificates, nil
}
