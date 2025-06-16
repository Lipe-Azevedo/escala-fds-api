package user

import (
	"escala-fds-api/internal/entity"

	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(user *entity.User) error
	FindUserByEmail(email string) (*entity.User, error)
	FindUserByID(id uint) (*entity.User, error)
	FindUsersByIDs(ids []uint) ([]entity.User, error)
	FindAllUsers() ([]entity.User, error)
	FindUsersByTeam(team entity.TeamName) ([]entity.User, error)
	FindUsersByTeamAndPosition(team entity.TeamName, position entity.PositionName) ([]entity.User, error)
	FindMasterUser() (*entity.User, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *repository) FindUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByID(id uint) (*entity.User, error) {
	var user entity.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUsersByIDs(ids []uint) ([]entity.User, error) {
	var users []entity.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *repository) FindAllUsers() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Order("first_name asc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) FindUsersByTeam(team entity.TeamName) ([]entity.User, error) {
	var users []entity.User
	err := r.db.Where("team = ?", team).Order("first_name asc").Find(&users).Error
	return users, err
}

func (r *repository) FindUsersByTeamAndPosition(team entity.TeamName, position entity.PositionName) ([]entity.User, error) {
	var users []entity.User
	err := r.db.Where("team = ? AND position = ?", team, position).Find(&users).Error
	return users, err
}

func (r *repository) FindMasterUser() (*entity.User, error) {
	var user entity.User
	err := r.db.Where("user_type = ?", entity.UserTypeMaster).First(&user).Error
	return &user, err
}

func (r *repository) UpdateUser(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *repository) DeleteUser(id uint) error {
	return r.db.Delete(&entity.User{}, id).Error
}
