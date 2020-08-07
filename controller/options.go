package controller

import (
	"github.com/neuronlabs/neuron/mapping"
)

// Options defines the configuration for the Options.
type Options struct {
	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lower_camel
	// - snake
	// - kebab
	NamingConvention mapping.NamingConvention
	// SynchronousConnections defines if the query relation includes would be taken concurrently.
	SynchronousConnections bool
	// UTCTimestamps is the flag that defines the format of the timestamps.
	UTCTimestamps bool
}

func defaultOptions() *Options {
	return &Options{
		NamingConvention: mapping.SnakeCase,
	}
}