package database

import (
	"time"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

type Options struct {
	// DefaultRepository is the default repository for given controller.
	DefaultRepository repository.Repository
	// RepositoryModels is the mapping between repository and it's models.
	RepositoryModels map[repository.Repository][]mapping.Model
	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// TimeFunc is the time function used by timestamps.
	TimeFunc func() time.Time
	// SynchronousConnections gets the synchronous connections.
	SynchronousConnections bool
	// MigrateModels are the models set up for database migration.
	MigrateModels []mapping.Model
}

// Option is an option function for the database settings.
type Option func(o *Options)

// WithDefaultRepository sets the default repository for all models without specified repository.
func WithDefaultRepository(r repository.Repository) Option {
	return func(o *Options) {
		o.DefaultRepository = r
	}
}

// WithRepositoryModels maps the repository 'r' to provided 'models'.
func WithRepositoryModels(r repository.Repository, models ...mapping.Model) Option {
	return func(o *Options) {
		o.RepositoryModels[r] = append(o.RepositoryModels[r], models...)
	}
}

// WithModelMap with model map
func WithModelMap(m *mapping.ModelMap) Option {
	return func(o *Options) {
		o.ModelMap = m
	}
}

// WithTimeFunc sets the function used for the timestamps.
func WithTimeFunc(timeFunc func() time.Time) Option {
	return func(o *Options) {
		o.TimeFunc = timeFunc
	}
}

// WithSynchronousConnections sets the synchronous connections for the db.
func WithSynchronousConnections() Option {
	return func(o *Options) {
		o.SynchronousConnections = true
	}
}

// WithMigrateModels sets tup the models to migrate into database structures.
func WithMigrateModels(models ...mapping.Model) Option {
	return func(o *Options) {
		o.MigrateModels = append(o.MigrateModels, models...)
	}
}
