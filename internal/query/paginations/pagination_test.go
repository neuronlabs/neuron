package paginations

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetLimitOffset tests the pagination GetLimitOffset method
func TestGetLimitOffset(t *testing.T) {
	p := &Pagination{tp: TpOffset, Offset: 1, Limit: 10}
	limit, offset := p.GetLimitOffset()
	assert.Equal(t, p.Limit, limit)
	assert.Equal(t, p.Offset, offset)

	p = &Pagination{tp: TpPage, PageNumber: 3, PageSize: 9}
	limit, offset = p.GetLimitOffset()

	assert.Equal(t, p.PageNumber*p.PageSize, offset)
	assert.Equal(t, p.PageSize, limit)

	p = &Pagination{tp: TpCursor}
	p.GetLimitOffset()
}
