package paginations

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLimitOffset(t *testing.T) {
	p := &Pagination{Type: TpOffset, Offset: 1, Limit: 10}
	limit, offset := p.GetLimitOffset()
	assert.Equal(t, p.Limit, limit)
	assert.Equal(t, p.Offset, offset)

	p = &Pagination{Type: TpPage, PageNumber: 3, PageSize: 9}
	limit, offset = p.GetLimitOffset()

	assert.Equal(t, p.PageNumber*p.PageSize, offset)
	assert.Equal(t, p.PageSize, limit)

	p = &Pagination{Type: TpCursor}
	p.GetLimitOffset()
}
