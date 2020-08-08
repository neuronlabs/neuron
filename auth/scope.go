package auth

import (
	"context"

	"github.com/neuronlabs/neuron/database"
)

// ScopeCollection is collection for the authorization scope.
type ScopeCollection interface {
	CreateScope(ctx context.Context, db database.DB, scope, description string) (Scope, error)
	GetScope(ctx context.Context, db database.DB, scope string) (Scope, error)
	DeleteScope(ctx context.Context, db database.DB, scope string) error
}

// Scope is an interface that defines authorization scope.
type Scope interface {
	ScopeName() string
}
