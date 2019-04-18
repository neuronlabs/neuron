package scope

import (
	"fmt"
	"github.com/neuronlabs/neuron/errors"
	"reflect"
)

func errNoRelationship(jsonapiType, included string) *errors.ApiError {
	err := errors.ErrInvalidResourceName.Copy()
	err.Detail = fmt.Sprintf("Object: '%v', has no relationship named: '%v'.",
		jsonapiType, included)
	return err
}

func errNoModelMappedForRel(model, relatedTo reflect.Type, fieldName string) error {
	err := fmt.Errorf("Model '%v', not mapped! Relationship to '%s' is set for '%s' field.",
		model, relatedTo, fieldName,
	)
	return err
}

func errNoRelationshipInModel(sFieldType, modelType reflect.Type, relationship string) error {
	err := fmt.Errorf("Errorfield of type: '%s' has no relationship within model: '%s', in relationship named: %v", sFieldType, modelType, relationship)
	return err
}
