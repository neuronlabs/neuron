package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaginationFormatQuery tests the Pagination.FormatQuery method.
func TestPaginationFormatQuery(t *testing.T) {
	t.Run("Limited", func(t *testing.T) {
		p := &Pagination{Limit: 10}

		require.NoError(t, p.Validate())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
	})

	t.Run("Offseted", func(t *testing.T) {
		p := &Pagination{Offset: 10}

		require.NoError(t, p.Validate())

		q := p.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "10", q.Get(ParamPageOffset))
	})

	t.Run("LimitOffset", func(t *testing.T) {
		p := &Pagination{Limit: 10, Offset: 140}

		require.NoError(t, p.Validate())

		q := p.FormatQuery()
		require.Len(t, q, 2)

		assert.Equal(t, "10", q.Get(ParamPageLimit))
		assert.Equal(t, "140", q.Get(ParamPageOffset))
	})
}

// TestPaginationNext tests the pagination next function.
func TestPaginationNext(t *testing.T) {
	t.Run("InvalidTotal", func(t *testing.T) {
		p := &Pagination{Limit: 10, Offset: 0}
		_, err := p.Next(-1)
		assert.Error(t, err)
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Limit: 10, Offset: -1}
		_, err := p.Next(100)
		assert.Error(t, err)
	})

	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 10}
			next, err := p.Next(100)
			require.NoError(t, err)

			assert.Equal(t, int64(10), next.Limit)
			assert.Equal(t, int64(20), next.Offset)
		})

		t.Run("Last", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 90}
			next, err := p.Next(100)
			require.NoError(t, err)

			// the pagination should not change
			assert.Equal(t, p, next)
		})

		t.Run("PartialLast", func(t *testing.T) {
			p := &Pagination{Limit: 5, Offset: 22}
			next, err := p.Next(30)
			require.NoError(t, err)

			assert.Equal(t, int64(22+5), next.Offset)
			assert.Equal(t, int64(5), next.Limit)
		})
	})
}

// TestPaginationPrevious tests the pagination 'Previous' function.
func TestPaginationPrevious(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple#01", func(t *testing.T) {
			p := Pagination{Offset: 20, Limit: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, int64(10), prev.Limit)
			assert.Equal(t, int64(10), prev.Offset)
		})

		t.Run("Simple#02", func(t *testing.T) {
			p := Pagination{Offset: 10, Limit: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, int64(10), prev.Limit)
			assert.Equal(t, int64(0), prev.Offset)
		})

		t.Run("CurrentFirst", func(t *testing.T) {
			p := Pagination{Offset: 0, Limit: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, &p, prev)
		})

		t.Run("Partial", func(t *testing.T) {
			p := Pagination{Offset: 5, Limit: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, int64(5), prev.Limit)
			assert.Equal(t, int64(0), prev.Offset)
		})
	})
}

// TestPaginationLast tests the pagination 'Last' function.
func TestPaginationLast(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 50}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				assert.Equal(t, int64(90), last.Offset)
				assert.Equal(t, int64(10), last.Limit)
			}
		})

		t.Run("CurrentLast", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 90}
			last, err := p.Last(100)
			require.NoError(t, err)

			assert.Equal(t, p, last)
		})

		t.Run("Partial#01", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 55}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				assert.Equal(t, int64(95), last.Offset)
				// page size doesn't change even though there is no more values
				assert.Equal(t, int64(10), last.Limit)
			}
		})
		t.Run("Partial#02", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 57}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				assert.Equal(t, int64(97), last.Offset)
				// page size doesn't change even though there is no more values
				assert.Equal(t, int64(10), last.Limit)
			}
		})

		t.Run("TotalZero", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 10}
			last, err := p.Last(0)
			require.NoError(t, err)

			assert.Equal(t, int64(10), last.Limit)
			assert.Equal(t, int64(0), last.Offset)
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Limit: -10, Offset: 0}
		_, err := p.Last(10)
		assert.Error(t, err)
	})

	t.Run("InvalidTotal", func(t *testing.T) {
		p := &Pagination{Limit: 10, Offset: 0}
		_, err := p.Last(-1)
		assert.Error(t, err)
	})
}

// TestPaginationFirst tests the pagination 'First' function.
func TestPaginationFirst(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 200}
			first, err := p.First()
			require.NoError(t, err)

			assert.Equal(t, int64(10), first.Limit)
			assert.Equal(t, int64(0), first.Offset)
		})

		t.Run("Current", func(t *testing.T) {
			p := &Pagination{Limit: 10, Offset: 0}
			first, err := p.First()
			require.NoError(t, err)

			assert.Equal(t, p, first)
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Limit: 10, Offset: -1}
		_, err := p.First()
		assert.Error(t, err)
	})
}

// TestPaginationString tests the String function of the pagination.
func TestPaginationString(t *testing.T) {
	p := &Pagination{Limit: 10, Offset: 10}
	s := p.String()
	assert.Contains(t, s, "Limit: 10")
	assert.Contains(t, s, "Offset: 10")
}
