package class

import (
	"github.com/neuronlabs/errors"
)

// MjrInternal is the major internal error classfication.
var MjrInternal errors.Major

func registerInternalClasses() {
	MjrInternal = errors.MustNewMajor()

	registerInternalQuery()
	registerInternalEncoding()
	registerInternalModel()
	registerInternalRepository()
	registerInternalCommon()
	registerInternalInit()
}

/**

Internal Query

*/
var (
	// MnrInternalQuery is the 'MjrInternal' minor classification for the internal
	// query errors.
	MnrInternalQuery errors.Minor

	// InternalQueryNoStoredValue is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for queries with no or invalid internal context values.
	InternalQueryNoStoredValue errors.Class

	// InternalQueryIncluded is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for the internal included fields issues.
	InternalQueryIncluded errors.Class

	// InternalQueryInvalidField is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for queries with invalid fields.
	InternalQueryInvalidField errors.Class

	// InternalQueryNoSuchModel is the 'MjrInternal', 'MnrInternalQuery' error classification
	// when the model is not found for internal query operations.
	InternalQueryNoSuchModel errors.Class

	// InternalQueryNilValue is the 'MjrInternal', 'MnrInternalQuery' error classification
	// when the internal query operations contains nil value.
	InternalQueryNilValue errors.Class

	// InternalQueryInvalidValue is the 'MjrInternal', 'MnrInternalQuery' error classification
	// when the internal query operations contains invalid query value.
	InternalQueryInvalidValue errors.Class

	// InternalQuerySort is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for invalid query sort fields.
	InternalQuerySort errors.Class

	// InternalQueryFilter is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for invalid internal query filter errors.
	InternalQueryFilter errors.Class

	// InternalQuerySelectedField is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with query selected fields.
	InternalQuerySelectedField errors.Class

	// InternalQueryModelMismatch is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with model type mismatch within internal methods.
	InternalQueryModelMismatch errors.Class

	// InternalQueryValidation is the 'MjrInternal', 'MnrInternalQuery' error classification
	// for internal errors related with query validator.
	InternalQueryValidation errors.Class

	// InternalQueryCardinality is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the provided query violates the query cardinality.
	InternalQueryCardinality errors.Class // i.e. SELECT * FROM collection WHERE id IN (SELECT too,many,fields FROM other)
)

func registerInternalQuery() {
	MnrInternalQuery = errors.MustNewMinor(MjrInternal)

	mjr, mnr := MjrInternal, MnrInternalQuery
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	InternalQueryNoStoredValue = newClass()
	InternalQueryInvalidField = newClass()
	InternalQueryNoSuchModel = newClass()
	InternalQueryNilValue = newClass()
	InternalQueryInvalidValue = newClass()
	InternalQueryIncluded = newClass()
	InternalQuerySort = newClass()
	InternalQueryFilter = newClass()
	InternalQuerySelectedField = newClass()
	InternalQueryModelMismatch = newClass()
	InternalQueryValidation = newClass()
	InternalQueryCardinality = newClass()
}

/**

Internal Encoding

*/

var (
	// MnrInternalEncoding is the 'MjrInternal' minor error classification
	// for internal encoding errors.
	MnrInternalEncoding errors.Minor

	// InternalEncodingModelFieldType is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors when the model field type is invalid.
	InternalEncodingModelFieldType errors.Class

	// InternalEncodingValue is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with invalid value type.
	InternalEncodingValue errors.Class

	// InternalEncodingPayload is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with some payload errors.
	InternalEncodingPayload errors.Class

	// InternalEncodingIncludeScope is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with included scopes.
	InternalEncodingIncludeScope errors.Class

	// InternalEncodingUnsupportedID is the 'MjrInternal' 'MnrInternalEncoding' error classification
	// for internal encoding errors with invalid primary field type or value.
	InternalEncodingUnsupportedID errors.Class
)

