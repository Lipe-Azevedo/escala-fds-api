package user

import (
	"escala-fds-api/internal/constants"
	"escala-fds-api/internal/entity"
)

type CreateUserRequest struct {
	Email             string                `json:"email" binding:"required,email"`
	Password          string                `json:"password" binding:"required,min=6"`
	FirstName         string                `json:"firstName" binding:"required"`
	LastName          string                `json:"lastName" binding:"required"`
	PhoneNumber       string                `json:"phoneNumber" binding:"required"`
	UserType          entity.UserType       `json:"userType" binding:"required"`
	Team              entity.TeamName       `json:"team"`
	Position          entity.PositionName   `json:"position"`
	Shift             entity.ShiftName      `json:"shift"`
	WeekdayOff        entity.WeekdayName    `json:"weekdayOff"`
	InitialWeekendOff entity.WeekendDayName `json:"initialWeekendOff"`
	SuperiorID        *uint                 `json:"superiorId"`
}

type UpdatePersonalDataRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password" binding:"omitempty,min=6"`
}

type UpdateWorkDataRequest struct {
	Team              entity.TeamName       `json:"team" binding:"required"`
	Position          entity.PositionName   `json:"position" binding:"required"`
	Shift             entity.ShiftName      `json:"shift" binding:"required"`
	WeekdayOff        entity.WeekdayName    `json:"weekdayOff" binding:"required"`
	InitialWeekendOff entity.WeekendDayName `json:"initialWeekendOff" binding:"required"`
	SuperiorID        *uint                 `json:"superiorId"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID                uint                  `json:"id"`
	Email             string                `json:"email"`
	FirstName         string                `json:"firstName"`
	LastName          string                `json:"lastName"`
	PhoneNumber       string                `json:"phoneNumber"`
	UserType          entity.UserType       `json:"userType"`
	Team              entity.TeamName       `json:"team,omitempty"`
	Position          entity.PositionName   `json:"position,omitempty"`
	Shift             entity.ShiftName      `json:"shift,omitempty"`
	WeekdayOff        entity.WeekdayName    `json:"weekdayOff,omitempty"`
	InitialWeekendOff entity.WeekendDayName `json:"initialWeekendOff,omitempty"`
	SuperiorID        *uint                 `json:"superiorId,omitempty"`
	CreatedAt         string                `json:"createdAt"`
	UpdatedAt         string                `json:"updatedAt"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		PhoneNumber:       user.PhoneNumber,
		UserType:          user.UserType,
		Team:              user.Team,
		Position:          user.Position,
		Shift:             user.Shift,
		WeekdayOff:        user.WeekdayOff,
		InitialWeekendOff: user.InitialWeekendOff,
		SuperiorID:        user.SuperiorID,
		CreatedAt:         user.CreatedAt.Format(constants.ApiTimestampLayout),
		UpdatedAt:         user.UpdatedAt.Format(constants.ApiTimestampLayout),
	}
}
