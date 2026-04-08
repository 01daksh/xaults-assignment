package enums

// HealthStatus represents the health state of a monitored microservice.
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusDown     HealthStatus = "down"
	HealthStatusUnknown  HealthStatus = "unknown"
)

func (h HealthStatus) IsValid() bool {
	switch h {
	case HealthStatusHealthy, HealthStatusDegraded, HealthStatusDown, HealthStatusUnknown:
		return true
	}
	return false
}

func (h HealthStatus) String() string {
	return string(h)
}