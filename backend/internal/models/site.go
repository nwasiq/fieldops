package models

import (
	"time"

	"gorm.io/gorm"
)

// Site is a customer location that technicians visit (§2.1).
type Site struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:200;not null"`
	Address      string `gorm:"size:500;not null"`
	ContactPhone string `gorm:"size:50"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
