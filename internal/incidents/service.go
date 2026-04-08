package incidents

import (
	"context"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	"xaults-assignment/customError"
	"xaults-assignment/enums"
	"xaults-assignment/internal/interfaces"
	internalmetrics "xaults-assignment/internal/metrics"
	"xaults-assignment/models"
)

type incidentService struct {
	IncidentRepo interfaces.IncidentRepository
	ServiceRepo  interfaces.ServiceRepository
}

func NewIncidentService(
	incidentRepo interfaces.IncidentRepository,
	serviceRepo interfaces.ServiceRepository,
) interfaces.IncidentService {
	return &incidentService{
		IncidentRepo: incidentRepo,
		ServiceRepo:  serviceRepo,
	}
}

func (s *incidentService) ReportIncident(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
	if _, err := s.ServiceRepo.FindByID(ctx, serviceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customError.ErrServiceNotFound
		}
		return nil, fmt.Errorf("validate service: %w", err)
	}

	incident := &models.Incident{
		ServiceID:   serviceID,
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      enums.IncidentStatusOpen,
	}
	if err := s.IncidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// 3. Update Prometheus gauge.
	internalmetrics.OpenIncidentsTotal.WithLabelValues(string(severity)).Inc()

	// 4. Recalculate and propagate health status (non-fatal).
	if err := s.recalculateHealth(ctx, serviceID); err != nil {
		log.Printf("warn: recalculate health for service %d: %v", serviceID, err)
	}

	return incident, nil
}

func (s *incidentService) recalculateHealth(ctx context.Context, serviceID uint) error {
	openIncidents, err := s.IncidentRepo.FindOpenByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("fetch open incidents: %w", err)
	}
	return s.ServiceRepo.UpdateHealthStatus(ctx, serviceID, deriveHealthStatus(openIncidents))
}

func deriveHealthStatus(open []models.Incident) enums.HealthStatus {
	if len(open) == 0 {
		return enums.HealthStatusHealthy
	}
	for _, inc := range open {
		if inc.Severity == enums.SeverityCritical || inc.Severity == enums.SeverityHigh {
			return enums.HealthStatusDown
		}
	}
	return enums.HealthStatusDegraded
}

func (s *incidentService) ListIncidents(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	if _, err := s.ServiceRepo.FindByID(ctx, serviceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customError.ErrServiceNotFound
		}
		return nil, fmt.Errorf("validate service: %w", err)
	}

	return s.IncidentRepo.FindByServiceID(ctx, serviceID)
}
