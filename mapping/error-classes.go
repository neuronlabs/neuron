package mapping

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrMapping is the major error classification for the mapping package.
	MjrMapping errors.Major

	// ClassInvalidFieldValue is an invalid field value class.
	ClassInvalidFieldValue errors.Class

	// MnrRepository is the minor error classification related to mapping repositories.
	MnrRepository errors.Minor
	// ClassInvalidRelationValue is the error for providing invalid relation value.
	ClassInvalidRelationValue errors.Class
	// ClassInvalidRelationField is the error for providing invalid relation field.
	ClassInvalidRelationField errors.Class
	// ClassInvalidRelationIndex is the error for providing invalid relation index.
	ClassInvalidRelationIndex errors.Class

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
	// ClassInvalidModelField is the error classification for invalid model field.
	ClassInvalidModelField errors.Class

	// ClassInternal is the error class for internal mapping errors.
	ClassInternal errors.Class

	ClassNamingConvention errors.Class
)

func init() {
	MjrMapping = errors.MustNewMajor()

	// Value
	ClassInvalidFieldValue = errors.MustNewMajorClass(MjrMapping)

	// Repository
	MnrRepository = errors.MustNewMinor(MjrMapping)
	ClassInvalidRelationValue = errors.MustNewClassWIndex(MjrMapping, MnrRepository)
	ClassInvalidRelationIndex = errors.MustNewClassWIndex(MjrMapping, MnrRepository)
	ClassInvalidRelationField = errors.MustNewClassWIndex(MjrMapping, MnrRepository)

	// Model
	MnrModel = errors.MustNewMinor(MjrMapping)
	ClassModelContainer = errors.MustNewClassWIndex(MjrMapping, MnrModel)
	ClassModelDefinition = errors.MustNewClassWIndex(MjrMapping, MnrModel)
	ClassModelNotFound = errors.MustNewClassWIndex(MjrMapping, MnrModel)
	ClassInvalidModelField = errors.MustNewClassWIndex(MjrMapping, MnrModel)

	// Internal
	ClassInternal = errors.MustNewMajorClass(MjrMapping)

	ClassNamingConvention = errors.MustNewMajorClass(MjrMapping)
}
