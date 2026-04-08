package interfaces

import (
	"context"

	"github.com/labstack/echo/v5"

	"xaults-assignment/enums"
	"xaults-assignment/models"
)

type IncidentController interface {
	ReportIncident(c *echo.Context) error
	ListIncidents(c *echo.Context) error
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *models.Incident) error
	FindByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error)
	FindOpenByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error)
}

type IncidentService interface {
	ReportIncident(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error)
	ListIncidents(ctx context.Context, serviceID uint) ([]models.Incident, error)
}
