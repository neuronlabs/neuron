package query

import (
	"net/url"
	"strconv"

	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/query/paginations"
)

// Pagination constants
const (
	// ParamPage is a JSON API query parameter used as for pagination.
	ParamPage = "page"

	// ParamPageNumber is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with ParamPageSize.
	ParamPageNumber = "page[number]"
	// ParamPageSize is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with ParamPageNumber.
	ParamPageSize = "page[size]"

	// ParamPageOffset is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageLimit.
	ParamPageOffset = "page[offset]"
	// ParamPageLimit is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageOffset.
	ParamPageLimit = "page[limit]"

	// ParamPageCursor is a JSON API query parameter used with a cursor-based
	// strategy.
	ParamPageCursor = "page[cursor]"

	// ParamPageTotal is a JSON API query parameter used in pagination
	// It tells to API to add information about total-pages or total-count
	// (depending on the current strategy).
	ParamPageTotal = "page[total]"
)

// PaginationType defines the pagination type.
type PaginationType int

const (
	// TpLimitOffset is the pagination type that defines limit or (and) offset.
	TpLimitOffset PaginationType = iota

	// TpPage is the pagination type that uses page type pagination i.e. page=1 page size = 10.
	TpPage
)

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
type Pagination paginations.Pagination

// Check checks if the pagination is well formed.
func (p *Pagination) Check() error {
	return (*paginations.Pagination)(p).Check()
}

// FormatQuery formats the pagination for the url query.
func (p *Pagination) FormatQuery(q ...url.Values) url.Values {
	var query url.Values
	if len(q) != 0 {
		query = q[0]
	}

	if query == nil {
		query = url.Values{}
	}

	var k, v string

	switch p.Type() {
	case TpLimitOffset:
		limit, offset := p.GetLimitOffset()
		if limit != 0 {
			k = ParamPageLimit
			v = strconv.Itoa(limit)
			query.Set(k, v)
		}
		if offset != 0 {
			k = ParamPageOffset
			v = strconv.Itoa(offset)
			query.Set(k, v)
		}
	case TpPage:
		number, size := (*paginations.Pagination)(p).GetNumberSize()
		if number != 0 {
			k = ParamPageNumber
			v = strconv.Itoa(number)
			query.Set(k, v)
		}

		if size != 0 {
			k = ParamPageSize
			v = strconv.Itoa(size)
			query.Set(k, v)
		}

	default:
		log.Debugf("Pagination with invalid pagination type: '%s'", p.Type())
	}
	return query
}

// GetNumberSize gets the page number and page size from the provided pagination.
func (p *Pagination) GetNumberSize() (number, size int) {
	return (*paginations.Pagination)(p).GetNumberSize()
}

// GetLimitOffset gets the limit and offset from the current pagination.
func (p *Pagination) GetLimitOffset() (limit int, offset int) {
	return (*paginations.Pagination)(p).GetLimitOffset()
}

// SetLimit sets the limit for the pagination.
func (p *Pagination) SetLimit(limit int) {
	(*paginations.Pagination)(p).SetValue(limit, paginations.ParamLimit)
}

// SetOffset sets the offset for the pagination.
func (p *Pagination) SetOffset(offset int) {
	(*paginations.Pagination)(p).SetValue(offset, paginations.ParamOffset)
}

// SetPageNumber sets the page number for the pagination.
func (p *Pagination) SetPageNumber(pageNumber int) {
	(*paginations.Pagination)(p).SetValue(pageNumber, paginations.ParamNumber)
}

// SetPageSize sets the page number for the pagination.
func (p *Pagination) SetPageSize(pageSize int) {
	(*paginations.Pagination)(p).SetValue(pageSize, paginations.ParamSize)
}

// String implements fmt.Stringer interface.
func (p *Pagination) String() string {
	return (*paginations.Pagination)(p).String()
}

// Type returns pagination type.
func (p *Pagination) Type() PaginationType {
	return PaginationType((*paginations.Pagination)(p).Type())
}

// LimitOffsetPagination createw new limit, offset type pagination.
func LimitOffsetPagination(limit, offset int) *Pagination {
	return newLimitOffset(limit, offset)
}

// PagedPagination creates new paged type Pagination.
func PagedPagination(number, size int) *Pagination {
	return newPaged(number, size)
}

func newPaged(number, size int) *Pagination {
	return (*Pagination)(paginations.NewPaged(number, size))
}

func newLimitOffset(limit, offset int) *Pagination {
	return (*Pagination)(paginations.NewLimitOffset(limit, offset))
}
