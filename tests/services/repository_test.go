package services_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"

	"xaults-assignment/enums"
	"xaults-assignment/internal/services"
	"xaults-assignment/models"
	"xaults-assignment/tests/testutil"
)

func TestServiceRepositoryCreate(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := services.NewServiceRepository()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "services" ("name","description","health_status","created_at","updated_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`)).
		WithArgs("payments", "handles payments", "healthy", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	service := &models.Service{Name: "payments", Description: "handles payments", HealthStatus: "healthy"}
	if err := repo.Create(context.Background(), service); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestServiceRepositoryFindAll(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := services.NewServiceRepository()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "health_status", "created_at", "updated_at"}).
			AddRow(1, "payments", "handles payments", "healthy", time.Unix(1, 0).UTC(), time.Unix(1, 0).UTC()).
			AddRow(2, "checkout", "handles checkout", "degraded", time.Unix(2, 0).UTC(), time.Unix(2, 0).UTC()))

	got, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected service count: %d", len(got))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestServiceRepositoryFindByID(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := services.NewServiceRepository()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE id = $1 ORDER BY "services"."id" LIMIT $2`)).
		WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "health_status", "created_at", "updated_at"}).
			AddRow(7, "payments", "handles payments", "healthy", time.Unix(1, 0).UTC(), time.Unix(1, 0).UTC()))

	got, err := repo.FindByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if got.ID != 7 || got.Name != "payments" {
		t.Fatalf("unexpected service: %#v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestServiceRepositoryFindByIDNotFound(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := services.NewServiceRepository()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE id = $1 ORDER BY "services"."id" LIMIT $2`)).
		WithArgs(uint(7), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByID(context.Background(), 7); err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestServiceRepositoryUpdateHealthStatus(t *testing.T) {
	_, mock, _ := testutil.MockGormDB(t)
	repo := services.NewServiceRepository()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "services" SET "health_status"=$1,"updated_at"=$2 WHERE id = $3`)).
		WithArgs(enums.HealthStatusDown, sqlmock.AnyArg(), uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.UpdateHealthStatus(context.Background(), 7, enums.HealthStatusDown); err != nil {
		t.Fatalf("UpdateHealthStatus returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
