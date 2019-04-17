package pagination

import (
	"github.com/kucjac/jsonapi/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

// TestFormatQuery tests the FormatQuery method
func TestFormatQuery(t *testing.T) {

	t.Run("Paged", func(t *testing.T) {
		p := NewPaged(10, 2)

		require.NoError(t, p.Check())

		q := url.Values{}

		p.FormatQuery(q)

		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(internal.QueryParamPageNumber))
		assert.Equal(t, "2", q.Get(internal.QueryParamPageSize))
	})

	t.Run("Limited", func(t *testing.T) {
		p := NewLimitOffset(10, 0)

		require.NoError(t, p.Check())

		q := p.FormatQuery()

		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(internal.QueryParamPageLimit))
	})

	t.Run("Offseted", func(t *testing.T) {
		p := NewLimitOffset(0, 10)

		require.NoError(t, p.Check())

		q := p.FormatQuery()

		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(internal.QueryParamPageOffset))
	})

	t.Run("LimitOffset", func(t *testing.T) {
		p := NewLimitOffset(10, 140)

		require.NoError(t, p.Check())

		q := p.FormatQuery()

		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(internal.QueryParamPageLimit))
		assert.Equal(t, "140", q.Get(internal.QueryParamPageOffset))
	})

}
