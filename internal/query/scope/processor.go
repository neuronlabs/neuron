package scope

import (
	"context"
)

// Processor is the interface used to create the scope's of the queries
type Processor interface {
	// Create runs creation processes
	Create(ctx context.Context, s *Scope) error

	// Delete runs delete processes
	Delete(ctx context.Context, s *Scope) error

	// Get runs get processes
	Get(ctx context.Context, s *Scope) error

	// List runs the list processes
	List(ctx context.Context, s *Scope) error

	// Patch runs the patch processes
	Patch(ctx context.Context, s *Scope) error
}
