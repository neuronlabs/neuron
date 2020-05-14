package repository

import (
	"github.com/neuronlabs/neuron/config"
)

// Factory is the interface used for creating the repositories.
type Factory interface {
	// DriverName gets the driver name for given factory.
	DriverName() string
	// New creates new Input of the Repository for given 'model'.
	New(config *config.Repository) (Repository, error)
}
