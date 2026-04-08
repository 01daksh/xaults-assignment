package router

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func RegisterHealthRoutes(e *echo.Echo, db *gorm.DB) {
	e.GET("/healthz", NewHealthHandler(db))
}

func NewHealthHandler(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
				"reason": err.Error(),
			})
		}
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
				"reason": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	}
}
