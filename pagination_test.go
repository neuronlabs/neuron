package jsonapi

import (
	"testing"
)

func TestGetLimitOffset(t *testing.T) {
	p := &Pagination{Type: OffsetPaginate, Offset: 1, Limit: 10}
	limit, offset := p.GetLimitOffset()
	assertEqual(t, p.Limit, limit)
	assertEqual(t, p.Offset, offset)

	p = &Pagination{Type: PagePaginate, PageNumber: 3, PageSize: 9}
	limit, offset = p.GetLimitOffset()

	assertEqual(t, p.PageNumber*p.PageSize, offset)
	assertEqual(t, p.PageSize, limit)

	p = &Pagination{Type: CursorPaginate}
	p.GetLimitOffset()
}
