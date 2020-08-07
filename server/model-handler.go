package server

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

type Params struct {
	Ctx           context.Context
	Authorizer    auth.Authorizer
	Authenticator auth.Authenticator
	DB            database.DB
}

//
// Insert
//

// InsertMiddlewarer is an interface that gets the middlewares for the insert endpoint.
type InsertMiddlewarer interface {
	InsertMiddlewares() []Middleware
}

// BeforeInsertHandler is a hook handler interface before API insert query.
type BeforeInsertHandler interface {
	HandleBeforeInsert(params *Params, input *codec.Payload) error
}

// InsertHandler is a handler of an API insert query.
type InsertHandler interface {
	HandleInsert(ctx context.Context, params Params, input *codec.Payload) (*codec.Payload, error)
}

// AfterInsertHandler is a hook handler interface after API insert query.
type AfterInsertHandler interface {
	HandleAfterInsert(params *Params, input *codec.Payload) error
}

//
// update
//

// UpdateMiddlewarer is an interface that gets middlewares for the update endpoint.
type UpdateMiddlewarer interface {
	UpdateMiddlewares() []Middleware
}

// BeforeUpdateHandler is a hook handler interface before API update query.
type BeforeUpdateHandler interface {
	HandleBeforeUpdate(params *Params, input *codec.Payload) error
}

type UpdateHandler interface {
	HandleUpdate(ctx context.Context, params Params, input *codec.Payload) (*codec.Payload, error)
}

// AfterUpdateHandler is a hook handler interface after API update query.
type AfterUpdateHandler interface {
	HandleAfterUpdate(params *Params, input *codec.Payload) error
}

//
// Get
//

// GetMiddlewarer is an interface that gets middlewares for the 'Get' endpoint.
type GetMiddlewarer interface {
	GetMiddlewares() []Middleware
}

// BeforeGetHandler is the api hook handler before handling the query.
type BeforeGetHandler interface {
	HandleBeforeGet(params *Params, q *query.Scope) error
}

// AfterGetHandler is the api hook after handling get query.
type AfterGetHandler interface {
	HandleAfterGet(params *Params, payload *codec.Payload) error
}

// GetHandler is the handler used to get the model
type GetHandler interface {
	HandleGet(ctx context.Context, params Params, q *query.Scope) (*codec.Payload, error)
}

//
// List
//

// ListMiddlewarer is an interface that gets middlewares for the 'List' endpoint.
type ListMiddlewarer interface {
	ListMiddlewares() []Middleware
}

// BeforeListHandler is the api hook handler before handling the list query.
type BeforeListHandler interface {
	HandleBeforeList(params *Params, q *query.Scope) error
}

// AfterListHandler is the api hook after handling list query.
type AfterListHandler interface {
	HandleAfterList(params *Params, payload *codec.Payload) error
}

// ListHandler is an interface used for handling the API list request.
type ListHandler interface {
	HandleList(ctx context.Context, params Params, q *query.Scope) (*codec.Payload, error)
}

//
// deleteQuery
//

// DeleteMiddlewarer is an interface that gets middlewares for the 'deleteQuery' endpoint.
type DeleteMiddlewarer interface {
	DeleteMiddlewares() []Middleware
}

// BeforeDeleteHandler is a hook interface used to execute before API deleteQuery request.
type BeforeDeleteHandler interface {
	HandleBeforeDelete(params *Params, q *query.Scope) error
}

// AfterDeleteHandler is a hook interface used to execute after API deleteQuery request.
type AfterDeleteHandler interface {
	HandleAfterDelete(params *Params, q *query.Scope, result *codec.Payload) error
}

// DeleteHandler is an interface used for handling API 'deleteQuery' requests.
type DeleteHandler interface {
	HandleDelete(ctx context.Context, params Params, q *query.Scope) (*codec.Payload, error)
}

//
// GetRelated
//

// GetRelationMiddlewarer is an interface that gets middlewares for the 'GetRelated' endpoint.
type GetRelationMiddlewarer interface {
	GetRelatedMiddlewares() []Middleware
}

// BeforeGetRelationHandler is an interface used as a hook before API request for getting relations.
type BeforeGetRelationHandler interface {
	HandleBeforeGetRelation(params *Params, q, relatedQuery *query.Scope, relation *mapping.StructField) error
}

// AfterGetRelationHandler is an interface used as a hook after API request for getting relations.
type AfterGetRelationHandler interface {
	HandleAfterGetRelation(params *Params, result *codec.Payload) error
}

// GetRelationHandler is an API interface used for handling 'GetRelation' queries.
type GetRelationHandler interface {
	HandleGetRelation(ctx context.Context, params Params, q, relatedQuery *query.Scope, relation *mapping.StructField) (*codec.Payload, error)
}

// SetRelationsHandler is an API interface that handles request for setting 'newRelations' for 'model' in field 'relation.
// This handler is used for all the requests (deleteQuery, update, Insert) relations.
type SetRelationsHandler interface {
	HandleSetRelations(ctx context.Context, params Params, model mapping.Model, newRelations []mapping.Model, relation *mapping.StructField) (*codec.Payload, error)
}

//
// deleteQuery Relations
//

// DeleteRelationsMiddlewarer is an interface that gets middlewares for the 'DeleteRelation' endpoint.
type DeleteRelationsMiddlewarer interface {
	DeleteRelationsMiddlewares() []Middleware
}

// BeforeDeleteRelationsHandler is a hook interface used to execute before API delete relations request.
type BeforeDeleteRelationsHandler interface {
	HandleBeforeDeleteRelations(params *Params, root mapping.Model, input *codec.Payload) error
}

// AfterDeleteRelationsHandler is a hook interface used to execute after API delete relations request.
type AfterDeleteRelationsHandler interface {
	HandleAfterDeleteRelations(params *Params, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}

//
// Insert Relations
//

// InsertRelationsMiddlewarer is an interface that gets middlewares for the 'InsertRelation' endpoint.
type InsertRelationsMiddlewarer interface {
	InsertRelationsMiddlewares() []Middleware
}

// BeforeInsertRelationsHandler is a hook interface used to execute before API insert relations request.
type BeforeInsertRelationsHandler interface {
	HandleBeforeInsertRelations(params *Params, root mapping.Model, input *codec.Payload) error
}

// AfterInsertRelationsHandler is a hook interface used to execute after API insert relations request.
type AfterInsertRelationsHandler interface {
	HandleAfterInsertRelations(params *Params, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}

//
// update Relations
//

// UpdateRelationsMiddlewarer is an interface that gets middlewares for the 'UpdateRelation' endpoint.
type UpdateRelationsMiddlewarer interface {
	UpdateRelationsMiddlewares() []Middleware
}

// BeforeUpdateRelationsHandler is a hook interface used to execute before API update relations request.
type BeforeUpdateRelationsHandler interface {
	HandleBeforeUpdateRelations(params *Params, root mapping.Model, input *codec.Payload) error
}

// AfterUpdateRelationsHandler is a hook interface used to execute after API update relations request.
type AfterUpdateRelationsHandler interface {
	HandleAfterUpdateRelations(params *Params, root mapping.Model, newRelations []mapping.Model, output *codec.Payload) error
}
