package holiday

import (
	"escala-fds-api/internal/entity"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateHoliday(holiday *entity.Holiday) error
	FindHolidayByID(id uint) (*entity.Holiday, error)
	FindHolidaysByDateRange(startDate, endDate time.Time) ([]entity.Holiday, error)
	FindAllHolidays() ([]entity.Holiday, error)
	IsHoliday(date time.Time) (bool, error)
	UpdateHoliday(holiday *entity.Holiday) error
	DeleteHoliday(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateHoliday(holiday *entity.Holiday) error {
	return r.db.Create(holiday).Error
}

func (r *repository) FindHolidayByID(id uint) (*entity.Holiday, error) {
	var holiday entity.Holiday
	if err := r.db.First(&holiday, id).Error; err != nil {
		return nil, err
	}
	return &holiday, nil
}

func (r *repository) FindHolidaysByDateRange(startDate, endDate time.Time) ([]entity.Holiday, error) {
	var holidays []entity.Holiday
	err := r.db.Where("date BETWEEN ? AND ?", startDate, endDate).Order("date asc").Find(&holidays).Error
	return holidays, err
}

func (r *repository) FindAllHolidays() ([]entity.Holiday, error) {
	var holidays []entity.Holiday
	err := r.db.Order("date asc").Find(&holidays).Error
	return holidays, err
}

func (r *repository) IsHoliday(date time.Time) (bool, error) {
	var count int64
	// Compare only the date part, ignoring time
	err := r.db.Model(&entity.Holiday{}).Where("DATE(date) = ?", date.Format("2006-01-02")).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) UpdateHoliday(holiday *entity.Holiday) error {
	return r.db.Save(holiday).Error
}

func (r *repository) DeleteHoliday(id uint) error {
	return r.db.Delete(&entity.Holiday{}, id).Error
}
