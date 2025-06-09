package comment

import (
	"escala-fds-api/internal/entity"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateComment(comment *entity.Comment) error
	FindCommentByID(id uint) (*entity.Comment, error)
	FindCommentsForUserInDateRange(userID uint, startDate, endDate time.Time) ([]entity.Comment, error)
	FindAllComments() ([]entity.Comment, error)
	UpdateComment(comment *entity.Comment) error
	DeleteComment(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateComment(comment *entity.Comment) error {
	return r.db.Create(comment).Error
}

func (r *repository) FindCommentByID(id uint) (*entity.Comment, error) {
	var comment entity.Comment
	if err := r.db.Preload("Collaborator").Preload("Author").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *repository) FindCommentsForUserInDateRange(userID uint, startDate, endDate time.Time) ([]entity.Comment, error) {
	var comments []entity.Comment
	err := r.db.
		Preload("Author").
		Where("collaborator_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date desc").
		Find(&comments).Error
	return comments, err
}

func (r *repository) FindAllComments() ([]entity.Comment, error) {
	var comments []entity.Comment
	err := r.db.Preload("Collaborator").Preload("Author").Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r *repository) UpdateComment(comment *entity.Comment) error {
	return r.db.Save(comment).Error
}

func (r *repository) DeleteComment(id uint) error {
	return r.db.Delete(&entity.Comment{}, id).Error
}
