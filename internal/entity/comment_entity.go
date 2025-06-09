package entity

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	CollaboratorID uint      `gorm:"not null;index"`
	Collaborator   User      `gorm:"foreignKey:CollaboratorID"`
	AuthorID       uint      `gorm:"not null;index"`
	Author         User      `gorm:"foreignKey:AuthorID"`
	Text           string    `gorm:"type:text;not null"`
	Date           time.Time `gorm:"type:date;not null;index"`
}
