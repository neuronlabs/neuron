package mapping

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrMapping is the major error classification for the mapping package.
	ErrMapping = errors.New("mapping")

	// ErrFieldValue is an invalid field value class.
	ErrFieldValue = errors.Wrap(ErrMapping, "field value")

	// ErrRelation is the minor error classification related to mapping repositories.
	ErrRelation = errors.Wrap(ErrMapping, "relation")
	// ErrInvalidRelationValue is the error for providing invalid relation value.
	ErrInvalidRelationValue = errors.Wrap(ErrRelation, "invalid value")
	// ErrInvalidRelationField is the error for providing invalid relation field.
	ErrInvalidRelationField = errors.Wrap(ErrRelation, "invalid field")
	// ErrInvalidRelationIndex is the error for providing invalid relation index.
	ErrInvalidRelationIndex = errors.Wrap(ErrRelation, "invalid index")

	// ErrModel is the minor error classification related to the models.
	ErrModel = errors.Wrap(ErrMapping, "model")
	// ErrNilModel is an error when the input model is nil.
	ErrNilModel = errors.Wrap(ErrModel, "nil")
	// ErrModelNotMatch is an error where the model doesn't match with it's relationship.
	ErrModelNotMatch = errors.Wrap(ErrMapping, "not match")
	// ErrModelContainer is the error classification with models mapping container.
	ErrModelContainer = errors.Wrap(ErrModel, "container")
	// ErrModelDefinition is the error classification for model without fields defined.
	ErrModelDefinition = errors.Wrap(ErrModel, "definition")
	// ErrModelNotFound is the error classification for models that are not mapped.
	ErrModelNotFound = errors.Wrap(ErrModel, "not found")
	// ErrModelNotImplements is the error classification when model doesn't implement some interface.
	ErrModelNotImplements = errors.Wrap(ErrModel, "not implements")
	// ErrInvalidModelField is the error classification for invalid model field.
	ErrInvalidModelField = errors.Wrap(ErrModel, "invalid field")
	// ErrFieldNotParser is the error classification when the field is not a string parser.
	ErrFieldNotParser = errors.Wrap(ErrModel, "field not parser")
	// ErrFieldNotNullable is the error classification when the field is not nullable.
	ErrFieldNotNullable = errors.Wrap(ErrModel, "field not nullable")

	// ErrNamingConvention is an error classification with errors related with naming convention.
	ErrNamingConvention = errors.Wrap(ErrMapping, "naming convention")

	// ErrInternal is the error class for internal mapping errors.
	ErrInternal = errors.Wrap(errors.ErrInternal, "mapping")
)
