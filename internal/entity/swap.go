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
	Requester              User       `gorm:"foreignKey:RequesterID"`
	InvolvedCollaboratorID *uint      `gorm:"index"`
	InvolvedCollaborator   *User      `gorm:"foreignKey:InvolvedCollaboratorID"`
	OriginalDate           time.Time  `gorm:"type:date;not null"`
	NewDate                time.Time  `gorm:"type:date;not null"`
	OriginalShift          ShiftName  `gorm:"type:varchar(20);not null"`
	NewShift               ShiftName  `gorm:"type:varchar(20);not null"`
	Reason                 string     `gorm:"type:text"`
	Status                 SwapStatus `gorm:"type:varchar(20);default:'pending';not null;index"`
	ApprovedByID           *uint
	ApprovedBy             *User `gorm:"foreignKey:ApprovedByID"`
	ApprovedAt             *time.Time
}
