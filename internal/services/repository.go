package services

import (
	"context"
	"fmt"

	"xaults-assignment/enums"
	"xaults-assignment/internal/database"
	"xaults-assignment/internal/interfaces"
	"xaults-assignment/models"
)

type ServiceRepository struct {
}

func NewServiceRepository() interfaces.ServiceRepository {
	return &ServiceRepository{}
}

func (r *ServiceRepository) Create(ctx context.Context, service *models.Service) error {
	db, err := database.GetDBConnection()
	if err != nil {
		return fmt.Errorf("couldn't get the db connection: %w", err)
	}

	if err := db.WithContext(ctx).Create(service).Error; err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) FindAll(ctx context.Context) ([]models.Service, error) {
	var services []models.Service
	db, err := database.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get the db connection: %w", err)
	}

	if err := db.WithContext(ctx).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return services, nil
}

func (r *ServiceRepository) FindByID(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	db, err := database.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get the db connection: %w", err)
	}
	if err := db.WithContext(ctx).Where("id = ?", id).First(&service).Error; err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	return &service, nil
}

func (r *ServiceRepository) UpdateHealthStatus(ctx context.Context, id uint, status enums.HealthStatus) error {
	db, err := database.GetDBConnection()
	if err != nil {
		return fmt.Errorf("couldn't get the db connection: %w", err)
	}
	result := db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id = ?", id).
		Update("health_status", status)
	if result.Error != nil {
		return fmt.Errorf("update health status: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("service %d not found", id)
	}
	return nil
}
