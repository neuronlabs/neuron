package pagination

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/paginations"
)

type Type int

const (
	TpLimitOffset Type = iota
	TpPage
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
		// not implemented yet
	}
	p.SetType(paginations.Type(tp))

	return (*Pagination)(p)
}

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
type Pagination paginations.Pagination

// GetLimitOffset gets the limit and offset from the current pagination
func (p *Pagination) GetLimitOffset() (int, int) {
	return (*paginations.Pagination)(p).GetLimitOffset()
}

// Check checks if the pagination is well formed
func (p *Pagination) Check() error {
	return (*paginations.Pagination)(p).Check()
}
