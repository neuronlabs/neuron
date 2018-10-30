package jsonapi

const (
	// StructTag annotation strings
	annotationJSONAPI = "jsonapi"

	// type tag
	annotationPrimary   = "primary"
	annotationID        = "id"
	annotationClientID  = "client-id"
	annotationLanguage  = "langtag"
	annotationAttribute = "attr"
	annotationRelation  = "relation"

	// relation tag
	annotationManyToMany     = "many2many"
	annotationRelationNoSync = "nosync"
	annotaitonRelationSync   = "sync"
	annotationForeignKey     = "foreign"

	annotationRelationRestrict = "restrict"
	annotationRelationNoAction = "no-action"
	annotationRelationCascade  = "cascade"
	annotationRelationSetNull  = "set-null"

	annotationDefault = "default"

	// name tag
	annotationName = "name"

	// flags tag
	annotationHidden      = "hidden"
	annotationISO8601     = "iso8601"
	annotationOmitEmpty   = "omitempty"
	annotationI18n        = "i18n"
	annotationFieldType   = "type"
	annotationFlags       = "flags"
	annotationNoFilter    = "nofilter"
	annotationNotSortable = "nosort"

	annotationSeperator         = ","
	annotationRelationSeperator = ":"
	annotationTagSeperator      = ";"
	annotationTagEqual          = '='

	annotationNestedSeperator = "."
	annotationOpenedBracket   = '['
	annotationClosedBracket   = ']'

	// FILTERS
	// disable for filtering purpose

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

	headerAcceptLanguage  = "Accept-Language"
	headerContentLanguage = "Content-Language"

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

	// QueryParamLanguage is a JSONAPI query parameter used in selecting a language tag for the
	// model.
	QueryParamLanguage = "language"

	// QueryParamLinks is a JSONAPI query parameter used in displaying the links of the relationships
	QueryParamLinks = "links"
)
