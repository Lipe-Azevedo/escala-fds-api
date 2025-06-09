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
	return holidays, err // <-- CORRIGIDO AQUI (e em outras funções para garantir)
}

func (r *repository) UpdateHoliday(holiday *entity.Holiday) error {
	return r.db.Save(holiday).Error
}

func (r *repository) DeleteHoliday(id uint) error {
	return r.db.Delete(&entity.Holiday{}, id).Error
}
