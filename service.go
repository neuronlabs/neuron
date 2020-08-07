package neuron

import (
	"github.com/neuronlabs/neuron/core"
)

// New creates new neuron service with provided options.
func New(options ...core.Option) *core.Service {
	return core.New(options...)
}
