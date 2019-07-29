package scope

import (
	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
)

func errNoRelationship(neuronName, included string) errors.DetailedError {
	err := errors.NewDet(class.QueryFilterUnknownField, "no relationship field found")
	err.SetDetailsf("Object: '%v', has no relationship named: '%v'.", neuronName, included)
	return err
}
