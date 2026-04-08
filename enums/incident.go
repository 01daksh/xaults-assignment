package enums

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow:
		return true
	}
	return false
}

type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusMonitoring IncidentStatus = "monitoring"
	IncidentStatusResolved   IncidentStatus = "resolved"
)

func (s IncidentStatus) String() string {
	return string(s)
}

func (s IncidentStatus) IsValid() bool {
	switch s {
	case IncidentStatusOpen, IncidentStatusMonitoring, IncidentStatusResolved:
		return true
	}
	return false
}
