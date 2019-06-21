package paginations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetLimitOffset tests the pagination GetLimitOffset method.
func TestGetLimitOffset(t *testing.T) {
	p := &Pagination{tp: TpOffset, second: 1, first: 10}
	limit, offset := p.GetLimitOffset()
	assert.Equal(t, p.first, limit)
	assert.Equal(t, p.second, offset)

	p = &Pagination{tp: TpPage, first: 3, second: 9}
	limit, offset = p.GetLimitOffset()

	assert.Equal(t, p.first*p.second, offset)
	assert.Equal(t, p.second, limit)

}
