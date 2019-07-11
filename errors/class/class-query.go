package class

// MjrQuery - major that classifies all the errors related with
// creating, operating on or changing the queries.
var MjrQuery Major

func registerQueryClasses() {
	MjrQuery = MustRegisterMajor("Query")

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
}

/**

Query Fieldset

*/

var (
	// MnrQueryFieldset is the 'MjrQuery' minor error classification
	// for query fieldset related issues.
	MnrQueryFieldset Minor

	// QueryFieldsetTooBig is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// used when provided more than possible fields in the fieldset.
	QueryFieldsetTooBig Class

	// QueryFieldsetUnknownField is the 'MjrQuery', 'MnrQueryFieldset' error classificatio
	// used when the provided fieldset field is not found within the model's collection.
	QueryFieldsetUnknownField Class

	// QueryFieldsetDuplicate is the 'MjrQuery', 'MnrQueryFieldset' error classificatio
	// for duplicated fieldset field's name.
	QueryFieldsetDuplicate Class

	// QueryFieldsetInvalid is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// for invalid fieldset issues.
	QueryFieldsetInvalid Class

	// QueryFieldsetEmpty is the 'MjrQuery', 'MnrQueryFieldset' error classification
	// for empty fieldset issues.
	QueryFieldsetEmpty Class
)

func registerQueryFieldset() {
	MnrQueryFieldset = MjrQuery.MustRegisterMinor("Fieldset", "issues related with the query fieldset")

	QueryFieldsetDuplicate = MnrQueryFieldset.MustRegisterIndex("Duplicates", "duplicated field found within the fieldset").Class()
	QueryFieldsetEmpty = MnrQueryFieldset.MustRegisterIndex("Empty", "empty fieldset issues").Class()
	QueryFieldsetInvalid = MnrQueryFieldset.MustRegisterIndex("Invalid", "invalid fieldset field type").Class()
	QueryFieldsetTooBig = MnrQueryFieldset.MustRegisterIndex("Too Big", "too many fields tried to be set within the query fieldset").Class()
	QueryFieldsetUnknownField = MnrQueryFieldset.MustRegisterIndex("Unknown Field", "provided fieldset field not found within the model's collection").Class()
}

/**

Query Selected Fields

*/

var (
	// MnrQuerySelectedFields is the 'MjrQuery' minor error classifcation
	// for the selected fields issues.
	MnrQuerySelectedFields Minor

	// QuerySelectedFieldsNotFound is the 'MjrQuery', 'MnrQuerySelectedFields' error classifcation
	// when the selected fields is not found.
	QuerySelectedFieldsNotFound Class

	// QuerySelectedFieldsNotSelected is the 'MjrQuery', 'MnrQuerySelectedFields' error classifcation
	// when unselecting fields which were not selected.
	QuerySelectedFieldsNotSelected Class

	// QuerySelectedFieldsInvalidModel is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// when provided selected field is of invalid or non matched model.
	QuerySelectedFieldsInvalidModel Class

	// QuerySelectedFieldInvalid is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// when the selected field is not valid.
	QuerySelectedFieldInvalid Class

	// QuerySelectedFieldAlreadyUsed is the 'MjrQuery', 'MnrQuerySelectedFields' error classification
	// whenthe selected field is already 'selected'.
	QuerySelectedFieldAlreadyUsed Class
)

func registerQuerySelectedFields() {
	MnrQuerySelectedFields = MjrQuery.MustRegisterMinor("Selected Fields", "issues related with the query selected fields")

	QuerySelectedFieldsNotFound = MnrQuerySelectedFields.MustRegisterIndex("Not Found", "selected fields not found").Class()
	QuerySelectedFieldsNotSelected = MnrQuerySelectedFields.MustRegisterIndex("Not Selected", "unselecting fields which were not selected").Class()
	QuerySelectedFieldsInvalidModel = MnrQuerySelectedFields.MustRegisterIndex("Invalid Model", "provided seleceted field's model is invalid or doesn't match given query model").Class()
	QuerySelectedFieldInvalid = MnrQuerySelectedFields.MustRegisterIndex("Invalid", "invalid selected field").Class()
	QuerySelectedFieldAlreadyUsed = MnrQuerySelectedFields.MustRegisterIndex("AlreadyUsed", "field is already selected").Class()
}

