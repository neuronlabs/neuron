package query

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

// DB is the interface that allows to start new queries.

// Creator is the default query composer that implements DB interface.
type Creator struct {
	c *controller.Controller
}

// Controller implements DB interface.
func (c *Creator) Controller() *controller.Controller {
	return c.c
}

// Query creates new query builder for given 'model' and it's optional instances 'models'.
func (c *Creator) Query(model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return c.query(context.Background(), model, models...)
}

// QueryCtx creates new query builder for given 'model' and it's optional instances 'models'.
func (c *Creator) QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return c.query(ctx, model, models...)
}

// NewComposer creates new query composer.
func NewComposer(c *controller.Controller) *Creator {
	return &Creator{c: c}
}

func (c *Creator) query(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) *query {
	b := &query{ctx: ctx}
	b.scope = newQueryScope(c, model, models...)
	return b
}

// query is the query builder that allows to execute Queries in a callback manner.
type query struct {
	ctx   context.Context
	scope *Scope
	err   error
}

// Ctx returns query context.
func (b *query) Ctx() context.Context {
	return b.ctx
}

// Scope returns query scope.
func (b *query) Scope() *Scope {
	return b.scope
}

// Err returns query error.
func (b *query) Err() error {
	return b.err
}

/**
 *
 * Execute Methods
 *
 */

// Count returns the number of the values for the provided query scope.
func (b *query) Count() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return b.scope.Count(b.ctx)
}

// Insert stores the values within the given scope's value repository, by starting
// the create process.
func (b *query) Insert() error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return b.scope.Insert(b.ctx)
}

// Update updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (b *query) Update() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return b.scope.Update(b.ctx)
}

// Find gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (b *query) Find() ([]mapping.Model, error) {
	if b.err != nil {
		return nil, b.err
	}
	b.setEmptyContext()
	return b.scope.Find(b.ctx)
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (b *query) Get() (mapping.Model, error) {
	if b.err != nil {
		return nil, b.err
	}
	b.setEmptyContext()
	return b.scope.Get(b.ctx)
}

// Exists returns true or false depending if there are any rows matching the query.
func (b *query) Exists() (bool, error) {
	if b.err != nil {
		return false, b.err
	}
	b.setEmptyContext()
	return b.scope.Exists(b.ctx)
}

// Delete deletes the values provided in the query's scope.
func (b *query) Delete() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return b.scope.Delete(b.ctx)
}

/**
 *
 * CallBack functions
 *
 */

// Filter adds 'filter' *Filter to the query.
func (b *query) Filter(filter *FilterField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Filter(filter)
	return b
}

// Where parses the filter into the filters.Filter and adds it to the given scope.
func (b *query) Where(filter string, values ...interface{}) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Where(filter, values...)
	return b
}

// Include includes 'relation' into given query.
func (b *query) Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Include(relation, relationFieldset...)
	return b
}

// Limit sets the maximum number of objects returned by the Find process,
func (b *query) Limit(limit int64) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Limit(limit)
	return b
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
func (b *query) Offset(offset int64) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Offset(offset)
	return b
}

// Select adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (b *query) Select(fields ...*mapping.StructField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Select(fields...)
	return b
}

// OrderBy creates the sort order of the result.
func (b *query) OrderBy(fields ...*SortField) Builder {
	if b.err != nil {
		return b
	}
	b.scope.SortingOrder = fields
	return b
}

//
// Relations
//

// AddRelations adds the relation models to the main model's 'relationField'.
func (b *query) AddRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return b.scope.addRelations(b.ctx, relationField, relations...)
}

// SetRelations removes and adds relation models to the main model's 'relationField'.
func (b *query) SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return b.scope.setRelations(b.ctx, relationField, relations...)
}

// RemoveRelations removes relations from provided main model for given 'relationField'.
func (b *query) RemoveRelations(relationField *mapping.StructField) (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return b.scope.removeRelations(b.ctx, relationField)
}

func (b *query) setEmptyContext() {
	if b.ctx == nil {
		b.ctx = context.Background()
	}
}
