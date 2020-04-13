package neuron

import (
	"context"

	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/query"
)

// Compile time Queryer interface implementations check.
var _ query.ORM = &Neuron{}

// Query creates new query for the provided 'model' using the default controller.
func Query(model interface{}) query.Builder {
	return query.NewQuery(context.Background(), controller.Default(), model)
}

// QueryCtx creates new query for the provided 'model' using the default controller with respect to the context 'ctx'.
func QueryCtx(ctx context.Context, model interface{}) query.Builder {
	return query.NewQuery(ctx, controller.Default(), model)
}

// QueryC creates new query for the provided 'model' using 'c' controller.
func QueryC(c *controller.Controller, model interface{}) query.Builder {
	return query.NewQuery(context.Background(), c, model)
}

// QueryCtxC creates new query for the provided 'model' using 'c' controller with respect to the context 'ctx'.
func QueryCtxC(ctx context.Context, c *controller.Controller, model interface{}) query.Builder {
	return query.NewQuery(ctx, c, model)
}
