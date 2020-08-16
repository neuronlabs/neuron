package core

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/store"
)

// Options is the structure that contains service options.
type Options struct {
	Name              string
	Version           string
	NamingConvention  mapping.NamingConvention
	DefaultRepository repository.Repository
	RepositoryModels  map[repository.Repository][]mapping.Model
	DefaultStore      store.Store
	Stores            map[string]store.Store
	Collections       []database.Collection
	Models            []mapping.Model
	MigrateModels     []mapping.Model
	DefaultNotNull    bool
	Server            server.Server
	Authorizer        auth.Verifier
	Authenticator     auth.Authenticator
	HandleSignals     bool
	Context           context.Context
	SynchronousORM    bool
	UTCTimestamps     bool
}

func (o *Options) controllerOptions() *controller.Options {
	cfg := &controller.Options{
		NamingConvention:       o.NamingConvention,
		SynchronousConnections: o.SynchronousORM,
		UTCTimestamps:          o.UTCTimestamps,
	}
	return cfg
}

func defaultOptions() *Options {
	return &Options{
		HandleSignals:    true,
		NamingConvention: mapping.SnakeCase,
		RepositoryModels: map[repository.Repository][]mapping.Model{},
		Stores:           map[string]store.Store{},
	}
}

// Option is the function that sets the options for the service.
type Option func(o *Options)
