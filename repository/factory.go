package repository

import (
	"github.com/neuronlabs/neuron-core/config"
)

// Factory is the interface used for creating the repositories.
type Factory interface {
	// DriverName gets the driver name for given factory.
	DriverName() string
	// New creates new instance of the Repository for given 'model'.
	New(c Controller, config *config.Repository) (Repository, error)
}
