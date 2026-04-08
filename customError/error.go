package customError

import "errors"

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrIncidentNotFound = errors.New("incident not found")
)
