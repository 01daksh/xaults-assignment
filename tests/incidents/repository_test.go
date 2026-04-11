package incidents_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"xaults-assignment/enums"
	"xaults-assignment/internal/incidents"
	"xaults-assignment/models"
	"xaults-assignment/tests/testutil"
)

func TestIncidentRepositoryCreate(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := incidents.NewIncidentRepository()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "incidents" ("service_id","title","description","severity","status","reported_at","resolved_at","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
		WithArgs(uint(7), "db down", "payments unavailable", enums.SeverityCritical, enums.IncidentStatusOpen, sqlmock.AnyArg(), nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	incident := &models.Incident{
		ServiceID:   7,
		Title:       "db down",
		Description: "payments unavailable",
		Severity:    enums.SeverityCritical,
		Status:      enums.IncidentStatusOpen,
	}
	if err := repo.Create(context.Background(), incident); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestIncidentRepositoryFindByServiceID(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := incidents.NewIncidentRepository()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "incidents" WHERE service_id = $1 ORDER BY reported_at DESC`)).
		WithArgs(uint(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "title", "description", "severity", "status", "reported_at", "resolved_at", "created_at", "updated_at"}).
			AddRow(2, 7, "db down", "payments unavailable", "critical", "open", time.Unix(2, 0).UTC(), nil, time.Unix(2, 0).UTC(), time.Unix(2, 0).UTC()).
			AddRow(1, 7, "latency high", "checkout slow", "medium", "open", time.Unix(1, 0).UTC(), nil, time.Unix(1, 0).UTC(), time.Unix(1, 0).UTC()))

	got, err := repo.FindByServiceID(context.Background(), 7)
	if err != nil {
		t.Fatalf("FindByServiceID returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected incident count: %d", len(got))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestIncidentRepositoryFindOpenByServiceID(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := incidents.NewIncidentRepository()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "incidents" WHERE service_id = $1 AND status != $2`)).
		WithArgs(uint(7), enums.IncidentStatusResolved.String()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "title", "description", "severity", "status", "reported_at", "resolved_at", "created_at", "updated_at"}).
			AddRow(2, 7, "db down", "payments unavailable", "critical", "open", time.Unix(2, 0).UTC(), nil, time.Unix(2, 0).UTC(), time.Unix(2, 0).UTC()))

	got, err := repo.FindOpenByServiceID(context.Background(), 7)
	if err != nil {
		t.Fatalf("FindOpenByServiceID returned error: %v", err)
	}
	if len(got) != 1 || got[0].Status != enums.IncidentStatusOpen {
		t.Fatalf("unexpected incidents: %#v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
