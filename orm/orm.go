package orm

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

// DB is the common interface that allows to do the queries.
type DB interface {
	// Controller returns orm based controller.
	Controller() *controller.Controller
	// Query creates a new query for provided 'model'.
	Query(model *mapping.ModelStruct, models ...mapping.Model) Builder
	// QueryCtx creates a new query for provided 'model'. The query should take a context on it's
	QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder

	// Insert inserts provided models into mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to insert multiple models at once.
	Insert(ctx context.Context, models ...mapping.Model) error
	// Update updates provided models in their mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to update multiple models at once.
	Update(ctx context.Context, models ...mapping.Model) (int64, error)
	// Delete deletes provided models from their mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to delete multiple model values at once.
	Delete(ctx context.Context, models ...mapping.Model) (int64, error)
}
