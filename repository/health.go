package repository

import (
	"context"
	"encoding"
	"strings"
)

// HealthChecker is the interface used to check the repository health.
type HealthChecker interface {
	// HealthCheck defines the health status of the repository.
	HealthCheck(ctx context.Context) (*HealthResponse, error)
}

// HealthResponse is the response for the health check.
type HealthResponse struct {
	Status HealthStatus
	Output string
	Notes  []string
}

// HealthStatus is the status of the health check
type HealthStatus int

// enum values for the health statuses
const (
	// StatusPass defines healthy status
	StatusPass HealthStatus = iota
	// StatusFail defines unhealthy result
	StatusFail
	// StatusWarn defines
	StatusWarn
)

func (s HealthStatus) String() string {
	switch s {
	case StatusPass:
		return "Pass"
	case StatusFail:
		return "Fail"
	case StatusWarn:
		return "Warn"
	default:
		return "Unknown"
	}
}

// compile time check for the health status
var _ encoding.TextMarshaler = HealthStatus(0)

// MarshalText implements encoding.TextMarshaler interface.
func (s HealthStatus) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
}
