package database

import (
	"context"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// Builder is the interface used to build queries.
type Builder interface {
	// Scope returns current query scope.
	Scope() *query.Scope
	// Err returns error for given query.
	Err() error
	// Ctx returns current query context.
	Ctx() context.Context

	// Count returns the number of model instances in the repository matching given query.
	Count() (int64, error)
	// Insert models into repository. Returns error with class ErrViolationUnique if the model with given primary key already exists.
	Insert() error
	// update updates query models in the repository. Input models with non zero primary key the function would take
	// update these models instances. For a non zero single model with filters it would update all models that match
	// given query. In order to update all models in the repository use UpdateAll method.
	Update() (int64, error)
	// Exists check if input model exists.
	Exists() (bool, error)
	// Get gets single model that matches given query.
	Get() (mapping.Model, error)
	// Find gets all models that matches given query. For models with DeletedAt timestamp field, adds the
	// filter for not nulls DeletedAt models, unless there is other DeletedAt filter.
	Find() ([]mapping.Model, error)
	// deleteQuery deletes models matching given query. If the model has DeletedAt timestamp field it would do the
	// soft delete - update all models matching given query and set their deleted_at field with current timestamp.
	Delete() (int64, error)
	// refreshQuery gets all input models and refreshes all selected fields and included relations.
	Refresh() error

	// Select model fields for the query fieldSet.
	Select(fields ...*mapping.StructField) Builder
	// Where adds the query 'filter' with provided arguments. The query should be composed of the field name and it's
	// operator. In example: 'ID in', 3,4,5 - would result with a filter on primary key field named 'ID' with the
	// IN operator and it's value equal to 3,4 or 5.
	Where(filter string, arguments ...interface{}) Builder
	// Include adds the 'relation' with it's optional fieldset to get for each resulting model to find.
	Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) Builder
	// OrderBy sorts the results by the fields in the given order.
	OrderBy(fields ...query.Sort) Builder
	// Limit sets the maximum number of the results for given query.
	Limit(limit int64) Builder
	// Offset sets the number of the results to omit in the query.
	Offset(offset int64) Builder
	// Filter adds the filter field to the given query.
	Filter(filter filter.Filter) Builder
	// WhereOr creates a OrGroup filter for provided simple filters.
	WhereOr(filters ...filter.Simple) Builder

	// AddRelations adds the 'relations' models to input models for 'relationField'
	AddRelations(relationField *mapping.StructField, relations ...mapping.Model) error
	// querySetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
	// The relation's foreign key must be allowed to set to null.
	SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error
	// ClearRelations clears all 'relationField' relations for given input models.
	// The relation's foreign key must be allowed to set to null.
	RemoveRelations(relationField *mapping.StructField) (int64, error)
	// IncludeRelation adds the 'relations' models to input models for 'relationField'
	GetRelations(relationField *mapping.StructField, relationFieldSet ...*mapping.StructField) ([]mapping.Model, error)
}

// query is the query builder that allows to execute Queries in a callback manner.
type dbQuery struct {
	ctx   context.Context
	db    *base
	scope *query.Scope
	err   error
}

// Ctx returns query context.
func (b *dbQuery) Ctx() context.Context {
	return b.ctx
}

// Scope returns query scope.
func (b *dbQuery) Scope() *query.Scope {
	return b.scope
}

// Err returns query error.
func (b *dbQuery) Err() error {
	return b.err
}

/**
 *
 * Execute Methods
 *
 */

// Count returns the number of the values for the provided query scope.
func (b *dbQuery) Count() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return Count(b.ctx, b.db, b.scope)
}

// Insert stores the values within the given scope's value repository, by starting
// the create process.
func (b *dbQuery) Insert() error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return queryInsert(b.ctx, b.db, b.scope)
}

// update updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (b *dbQuery) Update() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return queryUpdate(b.ctx, b.db, b.scope)
}

// Find gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (b *dbQuery) Find() ([]mapping.Model, error) {
	if b.err != nil {
		return nil, b.err
	}
	b.setEmptyContext()
	return queryFind(b.ctx, b.db, b.scope)
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (b *dbQuery) Get() (mapping.Model, error) {
	if b.err != nil {
		return nil, b.err
	}
	b.setEmptyContext()
	return queryGet(b.ctx, b.db, b.scope)
}

// Exists returns true or false depending if there are any rows matching the query.
func (b *dbQuery) Exists() (bool, error) {
	if b.err != nil {
		return false, b.err
	}
	b.setEmptyContext()
	return Exists(b.ctx, b.db, b.scope)
}

// deleteQuery the models provided in the query's scope.
func (b *dbQuery) Delete() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return deleteQuery(b.ctx, b.db, b.scope)
}

// refreshQuery deletes all the models in the collection.
func (b *dbQuery) Refresh() error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return refreshQuery(b.ctx, b.db, b.scope)
}

/**
 *
 * CallBack functions
 *
 */

// Filter adds 'filter' *Filter to the query.
func (b *dbQuery) Filter(f filter.Filter) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Filter(f)
	return b
}

// Where parses the filter into the filters.Filter and adds it to the given scope.
func (b *dbQuery) Where(where string, values ...interface{}) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Where(where, values...)
	return b
}

// WhereOr implements Builder interface.
func (b *dbQuery) WhereOr(filters ...filter.Simple) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.WhereOr(filters...)
	return b
}

// Include includes 'relation' into given query.
func (b *dbQuery) Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Include(relation, relationFieldset...)
	return b
}

// Limit sets the maximum number of objects returned by the Find process,
func (b *dbQuery) Limit(limit int64) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Limit(limit)
	return b
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
func (b *dbQuery) Offset(offset int64) Builder {
	if b.err != nil {
		return b
	}
	b.scope.Offset(offset)
	return b
}

// Select adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (b *dbQuery) Select(fields ...*mapping.StructField) Builder {
	if b.err != nil {
		return b
	}
	b.err = b.scope.Select(fields...)
	return b
}

// OrderBy creates the sort order of the result.
func (b *dbQuery) OrderBy(fields ...query.Sort) Builder {
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
func (b *dbQuery) AddRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return queryAddRelations(b.ctx, b.db, b.scope, relationField, relations...)
}

// querySetRelations removes and adds relation models to the main model's 'relationField'.
func (b *dbQuery) SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.err != nil {
		return b.err
	}
	b.setEmptyContext()
	return querySetRelations(b.ctx, b.db, b.scope, relationField, relations...)
}

// ClearRelations removes relations from provided main model for given 'relationField'.
func (b *dbQuery) RemoveRelations(relationField *mapping.StructField) (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.setEmptyContext()
	return queryClearRelations(b.ctx, b.db, b.scope, relationField)
}

// GetRelations gets models for 'relationField' with optional 'relationFieldset'.
func (b *dbQuery) GetRelations(relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	if b.err != nil {
		return []mapping.Model{}, b.err
	}
	b.setEmptyContext()
	return queryGetRelations(b.ctx, b.db, b.scope.ModelStruct, b.scope.Models, relationField, relationFieldset...)
}

func (b *dbQuery) setEmptyContext() {
	if b.ctx == nil {
		b.ctx = context.Background()
	}
}
