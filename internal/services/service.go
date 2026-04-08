package services

import (
	"context"
	"fmt"

	"xaults-assignment/enums"
	"xaults-assignment/internal/interfaces"
	internalmetrics "xaults-assignment/internal/metrics"
	"xaults-assignment/models"
)

type ServiceService struct {
	Repo interfaces.ServiceRepository
}


func NewServiceService(repo interfaces.ServiceRepository) interfaces.ServiceService {
	return &ServiceService{Repo: repo}
}

func (s *ServiceService) CreateService(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
	if name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	
	health := enums.HealthStatusUnknown.String()
	if healthStatus != nil {
		health = *healthStatus
	}
	
	svc := &models.Service{
		Name:         name,
		Description:  description,
		HealthStatus: health,
	}

	if err := s.Repo.Create(ctx, svc); err != nil {
		return nil, err
	}

	internalmetrics.ActiveServicesTotal.Inc()

	return svc, nil
}

func (s *ServiceService) ListServices(ctx context.Context) ([]models.Service, error) {
	return s.Repo.FindAll(ctx)
}
