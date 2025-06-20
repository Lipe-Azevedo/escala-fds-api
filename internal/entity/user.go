package entity

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserType string
type TeamName string
type PositionName string
type ShiftName string
type WeekdayName string
type WeekendDayName string

const (
	UserTypeMaster       UserType = "master"
	UserTypeCollaborator UserType = "collaborator"

	TeamSecurity        TeamName = "Security"
	TeamSupport         TeamName = "Support"
	TeamCustomerService TeamName = "CustomerService"

	PositionSecurity     PositionName = "Security"
	PositionSupervisorI  PositionName = "SupervisorI"
	PositionSupervisorII PositionName = "SupervisorII"
	PositionDevBackend   PositionName = "BackendDeveloper"
	PositionDevFrontend  PositionName = "FrontendDeveloper"
	PositionAttendant    PositionName = "Attendant"
	PositionMaster       PositionName = "Master"

	ShiftMorning   ShiftName   = "06:00-14:00"
	ShiftAfternoon ShiftName   = "14:00-22:00"
	ShiftNight     ShiftName   = "22:00-06:00"
	WeekdayMonday  WeekdayName = "monday"

	WeekdayTuesday   WeekdayName    = "tuesday"
	WeekdayWednesday WeekdayName    = "wednesday"
	WeekdayThursday  WeekdayName    = "thursday"
	WeekdayFriday    WeekdayName    = "friday"
	WeekendSaturday  WeekendDayName = "saturday"
	WeekendSunday    WeekendDayName = "sunday"
)

type User struct {
	gorm.Model
	Email             string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password          string         `gorm:"type:varchar(255);not null"`
	FirstName         string         `gorm:"type:varchar(50);not null"`
	LastName          string         `gorm:"type:varchar(50);not null"`
	PhoneNumber       string         `gorm:"type:varchar(20);not null"`
	Birthday          *time.Time     `gorm:"type:date"`
	UserType          UserType       `gorm:"type:varchar(20);not null"`
	Team              TeamName       `gorm:"type:varchar(50)"`
	Position          PositionName   `gorm:"type:varchar(50)"`
	Shift             ShiftName      `gorm:"type:varchar(20)"`
	WeekdayOff        WeekdayName    `gorm:"type:varchar(20)"`
	InitialWeekendOff WeekendDayName `gorm:"type:varchar(20)"`
	SuperiorID        *uint          `gorm:"index"`
}

func (u *User) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
