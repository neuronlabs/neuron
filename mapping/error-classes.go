package mapping

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	MjrClMapping errors.Major

	// MnrClValue is the minor error classification related to mapping values.
	MnrClValue errors.Minor
	// ClassInvalidFieldValue is an invalid field value class.
	ClassInvalidFieldValue errors.Class
	// ClassInvalidRelationValue is the error for providing invalid relation value,
	ClassInvalidRelationValue errors.Class

	// MnrClModel is the minor error classification related to the models.
	MnrClModel errors.Minor

	// ClassModelContainer is the error classification with models mapping container.
	ClassModelContainer errors.Class
	// ClassModelDefinition is the error classification for model without fields defined.
	ClassModelDefinition errors.Class
	// ClassModelNotFound is the error classification for models that are not mapped.
	ClassModelNotFound errors.Class
	// ClassModelNotImplements is the error classification when model doesn't implement some interface.
	ClassModelNotImplements errors.Class

	// ClassConfig is the error classification for the config errors.
	ClassConfig errors.Class

	// ClassInternal is the error class for internal mapping errors.
	ClassInternal errors.Class
)

func init() {
	MjrClMapping = errors.MustNewMajor()

	// Value
	MnrClValue = errors.MustNewMinor(MjrClMapping)
	ClassInvalidFieldValue = errors.MustNewClassWIndex(MjrClMapping, MnrClValue)
	ClassInvalidRelationValue = errors.MustNewClassWIndex(MjrClMapping, MnrClValue)

	// Model
	MnrClModel = errors.MustNewMinor(MjrClMapping)
	ClassModelContainer = errors.MustNewClassWIndex(MjrClMapping, MnrClModel)
	ClassModelDefinition = errors.MustNewClassWIndex(MjrClMapping, MnrClModel)
	ClassModelNotFound = errors.MustNewMajorClass(MjrClMapping, MnrClModel)

	// Config
	ClassConfig = errors.MustNewMajorClass(MjrClMapping)
	// Internal
	ClassInternal = errors.MustNewMajorClass(MjrClMapping)
}
