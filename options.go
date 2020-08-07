package neuron

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
)

// Authorizer sets the authorizer for the service.
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

// Collections adds the collections to the initialization process.
func Collections(collections ...database.Collection) core.Option {
	return func(o *core.Options) {
		o.Collections = append(o.Collections, collections...)
	}
}

// DefaultRepository sets the default repository for all models without specified repository.
func DefaultRepository(r repository.Repository) core.Option {
	return func(o *core.Options) {
		o.DefaultRepository = r
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
func NamingConvention(convention mapping.NamingConvention) core.Option {
	return func(o *core.Options) {
		o.NamingConvention = convention
	}
}

// RepositoryModel maps the repository 'r' to the 'model'.
func RepositoryModels(r repository.Repository, models ...mapping.Model) core.Option {
	return func(o *core.Options) {
		o.RepositoryModels[r] = append(o.RepositoryModels[r], models...)
	}
}

// Server sets the service server option.
func Server(s server.Server) core.Option {
	return func(o *core.Options) {
		o.Server = s
	}
}

// SynchronousConnections defines if the service should query repositories synchronously.
func SynchronousConnections(sync bool) core.Option {
	return func(o *core.Options) {
		o.SynchronousORM = sync
	}
}

// UTCTimestamps would set the timestamps of the service to UTC time zoned.
func UTCTimestamps(utcTimestamps bool) core.Option {
	return func(o *core.Options) {
		o.UTCTimestamps = utcTimestamps
	}
}
