package user

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Service interface {
	CreateUser(user entity.User, creatorType entity.UserType) (*entity.User, *ierr.RestErr)
	Login(email, password string) (string, *entity.User, *ierr.RestErr)
	FindUserByID(id uint) (*entity.User, *ierr.RestErr)
	FindAllUsers(requestorType entity.UserType, requestorTeam entity.TeamName) ([]entity.User, *ierr.RestErr)
	UpdatePersonalData(id, requestorId uint, requestorType entity.UserType, userUpdates entity.User) (*entity.User, *ierr.RestErr)
	UpdateWorkData(id uint, requestorType entity.UserType, userUpdates entity.User) (*entity.User, *ierr.RestErr)
	DeleteUser(id uint, requestorType entity.UserType) *ierr.RestErr
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

func (s *service) Login(email, password string) (string, *entity.User, *ierr.RestErr) {
	cleanEmail := strings.TrimSpace(email)
	cleanPassword := strings.TrimSpace(password)

	log.Printf("--- LOGIN ATTEMPT: Email=[%s] ---", cleanEmail)

	user, err := s.repo.FindUserByEmail(cleanEmail)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("LOGIN FAILED: User not found in DB for email: %s", cleanEmail)
			return "", nil, ierr.NewUnauthorizedError("invalid credentials")
		}
		log.Printf("LOGIN ERROR: Database error finding user: %v", err)
		return "", nil, ierr.NewInternalServerError("error finding user")
	}

	log.Printf("LOGIN DEBUG: User found. Stored Hash = [%s]", user.Password)
	log.Printf("LOGIN DEBUG: Password from request (after trim) = [%s]", cleanPassword)

	if !user.CheckPasswordHash(cleanPassword) {
		log.Printf("LOGIN FAILED: Password check failed for user %s.", cleanEmail)
		return "", nil, ierr.NewUnauthorizedError("invalid credentials")
	}

	log.Printf("LOGIN SUCCESS: Password check passed for user %s.", cleanEmail)

	claims := jwt.MapClaims{
		"id":        user.ID,
		"user_type": user.UserType,
		"team":      user.Team,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Printf("LOGIN ERROR: Token generation failed: %v", err)
		return "", nil, ierr.NewInternalServerError("error generating token")
	}

	return tokenString, user, nil
}

func (s *service) UpdatePersonalData(id, requestorId uint, requestorType entity.UserType, userUpdates entity.User) (*entity.User, *ierr.RestErr) {
	isUnauthenticatedDebugCall := requestorId == 0

	if !isUnauthenticatedDebugCall {
		if requestorType != entity.UserTypeMaster && id != requestorId {
			return nil, ierr.NewForbiddenError("you can only update your own personal data")
		}
	}

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
		log.Printf("DEBUG: New password hash generated and set for user ID %d", id)
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, ierr.NewInternalServerError("error updating user")
	}
	return user, nil
}

