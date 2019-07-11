package class

// MjrModel - major that classifies errors related with the models mapping.
var MjrModel Major

func registerModelClasses() {
	MjrModel = MustRegisterMajor("Models")

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
	MnrModelField Minor

	// ModelFieldForeignKeyNotFound is the 'MjrModel', 'MnrModelField' error classification
	// when the model field's foreign key is not found issue.
	ModelFieldForeignKeyNotFound Class

	// ModelFieldName is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's name issue.
	ModelFieldName Class

	// ModelFieldNestedType is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's nested field type issue.
	ModelFieldNestedType Class

	// ModelFieldNotFound is the 'MjrModel', 'MnrModelField' error classification
	// when the model field is not found issue.
	ModelFieldNotFound Class

	// ModelFieldTag is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's tag issue.
	ModelFieldTag Class

	// ModelFieldType is the 'MjrModel', 'MnrModelField' error classification
	// on the model field's type issue.
	ModelFieldType Class
)

func registerModelField() {
	MnrModelField = MjrModel.MustRegisterMinor("Field", "model's fields related issues")

	ModelFieldForeignKeyNotFound = MnrModelField.MustRegisterIndex("FK Not Found", "foreign key is not found within model").Class()
	ModelFieldName = MnrModelField.MustRegisterIndex("Name", "field's name is invalid").Class()
	ModelFieldNestedType = MnrModelField.MustRegisterIndex("Nested Type", "invalid field's nested field type").Class()
	ModelFieldNotFound = MnrModelField.MustRegisterIndex("Not Found", "field is not found within model").Class()
	ModelFieldTag = MnrModelField.MustRegisterIndex("Tag", "invalid field's tag").Class()
	ModelFieldType = MnrModelField.MustRegisterIndex("Type", "invalid field's type").Class()
}

/**

Model Schema

*/
var (
	// MnrModelSchema is the 'MjrModel' minor error classification
	// on the model's schema issues.
	MnrModelSchema Minor

	// ModelInSchemaAlreadyRegistered is the 'MjrModel', 'MnrModelSchema' error classifcation
	// when the model is already registered within given schema.
	ModelInSchemaAlreadyRegistered Class

	// ModelSchemaNotFound is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model schema is not found.
	ModelSchemaNotFound Class

	// ModelNotMapped is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model is not mapped within schema.
	ModelNotMappedInSchema Class
)

func registerModelSchema() {
	MnrModelSchema = MjrModel.MustRegisterMinor("Schema", "model's schema related issues")

	ModelInSchemaAlreadyRegistered = MnrModelSchema.MustRegisterIndex("Already Registered", "model is already registered within given schema").Class()
	ModelSchemaNotFound = MnrModelSchema.MustRegisterIndex("Not Found", "model schema not found").Class()
	ModelNotMappedInSchema = MnrModelSchema.MustRegisterIndex("Not Mapped", "model not mapped within schema").Class()
}

/**

Model Relationship

*/
var (
	// MnrModelRelationship is the 'MjrModel' minor error classification
	// on the issues with model relationships.
	MnrModelRelationship Minor

	// ModelRelationshipType is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's type.
	ModelRelationshipType Class

	// ModelRelationshipForeign is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's invalid foreign key.
	ModelRelationshipForeign Class

	// ModelRelationshipJoinModel is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's join model.
	ModelRelationshipJoinModel Class

	// ModelRelationshipBackreference is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship's back reference field.
	ModelRelationshipBackreference Class

	// ModelRelationshipOptions is the 'MjrModel', 'MnrModelRelationship' error classification
	// on the issues with relationship options like: 'on error', 'on patch', 'on delete'.
	ModelRelationshipOptions Class
)

func registerModelRelationship() {
	MnrModelRelationship = MjrModel.MustRegisterMinor("Relationship", "issues related with model's relationships")

	ModelRelationshipType = MnrModelRelationship.MustRegisterIndex("Type", "invalid relationship's type").Class()
	ModelRelationshipForeign = MnrModelRelationship.MustRegisterIndex("Foreign", "invalid relationship's foreign key").Class()
	ModelRelationshipJoinModel = MnrModelRelationship.MustRegisterIndex("Join Model", "issues related with relationship's join model").Class()
	ModelRelationshipBackreference = MnrModelRelationship.MustRegisterIndex("Backreference", "issues related with backerefernece fields").Class()
	ModelRelationshipOptions = MnrModelRelationship.MustRegisterIndex("Options", "options related issues").Class()
}

var (
	// MnrModelValue is the 'MjrModel' minor error classification used for model value issues.
	MnrModelValue Minor

	// ModelValueNil is the 'MjrModel', 'MnrModelValue' error classification
	// for nil model values - while getting the model struct.
	ModelValueNil Class
)

func registerModelValues() {
	MnrModelValue = MjrModel.MustRegisterMinor("Value", "issues related to model value")

	ModelValueNil = MnrModelValue.MustRegisterIndex("Nil", "getting model struct with nil value").Class()
}

var (
	// MnrModelMapping is the 'MjrModel' minor error classification for model mapping issues.
	MnrModelMapping Minor

	// ModelMappingNoFields is the 'MjrModel', 'MnrModelMapping' error classification
	// for model without required field or no fields at all.
	ModelMappingNoFields Class

	// ModelMappingInvalidType is the 'MjrModel', 'MnrModelMapping' error classification
	// for invalid types (i.e. field).
	ModelMappingInvalidType Class

	// ModelNotMapped is the 'MjrModel', 'MnrModelSchema' error classification
	// used when the model is not mapped within schema.
	ModelNotMapped Class
)

func registerModelMapping() {
	MnrModelMapping = MjrModel.MustRegisterMinor("Mapping", "issues related with model mappings")

	ModelMappingNoFields = MnrModelMapping.MustRegisterIndex("No Fields", "issues for models without required field or no fields at all").Class()
	ModelMappingInvalidType = MnrModelMapping.MustRegisterIndex("Invalid Type", "issues with the models (fields) type while mapping").Class()
	ModelNotMapped = MnrModelMapping.MustRegisterIndex("Not Mapped", "issues with non mapped models").Class()
}
