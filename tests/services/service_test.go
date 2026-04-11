package services_test

import (
	"context"
	"errors"
	"testing"

	"xaults-assignment/enums"
	"xaults-assignment/internal/services"
	"xaults-assignment/models"
)

type serviceRepoStub struct {
	createFn func(ctx context.Context, service *models.Service) error
	findAllFn func(ctx context.Context) ([]models.Service, error)
	findByIDFn func(ctx context.Context, id uint) (*models.Service, error)
	updateHealthStatusFn func(ctx context.Context, id uint, status enums.HealthStatus) error
}

func (s serviceRepoStub) Create(ctx context.Context, service *models.Service) error {
	return s.createFn(ctx, service)
}

func (s serviceRepoStub) FindAll(ctx context.Context) ([]models.Service, error) {
	return s.findAllFn(ctx)
}

func (s serviceRepoStub) FindByID(ctx context.Context, id uint) (*models.Service, error) {
	return s.findByIDFn(ctx, id)
}

func (s serviceRepoStub) UpdateHealthStatus(ctx context.Context, id uint, status enums.HealthStatus) error {
	return s.updateHealthStatusFn(ctx, id, status)
}

func TestServiceServiceCreateServiceDefaultsUnknownHealth(t *testing.T) {
	svc := services.NewServiceService(serviceRepoStub{
		createFn: func(ctx context.Context, service *models.Service) error {
			if service.Name != "payments" {
				t.Fatalf("unexpected name: %s", service.Name)
			}
			if service.HealthStatus != enums.HealthStatusUnknown.String() {
				t.Fatalf("unexpected health status: %s", service.HealthStatus)
			}
			return nil
		},
		findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
		findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) { return nil, nil },
		updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
	})

	got, err := svc.CreateService(context.Background(), "payments", "handles payments", nil)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}
	if got.Name != "payments" || got.HealthStatus != enums.HealthStatusUnknown.String() {
		t.Fatalf("unexpected service: %#v", got)
	}
}

func TestServiceServiceCreateServiceUsesProvidedHealth(t *testing.T) {
	health := enums.HealthStatusHealthy.String()
	svc := services.NewServiceService(serviceRepoStub{
		createFn: func(ctx context.Context, service *models.Service) error {
			if service.HealthStatus != health {
				t.Fatalf("unexpected health status: %s", service.HealthStatus)
			}
			return nil
		},
		findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
		findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) { return nil, nil },
		updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
	})

	got, err := svc.CreateService(context.Background(), "payments", "", &health)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}
	if got.HealthStatus != health {
		t.Fatalf("unexpected service: %#v", got)
	}
}

func TestServiceServiceCreateServiceRequiresName(t *testing.T) {
	svc := services.NewServiceService(serviceRepoStub{
		createFn: func(ctx context.Context, service *models.Service) error {
			t.Fatal("Create should not be called")
			return nil
		},
		findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
		findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) { return nil, nil },
		updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
	})

	_, err := svc.CreateService(context.Background(), "", "", nil)
	if err == nil || err.Error() != "service name is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceServiceCreateServicePropagatesRepositoryError(t *testing.T) {
	svc := services.NewServiceService(serviceRepoStub{
		createFn: func(ctx context.Context, service *models.Service) error {
			return errors.New("insert failed")
		},
		findAllFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
		findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) { return nil, nil },
		updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
	})

	_, err := svc.CreateService(context.Background(), "payments", "", nil)
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceServiceListServices(t *testing.T) {
	want := []models.Service{
		{ID: 1, Name: "payments", HealthStatus: enums.HealthStatusHealthy.String()},
		{ID: 2, Name: "checkout", HealthStatus: enums.HealthStatusDegraded.String()},
	}
	svc := services.NewServiceService(serviceRepoStub{
		createFn: func(ctx context.Context, service *models.Service) error { return nil },
		findAllFn: func(ctx context.Context) ([]models.Service, error) { return want, nil },
		findByIDFn: func(ctx context.Context, id uint) (*models.Service, error) { return nil, nil },
		updateHealthStatusFn: func(ctx context.Context, id uint, status enums.HealthStatus) error { return nil },
	})

	got, err := svc.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices returned error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d", len(got), len(want))
	}
}
