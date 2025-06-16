package entity

import (
	"time"

	"gorm.io/gorm"
)

type CertificateStatus string

const (
	CertificateStatusPending  CertificateStatus = "pending"
	CertificateStatusApproved CertificateStatus = "approved"
	CertificateStatusRejected CertificateStatus = "rejected"
)

type Certificate struct {
	gorm.Model
	CollaboratorID uint              `gorm:"not null;index"`
	StartDate      time.Time         `gorm:"type:date;not null"`
	EndDate        time.Time         `gorm:"type:date;not null"`
	Reason         string            `gorm:"type:text;not null"`
	Status         CertificateStatus `gorm:"type:varchar(20);default:'pending';not null;index"`
	ApprovedByID   *uint
	ApprovedAt     *time.Time
}
