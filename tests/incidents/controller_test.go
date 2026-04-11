package incidents_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"xaults-assignment/customError"
	"xaults-assignment/enums"
	"xaults-assignment/internal/incidents"
	"xaults-assignment/models"
	"xaults-assignment/tests/testutil"
)

type incidentServiceStub struct {
	reportFn func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error)
	listFn   func(ctx context.Context, serviceID uint) ([]models.Incident, error)
}

func (s incidentServiceStub) ReportIncident(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
	return s.reportFn(ctx, serviceID, title, description, severity)
}

func (s incidentServiceStub) ListIncidents(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	return s.listFn(ctx, serviceID)
}

func TestReportIncidentSuccess(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			if serviceID != 42 {
				t.Fatalf("unexpected serviceID: %d", serviceID)
			}
			if title != "database down" || description != "payments db unavailable" || severity != enums.SeverityCritical {
				t.Fatalf("unexpected incident request: title=%q description=%q severity=%q", title, description, severity)
			}

			return &models.Incident{
				ID:          1,
				ServiceID:   serviceID,
				Title:       title,
				Description: description,
				Severity:    severity,
				Status:      enums.IncidentStatusOpen,
				ReportedAt:  time.Unix(1, 0).UTC(),
				CreatedAt:   time.Unix(1, 0).UTC(),
				UpdatedAt:   time.Unix(1, 0).UTC(),
			}, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"title":"database down","description":"payments db unavailable","severity":"critical"}`)

	testutil.AssertStatus(t, rec, http.StatusCreated)
	got := testutil.DecodeJSON[models.Incident](t, rec)
	if got.ServiceID != 42 || got.Severity != enums.SeverityCritical {
		t.Fatalf("unexpected response body: %#v", got)
	}
}

func TestReportIncidentInvalidServiceID(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			t.Fatal("ReportIncident should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/not-a-number/incidents", `{"title":"database down","severity":"critical"}`)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertBody(t, rec.Body.String(), `{"error":"invalid service id"}`)
}

func TestReportIncidentInvalidJSON(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			t.Fatal("ReportIncident should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"title":`)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertBody(t, rec.Body.String(), `{"error":"invalid request body"}`)
}

func TestReportIncidentMissingTitle(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			t.Fatal("ReportIncident should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"severity":"high"}`)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertBody(t, rec.Body.String(), `{"error":"title is required"}`)
}

func TestReportIncidentInvalidSeverity(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			t.Fatal("ReportIncident should not be called")
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"title":"database down","severity":"urgent"}`)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertBody(t, rec.Body.String(), `{"error":"severity must be one of: critical, high, medium, low"}`)
}

func TestReportIncidentServiceNotFound(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, customError.ErrServiceNotFound
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"title":"database down","severity":"high"}`)

	testutil.AssertStatus(t, rec, http.StatusNotFound)
	assertBody(t, rec.Body.String(), `{"error":"service not found"}`)
}

func TestReportIncidentInternalError(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, errors.New("insert failed")
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) { return nil, nil },
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodPost, "/services/42/incidents", `{"title":"database down","severity":"high"}`)

	testutil.AssertStatus(t, rec, http.StatusInternalServerError)
	assertBody(t, rec.Body.String(), `{"error":"insert failed"}`)
}

func TestListIncidentsSuccess(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) {
			if serviceID != 42 {
				t.Fatalf("unexpected serviceID: %d", serviceID)
			}
			return []models.Incident{
				{ID: 1, ServiceID: 42, Title: "database down", Severity: enums.SeverityCritical, Status: enums.IncidentStatusOpen},
				{ID: 2, ServiceID: 42, Title: "latency high", Severity: enums.SeverityMedium, Status: enums.IncidentStatusOpen},
			}, nil
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services/42/incidents", "")

	testutil.AssertStatus(t, rec, http.StatusOK)
	got := testutil.DecodeJSON[[]models.Incident](t, rec)
	if len(got) != 2 {
		t.Fatalf("unexpected incident count: got %d", len(got))
	}
}

func TestListIncidentsInvalidServiceID(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) {
			t.Fatal("ListIncidents should not be called")
			return nil, nil
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services/not-a-number/incidents", "")

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
	assertBody(t, rec.Body.String(), `{"error":"invalid service id"}`)
}

func TestListIncidentsServiceNotFound(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) {
			return nil, customError.ErrServiceNotFound
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services/42/incidents", "")

	testutil.AssertStatus(t, rec, http.StatusNotFound)
	assertBody(t, rec.Body.String(), `{"error":"service not found"}`)
}

func TestListIncidentsInternalError(t *testing.T) {
	e := testutil.NewEcho()
	controller := incidents.NewIncidentController(incidentServiceStub{
		reportFn: func(ctx context.Context, serviceID uint, title, description string, severity enums.Severity) (*models.Incident, error) {
			return nil, nil
		},
		listFn: func(ctx context.Context, serviceID uint) ([]models.Incident, error) {
			return nil, errors.New("query failed")
		},
	})
	controller.RegisterRoutes(e)

	rec := testutil.JSONRequest(t, e, http.MethodGet, "/services/42/incidents", "")

	testutil.AssertStatus(t, rec, http.StatusInternalServerError)
	assertBody(t, rec.Body.String(), `{"error":"query failed"}`)
}

func assertBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want+"\n" {
		t.Fatalf("unexpected response body: got %q want %q", got, want+"\n")
	}
}
