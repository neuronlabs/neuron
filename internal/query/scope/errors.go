package scope

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
)

func errNoRelationship(neuronName, included string) *errors.Error {
	err := errors.New(class.QueryFilterUnknownField, "no relationship field found")
	err.SetDetailf("Object: '%v', has no relationship named: '%v'.", neuronName, included)
	return err
}
