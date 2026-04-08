package router

import (
	"fmt"
	"xaults-assignment/internal/incidents"
	internalmetrics "xaults-assignment/internal/metrics"
	"xaults-assignment/internal/services"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	applyMiddleware(e)

	services.NewServiceWire().RegisterRoutes(e)
	incidents.NewIncidentWire().RegisterRoutes(e)

	RegisterHealthRoutes(e, db)

	// Prometheus metrics scrape endpoint.
	e.GET("/metrics", func(c *echo.Context) error {
		internalmetrics.Handler().ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func applyMiddleware(e *echo.Echo) {

	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c *echo.Context, v echomiddleware.RequestLoggerValues) error {
			fmt.Printf("REQUEST uri=%s status=%v\n", v.URI, v.Status)
			return nil
		},
	}))
	e.Use(internalmetrics.Middleware())
}
