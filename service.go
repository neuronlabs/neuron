package neuron

import (
	"github.com/neuronlabs/neuron/service"
)

// New creates new neuron service with provided options.
func New(options ...service.Option) *service.Service {
	return service.New(options...)
}