/**

Query Filters

*/

var (
	// MnrQueryFilter is the 'MjrQuery' minor that classifies errors related with the query filters.
	MnrQueryFilter Minor

	// QueryFilterInvalidField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query filters field - name.
	QueryFilterInvalidField Class

	// QueryFilterInvalidFormat is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query filter format.
	QueryFilterInvalidFormat Class

	// QueryFilterLanguage is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid query language filter.	.
	QueryFilterLanguage Class

	// QueryFilterMissingRequired is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for missing query filters.
	QueryFilterMissingRequired Class

	// QueryFilterUnknownCollection is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown query filter collection provided.
	QueryFilterUnknownCollection Class

	// QueryFilterUnknownField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown query filter field.
	QueryFilterUnknownField Class

	// QueryFilterUnknownOperator is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unknown - invalid filter operator.
	QueryFilterUnknownOperator Class

	// QueryFilterUnsupportedField is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unsupported field types.
	QueryFilterUnsupportedField Class

	// QueryFilterUnsupportedOperator is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for unsupported query filter operator.
	QueryFilterUnsupportedOperator Class

	// QueryFilterValue is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid filter values.
	QueryFilterValue Class

	// QueryFitlerNonMatched is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for non matched model structures or field structures to models.
	QueryFitlerNonMatched Class

	// QueryFilterFieldKind is the 'MjrQuery', 'MnrQueryFilter' error classification
	// for invalid, unknown or unsupported filter field kind.
	QueryFilterFieldKind Class
)

func registerQueryFilters() {
	MnrQueryFilter = MjrQuery.MustRegisterMinor("Filters", "defines errors related with the query filters")

	QueryFilterInvalidField = MnrQueryFilter.MustRegisterIndex("Invalid Field", "define errors related with invalid filter field").Class()
	QueryFilterInvalidFormat = MnrQueryFilter.MustRegisterIndex("Invalid Format", "define errors related with invalid query filter format").Class()
	QueryFilterLanguage = MnrQueryFilter.MustRegisterIndex("Language", "language filter related issues").Class()
	QueryFilterMissingRequired = MnrQueryFilter.MustRegisterIndex("Missing Required", "missing required filter field").Class()
	QueryFilterUnknownCollection = MnrQueryFilter.MustRegisterIndex("Unknown Collection", "define errors related with unknown/invalid query filter operator").Class()
	QueryFilterUnknownField = MnrQueryFilter.MustRegisterIndex("Unknown Field", "defines errors related with unknown/invalid filter query field").Class()
	QueryFilterUnknownOperator = MnrQueryFilter.MustRegisterIndex("Unknown Operator", "define errors related with unknown query filter operator").Class()
	QueryFilterUnsupportedField = MnrQueryFilter.MustRegisterIndex("Unsupported Field", "unsupported field type issues").Class()
	QueryFilterUnsupportedOperator = MnrQueryFilter.MustRegisterIndex("Unsupported Operator", "define errors related with unsupported query filter operator").Class()
	QueryFilterValue = MnrQueryFilter.MustRegisterIndex("Value", "define errors related with invalid query filter value").Class()
	QueryFitlerNonMatched = MnrQueryFilter.MustRegisterIndex("NonMatched", "model structures doesn't match").Class()
}

/**

Query Sorts

*/

var (
	// MnrQuerySorts is the 'MjrQuery' minor that classifies errors related with the sorts.
	MnrQuerySorts Minor

	// QuerySortField is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for unknown, unsupported fields.
	QuerySortField Class

	// QuerySortFormat is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for invalid sorting format.
	QuerySortFormat Class

	// QuerySortTooManyFields is the 'MjrQuery', 'MnrQuerySorts' error classification
	// when too many sorting fields provided.
	QuerySortTooManyFields Class

	// QuerySortRelatedFields is the 'MjrQuery', 'MnrQuerySorts' error classification
	// for unsupported sorting by related fields.
	QuerySortRelatedFields Class
)

