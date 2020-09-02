package service

import (
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/filestore"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/store"
)

// Options is the structure that contains service options.
type Options struct {
	// Database Layer
	MigrateModels          []mapping.Model
	DefaultRepository      repository.Repository
	RepositoryModels       map[repository.Repository][]mapping.Model
	SynchronousConnections bool

	// Key-value stores.
	DefaultStore store.Store
	Stores       map[string]store.Store

	// File stores.
	DefaultFileStore filestore.Store
	FileStores       map[string]filestore.Store

	// Model options.
	NamingConvention mapping.NamingConvention
	Models           []mapping.Model
	DefaultNotNull   bool

	// Service server.
	Server server.Server

	// Auth abstractions.
	Verifier      auth.Verifier
	Authenticator auth.Authenticator
	Tokener       auth.Tokener

	// Service detailed options.
	HandleSignals bool
	// TimeFunc sets the time function for the service.
	TimeFunc func() time.Time
}

func defaultOptions() *Options {
	return &Options{
		HandleSignals:    true,
		NamingConvention: mapping.SnakeCase,
		RepositoryModels: map[repository.Repository][]mapping.Model{},
		Stores:           map[string]store.Store{},
		FileStores:       map[string]filestore.Store{},
		TimeFunc:         time.Now,
	}
}

// Option is the function that sets the options for the service.
type Option func(o *Options)

// WithAuthenticator sets the authenticator option.
func WithAuthenticator(authenticator auth.Authenticator) Option {
	return func(o *Options) {
		o.Authenticator = authenticator
	}
}

// WithDefaultRepository sets the default repository for all models without specified repository.
func WithDefaultRepository(r repository.Repository) Option {
	return func(o *Options) {
		o.DefaultRepository = r
	}
}

// WithDefaultStore sets the default store for the service.
func WithDefaultStore(s store.Store) Option {
	return func(o *Options) {
		o.DefaultStore = s
	}
}

// WithDefaultFileStore sets the default file store for the service.
func WithDefaultFileStore(s filestore.Store) Option {
	return func(o *Options) {
		o.DefaultFileStore = s
	}
}

// DefaultNotNullFields sets the non-pointer model fields to not null by default.
// By default all fields are set as nullable in the model's definition.
// WithThis might be helpful in the model migration.
func WithDefaultNotNullFields(o *Options) {
	o.DefaultNotNull = true
}

// WithFileStore sets the service store 's' obtainable at 'name'.
func WithFileStore(name string, s filestore.Store) Option {
	return func(o *Options) {
		o.FileStores[name] = s
	}
}

// WithHandleSignal is the option that determines if the os signals should be handled by the service.
func WithHandleSignal(handle bool) Option {
	return func(o *Options) {
		o.HandleSignals = handle
	}
}

// WithMigrateModels is the option that sets the models to migrate in their repositories.
func WithMigrateModels(models ...mapping.Model) Option {
	return func(o *Options) {
		o.MigrateModels = append(o.MigrateModels, models...)
	}
}

// WithModels is the option that sets the models for given service.
func WithModels(models ...mapping.Model) Option {
	return func(o *Options) {
		o.Models = append(o.Models, models...)
	}
}

// WithNamingConvention sets the naming convention option.
func WithNamingConvention(convention mapping.NamingConvention) Option {
	return func(o *Options) {
		o.NamingConvention = convention
	}
}

// WithRepositoryModels maps the repository 'r' to provided 'models'.
func WithRepositoryModels(r repository.Repository, models ...mapping.Model) Option {
	return func(o *Options) {
		o.RepositoryModels[r] = append(o.RepositoryModels[r], models...)
	}
}

// WithServer sets the service server option.
func WithServer(s server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

// WithStore sets the service store 's' obtainable at 'name'.
func WithStore(name string, s store.Store) Option {
	return func(o *Options) {
		o.Stores[name] = s
	}
}

// WithSynchronousConnections defines if the service should query repositories synchronously.
func WithSynchronousConnections(sync bool) Option {
	return func(o *Options) {
		o.SynchronousConnections = sync
	}
}

// WithTokener sets provided tokener for the service options.
func WithTokener(tokener auth.Tokener) Option {
	return func(o *Options) {
		o.Tokener = tokener
	}
}

// WithTimeFunc sets the time function for the service.
func WithTimeFunc(timeFunc func() time.Time) Option {
	return func(o *Options) {
		o.TimeFunc = timeFunc
	}
}

// WithVerifier sets the authorization verifier for the service.
func WithVerifier(authorizer auth.Verifier) Option {
	return func(o *Options) {
		o.Verifier = authorizer
	}
}
