package query

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaginationFormatQuery tests the Pagination.FormatQuery method.
func TestPaginationFormatQuery(t *testing.T) {
	t.Run("Paged", func(t *testing.T) {
		p := &Pagination{Size: 10, Offset: 2, Type: PageNumberPagination}
		require.NoError(t, p.IsValid())

		q := url.Values{}

		p.FormatQuery(q)
		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(ParamPageSize))
		assert.Equal(t, "2", q.Get(ParamPageNumber))
	})

	t.Run("Limited", func(t *testing.T) {
		p := &Pagination{Size: 10}

		require.NoError(t, p.IsValid())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
	})

	t.Run("Offseted", func(t *testing.T) {
		p := &Pagination{Offset: 10}

		require.NoError(t, p.IsValid())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageOffset))
	})

	t.Run("LimitOffset", func(t *testing.T) {
		p := &Pagination{10, 140, LimitOffsetPagination}

		require.NoError(t, p.IsValid())

		q := p.FormatQuery()
		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
		assert.Equal(t, "140", q.Get(ParamPageOffset))
	})
}

// TestGetLimitOffset tests the pagination GetLimitOffset method.
func TestGetLimitOffset(t *testing.T) {
	// At first check the default 'LimitOffsetPagination'.
	p := &Pagination{Size: 1, Offset: 10}
	limit, offset := p.GetLimitOffset()
	assert.Equal(t, int64(1), limit)
	assert.Equal(t, int64(10), offset)

	// Check the pagination with PageNumberPagination type.
	t.Run("PageNumber", func(t *testing.T) {
		p = &Pagination{Type: PageNumberPagination, Size: 3, Offset: 9}
		limit, offset = p.GetLimitOffset()

		// the offset would be (pageNumber - 1) * pageSize = 8 * 3 = 24
		assert.Equal(t, int64(24), offset)
		assert.Equal(t, int64(3), limit)

		// Check the pagination without page number defined.
		p = &Pagination{Type: PageNumberPagination, Size: 10}
		limit, offset = p.GetLimitOffset()

		assert.Equal(t, int64(0), offset)
		assert.Equal(t, int64(10), limit)
	})
}

// TestGetNumberSize tests the pagination GetNumberSize method.
func TestGetNumberSize(t *testing.T) {
	// At first check the 'PageNumberPagination'.
	p := &Pagination{Type: PageNumberPagination, Size: 3, Offset: 9}
	number, size := p.GetNumberSize()

	assert.Equal(t, int64(9), number)
	assert.Equal(t, int64(3), size)

	t.Run("LimitOffsetTyped", func(t *testing.T) {
		// Check the 'LimitOffsetPagination' with defined fields.
		p = &Pagination{Size: 3, Offset: 10}
		number, size = p.GetNumberSize()
		assert.Equal(t, int64(3), size)
		assert.Equal(t, int64(4), number)

		// Check the values with zero values.
		p = &Pagination{Size: 10}
		number, size = p.GetNumberSize()
		assert.Equal(t, int64(10), size)
		assert.Equal(t, int64(1), number)
	})
}
