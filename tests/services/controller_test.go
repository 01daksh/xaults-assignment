package services_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"xaults-assignment/internal/services"
	"xaults-assignment/models"
	"xaults-assignment/tests/testutil"
)

type serviceServiceStub struct {
	createFn func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error)
	listFn   func(ctx context.Context) ([]models.Service, error)
}

func (s serviceServiceStub) CreateService(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
	return s.createFn(ctx, name, description, healthStatus)
}

func (s serviceServiceStub) ListServices(ctx context.Context) ([]models.Service, error) {
	return s.listFn(ctx)
}

func TestCreateServiceSuccess(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			if name != "payments" {
				t.Fatalf("unexpected name: %s", name)
			}
			if description != "handles payments" {
				t.Fatalf("unexpected description: %s", description)
			}
			if healthStatus == nil || *healthStatus != "healthy" {
				t.Fatalf("unexpected health status: %#v", healthStatus)
			}

			return &models.Service{
				ID:           1,
				Name:         name,
				Description:  description,
				HealthStatus: *healthStatus,
				CreatedAt:    time.Unix(1, 0).UTC(),
				UpdatedAt:    time.Unix(1, 0).UTC(),
			}, nil
		},
		listFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services", `{"name":"payments","description":"handles payments","health_status":"healthy"}`)

	testutil.AssertStatus(t, rec, http.StatusCreated)
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("unexpected content type: %q", got)
	}

	var got models.Service
	got = testutil.DecodeJSON[models.Service](t, rec)
	if got.Name != "payments" || got.HealthStatus != "healthy" {
		t.Fatalf("unexpected response body: %#v", got)
	}
}

func TestCreateServiceInvalidJSON(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			t.Fatal("CreateService should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services", `{"name":`)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertMessageBody(t, rec.Body.String(), `{"message":"invalid request body"}`)
}

func TestCreateServiceMissingName(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			t.Fatal("CreateService should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services", `{"description":"handles payments"}`)

	testutil.AssertStatus(t, rec, http.StatusUnprocessableEntity)
	assertMessageBody(t, rec.Body.String(), `{"message":"name is required"}`)
}

func TestCreateServiceInternalError(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			return nil, errors.New("insert failed")
		},
		listFn: func(ctx context.Context) ([]models.Service, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services", `{"name":"payments"}`)

	testutil.AssertStatus(t, rec, http.StatusInternalServerError)
	assertMessageBody(t, rec.Body.String(), `{"message":"insert failed"}`)
}

func TestListServicesSuccess(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context) ([]models.Service, error) {
			return []models.Service{
				{ID: 1, Name: "payments", HealthStatus: "healthy"},
				{ID: 2, Name: "checkout", HealthStatus: "degraded"},
			}, nil
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services", "")

	testutil.AssertStatus(t, rec, http.StatusOK)
	got := testutil.DecodeJSON[[]models.Service](t, rec)
	if len(got) != 2 {
		t.Fatalf("unexpected service count: got %d", len(got))
	}
	if got[0].Name != "payments" || got[1].Name != "checkout" {
		t.Fatalf("unexpected response body: %#v", got)
	}
}

func TestListServicesInternalError(t *testing.T) {
	e := testutil.NewEcho()
	controller := services.NewServiceController(serviceServiceStub{
		createFn: func(ctx context.Context, name, description string, healthStatus *string) (*models.Service, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context) ([]models.Service, error) {
			return nil, errors.New("query failed")
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services", "")

	testutil.AssertStatus(t, rec, http.StatusInternalServerError)
	assertMessageBody(t, rec.Body.String(), `{"message":"query failed"}`)
}
func assertMessageBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want+"\n" {
		t.Fatalf("unexpected response body: got %q want %q", got, want+"\n")
	}
}
