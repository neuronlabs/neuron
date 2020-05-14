package service

import (
	"context"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/server"
)

// Options is the structure that contains service options.
type Options struct {
	Name                      string
	NamingConvention          namingConvention
	Repositories              []Repository
	Models                    []mapping.Model
	MigrateModels             []mapping.Model
	Server                    server.Server
	HandleSignals             bool
	Context                   context.Context
	DefaultRepositoryName     string
	AsynchronousConnections   bool
	DisallowDefaultRepository bool
	UTCTimestamps             bool
	ExternalController        bool
}

func (o *Options) config() (*config.Controller, error) {
	cfg := &config.Controller{
		Repositories:              make(map[string]*config.Repository),
		NamingConvention:          o.NamingConvention.String(),
		DefaultRepositoryName:     o.DefaultRepositoryName,
		DisallowDefaultRepository: o.DisallowDefaultRepository,
		AsynchronousIncludes:      o.AsynchronousConnections,
		UTCTimestamps:             o.UTCTimestamps,
	}
	for _, repo := range o.Repositories {
		_, ok := cfg.Repositories[repo.Name]
		if ok {
			return nil, errors.Newf(controller.ClassRepositoryAlreadyRegistered, "duplicated repository with name: '%s'", repo.Name)
		}
		cfg.Repositories[repo.Name] = repo.Config
	}
	return cfg, nil
}

func defaultOptions() *Options {
	return &Options{
		HandleSignals:    true,
		Context:          context.Background(),
		NamingConvention: SnakeCase,
	}
}

// Option is the function that sets the options for the service.
type Option func(o *Options)

// Repository is the pair of the name with related config used by the service.
type Repository struct {
	Name   string
	Config *config.Repository
}

// AsynchronousConnections defines if the service should query repositories asynchronously.
func AsynchronousConnections(async bool) Option {
	return func(o *Options) {
		o.AsynchronousConnections = async
	}
}

// Context sets the context for the service to
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// DefaultRepositoryName sets default repository name for the service. All the models without repository name defined
// would be assigned to this repository.
func DefaultRepositoryName(name string) Option {
	return func(o *Options) {
		o.DefaultRepositoryName = name
	}
}

// DisallowDefaultRepository marks if the service should disallow default repository.
func DisallowDefaultRepository(disallow bool) Option {
	return func(o *Options) {
		o.DisallowDefaultRepository = disallow
	}
}

// ExternalController sets the non default external controller.
func ExternalController(external bool) Option {
	return func(o *Options) {
		o.ExternalController = external
	}
}

// HandleSignal is the option that determines if the os signals should be handled by the service.
func HandleSignal(handle bool) Option {
	return func(o *Options) {
		o.HandleSignals = handle
	}
}

// MigrateModels is the option that sets the models to migrate in their repositories.
func MigrateModels(models ...mapping.Model) Option {
	return func(o *Options) {
		o.MigrateModels = models
	}
}

// Models is the option that sets the models for given service.
func Models(models ...mapping.Model) Option {
	return func(o *Options) {
		o.Models = models
	}
}

// Name is the name option for the service.
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// NamingConvention sets the naming convention option.
func NamingConvention(naming namingConvention) Option {
	return func(o *Options) {
		o.NamingConvention = naming
	}
}

// Repositories is the option that sets given repositories.
func Repositories(repositories ...Repository) Option {
	return func(o *Options) {
		o.Repositories = repositories
	}
}

// UTCTimestamps would set the timestamps of the service to UTC time zoned.
func UTCTimestamps(utcTimestamps bool) Option {
	return func(o *Options) {
		o.UTCTimestamps = utcTimestamps
	}
}
