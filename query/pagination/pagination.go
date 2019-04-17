package pagination

import (
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/internal/query/paginations"
	"github.com/kucjac/jsonapi/log"
	"net/url"
	"strconv"
)

// Type defines the pagination type
type Type int

const (
	// TpLimitOffset is the pagination type that defines limit or (and) offset
	TpLimitOffset Type = iota

	// TpPage is the pagination type that uses page type pagination i.e. page=1 page size = 10
	TpPage

	// TpCursor is the current cursor pagination type
	TpCursor
)

// New creates new pagination for the given query
// The arguments 'a' and 'b' means as follows:
// - 'limit' and 'offset' for the type 'TpLimitOffset'
// - 'pageSize' and 'pageNumber' for the type 'TpPage'
func New(a, b int, tp Type) *Pagination {
	p := &paginations.Pagination{}
	switch tp {
	case TpLimitOffset:
		p.Limit = a
		p.Offset = b
	case TpPage:
		p.PageSize = a
		p.PageNumber = b
	case TpCursor:
		log.Debugf("Cursor Pagination type not implemented yet.")
		// not implemented yet
	}
	p.SetType(paginations.Type(tp))

	return (*Pagination)(p)
}

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
type Pagination paginations.Pagination

// NewPaged creates new paged type Pagination
func NewPaged(number, size int) *Pagination {
	p := &paginations.Pagination{
		PageNumber: number,
		PageSize:   size,
	}
	p.SetType(paginations.TpPage)
	return (*Pagination)(p)
}

// NewLimitOffset createw new limit, offset type pagination
func NewLimitOffset(limit, offset int) *Pagination {
	p := &paginations.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	p.SetType(paginations.TpOffset)
	return (*Pagination)(p)
}

// GetLimitOffset gets the limit and offset from the current pagination
func (p *Pagination) GetLimitOffset() (int, int) {
	return (*paginations.Pagination)(p).GetLimitOffset()
}

// Check checks if the pagination is well formed
func (p *Pagination) Check() error {
	return (*paginations.Pagination)(p).Check()
}

// String implements Stringer interface
func (p *Pagination) String() string {
	return (*paginations.Pagination)(p).String()
}

// Type returns pagination type
func (p *Pagination) Type() Type {
	return Type((*paginations.Pagination)(p).Type())
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
			k = internal.QueryParamPageLimit
			v = strconv.Itoa(limit)
			query.Set(k, v)
		}
		if offset != 0 {
			k = internal.QueryParamPageOffset
			v = strconv.Itoa(offset)
			query.Set(k, v)
		}
	case TpPage:
		if p.PageNumber != 0 {
			k = internal.QueryParamPageNumber
			v = strconv.Itoa(p.PageNumber)
			query.Set(k, v)
		}

		if p.PageSize != 0 {
			k = internal.QueryParamPageSize
			v = strconv.Itoa(p.PageSize)
			query.Set(k, v)
		}

	case TpCursor:
		log.Debugf("Cursor Pagination not implemented yet!")
	}
	return query
}
