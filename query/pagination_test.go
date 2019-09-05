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
		require.NoError(t, p.Check())

		q := url.Values{}

		p.FormatQuery(q)
		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(ParamPageSize), "%+v", p)
		assert.Equal(t, "2", q.Get(ParamPageNumber))
	})

	t.Run("Limited", func(t *testing.T) {
		p := &Pagination{Size: 10}

		require.NoError(t, p.Check())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
	})

	t.Run("Offseted", func(t *testing.T) {
		p := &Pagination{Offset: 10}

		require.NoError(t, p.Check())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageOffset))
	})

	t.Run("LimitOffset", func(t *testing.T) {
		p := &Pagination{10, 140, LimitOffsetPagination}

		require.NoError(t, p.Check())

		q := p.FormatQuery()
		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
		assert.Equal(t, "140", q.Get(ParamPageOffset))
	})
}
