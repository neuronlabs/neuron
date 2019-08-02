package class

import (
	"github.com/neuronlabs/errors"
)

// MjrModel - major that classifies errors related with the models mapping.
var MjrModel errors.Major

func registerModelClasses() {
	MjrModel = errors.MustNewMajor()

	registerModelField()
	registerModelSchema()
	registerModelRelationship()
	registerModelValues()
	registerModelMapping()
}

/**

Model Field

*/
var (
	// MnrModelField is the 'MjrModel' minor error classification
	// on the model field's definitions.
	MnrModelField errors.Minor

	// ModelFieldForeignKeyNotFound is the 'MjrModel', 'MnrModelField' error classification
	// when the model field's foreign key is not found issue.
	ModelFieldForeignKeyNotFound errors.Class

	// ModelFieldName is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's name issue.
	ModelFieldName errors.Class

	// ModelFieldNestedType is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's nested field type issue.
	ModelFieldNestedType errors.Class

	// ModelFieldNotFound is the 'MjrModel', 'MnrModelField' error classification
	// when the model field is not found issue.
	ModelFieldNotFound errors.Class

	// ModelFieldTag is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's tag issue.
	ModelFieldTag errors.Class

	// ModelFieldType is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's type issue.
	ModelFieldType errors.Class
)

func registerModelField() {
	MnrModelField = errors.MustNewMinor(MjrModel)

	mjr, mnr := MjrModel, MnrModelField
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	ModelFieldForeignKeyNotFound = newClass()
	ModelFieldName = newClass()
	ModelFieldNestedType = newClass()
	ModelFieldNotFound = newClass()
	ModelFieldTag = newClass()
	ModelFieldType = newClass()
}

/**

Model Schema

*/
var (
	// MnrModelSchema is the 'MjrModel' minor error classification
	// on the model's schema issues.
	MnrModelSchema errors.Minor

	// ModelInSchemaAlreadyRegistered is the 'MjrModel', 'MnrModelSchema' error classifcation
	// when the model is already registered within given schema.
	ModelInSchemaAlreadyRegistered errors.Class

	// ModelSchemaNotFound is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model schema is not found.
	ModelSchemaNotFound errors.Class

	// ModelNotMapped is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model is not mapped within schema.
	ModelNotMappedInSchema errors.Class
)

func registerModelSchema() {
	MnrModelSchema = errors.MustNewMinor(MjrModel)

	mjr, mnr := MjrModel, MnrModelSchema
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	ModelInSchemaAlreadyRegistered = newClass()
	ModelSchemaNotFound = newClass()
	ModelNotMappedInSchema = newClass()
}

/**

Model Relationship

*/
var (
	// MnrModelRelationship is the 'MjrModel' minor error classification
	// on the issues with model relationships.
	MnrModelRelationship errors.Minor

	// ModelRelationshipType is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's type.
	ModelRelationshipType errors.Class

	// ModelRelationshipForeign is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's invalid foreign key.
	ModelRelationshipForeign errors.Class

	// ModelRelationshipJoinModel is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's join model.
	ModelRelationshipJoinModel errors.Class

	// ModelRelationshipBackreference is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's back reference field.
	ModelRelationshipBackreference errors.Class

	// ModelRelationshipOptions is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship options like: 'on error', 'on patch', 'on delete'.
	ModelRelationshipOptions errors.Class
)

func registerModelRelationship() {
	MnrModelRelationship = errors.MustNewMinor(MjrModel)

	mjr, mnr := MjrModel, MnrModelRelationship
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	ModelRelationshipType = newClass()
	ModelRelationshipForeign = newClass()
	ModelRelationshipJoinModel = newClass()
	ModelRelationshipBackreference = newClass()
	ModelRelationshipOptions = newClass()
}

var (
	// MnrModelValue is the 'MjrModel' minor error classification used for model value issues.
	MnrModelValue errors.Minor

	// ModelValueNil is the 'MjrModel', 'MnrModelValue' error classification
	// for nil model values - while getting the model struct.
	ModelValueNil errors.Class
)

func registerModelValues() {
	MnrModelValue = errors.MustNewMinor(MjrModel)

	ModelValueNil = errors.MustNewClass(MjrModel, MnrModelValue, errors.MustNewIndex(MjrModel, MnrModelValue))
}

var (
	// MnrModelMapping is the 'MjrModel' minor error classification for model mapping issues.
	MnrModelMapping errors.Minor

	// ModelMappingNoFields is the 'MjrModel', 'MnrModelMapping' error classification
	// for model without required field or no fields at all.
	ModelMappingNoFields errors.Class

	// ModelMappingInvalidType is the 'MjrModel', 'MnrModelMapping' error classification
	// for invalid types (i.e. field).
	ModelMappingInvalidType errors.Class

	// ModelNotMapped is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model is not mapped within schema.
	ModelNotMapped errors.Class
)

func registerModelMapping() {
	MnrModelMapping = errors.MustNewMinor(MjrModel)

	mjr, mnr := MjrModel, MnrModelMapping
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	ModelMappingNoFields = newClass()
	ModelMappingInvalidType = newClass()
	ModelNotMapped = newClass()
}
