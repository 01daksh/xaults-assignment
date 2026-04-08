package incidents

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"xaults-assignment/customError"
	"xaults-assignment/enums"
	"xaults-assignment/internal/interfaces"
)

type IncidentController struct {
	Svc interfaces.IncidentService
}

func NewIncidentController(svc interfaces.IncidentService) *IncidentController {
	return &IncidentController{Svc: svc}
}

type reportIncidentRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Severity    enums.Severity `json:"severity"`
}

func (ic *IncidentController) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/services/:id/incidents")
	g.POST("", ic.ReportIncident)
	g.GET("", ic.ListIncidents)
}

func (ic *IncidentController) ReportIncident(c *echo.Context) error {
	serviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid service id",
		})
	}

	var req reportIncidentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "title is required",
		})
	}
	if !req.Severity.IsValid() {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "severity must be one of: critical, high, medium, low",
		})
	}

	incident, err := ic.Svc.ReportIncident(c.Request().Context(), uint(serviceID), req.Title, req.Description, req.Severity)
	if err != nil {
		if errors.Is(err, customError.ErrServiceNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "service not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, incident)
}

func (ic *IncidentController) ListIncidents(c *echo.Context) error {
	serviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid service id",
		})
	}

	list, err := ic.Svc.ListIncidents(c.Request().Context(), uint(serviceID))
	if err != nil {
		if errors.Is(err, customError.ErrServiceNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "service not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, list)
}