func registerQuerySorts() {
	MnrQuerySorts = MjrQuery.MustRegisterMinor("Sorts", "related with the query sorts")

	QuerySortField = MnrQuerySorts.MustRegisterIndex("Field", "related with the query sort field").Class()
	QuerySortFormat = MnrQuerySorts.MustRegisterIndex("Format", "related with the query sort format").Class()
	QuerySortTooManyFields = MnrQuerySorts.MustRegisterIndex("Too Many Fields", "too many fields while adding sorting fields").Class()
	QuerySortRelatedFields = MnrQuerySorts.MustRegisterIndex("Related Fields", "related with unsupported query sorts by related fields").Class()

}

/**

Query Pagination

*/

var (
	// MnrQueryPagination is the 'MjrQuery' minor that classifies errors related with the pagination.
	MnrQueryPagination Minor

	// QueryPaginationValue is the 'MjrQuery', 'MnrQueryPagination' error classification
	// for invalid pagination values. I.e. invalid limit, offset value.
	QueryPaginationValue Class

	// QueryPaginationType is the 'MjrQuery', 'MnrQueryPagination' error classification
	// for invalid pagination type provided.
	QueryPaginationType Class

	// QueryPaginationAlreadySet is the 'MjrQuery', 'MnrQueryPagination' error classification
	// while trying to add pagination when it is already set.
	QueryPaginationAlreadySet Class
)

func registerQueryPagination() {
	MnrQueryPagination = MjrQuery.MustRegisterMinor("Pagination", "defines errors related with query paginatinos")

	QueryPaginationValue = MnrQueryPagination.MustRegisterIndex("Value", "invalid query pagination's value - i.e. invalid / unsupported limit or offset value").Class()
	QueryPaginationType = MnrQueryPagination.MustRegisterIndex("Type", "invalid query pagination type").Class()
	QueryPaginationAlreadySet = MnrQueryPagination.MustRegisterIndex("Already Set", "pagination is already set").Class()
}

/**

Query Value

*/

var (
	// MnrQueryValue is the 'MjrQuery' minor that classifies errors related with the query values.
	MnrQueryValue Minor

	// QueryNoValue is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with no values provided.
	QueryNoValue Class

	// QueryValueMissingRequired is the 'MjrQuery', 'MnrQueryValue' error classifcation
	// occurred on missing required field values.
	QueryValueMissingRequired Class

	// QueryValueNoResult is the 'MjrQuery', 'MnrQueryValue' error classifaction
	// when query returns no results.
	QueryValueNoResult Class

	// QueryValuePrimary is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with invalid primary field value.
	QueryValuePrimary Class

	// QueryValueType is the 'MjrQuery', 'MnrQueryValue' error classification
	// for queries with invalid or unsupported value type provided.
	QueryValueType Class

	// QueryValueUnaddressable is the 'MjrQuery', 'MnrQueryValue' error classifcation
	// for queries with unadresable value type provided.
	QueryValueUnaddressable Class

	// QueryValueValidation is the 'MjrQuery', 'MnrQueryValue' error classifaction
	// for queries with non validated values - validator failed.
	QueryValueValidation Class
)

func registerQueryValue() {
	MnrQueryValue = MjrQuery.MustRegisterMinor("Value", "defines errors related with the query value")

	QueryNoValue = MnrQueryValue.MustRegisterIndex("Nil", "no value provided for the query").Class()
	QueryValueMissingRequired = MnrQueryValue.MustRegisterIndex("Missing Required", "missing required field values").Class()
	QueryValueNoResult = MnrQueryValue.MustRegisterIndex("No Result", "query returns no result value").Class()
	QueryValuePrimary = MnrQueryValue.MustRegisterIndex("Primary", "query primary field value is not valid").Class()
	QueryValueType = MnrQueryValue.MustRegisterIndex("Type", "invalid or unsupported query filter value type").Class()
	QueryValueUnaddressable = MnrQueryValue.MustRegisterIndex("Unaddressable", "provided unaddressable query value").Class()
	QueryValueValidation = MnrQueryValue.MustRegisterIndex("Validation", "query value not validated within the process").Class()
}

