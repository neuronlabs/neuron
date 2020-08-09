package database

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
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

// New creates new DB for given controller.
//nolint:golint
func New(c *controller.Controller) *db {
	return &db{c: c}
}

// Compile time check for the DB interface implementations.
var _ DB = &db{}

// Composer is the default query composer that implements DB interface.
type db struct {
	c *controller.Controller
}

// Begin starts new transaction with respect to the transaction context and transaction options with controller 'c'.
func (d *db) Begin(ctx context.Context, options *query.TxOptions) *Tx {
	return begin(ctx, d.c, options)
}

// Controller implements DB interface.
func (d *db) Controller() *controller.Controller {
	return d.c
}

// Query creates new query builder for given 'model' and it's optional instances 'models'.
func (d *db) Query(model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return d.query(context.Background(), model, models...)
}

// QueryCtx creates new query builder for given 'model' and it's optional instances 'models'.
func (d *db) QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return d.query(ctx, model, models...)
}

// QueryGet implements QueryGetter interface.
func (d *db) QueryGet(ctx context.Context, q *query.Scope) (mapping.Model, error) {
	return queryGet(ctx, d, q)
}

// QueryGet implements QueryGetter interface.
func (d *db) QueryFind(ctx context.Context, q *query.Scope) ([]mapping.Model, error) {
	return queryFind(ctx, d, q)
}

// Insert implements DB interface.
func (d *db) Insert(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if len(models) == 0 {
		return errors.Wrap(query.ErrNoModels, "nothing to insert")
	}
	s := query.NewScope(mStruct, models...)
	return queryInsert(ctx, d, s)
}

// InsertQuery implements QueryInserter interface.
func (d *db) InsertQuery(ctx context.Context, q *query.Scope) error {
	return queryInsert(ctx, d, q)
}

// Update implements DB interface.
func (d *db) Update(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.Wrap(query.ErrNoModels, "nothing to update")
	}
	s := query.NewScope(mStruct, models...)
	return queryUpdate(ctx, d, s)
}

// UpdateQuery implements QueryUpdater interface.
func (d *db) UpdateQuery(ctx context.Context, q *query.Scope) (int64, error) {
	return queryUpdate(ctx, d, q)
}

// deleteQuery implements DB interface.
func (d *db) Delete(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.Wrap(query.ErrNoModels, "nothing to delete")
	}
	s := query.NewScope(mStruct, models...)
	return deleteQuery(ctx, d, s)
}

// DeleteQuery implements QueryDeleter interface.
func (d *db) DeleteQuery(ctx context.Context, q *query.Scope) (int64, error) {
	return deleteQuery(ctx, d, q)
}

// Refresh implements DB interface.
func (d *db) Refresh(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if len(models) == 0 {
		return nil
	}
	q := query.NewScope(mStruct, models...)
	return refreshQuery(ctx, d, q)
}

// QueryRefresh implements QueryRefresher interface.
func (d *db) QueryRefresh(ctx context.Context, q *query.Scope) error {
	return refreshQuery(ctx, d, q)
}

//
// Relations
//

func (d *db) AddRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	mStruct, err := d.c.ModelStruct(model)
	if err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	return queryAddRelations(ctx, d, q, relationField, relations...)
}

// QueryAddRelations implements QueryRelationAdder interface.
func (d *db) QueryAddRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relations ...mapping.Model) error {
	return queryAddRelations(ctx, d, s, relationField, relations...)
}

// querySetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
// The relation's foreign key must be allowed to set to null.
func (d *db) SetRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	mStruct, err := d.c.ModelStruct(model)
	if err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	return querySetRelations(ctx, d, q, relationField, relations...)
}

var _ QueryRelationSetter = &db{}

// QuerySetRelations implements QueryRelationSetter interface.
func (d *db) QuerySetRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relations ...mapping.Model) error {
	return querySetRelations(ctx, d, s, relationField, relations...)
}

// ClearRelations clears all 'relationField' relations for given input models.
// The relation's foreign key must be allowed to set to null.
func (d *db) ClearRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField) (int64, error) {
	mStruct, err := d.c.ModelStruct(model)
	if err != nil {
		return 0, err
	}
	q := query.NewScope(mStruct, model)
	return queryClearRelations(ctx, d, q, relationField)
}

var _ QueryRelationClearer = &db{}

// QueryClearRelations implements QueryRelationClearer interface.
func (d *db) QueryClearRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField) (int64, error) {
	return queryClearRelations(ctx, d, s, relationField)
}

// IncludeRelation gets the relations at the 'relationField' for provided models. An optional relationFieldset might be provided.
func (d *db) IncludeRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	return queryIncludeRelation(ctx, d, mStruct, models, relationField, relationFieldset...)
}

// GetRelations implements DB interface.
func (d *db) GetRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	return queryGetRelations(ctx, d, mStruct, models, relationField, relationFieldset...)
}

func (d *db) query(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) *dbQuery {
	b := &dbQuery{ctx: ctx, db: d}
	b.scope = query.NewScope(model, models...)
	return b
}
