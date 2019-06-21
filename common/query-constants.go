package common

// Pagination constants
const (
	// QueryParamPage is a JSON API query parameter used as for pagination.
	QueryParamPage = "page"

	// QueryParamPageNumber is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with QueryParamPageSize.
	QueryParamPageNumber = "page[number]"
	// QueryParamPageSize is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with QueryParamPageNumber.
	QueryParamPageSize = "page[size]"

	// QueryParamPageOffset is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with QueryParamPageLimit.
	QueryParamPageOffset = "page[offset]"
	// QueryParamPageLimit is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with QueryParamPageOffset.
	QueryParamPageLimit = "page[limit]"

	// QueryParamPageCursor is a JSON API query parameter used with a cursor-based
	// strategy.
	QueryParamPageCursor = "page[cursor]"

	// QueryParamPageTotal is a JSON API query parameter used in pagination
	// It tells to API to add information about total-pages or total-count
	// (depending on the current strategy).
	QueryParamPageTotal = "page[total]"
)

// Sorts Constants
const (
	// QueryParamSort is the url query parameter name for the sorting fields.
	QueryParamSort = "sort"
)

const (
	// QueryParamLanguage is the language query parameter used in the url.Values.
	QueryParamLanguage = "lang"
)
