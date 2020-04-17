package class

import (
	"github.com/neuronlabs/errors"
)

// MjrQuery - major that classifies all the errors related with
// creating, operating on or changing the queries.
var MjrQuery errors.Major

func registerQueryClasses() {
	MjrQuery = errors.MustNewMajor()

	registerQueryFieldset()
	registerQuerySelectedFields()
	registerQueryFilters()
	registerQuerySorts()
	registerQueryPagination()
	registerQueryValue()
	registerQueryInclude()
	registerQueryTransactions()
	registerQueryViolation()
	registerQueryRelations()
	registerQueryProcessor()
}

/**

Query Fieldset

*/

var (
	// MnrQueryFieldset is the 'MjrQuery' minor error classification
	// for query fieldset related issues.
	MnrQueryFieldset errors.Minor

	// QueryFieldsetTooBig is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// used when provided more than possible fields in the fieldset.
	QueryFieldsetTooBig errors.Class

	// QueryFieldsetUnknownField is the 'MjrQuery', 'MnrQueryFieldset' error classificatio
	// used when the provided fieldset field is not found within the model's collection.
	QueryFieldsetUnknownField errors.Class

	// QueryFieldsetDuplicate is the 'MjrQuery', 'MnrQueryFieldset' error classificatio
	// for duplicated fieldset field's name.
	QueryFieldsetDuplicate errors.Class

	// QueryFieldsetInvalid is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// for invalid fieldset issues.
	QueryFieldsetInvalid errors.Class

	// QueryFieldsetEmpty is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// for empty fieldset issues.
	QueryFieldsetEmpty errors.Class
)

func registerQueryFieldset() {
	MnrQueryFieldset = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryFieldset
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryFieldsetDuplicate = newClass()
	QueryFieldsetEmpty = newClass()
	QueryFieldsetInvalid = newClass()
	QueryFieldsetTooBig = newClass()
	QueryFieldsetUnknownField = newClass()
}

/**

Query Selected Fields

*/

var (
	// MnrQuerySelectedFields is the 'MjrQuery' minor error classifcation
	// for the selected fields issues.
	MnrQuerySelectedFields errors.Minor

	// QuerySelectedFieldsNotFound is the 'MjrQuery', 'MnrQuerySelectedFields' error classifcation
	// when the selected fields is not found.
	QuerySelectedFieldsNotFound errors.Class

	// QuerySelectedFieldsNotSelected is the 'MjrQuery', 'MnrQuerySelectedFields' error classifcation
	// when unselecting fields which were not selected.
	QuerySelectedFieldsNotSelected errors.Class

	// QueryFieldsetInvalidModel is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// when provided selected field is of invalid or non matched model.
	QueryFieldsetInvalidModel errors.Class

	// QuerySelectedFieldInvalid is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// when the selected field is not valid.
	QuerySelectedFieldInvalid errors.Class

	// QuerySelectedFieldAlreadyUsed is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// whenthe selected field is already 'selected'.
	QuerySelectedFieldAlreadyUsed errors.Class
)

func registerQuerySelectedFields() {
	MnrQuerySelectedFields = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQuerySelectedFields
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QuerySelectedFieldsNotFound = newClass()
	QuerySelectedFieldsNotSelected = newClass()
	QueryFieldsetInvalidModel = newClass()
	QuerySelectedFieldInvalid = newClass()
	QuerySelectedFieldAlreadyUsed = newClass()
}

/**

Query Filters

*/

var (
	// MnrQueryFilter is the 'MjrQuery' minor that classifies errors related with the query filters.
	MnrQueryFilter errors.Minor

	// QueryFilterInvalidField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query filters field - name.
	QueryFilterInvalidField errors.Class

	// QueryFilterInvalidFormat is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query filter format.
	QueryFilterInvalidFormat errors.Class

	// QueryFilterLanguage is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query language filter.	.
	QueryFilterLanguage errors.Class

	// QueryFilterMissingRequired is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for missing query filters.
	QueryFilterMissingRequired errors.Class

	// QueryFilterUnknownCollection is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown query filter collection provided.
	QueryFilterUnknownCollection errors.Class

	// QueryFilterUnknownField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown query filter field.
	QueryFilterUnknownField errors.Class

	// QueryFilterUnknownOperator is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown - invalid filter operator.
	QueryFilterUnknownOperator errors.Class

	// QueryFilterUnsupportedField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unsupported field types.
	QueryFilterUnsupportedField errors.Class

	// QueryFilterUnsupportedOperator is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unsupported query filter operator.
	QueryFilterUnsupportedOperator errors.Class

	// QueryFilterValue is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid filter values.
	QueryFilterValue errors.Class

	// QueryFitlerNonMatched is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for non matched model structures or field structures to models.
	QueryFitlerNonMatched errors.Class

	// QueryFilterFieldKind is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid, unknown or unsupported filter field kind.
	QueryFilterFieldKind errors.Class
)

