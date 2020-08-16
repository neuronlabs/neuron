package neuron

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/store"
)

// AuthorizationVerifier sets the authorization verifier for the service.
func AuthorizationVerifier(authorizer auth.Verifier) core.Option {
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

// DefaultStore sets the default store for the service.
func DefaultStore(s store.Store) core.Option {
	return func(o *core.Options) {
		o.DefaultStore = s
	}
}

// DefaultNotNullFields sets the non-pointer model fields to not null by default.
// By default all fields are set as nullable in the model's definition.
// This might be helpful in the model migration.
func DefaultNotNullFields(o *core.Options) {
	o.DefaultNotNull = true
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

// RepositoryModels maps the repository 'r' to provided 'models'.
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

// Store sets the service store 's' obtainable at 'name'.
func Store(name string, s store.Store) core.Option {
	return func(o *core.Options) {
		o.Stores[name] = s
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
