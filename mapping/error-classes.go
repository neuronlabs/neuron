package mapping

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrMapping is the major error classification for the mapping package.
	MjrMapping errors.Major

	// MnrValue is the minor error classification related to mapping values.
	MnrValue errors.Minor
	// ClassInvalidFieldValue is an invalid field value class.
	ClassInvalidFieldValue errors.Class
	// ClassInvalidRelationValue is the error for providing invalid relation value,
	ClassInvalidRelationValue errors.Class

	// MnrModel is the minor error classification related to the models.
	MnrModel errors.Minor

	// ClassModelContainer is the error classification with models mapping container.
	ClassModelContainer errors.Class
	// ClassModelDefinition is the error classification for model without fields defined.
	ClassModelDefinition errors.Class
	// ClassModelNotFound is the error classification for models that are not mapped.
	ClassModelNotFound errors.Class
	// ClassModelNotImplements is the error classification when model doesn't implement some interface.
	ClassModelNotImplements errors.Class

	// ClassInternal is the error class for internal mapping errors.
	ClassInternal errors.Class
)

func init() {
	MjrMapping = errors.MustNewMajor()

	// Value
	MnrValue = errors.MustNewMinor(MjrMapping)
	ClassInvalidFieldValue = errors.MustNewClassWIndex(MjrMapping, MnrValue)
	ClassInvalidRelationValue = errors.MustNewClassWIndex(MjrMapping, MnrValue)

	// Model
	MnrModel = errors.MustNewMinor(MjrMapping)
	ClassModelContainer = errors.MustNewClassWIndex(MjrMapping, MnrModel)
	ClassModelDefinition = errors.MustNewClassWIndex(MjrMapping, MnrModel)
	ClassModelNotFound = errors.MustNewClassWIndex(MjrMapping, MnrModel)

	// Internal
	ClassInternal = errors.MustNewMajorClass(MjrMapping)
}