func registerQueryFilters() {
	MnrQueryFilter = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryFilter
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryFilterInvalidField = newClass()
	QueryFilterInvalidFormat = newClass()
	QueryFilterLanguage = newClass()
	QueryFilterMissingRequired = newClass()
	QueryFilterUnknownCollection = newClass()
	QueryFilterUnknownField = newClass()
	QueryFilterUnknownOperator = newClass()
	QueryFilterUnsupportedField = newClass()
	QueryFilterUnsupportedOperator = newClass()
	QueryFilterValue = newClass()
	QueryFitlerNonMatched = newClass()
}

/**

Query Sorts

*/

var (
	// MnrQuerySorts is the 'MjrQuery' minor that classifies errors related with the sorts.
	MnrQuerySorts errors.Minor

	// QuerySortField is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for unknown, unsupported fields.
	QuerySortField errors.Class

	// QuerySortFormat is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for invalid sorting format.
	QuerySortFormat errors.Class

	// QuerySortTooManyFields is the 'MjrQuery', 'MnrQuerySorts' error classification
	// when too many sorting fields provided.
	QuerySortTooManyFields errors.Class

	// QuerySortRelatedFields is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for unsupported sorting by related fields.
	QuerySortRelatedFields errors.Class
)

func registerQuerySorts() {
	MnrQuerySorts = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQuerySorts
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QuerySortField = newClass()
	QuerySortFormat = newClass()
	QuerySortTooManyFields = newClass()
	QuerySortRelatedFields = newClass()
}

/**

Query Pagination

*/

var (
	// MnrQueryPagination is the 'MjrQuery' minor that classifies errors related with the pagination.
	MnrQueryPagination errors.Minor

	// QueryPaginationValue is the 'MjrQuery', 'MnrQueryPagination' error classification
	// for invalid pagination values. I.e. invalid limit, offset value.
	QueryPaginationValue errors.Class

	// QueryPaginationType is the 'MjrQuery', 'MnrQueryPagination' error classification
	// for invalid pagination type provided.
	QueryPaginationType errors.Class

	// QueryPaginationAlreadySet is the 'MjrQuery', 'MnrQueryPagination' error classification
	// while trying to add pagination when it is already set.
	QueryPaginationAlreadySet errors.Class
)

func registerQueryPagination() {
	MnrQueryPagination = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryPagination
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryPaginationValue = newClass()
	QueryPaginationType = newClass()
	QueryPaginationAlreadySet = newClass()
}

/**

Query Value

*/

var (
	// MnrQueryValue is the 'MjrQuery' minor that classifies errors related with the query values.
	MnrQueryValue errors.Minor

	// QueryNilValue is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with no values provided.
	QueryNilValue errors.Class

	// QueryValueMissingRequired is the 'MjrQuery', 'MnrQueryValue' error classifcation
	// occurred on missing required field values.
	QueryValueMissingRequired errors.Class

	// QueryValueNoResult is the 'MjrQuery', 'MnrQueryValue' error classifaction
	// when query returns no results.
	QueryValueNoResult errors.Class

	// QueryValuePrimary is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with invalid primary field value.
	QueryValuePrimary errors.Class

	// QueryValueType is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with invalid or unsupported value type provided.
	QueryValueType errors.Class

	// QueryValueUnaddressable is the 'MjrQuery', 'MnrQueryValue' error classifcation
	// for queries with unadresable value type provided.
	QueryValueUnaddressable errors.Class

	// QueryValueValidation is the 'MjrQuery', 'MnrQueryValue' error classifaction
	// for queries with non validated values - validator failed.
	QueryValueValidation errors.Class
)

func registerQueryValue() {
	MnrQueryValue = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryValue
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryNilValue = newClass()
	QueryValueMissingRequired = newClass()
	QueryValueNoResult = newClass()
	QueryValuePrimary = newClass()
	QueryValueType = newClass()
	QueryValueUnaddressable = newClass()
	QueryValueValidation = newClass()
}

/**

Query Include

*/

var (
	// MnrQueryInclude is the 'MjrQuery' minor error classifcation
	// for the query includes issues.
	MnrQueryInclude errors.Minor

	// QueryIncludeTooMany is the 'MjrQuery', 'MnrQueryInclude' error classifcation
	// used on when provided too many includes or too deep included field.
	QueryIncludeTooMany errors.Class

	// QueryNotIncluded is the 'MjrQuery', 'MnrQueryInclude' error classification
	// used when the provided query collection / model is not included.
	QueryNotIncluded errors.Class
)

func registerQueryInclude() {
	MnrQueryInclude = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryInclude
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryIncludeTooMany = newClass()
	QueryNotIncluded = newClass()
}

/**

Query Transactions

*/