func registerInternalEncoding() {
	MnrInternalEncoding = errors.MustNewMinor(MjrInternal)

	mjr, mnr := MjrInternal, MnrInternalEncoding
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	InternalEncodingModelFieldType = newClass()
	InternalEncodingValue = newClass()
	InternalEncodingPayload = newClass()
	InternalEncodingIncludeScope = newClass()
	InternalEncodingUnsupportedID = newClass()
}

/**

Internal Model

*/
var (
	// MnrInternalModel is the 'MjrInternal' minor error classifcation
	// for internal model issues.
	MnrInternalModel errors.Minor

	// InternalModelRelationNotMapped is the 'MjrInternal', 'MnrInternalModel' error classification
	// for internal not mapped model relationships issues.
	InternalModelRelationNotMapped errors.Class

	// InternalModelNotCast is the 'MjrInternal', 'MnrInternalModel' error classification
	// for the internal model's that should cast to some interface.
	InternalModelNotCast errors.Class

	// InternalModelFlag is the 'MjrInternal', 'MnrInternalModel' error classification
	// for the internal model flag errors.
	InternalModelFlag errors.Class
)

func registerInternalModel() {
	MnrInternalModel = errors.MustNewMinor(MjrInternal)

	mjr, mnr := MjrInternal, MnrInternalModel
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	InternalModelRelationNotMapped = newClass()
	InternalModelNotCast = newClass()
	InternalModelFlag = newClass()
}

/**

Internal Repository

*/
var (
	// MnrInternalRepository is the 'MjrInternal' minor error classification
	// for the repository related internal issues.
	MnrInternalRepository errors.Minor

	// InternalRepository is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for internal repositories errors.
	InternalRepository errors.Class

	// InternalRepositoryResourceName is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for the errors related with the internal repository invalid resource name.
	InternalRepositoryResourceName errors.Class

	// InternalRepositorySyntax is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for the errors related with the internal repository syntax.
	InternalRepositorySyntax errors.Class

	// InternalRepositorySystemError is the 'MjrInternal', 'MnrInternalRepository' error classification
	// while repository system error occurred: i.e. not enough disk space.
	InternalRepositorySystemError errors.Class

	// InternalRepositoryClientMismatch is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for internal errors on matching the repository clients.
	InternalRepositoryClientMismatch errors.Class

	// InternalRepositoryIndex is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for internal errors with index.
	InternalRepositoryIndex errors.Class

	// InternalRepositoryOptions is the 'MjrInternal', 'MnrInternalRepository' error classification
	// for internal repository options errors.
	InternalRepositoryOptions errors.Class
)

func registerInternalRepository() {
	MnrInternalRepository = errors.MustNewMinor(MjrInternal)

	mjr, mnr := MjrInternal, MnrInternalRepository
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	InternalRepository = errors.MustNewMinorClass(mjr, mnr)

	InternalRepositoryResourceName = newClass()
	InternalRepositorySyntax = newClass()
	InternalRepositorySystemError = newClass()
	InternalRepositoryClientMismatch = newClass()
	InternalRepositoryIndex = newClass()
	InternalRepositoryOptions = newClass()
}

var (
	// MnrInternalCommon is the common 'MjrInternal' minor error classification
	// for the non classified internal errors.
	MnrInternalCommon errors.Minor

	// InternalCommon is the 'MjrInternal', 'MinorInternalCommon' error classification
	// for non classified internal errors.
	InternalCommon errors.Class
)

func registerInternalCommon() {
	MnrInternalCommon = errors.MustNewMinor(MjrInternal)

	InternalCommon = errors.MustNewMinorClass(MjrInternal, MnrInternalCommon)
}

var (
	// MnrInternalInit is the common 'MnrInternalInit' error classification for
	// initialization errors.
	MnrInternalInit errors.Minor

	// InternalInitController error class for controller initialization errors.
	InternalInitController errors.Class
)

func registerInternalInit() {
	MnrInternalInit = errors.MustNewMinor(MjrInternal)

	InternalInitController = errors.MustNewClass(MjrInternal, MnrInternalInit, errors.MustNewIndex(MjrInternal, MnrInternalInit))
}
