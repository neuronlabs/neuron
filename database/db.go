package database

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// DB is the common interface that allows to do the queries.
type DB interface {
	// Query creates a new query for provided 'model'.
	Query(model *mapping.ModelStruct, models ...mapping.Model) Builder
	// QueryCtx creates a new query for provided 'model'. The query should take a context on it's
	QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder

	// Insert inserts provided models into mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to insert multiple models at once.
	Insert(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error
	// update updates provided models in their mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to update multiple models at once.
	Update(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error)
	// deleteQuery deletes provided models from their mapped repositories. The DB would use models non zero fields. Provided models
	// must have non zero primary key values. The function allows to delete multiple model values at once.
	Delete(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error)
	// refreshQuery refreshes fields (attributes, foreign keys) for provided models.
	Refresh(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error

	//
	// Relations
	//

	// AddRelations adds the 'relations' models to input models for 'relationField'
	AddRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error
	// querySetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
	// The relation's foreign key must be allowed to set to null.
	SetRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error
	// ClearRelations clears all 'relationField' relations for given input models.
	// The relation's foreign key must be allowed to set to null.
	ClearRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField) (int64, error)
	// IncludeRelation includes the relations at the 'relationField' for provided models. An optional relationFieldset might be provided.
	IncludeRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) error
	// GetRelations gets the 'relatedField' Models for provided the input 'models. An optional relationFieldset might be provided for the relation models.
	GetRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error)

	// Now returns current time used by the database layer.
	Now() time.Time

	// ModelMap gets related model map.
	ModelMap() *mapping.ModelMap
}

type repositoryMapper interface {
	// Repositories gets the repository mapper.
	mapper() *RepositoryMapper
}

// QueryFinder is an interface that allows to list results from the query.
type QueryFinder interface {
	// QueryFind gets the values from the repository with respect to the query filters, sorts, pagination and included values.
	// Provided 'ctx' context.Context would be used while querying the repositories.
	QueryFind(ctx context.Context, q *query.Scope) ([]mapping.Model, error)
}

// QueryGetter is an interface that allows to get single result from query.
type QueryGetter interface {
	// QueryGet gets single value from the repository taking into account the scope filters and parameters.
	QueryGet(ctx context.Context, q *query.Scope) (mapping.Model, error)
}

// QueryRefresher is an interface that allows to refresh the models in the query scope.
type QueryRefresher interface {
	QueryRefresh(ctx context.Context, q *query.Scope) error
}

// QueryInserter is an interface that allows to store the values in the query scope.
type QueryInserter interface {
	// InsertQuery stores the values within the given scope's value repository.
	InsertQuery(ctx context.Context, q *query.Scope) error
}

// QueryUpdater is an interface that allows to update the values in the query scope.
type QueryUpdater interface {
	// UpdateQuery updates the models with selected fields or a single model with provided filters.
	UpdateQuery(ctx context.Context, q *query.Scope) (modelsAffected int64, err error)
}

// QueryDeleter is an interface that allows to delete the model values in the query scope.
type QueryDeleter interface {
	// DeleteQuery deletes the values provided in the query's scope.
	DeleteQuery(ctx context.Context, q *query.Scope) (int64, error)
}

// QueryRelationAdder is an interface that allows to get the relations from the provided query results.
type QueryRelationAdder interface {
	// AddRelations appends relationship 'relField' 'relationModels' to the scope values.
	QueryAddRelations(ctx context.Context, s *query.Scope, relField *mapping.StructField, relationModels ...mapping.Model) error
}

// QueryRelationSetter is an interface that allows to set the query relation for provided query scope.
type QueryRelationSetter interface {
	// querySetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
	// The relation's foreign key must be allowed to set to null.
	QuerySetRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relationModels ...mapping.Model) error
}

// QueryRelationClearer is an interface that allows to clear the relations for provided query scope.
type QueryRelationClearer interface {
	// ClearRelations clears all 'relationField' relations for given query.
	// The relation's foreign key must be allowed to set to null.
	QueryClearRelations(ctx context.Context, s *query.Scope, relField *mapping.StructField) (int64, error)
}

type synchronousConnector interface {
	synchronousConnections() bool
}
