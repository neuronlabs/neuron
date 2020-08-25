package core

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/mapping"
)

// Options defines the configuration for the Options.
type Options struct {
	AccountModel auth.Account
	// NamingConvention is the naming convention used while mapping the models.
	NamingConvention mapping.NamingConvention
	// SynchronousConnections defines if the query relation includes would be taken concurrently.
	SynchronousConnections bool
	// UTCTimestamps is the flag that defines the format of the timestamps.
	UTCTimestamps bool
	// DefaultNotNullFields defines if the model non-pointer fields should be marked as not null by default.
	DefaultNotNullFields bool
	// ModelNotNullFields defines not null fields for specified model only.
	ModelNotNullFields map[mapping.Model]struct{}
}

func defaultOptions() *Options {
	return &Options{
		NamingConvention: mapping.SnakeCase,
	}
}
