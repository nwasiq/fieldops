package models

import (
	"time"

	"gorm.io/gorm"
)

// User is a person who can log in. LicenceNumber is stored encrypted; the
// services layer is the only place that sees it in the clear.
type User struct {
	ID            uint   `gorm:"primaryKey"`
	Email         string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash  string `gorm:"size:255;not null"`
	FirstName     string `gorm:"size:100;not null"`
	LastName      string `gorm:"size:100;not null"`
	Role          string `gorm:"size:32;not null;index"`
	IsActive      bool   `gorm:"not null;default:true"`
	LicenceNumber string `gorm:"size:512"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// FullName is the display form used for attribution.
func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}