var (
	// MnrQueryTransaction is the 'MjrQuery' minor error classification that defines issues with the query transactions.
	MnrQueryTransaction errors.Minor

	// QueryTxBegin is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// for query tranasaction begins.
	QueryTxBegin errors.Class

	// QueryTxAlreadyBegin is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction had already begin.
	QueryTxAlreadyBegin errors.Class

	// QueryTxAlreadyResolved is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction had already resolved.
	QueryTxAlreadyResolved errors.Class

	// QueryTxNotFound is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction transaction not found for given query.
	QueryTxNotFound errors.Class

	// QueryTxRollback is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction with rollback issues.
	QueryTxRollback errors.Class

	// QueryTxFailed is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction failed.
	QueryTxFailed errors.Class

	// QueryTxUnknownState is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction state is unknown.
	QueryTxUnknownState errors.Class

	// QueryTxUnknownIsolationLevel is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction isolation level is unknown.
	QueryTxUnknownIsolationLevel errors.Class

	// QueryTxTermination is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction is incorrectly terminated.
	QueryTxTermination errors.Class

	// QueryTxLockTimeOut is the 'MjrQuery', 'MnrQueryTransaction' error classification
	// when query transaction lock had timed out.
	QueryTxLockTimeOut errors.Class

	// QueryTxDone is the Query Transaction error classification that states the transaction is finished.
	QueryTxDone errors.Class
)

func registerQueryTransactions() {
	MnrQueryTransaction = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryTransaction
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryTxAlreadyBegin = newClass()
	QueryTxAlreadyResolved = newClass()
	QueryTxBegin = newClass()
	QueryTxFailed = newClass()
	QueryTxNotFound = newClass()
	QueryTxRollback = newClass()
	QueryTxUnknownState = newClass()
	QueryTxUnknownIsolationLevel = newClass()
	QueryTxTermination = newClass()
	QueryTxLockTimeOut = newClass()
	QueryTxDone = newClass()
}

/**

Query Violation

*/

var (
	// MnrQueryViolation is the 'MjrQuery' minor error classifaction
	// related with query violations.
	MnrQueryViolation errors.Minor

	// QueryViolationIntegrityConstraint is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the query violates integrity constraint in example no foreign key exists for given insertion query.
	QueryViolationIntegrityConstraint errors.Class

	// QueryViolationNotNull is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the not null restriction is violated by the given insertion, patching query.
	QueryViolationNotNull errors.Class

	// QueryViolationForeignKey is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the foreign key is being violated.
	QueryViolationForeignKey errors.Class

	// QueryViolationUnique is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the uniqueness of the field is being violated. I.e. user defined primary key is not unique.
	QueryViolationUnique errors.Class

	// QueryViolationCheck is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the repository value check is violated. I.e. SQL Check value.
	QueryViolationCheck errors.Class

	// QueryViolationDataType is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the inserted data type violates the allowed data types.
	QueryViolationDataType errors.Class

	// QueryViolationRestrict is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// while doing restricted operation. I.e. SQL based repository with field marked as RESTRICT.
	QueryViolationRestrict errors.Class
)

func registerQueryViolation() {
	MnrQueryViolation = errors.MustNewMinor(MjrQuery)

	mjr, mnr := MjrQuery, MnrQueryViolation
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	QueryViolationIntegrityConstraint = newClass()
	QueryViolationNotNull = newClass()
	QueryViolationForeignKey = newClass()
	QueryViolationUnique = newClass()
	QueryViolationCheck = newClass()
	QueryViolationDataType = newClass()
	QueryViolationRestrict = newClass()
}

var (
	// MnrQueryProcessor is the 'MjrQuery' error classification related with
	// the issues with the processes and a processor.
	MnrQueryProcessor errors.Minor

	// QueryProcessorNotFound is the 'MjrQuery', 'MnrQueryProcessor' error classification
	// for the query processes not found.
	QueryProcessorNotFound errors.Class
)

func registerQueryProcessor() {
	MnrQueryProcessor = errors.MustNewMinor(MjrQuery)

	QueryProcessorNotFound = errors.MustNewClass(MjrQuery, MnrQueryProcessor, errors.MustNewIndex(MjrQuery, MnrQueryProcessor))
}

var (
	// MnrQueryRelation is the 'MjrQuery' minor error classification related
	// with the issues while querying the relations.
	MnrQueryRelation errors.Minor

	// QueryRelation is the 'MjrQuery', 'MnrQueryRelation' error classification
	// for general issues while querying the relations.
	QueryRelation errors.Class
	// QueryRelationNotFound is the 'MjrQuery', 'MnrQueryRelations' error classification
	// used when the relation query instance is not found.
	QueryRelationNotFound errors.Class
)

func registerQueryRelations() {
	MnrQueryRelation = errors.MustNewMinor(MjrQuery)

	QueryRelation = errors.MustNewMinorClass(MjrQuery, MnrQueryRelation)
	QueryRelationNotFound = errors.MustNewClass(MjrQuery, MnrQueryRelation, errors.MustNewIndex(MjrQuery, MnrQueryRelation))
}
