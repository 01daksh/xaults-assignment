package incidents

import (
	"context"
	"fmt"

	"xaults-assignment/enums"
	"xaults-assignment/internal/database"
	"xaults-assignment/internal/interfaces"
	"xaults-assignment/models"
)

type IncidentRepository struct {
}

func NewIncidentRepository() interfaces.IncidentRepository {
	return &IncidentRepository{}
}

func (r *IncidentRepository) Create(ctx context.Context, incident *models.Incident) error {
	db, err := database.GetDBConnection()
	if err != nil {
		return fmt.Errorf("get db connection: %w", err)
	}
	if err := db.WithContext(ctx).Create(incident).Error; err != nil {
		return fmt.Errorf("create incident: %w", err)
	}
	return nil
}

func (r *IncidentRepository) FindByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	var list []models.Incident
	db, err := database.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("get db connection: %w", err)
	}
	if err := db.WithContext(ctx).
		Where("service_id = ?", serviceID).Order("reported_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list incidents for service %d: %w", serviceID, err)
	}
	return list, nil
}

func (r *IncidentRepository) FindOpenByServiceID(ctx context.Context, serviceID uint) ([]models.Incident, error) {
	var list []models.Incident
	db, err := database.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("get db connection: %w", err)
	}
	if err := db.WithContext(ctx).
		Where("service_id = ? AND status != ?", serviceID, enums.IncidentStatusResolved.String()).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("find open incidents for service %d: %w", serviceID, err)
	}
	return list, nil
}
