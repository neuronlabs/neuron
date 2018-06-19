package jsonapi

const (
	// StructTag annotation strings
	annotationJSONAPI           = "jsonapi"
	annotationPrimary           = "primary"
	annotationID                = "id"
	annotationClientID          = "client-id"
	annotationLanguage          = "langtag"
	annotationAttribute         = "attr"
	annotationRelation          = "relation"
	annotationOmitEmpty         = "omitempty"
	annotationI18n              = "i18n"
	annotationISO8601           = "iso8601"
	annotationSeperator         = ","
	annotationRelationSeperator = ":"
	annotationNestedSeperator   = "."
	annotationOpenedBracket     = '['
	annotationClosedBracket     = ']'

	// Hidden field
	annotationHidden = "hidden"

	annotationsManyToMany = "many2many"

	// FILTERS
	// disable for filtering purpose
	annotationNoFilter = "nofilter"

	annotationSortable = "sortable"

	// logical filters
	annotationEqual        = "eq"
	annotationIn           = "in"
	annotationNotEqual     = "ne"
	annotationNotIn        = "notin"
	annotationGreaterThan  = "gt"
	annotationGreaterEqual = "ge"
	annotationLessThan     = "lt"
	annotationLessEqual    = "le"

	// not used currently
	// annotationNot = "not"
	// annotationOr  = "or"
	// annotationAnd = "and"

	// string only filters
	annotationContains   = "contains"
	annotationStartsWith = "startswith"
	annotationEndsWith   = "endswith"

	iso8601TimeFormat = "2006-01-02T15:04:05Z"

	// MediaType is the identifier for the JSON API media type
	//
	// see http://jsonapi.org/format/#document-structure
	MediaType = "application/vnd.api+json"

	// Pagination Constants
	//
	// http://jsonapi.org/format/#fetching-pagination

	// KeyFirstPage is the key to the links object whose value contains a link to
	// the first page of data
	KeyFirstPage = "first"
	// KeyLastPage is the key to the links object whose value contains a link to
	// the last page of data
	KeyLastPage = "last"
	// KeyPreviousPage is the key to the links object whose value contains a link
	// to the previous page of data
	KeyPreviousPage = "prev"
	// KeyNextPage is the key to the links object whose value contains a link to
	// the next page of data
	KeyNextPage = "next"

	// KeyTotalPages is the key to the meta object whose value contains the total
	// pagination pages
	KeyTotalPages = "total-pages"

	PresetValue = "jsonapi_preset_value"

	// QueryParamPage is a JSON API query parameter used as for pagination.
	QueryParamPage = "page"

	// QueryParamPageNumber is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with QueryParamPageSize
	QueryParamPageNumber = "page[number]"
	// QueryParamPageSize is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with QueryParamPageNumber
	QueryParamPageSize = "page[size]"

	// QueryParamPageOffset is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with QueryParamPageLimit
	QueryParamPageOffset = "page[offset]"
	// QueryParamPageLimit is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with QueryParamPageOffset
	QueryParamPageLimit = "page[limit]"

	// QueryParamPageCursor is a JSON API query parameter used with a cursor-based
	// strategy
	QueryParamPageCursor = "page[cursor]"

	// QueryParamPageTotal is a JSON API query parameter used in pagination
	// It tells to API to add information about total-pages or total-count
	// (depending on the current strategy)
	QueryParamPageTotal = "page[total]"

	// QueryParamInclude
	QueryParamInclude = "include"

	// QueryParamSort
	QueryParamSort = "sort"

	// QueryParamFilter
	QueryParamFilter = "filter"

	// QueryParamFields
	QueryParamFields = "fields"

	// Preset query annotation
	QueryParamPreset = "preset"
)
