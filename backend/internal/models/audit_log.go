package models

import "time"

// AuditLog is one append-only record of a mutating request (§6). Changes holds
// a JSON object of {"field": {"before": …, "after": …}} for each changed field.
type AuditLog struct {
	ID           uint   `gorm:"primaryKey"`
	ActorID      uint   `gorm:"not null;index"`
	Action       string `gorm:"size:32;not null"`
	ResourceType string `gorm:"size:32;not null;index"`
	ResourceID   uint   `gorm:"not null;index"`
	Changes      string `gorm:"type:text;not null"`
	CreatedAt    time.Time
}
