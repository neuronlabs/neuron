package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/controller"
)

// Builder is the interface used to build queries.
type Builder interface {
	Scope() *Scope
	Err() error
	Ctx() context.Context

	Count() (int64, error)
	Create() error
	Patch() error
	List() error
	Get() error
	Delete() error

	AddFilterField(filter *FilterField) Builder
	Filter(filter string, values ...interface{}) Builder
	Include(relation string, relationFieldset ...string) Builder
	Limit(limit int64) Builder
	Offset(offset int64) Builder
	PageSize(pageSize int64) Builder
	PageNumber(pageNumber int64) Builder
	Processor(processor *Processor) Builder
	Fields(fields ...interface{}) Builder
	Sort(fields ...string) Builder
}

// compile time check for the query builder.
var _ Builder = &Query{}

// Query is the query builder that allows to execute Queries in a callback manner.
type Query struct {
	ctx   context.Context
	scope *Scope
	err   error
}

// NewQuery creates new query Query for given 'model' and default controller.
func NewQuery(ctx context.Context, c *controller.Controller, model interface{}) *Query {
	return newQuery(ctx, c, model)
}

func newQuery(ctx context.Context, c *controller.Controller, model interface{}) *Query {
	b := &Query{ctx: ctx}
	b.scope, b.err = NewC(c, model)
	return b
}

// Ctx returns query context.
func (b *Query) Ctx() context.Context {
	return b.ctx
}

// Scope returns query scope.
func (b *Query) Scope() *Scope {
	return b.scope
}

// Err returns query error.
func (b *Query) Err() error {
	return b.err
}

/**
 *
 * Execute Methods
 *
 */

// Count returns the number of the values for the provided query scope.
func (b *Query) Count() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.CountContext(b.ctx)
}

// Create stores the values within the given scope's value repository, by starting
// the create process.
func (b *Query) Create() error {
	if b.err != nil {
		return b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.CreateContext(b.ctx)
}

// Patch updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (b *Query) Patch() error {
	if b.err != nil {
		return b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.PatchContext(b.ctx)
}

// List gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (b *Query) List() error {
	if b.err != nil {
		return b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.ListContext(b.ctx)
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (b *Query) Get() error {
	if b.err != nil {
		return b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.GetContext(b.ctx)
}

// Delete deletes the values provided in the query's scope.
func (b *Query) Delete() error {
	if b.err != nil {
		return b.err
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	return b.scope.DeleteContext(b.ctx)
}

/**
 *
 * CallBack functions
 *
 */

// AddFilterField adds 'filter' *FilterField to the query.
func (b *Query) AddFilterField(filter *FilterField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.FilterField(filter)
	return b
}

// Filter parses the filter into the filters.FilterField and adds it to the given scope.
func (b *Query) Filter(filter string, values ...interface{}) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Filter(filter, values...)
	return b
}

// Include includes 'relation' into given query.
func (b *Query) Include(relation string, relationFieldset ...string) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Include(relation, relationFieldset...)
	return b
}

// Limit sets the maximum number of objects returned by the List process,
func (b *Query) Limit(limit int64) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Limit(limit)
	return b
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
func (b *Query) Offset(offset int64) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Offset(offset)
	return b
}

// PageSize defines pagination page size - maximum amount of returned objects.
func (b *Query) PageSize(pageSize int64) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.PageSize(pageSize)
	return b
}

// PageNumber defines the pagination page number.
func (b *Query) PageNumber(pageNumber int64) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.PageNumber(pageNumber)
	return b
}

// Processor sets the query processor. Implements Builder interface.
func (b *Query) Processor(processor *Processor) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Processor = processor
	return b
}

// Fields adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (b *Query) Fields(fields ...interface{}) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.SetFields(fields...)
	return b
}

// Sort creates the sort order of the result.
func (b *Query) Sort(fields ...string) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Sort(fields...)
	return b
}
