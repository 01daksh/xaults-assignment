package incidents_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"xaults-assignment/customError"
	"xaults-assignment/enums"
	"xaults-assignment/internal/incidents"
	"xaults-assignment/models"
)

type incidentRepoStub struct {
	createFn func(ctx context.Context, incident *models.Incident) error
	findByServiceIDFn func(ctx context.Context, serviceID uint) ([]models.Incident, error)
	findOpenByServiceIDFn func(ctx context.Context, serviceID uint) ([]models.Incident, error)
}

func (s incidentRepoStub) Create(ctx context.Context, incident *models.Incident) error {
	return s.createFn(ctx, incident)
}

func (s incidentRepoStub) FindByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	return s.findByServiceIDFn(ctx, serviceID)
}

func (s incidentRepoStub) FindOpenByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	return s.findOpenByServiceIDFn(ctx, serviceID)
}

type incidentServiceRepoStub struct {
	createFn func(ctx context.Context, service *models.Service) error
	findAllFn func(ctx context.Context) ([]models.Service, error)
	findByIDFn func(ctx context.Context, id uint) (*models.Service, error)
	updateHealthStatusFn func(ctx context.Context, id uint, status enums.HealthStatus) error
}

func (s incidentServiceRepoStub) Create(ctx context.Context, service *models.Service) error {
	return s.createFn(ctx, service)
}

func (s incidentServiceRepoStub) FindAll(ctx context.Context) ([]models.Service, error) {
	return s.findAllFn(ctx)
}

func (s incidentServiceRepoStub) FindByID(ctx context.Context, id uint) (*models.Service, error) {
	return s.findByIDFn(ctx, id)
}

func (s incidentServiceRepoStub) UpdateHealthStatus(ctx context.Context, id uint, status enums.HealthStatus) error {
	return s.updateHealthStatusFn(ctx, id, status)
}

func TestIncidentServiceReportIncidentSetsOpenStatusAndUpdatesHealth(t *testing.T) {
	service := incidents.NewIncidentService(
		incidentRepoStub{
			createFn: func(ctx context.Context, incident *models.Incident) error {
				if incident.Status != enums.IncidentStatusOpen {
					t.Fatalf("unexpected status: %s", incident.Status)
				}
				return nil
			},
			findByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
			findOpenByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) {
				return []models.Incident{{Severity: enums.SeverityCritical}}, nil
			},
		},
		incidentServiceRepoStub{
			createFn: func(ctx context.Context, service *models.Service) error { return nil },
			findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
			findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) {
				return &models.Service{ID: id, Name: "payments"}, nil
			},
			updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error {
				if status != enums.HealthStatusDown {
					t.Fatalf("unexpected derived health status: %s", status)
				}
				return nil
			},
		},
	)

	got, err := service.ReportIncident(context.Background(), 7, "db down", "unavailable", enums.SeverityCritical)
	if err != nil {
		t.Fatalf("ReportIncident returned error: %v", err)
	}
	if got.ServiceID != 7 || got.Status != enums.IncidentStatusOpen {
		t.Fatalf("unexpected incident: %#v", got)
	}
}

func TestIncidentServiceReportIncidentMapsServiceNotFound(t *testing.T) {
	service := incidents.NewIncidentService(
		incidentRepoStub{
			createFn: func(ctx context.Context, incident *models.Incident) error { return nil },
			findByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
			findOpenByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
		},
		incidentServiceRepoStub{
			createFn: func(ctx context.Context, service *models.Service) error { return nil },
			findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
			findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) {
				return nil, gorm.ErrRecordNotFound
			},
			updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
		},
	)

	_, err := service.ReportIncident(context.Background(), 7, "db down", "", enums.SeverityHigh)
	if !errors.Is(err, customError.ErrServiceNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncidentServiceReportIncidentPropagatesCreateError(t *testing.T) {
	service := incidents.NewIncidentService(
		incidentRepoStub{
			createFn: func(ctx context.Context, incident *models.Incident) error { return errors.New("insert failed") },
			findByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
			findOpenByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
		},
		incidentServiceRepoStub{
			createFn: func(ctx context.Context, service *models.Service) error { return nil },
			findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
			findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) {
				return &models.Service{ID: id}, nil
			},
			updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
		},
	)

	_, err := service.ReportIncident(context.Background(), 7, "db down", "", enums.SeverityHigh)
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncidentServiceListIncidents(t *testing.T) {
	want := []models.Incident{
		{ID: 1, ServiceID: 7, Title: "db down", Severity: enums.SeverityCritical},
	}
	service := incidents.NewIncidentService(
		incidentRepoStub{
			createFn: func(ctx context.Context, incident *models.Incident) error { return nil },
			findByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return want, nil },
			findOpenByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
		},
		incidentServiceRepoStub{
			createFn: func(ctx context.Context, service *models.Service) error { return nil },
			findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
			findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) {
				return &models.Service{ID: id}, nil
			},
			updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
		},
	)

	got, err := service.ListIncidents(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListIncidents returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "db down" {
		t.Fatalf("unexpected incidents: %#v", got)
	}
}

func TestIncidentServiceListIncidentsMapsServiceNotFound(t *testing.T) {
	service := incidents.NewIncidentService(
		incidentRepoStub{
			createFn: func(ctx context.Context, incident *models.Incident) error { return nil },
			findByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
			findOpenByServiceIDFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
		},
		incidentServiceRepoStub{
			createFn: func(ctx context.Context, service *models.Service) error { return nil },
			findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
			findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) {
				return nil, gorm.ErrRecordNotFound
			},
			updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
		},
	)

	_, err := service.ListIncidents(context.Background(), 7)
	if !errors.Is(err, customError.ErrServiceNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
