package neuron

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/server"
)

// SynchronousConnections defines if the service should query repositories synchronously.
func SynchronousConnections(sync bool) core.Option {
	return func(o *core.Options) {
		o.SynchronousORM = sync
	}
}

// Collections adds the collections options.
func Collections(collections ...db.Collection) core.Option {
	return func(o *core.Options) {
		o.Collections = append(o.Collections, collections...)
	}
}

// Context sets the context for the service to
func Context(ctx context.Context) core.Option {
	return func(o *core.Options) {
		o.Context = ctx
	}
}

// DefaultRepositoryName sets default repository name for the service. All the models without repository name defined
// would be assigned to this repository.
func DefaultRepositoryName(name string) core.Option {
	return func(o *core.Options) {
		o.DefaultRepositoryName = name
	}
}

// DisallowDefaultRepository marks if the service should disallow default repository.
func DisallowDefaultRepository(disallow bool) core.Option {
	return func(o *core.Options) {
		o.DisallowDefaultRepository = disallow
	}
}

// HandleSignal is the option that determines if the os signals should be handled by the service.
func HandleSignal(handle bool) core.Option {
	return func(o *core.Options) {
		o.HandleSignals = handle
	}
}

// MigrateModels is the option that sets the models to migrate in their repositories.
func MigrateModels(models ...mapping.Model) core.Option {
	return func(o *core.Options) {
		o.MigrateModels = append(o.MigrateModels, models...)
	}
}

// Models is the option that sets the models for given service.
func Models(models ...mapping.Model) core.Option {
	return func(o *core.Options) {
		o.Models = append(o.Models, models...)
	}
}

// Name is the name option for the service.
func Name(name string) core.Option {
	return func(o *core.Options) {
		o.Name = name
	}
}

// NamingConvention sets the naming convention option.
func NamingConvention(naming string) core.Option {
	return func(o *core.Options) {
		if err := o.NamingConvention.Parse(naming); err != nil {
			panic(err)
		}
	}
}

// Repositories is the option that sets given repositories.
func Repositories(repositories ...core.Repository) core.Option {
	return func(o *core.Options) {
		o.Repositories = append(o.Repositories, repositories...)
	}
}

// UTCTimestamps would set the timestamps of the service to UTC time zoned.
func UTCTimestamps(utcTimestamps bool) core.Option {
	return func(o *core.Options) {
		o.UTCTimestamps = utcTimestamps
	}
}

func NonRepositoryModels(models ...mapping.Model) core.Option {
	return func(o *core.Options) {
		o.NonRepositoryModels = append(o.NonRepositoryModels, models...)
	}
}

func Authorizer(authorizer auth.Authorizer) core.Option {
	return func(o *core.Options) {
		o.Authorizer = authorizer
	}
}

// Authenticator sets the authenticator option.
func Authenticator(authenticator auth.Authenticator) core.Option {
	return func(o *core.Options) {
		o.Authenticator = authenticator
	}
}

// Server sets the service server option.
func Server(s server.Server) core.Option {
	return func(o *core.Options) {
		o.Server = s
	}
}