/**

Query Include

*/

var (
	// MnrQueryInclude is the 'MjrQuery' minor error classifcation
	// for the query includes issues.
	MnrQueryInclude Minor

	// QueryIncludeTooMany is the 'MjrQuery', 'MnrQueryInclude' error classifcation
	// used on when provided too many includes or too deep included field.
	QueryIncludeTooMany Class

	// QueryNotIncluded is the 'MjrQuery', 'MnrQueryInclude' error classification
	// used when the provided query collection / model is not included.
	QueryNotIncluded Class
)

func registerQueryInclude() {
	MnrQueryInclude = MjrQuery.MustRegisterMinor("Include", "issues related with the included fields")

	QueryIncludeTooMany = MnrQueryInclude.MustRegisterIndex("Too Many", "included too many or too deep than possible included fields").Class()
	QueryNotIncluded = MnrQueryInclude.MustRegisterIndex("Not Included", "given model is not included within given query").Class()
}

/**

Query Transactions

*/

var (
	// MnrQueryTransaction is the 'MjrQuery' minor error classification that defines issues with the query transactions.
	MnrQueryTransaction Minor

	// QueryTxBegin is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// for query tranasaction begins.
	QueryTxBegin Class

	// QueryTxAlreadyBegin is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction had already begin.
	QueryTxAlreadyBegin Class

	// QueryTxAlreadyResolved is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction had already resolved.
	QueryTxAlreadyResolved Class

	// QueryTxNotFound is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction transaction not found for given query.
	QueryTxNotFound Class

	// QueryTxRollback is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction with rollback issues.
	QueryTxRollback Class

	// QueryTxFailed is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction failed.
	QueryTxFailed Class

	// QueryTxUnknownState is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction state is unknown.
	QueryTxUnknownState Class

	// QueryTxUnknownIsolationLevel is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction isolation level is unknown.
	QueryTxUnknownIsolationLevel Class

	// QueryTxTermination is the 'MjrQuery', 'MnrQueryTransaction' error classifaction
	// when query transaction is invalidly terminated.
	QueryTxTermination Class

	// QueryTxLockTimeOut is the 'MjrQuery', 'MnrQueryTransaction' error classification
	// when query transaction lock had timed out.
	QueryTxLockTimeOut Class
)

func registerQueryTransactions() {
	MnrQueryTransaction = MjrQuery.MustRegisterMinor("Transaction", "query transaction issues")

	QueryTxAlreadyBegin = MnrQueryTransaction.MustRegisterIndex("Already Begin", "query transaction already begin").Class()
	QueryTxAlreadyResolved = MnrQueryTransaction.MustRegisterIndex("Already Resolved", "query transaction already resolved").Class()
	QueryTxBegin = MnrQueryTransaction.MustRegisterIndex("Begin", "query transaction begin").Class()
	QueryTxFailed = MnrQueryTransaction.MustRegisterIndex("Failed", "transaction failed").Class()
	QueryTxNotFound = MnrQueryTransaction.MustRegisterIndex("Not Found", "transaction not found for given query").Class()
	QueryTxRollback = MnrQueryTransaction.MustRegisterIndex("Rollback", "query transaction with rollback issues").Class()
	QueryTxUnknownState = MnrQueryTransaction.MustRegisterIndex("Unknown State", "transaction state unknown").Class()
	QueryTxUnknownIsolationLevel = MnrQueryTransaction.MustRegisterIndex("Unknown Isolation Level", "transaction isolation level unknown").Class()
	QueryTxTermination = MnrQueryTransaction.MustRegisterIndex("Termination", "invalid transaction termination").Class()
	QueryTxLockTimeOut = MnrQueryTransaction.MustRegisterIndex("Lock Timed Out", "transaction lock had timed out").Class()
}

