package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics exposed by this application.
var (
	// HTTPRequestsTotal counts HTTP requests by method, path template, and
	// response status code.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests partitioned by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDurationSeconds records per-route request latency.
	HTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// ActiveServicesTotal is the current count of registered services.
	ActiveServicesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_services_total",
			Help: "Current number of registered services.",
		},
	)

	// OpenIncidentsTotal tracks currently open incidents by severity.
	OpenIncidentsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "open_incidents_total",
			Help: "Current number of open incidents partitioned by severity.",
		},
		[]string{"severity"},
	)
)

// Middleware returns an Echo middleware that instruments every request with
// Prometheus counters and histograms.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			// Echo may defer converting returned errors into an HTTP response
			// until after middleware unwinds. Force the HTTP error handler
			// to run now so metrics observe the final status code.
			if err != nil {
				c.Echo().HTTPErrorHandler(c,err)
				err = nil
			}

			elapsed := time.Since(start).Seconds()

			req := c.Request()

			// Use the matched route pattern (e.g. "/services/:id") rather than
			// the raw URL path to avoid a high-cardinality label explosion.
			// RouteInfo.Path is a plain string field in Echo v5.
			path := c.RouteInfo().Path
			if path == "" {
				path = req.URL.Path
			}

			// c.Response() returns http.ResponseWriter; the concrete type is
			// *echo.Response which carries the actual Status int field.
			statusCode := http.StatusOK
			if r, ok := c.Response().(*echo.Response); ok {
				if r.Status != 0 {
					statusCode = r.Status
				}
			}

			status := strconv.Itoa(statusCode)

			HTTPRequestsTotal.WithLabelValues(req.Method, path, status).Inc()
			HTTPRequestDurationSeconds.WithLabelValues(req.Method, path).Observe(elapsed)

			return err
		}
	}
}

// Handler returns the standard Prometheus HTTP handler for /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
