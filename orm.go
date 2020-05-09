package neuron

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/query"
)

// Compile time Queryer interface implementations check.
var _ query.DB = &Neuron{}

// query creates new query for the provided 'model' using the default controller.
func Query(model interface{}) query.Builder {
	return query.NewCtx(context.Background(), controller.Default(), model)
}

// QueryCtx creates new query for the provided 'model' using the default controller with respect to the context 'ctx'.
func QueryCtx(ctx context.Context, model interface{}) query.Builder {
	return query.NewCtx(ctx, controller.Default(), model)
}
