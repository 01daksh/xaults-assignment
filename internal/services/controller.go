package services

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"xaults-assignment/internal/interfaces"
)

type ServiceController struct {
	Svc interfaces.ServiceService
}

func NewServiceController(svc interfaces.ServiceService) *ServiceController {
	return &ServiceController{Svc: svc}
}

type createServiceRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	HealthStatus *string `json:"health_status"`
}

func (sc *ServiceController) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/services")
	g.POST("", sc.CreateService)
	g.GET("", sc.ListServices)
}

func (sc *ServiceController) CreateService(c *echo.Context) error {
	var req createServiceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "name is required")
	}

	svc, err := sc.Svc.CreateService(c.Request().Context(), req.Name, req.Description, req.HealthStatus)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, svc)
}

func (sc *ServiceController) ListServices(c *echo.Context) error {
	services, err := sc.Svc.ListServices(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, services)
}
