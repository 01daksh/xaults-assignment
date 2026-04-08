package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xaults-assignment/config"
	"xaults-assignment/models"
)

var db *gorm.DB

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	return NewPostgresDB(cfg)
}

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	var err error
	db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("retrieve underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Minute)

	if err := db.AutoMigrate(&models.Service{}, &models.Incident{}); err != nil {
		return nil, fmt.Errorf("run auto-migration: %w", err)
	}

	return db, nil
}

func GetDBConnection() (*gorm.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db, nil
}