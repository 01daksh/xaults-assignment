package models

import (
	"time"
)

// Service represents a monitored microservice
type Service struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`
	HealthStatus string    `gorm:"type:varchar(50);not null;default:'unknown'" json:"health_status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Service) TableName() string {
	return "services"
}
