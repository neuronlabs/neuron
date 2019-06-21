package class

// MjrInternal is the major internal error classfication.
var MjrInternal Major

func registerInternalClasses() {
	MjrInternal = MustRegisterMajor("Internal", "internal error classifcation")

	registerInternalQuery()
	registerInternalEncoding()
	registerInternalModel()
	registerInternalRepository()
	registerInternalCommon()
}

/**

Internal Query

*/
var (
	// MnrInternalQuery is the 'MjrInternal' minor classification for the internal
	// query errors.
	MnrInternalQuery Minor

	// InternalQueryNoStoredValue is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for queries with no or invalid internal context values.
	InternalQueryNoStoredValue Class

	// InternalQueryIncluded is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for the internal included fields issues.
	InternalQueryIncluded Class

	// InternalQueryInvalidField is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for queries with invalid fields.
	InternalQueryInvalidField Class

	// InternalQueryNoSuchModel is the 'MjrInternal', 'MnrInternalQuery' error classification
	// when the model is not found for internal query operations.
	InternalQueryNoSuchModel Class

	// InternalQueryNilValue is the 'MjrInternal', 'MnrInternalQuery' error classification
	// when the internal query operations contains nil value.
	InternalQueryNilValue Class

	// InternalQuerySort is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for invalid query sort fields.
	InternalQuerySort Class

	// InternalQueryFilter is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for invalid internal query filter errors.
	InternalQueryFilter Class

	// InternalQuerySelectedField is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with query selected fields.
	InternalQuerySelectedField Class

	// InternalQueryModelMismatch is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with model type mismatch within internal methods.
	InternalQueryModelMismatch Class

	// InternalQueryValidation is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with query validator.
	InternalQueryValidation Class

	// InternalQueryCardinality is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the provided query violates the query cardinality.
	InternalQueryCardinality Class // i.e. SELECT * FROM collection WHERE id IN (SELECT too,many,fields FROM other)
)

func registerInternalQuery() {
	MnrInternalQuery = MjrInternal.MustRegisterMinor("Query", "internal errors related with the queries")

	InternalQueryNoStoredValue = MnrInternalQuery.MustRegisterIndex("No Stored Value", "no or invalid internal context value for the given query").Class()
	InternalQueryInvalidField = MnrInternalQuery.MustRegisterIndex("Invalid Field", "invalid field provided for internal query operations").Class()
	InternalQueryNoSuchModel = MnrInternalQuery.MustRegisterIndex("No Such Model", "model not found for internal query operations").Class()
	InternalQueryNilValue = MnrInternalQuery.MustRegisterIndex("Nil Value", "nil value occurred on internal query operation").Class()
	InternalQueryIncluded = MnrInternalQuery.MustRegisterIndex("Included", "included field internal issues").Class()
	InternalQuerySort = MnrInternalQuery.MustRegisterIndex("Sort", "sorting fields internal issues").Class()
	InternalQueryFilter = MnrInternalQuery.MustRegisterIndex("Filter", "filter field internal issue").Class()
	InternalQuerySelectedField = MnrInternalQuery.MustRegisterIndex("Selected Field", "selected field internal issue").Class()
	InternalQueryModelMismatch = MnrInternalQuery.MustRegisterIndex("Model Mismatch", "model within the internal method mismatches the model struct").Class()
	InternalQueryValidation = MnrInternalQuery.MustRegisterIndex("Validation", "internal validation error - related with the validator package").Class()
	InternalQueryCardinality = MnrInternalQuery.MustRegisterIndex("Cardinality", "query cardinality had failed").Class()
}

/**

Internal Encoding

*/

var (
	// MnrInternalEncoding is the 'MjrInternal' minor error classification
	// for internal encoding errors.
	MnrInternalEncoding Minor

	// InternalEncodingModelFieldType is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors when the model field type is invalid.
	InternalEncodingModelFieldType Class

	// InternalEncodingValue is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with invalid value type.
	InternalEncodingValue Class

	// InternalEncodingPayload is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with some payload errors.
	InternalEncodingPayload Class

	// InternalEncodingIncludeScope is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with included scopes.
	InternalEncodingIncludeScope Class

	// InternalEncodingUnsupportedID is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with invalid primary field type or value.
	InternalEncodingUnsupportedID Class
)

func registerInternalEncoding() {
	MnrInternalEncoding = MjrInternal.MustRegisterMinor("Encoding", "internal encoding errors")

	InternalEncodingModelFieldType = MnrInternalEncoding.MustRegisterIndex("Model Field Type", "invalid model field type").Class()
	InternalEncodingValue = MnrInternalEncoding.MustRegisterIndex("Value", "invalid encoding value ").Class()
	InternalEncodingPayload = MnrInternalEncoding.MustRegisterIndex("Payload", "encoding the payload").Class()
	InternalEncodingIncludeScope = MnrInternalEncoding.MustRegisterIndex("Include Scope", "encoding the included scope").Class()
	InternalEncodingUnsupportedID = MnrInternalEncoding.MustRegisterIndex("Unsupported ID", "unsupported primary field type or value").Class()
}

/**

Internal Model

*/
var (
	// MnrInternalModel is the 'MjrInternal' minor error classifcation
	// for internal model issues.
	MnrInternalModel Minor

	// InternalModelRelationNotMapped is the 'MjrInternal', 'MnrInternalModel' error classification
	// for internal not mapped model relationships issues.
	InternalModelRelationNotMapped Class

	// InternalModelNotCast is the 'MjrInternal', 'MnrInternalModel' error classification
	// for the internal model's that should cast to some interface.
	InternalModelNotCast Class
)

func registerInternalModel() {
	MnrInternalModel = MjrInternal.MustRegisterMinor("Model", "model related internal errors")

	InternalModelRelationNotMapped = MnrInternalModel.MustRegisterIndex("Not Mapped Relation", "not mapped relationship internal issue").Class()
	InternalModelNotCast = MnrInternalModel.MustRegisterIndex("Not Cast", "model should cast into some interface").Class()
}

/**

Internal Repository

*/
var (
	// MnrInternalRepository is the 'MjrInternal' minor error classification
	// for the repository related internal issues.
	MnrInternalRepository Minor

	// InternalRepositorySyntax is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for the errors related with the internal repository syntax.
	InternalRepositorySyntax Class

	// InternalRepositoryResourceName	 is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for the errors related with the internal repository invalid resource name.
	InternalRepositoryResourceName Class
)

func registerInternalRepository() {
	MnrInternalRepository = MjrInternal.MustRegisterMinor("Repository", "repository related internal issues")

	InternalRepositorySyntax = MnrInternalRepository.MustRegisterIndex("Syntax", "invalid syntax for one of the repository queries").Class()
	InternalRepositoryResourceName = MnrInternalRepository.MustRegisterIndex("ResourceName", "invalid resource name for one of the repositories are queried").Class()
}

var (
	// MnrInternalCommon is the common 'MjrInternal' minor error classification
	// for the non classified internal errors.
	MnrInternalCommon Minor

	// InternalCommon is the 'MjrInternal', 'MinorInternalCommon' error classification
	// for non classified internal errors.
	InternalCommon Class
)

func registerInternalCommon() {
	MnrInternalCommon = MjrInternal.MustRegisterMinor("Common", "non classified internal errors")

	InternalCommon = MustNewMinorClass(MnrInternalCommon)
}
