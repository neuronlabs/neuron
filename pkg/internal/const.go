package internal

const (
	// StructTag Annotation strings
	AnnotationJSONAPI = "jsonapi"

	// type tag
	AnnotationPrimary     = "primary"
	AnnotationID          = "id"
	AnnotationClientID    = "client-id"
	AnnotationLanguage    = "langtag"
	AnnotationAttribute   = "attr"
	AnnotationRelation    = "relation"
	AnnotationFilterKey   = "filterkey"
	AnnotationForeignKey  = "foreign"
	AnnotationNestedField = "nested"

	// relation tag
	AnnotationManyToMany     = "many2many"
	AnnotationRelationNoSync = "nosync"
	annotaitonRelationSync   = "sync"

	AnnotationRelationRestrict = "restrict"
	AnnotationRelationNoAction = "no-action"
	AnnotationRelationCascade  = "cascade"
	AnnotationRelationSetNull  = "set-null"

	AnnotationDefault = "default"

	// name tag
	AnnotationName = "name"

	// flags tag
	AnnotationHidden      = "hidden"
	AnnotationISO8601     = "iso8601"
	AnnotationOmitEmpty   = "omitempty"
	AnnotationI18n        = "i18n"
	AnnotationFieldType   = "type"
	AnnotationFlags       = "flags"
	AnnotationNoFilter    = "nofilter"
	AnnotationNotSortable = "nosort"

	AnnotationSeperator         = ","
	AnnotationRelationSeperator = ":"
	AnnotationTagSeperator      = ";"
	AnnotationTagEqual          = '='

	AnnotationNestedSeperator = "."
	AnnotationOpenedBracket   = '['
	AnnotationClosedBracket   = ']'

	// FILTERS
	// disable for filtering purpose

	// not used currently
	// AnnotationNot = "not"
	// AnnotationOr  = "or"
	// AnnotationAnd = "and"

	HeaderAcceptLanguage  = "Accept-Language"
	HeaderContentLanguage = "Content-Language"

	// string only filters

	Iso8601TimeFormat = "2006-01-02T15:04:05Z"

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

	// Preset query Annotation
	QueryParamPreset = "preset"

	// QueryParamLanguage is a JSONAPI query parameter used in selecting a language tag for the
	// model.
	QueryParamLanguage = "language"

	// QueryParamLinks is a JSONAPI query parameter used in displaying the links of the relationships
	QueryParamLinks = "links"
)
