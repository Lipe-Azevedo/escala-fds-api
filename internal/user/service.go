package user

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Service interface {
	CreateUser(user entity.User) (*entity.User, *ierr.RestErr)
	Login(email, password string) (string, *entity.User, *ierr.RestErr)
	FindUserByID(id uint) (*entity.User, *ierr.RestErr)
	FindAllUsers() ([]entity.User, *ierr.RestErr)
	UpdatePersonalData(id uint, userUpdates entity.User) (*entity.User, *ierr.RestErr)
	UpdateWorkData(id uint, userUpdates entity.User) (*entity.User, *ierr.RestErr)
	DeleteUser(id uint) *ierr.RestErr
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository) Service {
	return &service{
		repo:      repo,
		jwtSecret: os.Getenv("JWT_SECRET_KEY"),
	}
}

func (s *service) CreateUser(user entity.User) (*entity.User, *ierr.RestErr) {
	existingUser, err := s.repo.FindUserByEmail(user.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, ierr.NewInternalServerError("error finding user by email")
	}
	if existingUser != nil && existingUser.ID != 0 {
		return nil, ierr.NewConflictError("user with this email already exists")
	}

	if user.SuperiorID != nil {
		_, err := s.repo.FindUserByID(*user.SuperiorID)
		if err != nil {
			return nil, ierr.NewBadRequestError("superior user not found")
		}
	}

	if err := user.HashPassword(); err != nil {
		return nil, ierr.NewInternalServerError("error hashing password")
	}

	if err := s.repo.CreateUser(&user); err != nil {
		return nil, ierr.NewInternalServerError("error creating user")
	}
	return &user, nil
}

func (s *service) Login(email, password string) (string, *entity.User, *ierr.RestErr) {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, ierr.NewUnauthorizedError("invalid credentials")
		}
		return "", nil, ierr.NewInternalServerError("error finding user")
	}

	if !user.CheckPasswordHash(password) {
		return "", nil, ierr.NewUnauthorizedError("invalid credentials")
	}

	claims := jwt.MapClaims{
		"id":        user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, ierr.NewInternalServerError("error generating token")
	}

	return tokenString, user, nil
}

func (s *service) FindUserByID(id uint) (*entity.User, *ierr.RestErr) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("user not found")
		}
		return nil, ierr.NewInternalServerError("error finding user")
	}
	return user, nil
}

func (s *service) FindAllUsers() ([]entity.User, *ierr.RestErr) {
	users, err := s.repo.FindAllUsers()
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding users")
	}
	return users, nil
}

func (s *service) UpdatePersonalData(id uint, userUpdates entity.User) (*entity.User, *ierr.RestErr) {
	user, restErr := s.FindUserByID(id)
	if restErr != nil {
		return nil, restErr
	}

	if userUpdates.FirstName != "" {
		user.FirstName = userUpdates.FirstName
	}
	if userUpdates.LastName != "" {
		user.LastName = userUpdates.LastName
	}
	if userUpdates.PhoneNumber != "" {
		user.PhoneNumber = userUpdates.PhoneNumber
	}
	if userUpdates.Password != "" {
		user.Password = userUpdates.Password
		if err := user.HashPassword(); err != nil {
			return nil, ierr.NewInternalServerError("error hashing password")
		}
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, ierr.NewInternalServerError("error updating user")
	}
	return user, nil
}

func (s *service) UpdateWorkData(id uint, userUpdates entity.User) (*entity.User, *ierr.RestErr) {
	user, restErr := s.FindUserByID(id)
	if restErr != nil {
		return nil, restErr
	}

	user.Team = userUpdates.Team
	user.Position = userUpdates.Position
	user.Shift = userUpdates.Shift
	user.WeekdayOff = userUpdates.WeekdayOff
	user.InitialWeekendOff = userUpdates.InitialWeekendOff
	user.SuperiorID = userUpdates.SuperiorID

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, ierr.NewInternalServerError("error updating user work data")
	}
	return user, nil
}

func (s *service) DeleteUser(id uint) *ierr.RestErr {
	if _, err := s.FindUserByID(id); err != nil {
		return err
	}
	if err := s.repo.DeleteUser(id); err != nil {
		return ierr.NewInternalServerError("error deleting user")
	}
	return nil
}
