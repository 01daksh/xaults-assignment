package models

import (
	"time"
	"xaults-assignment/enums"
)

// Incident => for a particular service
type Incident struct {
	ID          uint                 `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID   uint                 `gorm:"not null" json:"service_id"`
	Title       string               `gorm:"type:varchar(255);not null" json:"title"`
	Description string               `gorm:"type:text" json:"description,omitempty"`
	Severity    enums.Severity       `gorm:"type:varchar(50);not null" json:"severity"`
	Status      enums.IncidentStatus `gorm:"type:varchar(50);not null;default:'open'" json:"status"`
	ReportedAt  time.Time            `json:"reported_at"`
	ResolvedAt  *time.Time           `json:"resolved_at,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (i *Incident) TableName() string {
	return "incidents"
}
