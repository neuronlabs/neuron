package server

import (
	"context"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

//
// Insert
//

// InsertTransactioner is an interface that forces  model insertion in a transaction.
// Returned transaction options may be returned nil.
type InsertTransactioner interface {
	InsertWithTransaction() *query.TxOptions
}

// WithContextInserter is an interface used for changing insert context.
type WithContextInserter interface {
	InsertWithContext(ctx context.Context) (context.Context, error)
}

// InsertMiddlewarer is an interface that gets the middlewares for the insert endpoint.
type InsertMiddlewarer interface {
	InsertMiddlewares() []Middleware
}

// BeforeInsertHandler is a hook handler interface before API insert query.
type BeforeInsertHandler interface {
	HandleBeforeInsert(ctx context.Context, db database.DB, input *codec.Payload) error
}

// InsertHandler is a handler of an API insert query.
type InsertHandler interface {
	HandleInsert(ctx context.Context, db database.DB, input *codec.Payload) (*codec.Payload, error)
}

// AfterInsertHandler is a hook handler interface after API insert query.
type AfterInsertHandler interface {
	HandleAfterInsert(ctx context.Context, db database.DB, input *codec.Payload) error
}

//
// update
//

// UpdateTransactioner is an interface for the model that sets the update query in transaction with provided options.
type UpdateTransactioner interface {
	UpdateWithTransaction() *query.TxOptions
}

// WithContextUpdater is an interface used for updating the context of an update query.
type WithContextUpdater interface {
	UpdateWithContext(ctx context.Context) (context.Context, error)
}

// UpdateMiddlewarer is an interface that gets middlewares for the update endpoint.
type UpdateMiddlewarer interface {
	UpdateMiddlewares() []Middleware
}

// BeforeUpdateHandler is a hook handler interface before API update query.
type BeforeUpdateHandler interface {
	HandleBeforeUpdate(ctx context.Context, db database.DB, input *codec.Payload) error
}

// UpdateHandler is a handler interface for the API update query.
type UpdateHandler interface {
	HandleUpdate(ctx context.Context, db database.DB, input *codec.Payload) (*codec.Payload, error)
}

// AfterUpdateHandler is a hook handler interface after API update query.
type AfterUpdateHandler interface {
	HandleAfterUpdate(ctx context.Context, db database.DB, input *codec.Payload) error
}

//
// Get
//

// GetTransactioner is an interface that sets the transaction for the get query.
type GetTransactioner interface {
	GetWithTransaction() *query.TxOptions
}

// WithContextGetter is an interface that changes the get query context.
type WithContextGetter interface {
	GetWithContext(ctx context.Context) (context.Context, error)
}

// GetMiddlewarer is an interface that gets middlewares for the 'Get' endpoint.
type GetMiddlewarer interface {
	GetMiddlewares() []Middleware
}

// BeforeGetHandler is the api hook handler before handling the query.
type BeforeGetHandler interface {
	HandleBeforeGet(ctx context.Context, db database.DB, q *query.Scope) error
}

// AfterGetHandler is the api hook after handling get query.
type AfterGetHandler interface {
	HandleAfterGet(ctx context.Context, db database.DB, payload *codec.Payload) error
}

// GetHandler is the handler used to get the model
type GetHandler interface {
	HandleGet(ctx context.Context, db database.DB, q *query.Scope) (*codec.Payload, error)
}

//
// List
//

// ListTransactioner is an interface that sets the list query in transaction.
type ListTransactioner interface {
	ListWithTransaction() *query.TxOptions
}

// WithContextLister is an interface that sets the context for the list query.
type WithContextLister interface {
	ListWithContext(ctx context.Context) (context.Context, error)
}

// ListMiddlewarer is an interface that gets middlewares for the 'List' endpoint.
type ListMiddlewarer interface {
	ListMiddlewares() []Middleware
}

// BeforeListHandler is the api hook handler before handling the list query.
type BeforeListHandler interface {
	HandleBeforeList(ctx context.Context, db database.DB, q *query.Scope) error
}

// ListHandler is an interface used for handling the API list request.
type ListHandler interface {
	HandleList(ctx context.Context, db database.DB, q *query.Scope) (*codec.Payload, error)
}

// AfterListHandler is the api hook after handling list query.
type AfterListHandler interface {
	HandleAfterList(ctx context.Context, db database.DB, payload *codec.Payload) error
}

//
// Delete
//

// DeleteTransactioner is an interface that sets the delete query in transaction.
type DeleteTransactioner interface {
	DeleteWithTransaction() *query.TxOptions
}

// WithContextDeleter is an interface that sets the context for the delete query.
type WithContextDeleter interface {
	DeleteWithContext(ctx context.Context) (context.Context, error)
}

// DeleteMiddlewarer is an interface that gets middlewares for the 'deleteQuery' endpoint.
type DeleteMiddlewarer interface {
	DeleteMiddlewares() []Middleware
}

// BeforeDeleteHandler is a hook interface used to execute before API deleteQuery request.
type BeforeDeleteHandler interface {
	HandleBeforeDelete(ctx context.Context, db database.DB, q *query.Scope) error
}

// DeleteHandler is an interface used for handling API 'deleteQuery' requests.
type DeleteHandler interface {
	HandleDelete(ctx context.Context, db database.DB, q *query.Scope) (*codec.Payload, error)
}

// AfterDeleteHandler is a hook interface used to execute after API deleteQuery request.
type AfterDeleteHandler interface {
	HandleAfterDelete(ctx context.Context, db database.DB, q *query.Scope, result *codec.Payload) error
}

//
// GetRelated
//

// GetRelatedTransactioner is an interface that sets the get related query in transaction.
type GetRelatedTransactioner interface {
	GetRelatedWithTransaction() *query.TxOptions
}

// WithContextGetRelated is an interface that sets the context for the get related query.
type WithContextGetRelated interface {
	GetRelatedWithContext(ctx context.Context) (context.Context, error)
}

// GetRelationMiddlewarer is an interface that gets middlewares for the 'GetRelated' endpoint.
type GetRelationMiddlewarer interface {
	GetRelatedMiddlewares() []Middleware
}

// BeforeGetRelationHandler is an interface used as a hook before API request for getting relations.
type BeforeGetRelationHandler interface {
	HandleBeforeGetRelation(ctx context.Context, db database.DB, q, relatedQuery *query.Scope, relation *mapping.StructField) error
}

// AfterGetRelationHandler is an interface used as a hook after API request for getting relations.
type AfterGetRelationHandler interface {
	HandleAfterGetRelation(ctx context.Context, db database.DB, result *codec.Payload) error
}

// GetRelationHandler is an API interface used for handling 'GetRelation' queries.
type GetRelationHandler interface {
	HandleGetRelation(ctx context.Context, db database.DB, modelQuery, relatedQuery *query.Scope, relation *mapping.StructField) (*codec.Payload, error)
}

// SetRelationsHandler is an API interface that handles request for setting 'newRelations' for 'model' in field 'relation.
// This handler is used for all the requests (deleteQuery, update, Insert) relations.
type SetRelationsHandler interface {
	HandleSetRelations(ctx context.Context, db database.DB, model mapping.Model, newRelations []mapping.Model, relation *mapping.StructField) (*codec.Payload, error)
}

//
// Delete Relations
//

// WithContextDeleteRelationer is an interface that sets the context for the delete relations query.
type WithContextDeleteRelationer interface {
	DeleteRelationsWithContext(ctx context.Context) (context.Context, error)
}

// DeleteRelationsMiddlewarer is an interface that gets middlewares for the 'DeleteRelation' endpoint.
type DeleteRelationsMiddlewarer interface {
	DeleteRelationsMiddlewares() []Middleware
}

// BeforeDeleteRelationsHandler is a hook interface used to execute before API delete relations request.
type BeforeDeleteRelationsHandler interface {
	HandleBeforeDeleteRelations(ctx context.Context, db database.DB, root mapping.Model, input *codec.Payload) error
}

// AfterDeleteRelationsHandler is a hook interface used to execute after API delete relations request.
type AfterDeleteRelationsHandler interface {
	HandleAfterDeleteRelations(ctx context.Context, db database.DB, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}

//
// Insert Relations
//

// WithContextInsertRelationer is an interface that sets the context for the insert relations query.
type WithContextInsertRelationer interface {
	InsertRelationsWithContext(ctx context.Context) (context.Context, error)
}

// InsertRelationsMiddlewarer is an interface that gets middlewares for the 'InsertRelation' endpoint.
type InsertRelationsMiddlewarer interface {
	InsertRelationsMiddlewares() []Middleware
}

// BeforeInsertRelationsHandler is a hook interface used to execute before API insert relations request.
type BeforeInsertRelationsHandler interface {
	HandleBeforeInsertRelations(ctx context.Context, db database.DB, root mapping.Model, input *codec.Payload) error
}

// AfterInsertRelationsHandler is a hook interface used to execute after API insert relations request.
type AfterInsertRelationsHandler interface {
	HandleAfterInsertRelations(ctx context.Context, db database.DB, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}

//
// Update Relations
//

// WithContextUpdateRelationer is an interface that sets the context for the update relations query.
type WithContextUpdateRelationer interface {
	UpdateRelationsWithContext(ctx context.Context) (context.Context, error)
}

// UpdateRelationsMiddlewarer is an interface that gets middlewares for the 'UpdateRelation' endpoint.
type UpdateRelationsMiddlewarer interface {
	UpdateRelationsMiddlewares() []Middleware
}

// BeforeUpdateRelationsHandler is a hook interface used to execute before API update relations request.
type BeforeUpdateRelationsHandler interface {
	HandleBeforeUpdateRelations(ctx context.Context, db database.DB, root mapping.Model, input *codec.Payload) error
}

// AfterUpdateRelationsHandler is a hook interface used to execute after API update relations request.
type AfterUpdateRelationsHandler interface {
	HandleAfterUpdateRelations(ctx context.Context, db database.DB, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}
