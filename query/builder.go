package query

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

// DB is the common interface that allows to do the queries.
type DB interface {
	// Controller gets current controller.
	Controller() *controller.Controller
	// Query creates a new query for provided 'model'.
	Query(model *mapping.ModelStruct, models ...mapping.Model) Builder
	// QueryCtx creates a new query for provided 'model'. The query should take a context on it's
	QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder
}

// Compile time check for the DB interface implementations.
var (
	_ DB = &Tx{}
	_ DB = &Creator{}
)

// Builder is the interface used to build queries.
type Builder interface {
	Scope() *Scope
	Err() error
	Ctx() context.Context

	Count() (int64, error)
	Insert() error
	Update() (int64, error)
	Exists() (bool, error)
	Get() (mapping.Model, error)
	Find() ([]mapping.Model, error)
	Delete() (int64, error)

	Select(fields ...*mapping.StructField) Builder
	Where(filter string, values ...interface{}) Builder
	Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) Builder
	OrderBy(fields ...*SortField) Builder
	Limit(limit int64) Builder
	Offset(offset int64) Builder
	Filter(filter *FilterField) Builder

	AddRelations(relationField *mapping.StructField, relations ...mapping.Model) error
	SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error
	RemoveRelations(relationField *mapping.StructField) (int64, error)
}

// Compile time check for the query builder.
var _ Builder = &query{}
