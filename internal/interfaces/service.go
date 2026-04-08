package interfaces

import (
	"context"

	"github.com/labstack/echo/v5"

	"xaults-assignment/enums"
	"xaults-assignment/models"
)

type ServiceController interface {
	CreateService(c *echo.Context) error
	ListServices(c *echo.Context) error
}

type ServiceRepository interface {
	Create(ctx context.Context, service *models.Service) error
	FindAll(ctx context.Context) ([]models.Service, error)
	FindByID(ctx context.Context, id uint) (*models.Service, error)
	UpdateHealthStatus(ctx context.Context, id uint, status enums.HealthStatus) error
}

type ServiceService interface {
	CreateService(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error)
	ListServices(ctx context.Context) ([]models.Service, error)
}
