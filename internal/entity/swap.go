package entity

import (
	"time"

	"gorm.io/gorm"
)

type SwapStatus string

const (
	StatusPending  SwapStatus = "pending"
	StatusApproved SwapStatus = "approved"
	StatusRejected SwapStatus = "rejected"
)

type Swap struct {
	gorm.Model
	RequesterID            uint       `gorm:"not null;index"`
	InvolvedCollaboratorID *uint      `gorm:"index"`
	OriginalDate           time.Time  `gorm:"type:date;not null"`
	NewDate                time.Time  `gorm:"type:date;not null"`
	OriginalShift          ShiftName  `gorm:"type:varchar(20);not null"`
	NewShift               ShiftName  `gorm:"type:varchar(20);not null"`
	Reason                 string     `gorm:"type:text"`
	Status                 SwapStatus `gorm:"type:varchar(20);default:'pending';not null;index"`
	ApprovedByID           *uint
	ApprovedAt             *time.Time
}
