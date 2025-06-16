package certificate

import (
	"escala-fds-api/internal/entity"

	"gorm.io/gorm"
)

type Repository interface {
	Create(certificate *entity.Certificate) error
	FindByID(id uint) (*entity.Certificate, error)
	FindAll() ([]entity.Certificate, error)
	FindByCollaboratorID(collaboratorID uint) ([]entity.Certificate, error)
	Update(certificate *entity.Certificate) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(certificate *entity.Certificate) error {
	return r.db.Create(certificate).Error
}

func (r *repository) FindByID(id uint) (*entity.Certificate, error) {
	var certificate entity.Certificate
	err := r.db.Preload("Collaborator").Preload("ApprovedBy").First(&certificate, id).Error
	return &certificate, err
}

func (r *repository) FindAll() ([]entity.Certificate, error) {
	var certificates []entity.Certificate
	err := r.db.Preload("Collaborator").Order("created_at desc").Find(&certificates).Error
	return certificates, err
}

func (r *repository) FindByCollaboratorID(collaboratorID uint) ([]entity.Certificate, error) {
	var certificates []entity.Certificate
	err := r.db.Preload("Collaborator").Preload("ApprovedBy").Where("collaborator_id = ?", collaboratorID).Order("start_date desc").Find(&certificates).Error
	return certificates, err
}

func (r *repository) Update(certificate *entity.Certificate) error {
	return r.db.Save(certificate).Error
}
