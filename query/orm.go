package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/controller"
)

// ORM is the interface that allows to start new queries.
type ORM interface {
	// Controller gets current controller.
	Controller() *controller.Controller
	// Query creates a new query for provided 'model'.
	Query(model interface{}) Builder
	// QueryCtx creates a new query for provided 'model'. The query should take a context on it's
	QueryCtx(ctx context.Context, model interface{}) Builder
}

// Compile time Queryer interface implementations check.
var (
	_ ORM = &Tx{}
	_ ORM = &Scope{}
)
