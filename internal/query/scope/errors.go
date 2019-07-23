package scope

import (
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
)

func errNoRelationship(neuronName, included string) *errors.Error {
	err := errors.New(class.QueryFilterUnknownField, "no relationship field found")
	err = err.SetDetailf("Object: '%v', has no relationship named: '%v'.", neuronName, included)
	return err
}
