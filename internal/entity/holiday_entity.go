package entity

import (
	"time"

	"gorm.io/gorm"
)

type HolidayType string

const (
	HolidayTypeNational HolidayType = "national"
	HolidayTypeState    HolidayType = "state"
	HolidayTypeCity     HolidayType = "city"
)

type Holiday struct {
	gorm.Model
	Name string      `gorm:"type:varchar(100);not null"`
	Date time.Time   `gorm:"type:date;not null;uniqueIndex"`
	Type HolidayType `gorm:"type:varchar(20);not null"`
}
