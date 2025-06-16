package swap

import (
	"escala-fds-api/internal/entity"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateSwap(swap *entity.Swap) error
	FindSwapByID(id uint) (*entity.Swap, error)
	FindSwapsByUserID(userID uint, statusFilter string) ([]entity.Swap, error)
	FindApprovedSwapsForDateRange(userID uint, startDate, endDate time.Time) ([]entity.Swap, error)
	FindAllSwaps() ([]entity.Swap, error)
	UpdateSwap(swap *entity.Swap) error
	DeleteSwap(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateSwap(swap *entity.Swap) error {
	return r.db.Create(swap).Error
}

func (r *repository) FindSwapByID(id uint) (*entity.Swap, error) {
	var swap entity.Swap
	if err := r.db.First(&swap, id).Error; err != nil {
		return nil, err
	}
	return &swap, nil
}

func (r *repository) FindSwapsByUserID(userID uint, statusFilter string) ([]entity.Swap, error) {
	var swaps []entity.Swap
	query := r.db.
		Where("requester_id = ? OR involved_collaborator_id = ?", userID, userID)

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	err := query.Order("created_at desc").Find(&swaps).Error
	return swaps, err
}

func (r *repository) FindApprovedSwapsForDateRange(userID uint, startDate, endDate time.Time) ([]entity.Swap, error) {
	var swaps []entity.Swap
	err := r.db.
		Where("status = ?", entity.StatusApproved).
		Where(
			r.db.Where("requester_id = ? AND (original_date BETWEEN ? AND ? OR new_date BETWEEN ? AND ?)", userID, startDate, endDate, startDate, endDate).
				Or("involved_collaborator_id = ? AND (original_date BETWEEN ? AND ? OR new_date BETWEEN ? AND ?)", userID, startDate, endDate, startDate, endDate),
		).
		Find(&swaps).Error
	return swaps, err
}

func (r *repository) FindAllSwaps() ([]entity.Swap, error) {
	var swaps []entity.Swap
	err := r.db.Order("created_at desc").Find(&swaps).Error
	return swaps, err
}

func (r *repository) UpdateSwap(swap *entity.Swap) error {
	return r.db.Save(swap).Error
}

func (r *repository) DeleteSwap(id uint) error {
	return r.db.Delete(&entity.Swap{}, id).Error
}
