package service

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/filestore"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/store"
)

// Options is the structure that contains service options.
type Options struct {
	// Initializers.
	Initializers []core.Initializer
	// NamedInitializers provide the named initializers.
	NamedInitializers map[string]core.Initializer
	// Repositories and mapped models.
	DefaultRepository repository.Repository
	RepositoryModels  map[repository.Repository][]mapping.Model
	// Key-value stores.
	DefaultStore store.Store
	Stores       map[string]store.Store

	// File stores.
	DefaultFileStore filestore.Store
	FileStores       map[string]filestore.Store

	// Model options.
	NamingConvention mapping.NamingConvention
	AccountModel     auth.Account
	Models           []mapping.Model
	MigrateModels    []mapping.Model
	DefaultNotNull   bool
	// Service server.
	Server server.Server
	// Auth abstractions.
	Verifier      auth.Verifier
	Authenticator auth.Authenticator
	Tokener       auth.Tokener
	// Service detailed options.
	HandleSignals  bool
	SynchronousORM bool
	// UTCTimestamps sets the timestamps to be in UTC timezone.
	UTCTimestamps bool
}

func (o *Options) controllerOptions() *core.Options {
	cfg := &core.Options{
		NamingConvention:       o.NamingConvention,
		SynchronousConnections: o.SynchronousORM,
		UTCTimestamps:          o.UTCTimestamps,
	}
	return cfg
}

func defaultOptions() *Options {
	return &Options{
		HandleSignals:     true,
		NamingConvention:  mapping.SnakeCase,
		NamedInitializers: map[string]core.Initializer{},
		RepositoryModels:  map[repository.Repository][]mapping.Model{},
		Stores:            map[string]store.Store{},
		FileStores:        map[string]filestore.Store{},
	}
}

// Option is the function that sets the options for the service.
type Option func(o *Options)
