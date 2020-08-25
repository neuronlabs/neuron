package neuron

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/filestore"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/service"
	"github.com/neuronlabs/neuron/store"
)

// Authenticator sets the authenticator option.
func Authenticator(authenticator auth.Authenticator) service.Option {
	return func(o *service.Options) {
		o.Authenticator = authenticator
	}
}

// AccountModel sets the service account model.
func AccountModel(account auth.Account) service.Option {
	return func(o *service.Options) {
		o.AccountModel = account
	}
}

// Collections adds the collections to the initialization process.
func Collections(collections ...core.Initializer) service.Option {
	return func(o *service.Options) {
		o.Initializers = append(o.Initializers, collections...)
	}
}

// DefaultRepository sets the default repository for all models without specified repository.
func DefaultRepository(r repository.Repository) service.Option {
	return func(o *service.Options) {
		o.DefaultRepository = r
	}
}

// DefaultStore sets the default store for the service.
func DefaultStore(s store.Store) service.Option {
	return func(o *service.Options) {
		o.DefaultStore = s
	}
}

// DefaultFileStore sets the default file store for the service.
func DefaultFileStore(s filestore.Store) service.Option {
	return func(o *service.Options) {
		o.DefaultFileStore = s
	}
}

// DefaultNotNullFields sets the non-pointer model fields to not null by default.
// By default all fields are set as nullable in the model's definition.
// This might be helpful in the model migration.
func DefaultNotNullFields(o *service.Options) {
	o.DefaultNotNull = true
}

// FileStore sets the service store 's' obtainable at 'name'.
func FileStore(name string, s filestore.Store) service.Option {
	return func(o *service.Options) {
		o.FileStores[name] = s
	}
}

// HandleSignal is the option that determines if the os signals should be handled by the service.
func HandleSignal(handle bool) service.Option {
	return func(o *service.Options) {
		o.HandleSignals = handle
	}
}

// Initializers adds the abstractions that needs to get initialized with the controller.
func Initializers(initializers ...core.Initializer) service.Option {
	return func(o *service.Options) {
		o.Initializers = append(o.Initializers, initializers...)
	}
}

// MigrateModels is the option that sets the models to migrate in their repositories.
func MigrateModels(models ...mapping.Model) service.Option {
	return func(o *service.Options) {
		o.MigrateModels = append(o.MigrateModels, models...)
	}
}

// Models is the option that sets the models for given service.
func Models(models ...mapping.Model) service.Option {
	return func(o *service.Options) {
		o.Models = append(o.Models, models...)
	}
}

// NamingConvention sets the naming convention option.
func NamingConvention(convention mapping.NamingConvention) service.Option {
	return func(o *service.Options) {
		o.NamingConvention = convention
	}
}

// RepositoryModels maps the repository 'r' to provided 'models'.
func RepositoryModels(r repository.Repository, models ...mapping.Model) service.Option {
	return func(o *service.Options) {
		o.RepositoryModels[r] = append(o.RepositoryModels[r], models...)
	}
}

// Server sets the service server option.
func Server(s server.Server) service.Option {
	return func(o *service.Options) {
		o.Server = s
	}
}

// Store sets the service store 's' obtainable at 'name'.
func Store(name string, s store.Store) service.Option {
	return func(o *service.Options) {
		o.Stores[name] = s
	}
}

// SynchronousConnections defines if the service should query repositories synchronously.
func SynchronousConnections(sync bool) service.Option {
	return func(o *service.Options) {
		o.SynchronousORM = sync
	}
}

// Tokener sets provided tokener for the service options.
func Tokener(tokener auth.Tokener) service.Option {
	return func(o *service.Options) {
		o.Tokener = tokener
	}
}

// UTCTimestamps would set the timestamps of the service to UTC time zoned.
func UTCTimestamps(utcTimestamps bool) service.Option {
	return func(o *service.Options) {
		o.UTCTimestamps = utcTimestamps
	}
}

// Verifier sets the authorization verifier for the service.
func Verifier(authorizer auth.Verifier) service.Option {
	return func(o *service.Options) {
		o.Verifier = authorizer
	}
}