func (s *service) CreateUser(user entity.User, creatorType entity.UserType) (*entity.User, *ierr.RestErr) {
	if creatorType != entity.UserTypeMaster {
		return nil, ierr.NewForbiddenError("only masters can create new users")
	}

	if user.UserType == entity.UserTypeCollaborator {
		if err := s.validateWorkData(&user); err != nil {
			return nil, err
		}
		superiorID, err := s.determineSuperior(user.Team, user.Position)
		if err != nil {
			return nil, err
		}
		user.SuperiorID = superiorID
	}

	existingUser, err := s.repo.FindUserByEmail(user.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, ierr.NewInternalServerError("error finding user by email")
	}
	if existingUser != nil {
		return nil, ierr.NewConflictError("user with this email already exists")
	}

	if err := user.HashPassword(); err != nil {
		return nil, ierr.NewInternalServerError("error hashing password")
	}

	if err := s.repo.CreateUser(&user); err != nil {
		return nil, ierr.NewInternalServerError("error creating user")
	}
	return &user, nil
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

func (s *service) FindAllUsers(requestorType entity.UserType, requestorTeam entity.TeamName) ([]entity.User, *ierr.RestErr) {
	var users []entity.User
	var err error

	if requestorType == entity.UserTypeMaster {
		users, err = s.repo.FindAllUsers()
	} else {
		users, err = s.repo.FindUsersByTeam(requestorTeam)
	}

	if err != nil {
		return nil, ierr.NewInternalServerError("error finding users")
	}
	return users, nil
}

func (s *service) UpdateWorkData(id uint, requestorType entity.UserType, userUpdates entity.User) (*entity.User, *ierr.RestErr) {
	if requestorType != entity.UserTypeMaster {
		return nil, ierr.NewForbiddenError("only masters can update work data")
	}

	user, restErr := s.FindUserByID(id)
	if restErr != nil {
		return nil, restErr
	}

	if err := s.validateWorkData(&userUpdates); err != nil {
		return nil, err
	}

	user.Team = userUpdates.Team
	user.Position = userUpdates.Position
	user.Shift = userUpdates.Shift
	user.WeekdayOff = userUpdates.WeekdayOff
	user.InitialWeekendOff = userUpdates.InitialWeekendOff

	superiorID, err := s.determineSuperior(userUpdates.Team, userUpdates.Position)
	if err != nil {
		return nil, err
	}
	user.SuperiorID = superiorID

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, ierr.NewInternalServerError("error updating user work data")
	}
	return user, nil
}

func (s *service) DeleteUser(id uint, requestorType entity.UserType) *ierr.RestErr {
	if requestorType != entity.UserTypeMaster {
		return ierr.NewForbiddenError("only masters can delete users")
	}

	if _, err := s.FindUserByID(id); err != nil {
		return err
	}
	if err := s.repo.DeleteUser(id); err != nil {
		return ierr.NewInternalServerError("error deleting user")
	}
	return nil
}

func (s *service) validateWorkData(user *entity.User) *ierr.RestErr {
	positions, ok := validPositions[user.Team]
	if !ok {
		return ierr.NewBadRequestError(fmt.Sprintf("invalid team: %s", user.Team))
	}

	isValidPosition := false
	for _, pos := range positions {
		if pos == user.Position {
			isValidPosition = true
			break
		}
	}
	if !isValidPosition {
		return ierr.NewBadRequestError(fmt.Sprintf("position '%s' is not valid for team '%s'", user.Position, user.Team))
	}
	return nil
}

var validPositions = map[entity.TeamName][]entity.PositionName{
	entity.TeamSecurity:        {entity.PositionSecurity, entity.PositionSupervisorI, entity.PositionSupervisorII},
	entity.TeamSupport:         {entity.PositionDevBackend, entity.PositionDevFrontend},
	entity.TeamCustomerService: {entity.PositionAttendant, entity.PositionSupervisorI, entity.PositionSupervisorII},
}

func (s *service) determineSuperior(team entity.TeamName, position entity.PositionName) (*uint, *ierr.RestErr) {
	if position == entity.PositionDevBackend || position == entity.PositionDevFrontend || position == entity.PositionSupervisorII {
		master, err := s.repo.FindMasterUser()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, ierr.NewNotFoundError("master user not found to be set as superior")
			}
			return nil, ierr.NewInternalServerError("error finding master user")
		}
		return &master.ID, nil
	}

	var superiorPosition entity.PositionName
	switch position {
	case entity.PositionSecurity, entity.PositionAttendant:
		superiorPosition = entity.PositionSupervisorI
	case entity.PositionSupervisorI:
		superiorPosition = entity.PositionSupervisorII
	default:
		return nil, nil
	}

	superiors, err := s.repo.FindUsersByTeamAndPosition(team, superiorPosition)
	if err != nil {
		return nil, ierr.NewInternalServerError(fmt.Sprintf("error finding superior with position %s", superiorPosition))
	}
	if len(superiors) == 0 {
		return nil, ierr.NewNotFoundError(fmt.Sprintf("no superior found with position %s in team %s", superiorPosition, team))
	}

	return &superiors[0].ID, nil
}
