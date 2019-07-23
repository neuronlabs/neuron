package query

import (
	"net/url"
	"strconv"

	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/query/paginations"
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
			k = common.QueryParamPageLimit
			v = strconv.Itoa(limit)
			query.Set(k, v)
		}
		if offset != 0 {
			k = common.QueryParamPageOffset
			v = strconv.Itoa(offset)
			query.Set(k, v)
		}
	case TpPage:
		number, size := (*paginations.Pagination)(p).GetNumberSize()
		if number != 0 {
			k = common.QueryParamPageNumber
			v = strconv.Itoa(number)
			query.Set(k, v)
		}

		if size != 0 {
			k = common.QueryParamPageSize
			v = strconv.Itoa(size)
			query.Set(k, v)
		}

	default:
		log.Debugf("Pagination with invalid pagination type: '%s'", p.Type())
	}
	return query
}

// GetLimitOffset gets the limit and offset from the current pagination.
func (p *Pagination) GetLimitOffset() (int, int) {
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
