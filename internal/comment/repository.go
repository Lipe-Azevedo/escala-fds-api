package comment

import (
	"escala-fds-api/internal/entity"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateComment(comment *entity.Comment) error
	FindCommentByID(id uint) (*entity.Comment, error)
	Find(filters Filters) ([]entity.Comment, error)
	FindCommentsForUserInDateRange(userID uint, startDate, endDate time.Time) ([]entity.Comment, error)
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
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *repository) Find(filters Filters) ([]entity.Comment, error) {
	var comments []entity.Comment
	query := r.db.Table("comments").
		Joins("JOIN users AS collaborator ON collaborator.id = comments.collaborator_id")

	if filters.CollaboratorID != "" {
		query = query.Where("comments.collaborator_id = ?", filters.CollaboratorID)
	}
	if filters.AuthorID != "" {
		query = query.Where("comments.author_id = ?", filters.AuthorID)
	}
	if filters.Team != "" {
		query = query.Where("collaborator.team = ?", filters.Team)
	}
	if filters.StartDate != "" {
		if date, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("comments.date >= ?", date)
		}
	}
	if filters.EndDate != "" {
		if date, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			query = query.Where("comments.date <= ?", date)
		}
	}

	err := query.Order("comments.date desc").Find(&comments).Error
	return comments, err
}

func (r *repository) FindCommentsForUserInDateRange(userID uint, startDate, endDate time.Time) ([]entity.Comment, error) {
	var comments []entity.Comment
	err := r.db.
		Where("collaborator_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date desc").
		Find(&comments).Error
	return comments, err
}

func (r *repository) UpdateComment(comment *entity.Comment) error {
	return r.db.Save(comment).Error
}

func (r *repository) DeleteComment(id uint) error {
	return r.db.Delete(&entity.Comment{}, id).Error
}
