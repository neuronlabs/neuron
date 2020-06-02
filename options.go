package neuron

import (
	"context"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/mapping"
)

// SynchronousConnections defines if the service should query repositories synchronously.
func SynchronousConnections(sync bool) core.Option {
	return func(o *core.Options) {
		o.SynchronousORM = sync
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

// ExternalController sets the non default external controller.
func ExternalController(external bool) core.Option {
	return func(o *core.Options) {
		o.ExternalController = external
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
		o.MigrateModels = models
	}
}

// Models is the option that sets the models for given service.
func Models(models ...mapping.Model) core.Option {
	return func(o *core.Options) {
		o.Models = models
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
		o.NamingConvention.Parse(naming)
	}
}

// Repositories is the option that sets given repositories.
func Repositories(repositories ...core.Repository) core.Option {
	return func(o *core.Options) {
		o.Repositories = repositories
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
		o.NonRepositoryModels = models
	}
}
