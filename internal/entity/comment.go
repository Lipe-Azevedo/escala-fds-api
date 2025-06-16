package entity

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	CollaboratorID uint      `gorm:"not null;index"`
	AuthorID       uint      `gorm:"not null;index"`
	Text           string    `gorm:"type:text;not null"`
	Date           time.Time `gorm:"type:date;not null;index"`
}