/**

Query Violation

*/

var (
	// MnrQueryViolation is the 'MjrQuery' minor error classifaction
	// related with query violations.
	MnrQueryViolation Minor

	// QueryViolationIntegrityConstraint is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the query violates integrity constraint in example no foreign key exists for given insertion query.
	QueryViolationIntegrityConstraint Class

	// QueryViolationNotNull is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the not null restriction is violated by the given insertion, patching query.
	QueryViolationNotNull Class

	// QueryViolationForeignKey is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the foreign key is being violated.
	QueryViolationForeignKey Class

	// QueryViolationUnique is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the uniqueness of the field is being violated. I.e. user defined primary key is not unique.
	QueryViolationUnique Class

	// QueryViolationCheck is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the repository value check is violated. I.e. SQL Check value.
	QueryViolationCheck Class

	// QueryViolationDataType is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// when the inserted data type violates the allowed data types.
	QueryViolationDataType Class

	// QueryViolationRestrict is the 'MjrQuery', 'MnrQueryViolation' error classifcation
	// while doing restricted operation. I.e. SQL based repository with field marked as RESTRICT.
	QueryViolationRestrict Class
)

func registerQueryViolation() {
	MnrQueryViolation = MjrQuery.MustRegisterMinor("Violation", "query violates some constraints")

	QueryViolationIntegrityConstraint = MnrQueryViolation.MustRegisterIndex("Intergrity Constraint", "the integrity constraint is violated by the query").Class()
	QueryViolationNotNull = MnrQueryViolation.MustRegisterIndex("Not Null", "Not null constraint is violated by the query").Class()
	QueryViolationForeignKey = MnrQueryViolation.MustRegisterIndex("Foreign Key", "inserting or patching foreign key of non existing model instance").Class()
	QueryViolationUnique = MnrQueryViolation.MustRegisterIndex("Unique", "violating unique restriction on the field").Class()
	QueryViolationCheck = MnrQueryViolation.MustRegisterIndex("Check", "repository based data value check failed").Class()
	QueryViolationDataType = MnrQueryViolation.MustRegisterIndex("Data Type", "inserting or patching invalid data type value").Class()
	QueryViolationRestrict = MnrQueryViolation.MustRegisterIndex("Restrict", "restricted operation").Class()
}

var (
	// MnrQueryProcessor is the 'MjrQuery' error classification related with
	// the issues with the processes and a processor.
	MnrQueryProcessor Minor

	// QueryProcessorNotFound is the 'MjrQuery', 'MnrQueryProcessor' error classification
	// for the query processes not found.
	QueryProcessorNotFound Class
)

func registerQueryProcessor() {
	MnrQueryProcessor = MjrQuery.MustRegisterMinor("Processor", "issues related with the query processes / processor")

	QueryProcessorNotFound = MnrQueryProcessor.MustRegisterIndex("Not Found", "query process / processor not found").Class()
}

var (
	// MnrQueryRelation is the 'MjrQuery' minor error classification related
	// with the issues while querying the relations.
	MnrQueryRelation Minor

	// QueryRelation is the 'MjrQuery', 'MnrQueryRelation' error classification
	// for general issues while querying the relations.
	QueryRelation Class
	// QueryRelationNotFound is the 'MjrQuery', 'MnrQueryRelations' error classification
	// used when the relation query instance is not found.
	QueryRelationNotFound Class
)

func registerQueryRelations() {
	MnrQueryRelation = MjrQuery.MustRegisterMinor("Relation", "issues with querying, patching, creating relations")

	QueryRelation = MustNewMinorClass(MnrQueryRelation)
	QueryRelationNotFound = MnrQueryRelation.MustRegisterIndex("Not Found", "queried relation is not found").Class()
}
