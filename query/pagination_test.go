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

// TestPaginationNext tests the pagination next function.
func TestPaginationNext(t *testing.T) {
	t.Run("InvalidTotal", func(t *testing.T) {
		p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 0}
		_, err := p.Next(-1)
		assert.Error(t, err)
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: -1}
		_, err := p.Next(100)
		assert.Error(t, err)
	})

	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 10}
			next, err := p.Next(100)
			require.NoError(t, err)

			assert.Equal(t, int64(10), next.Size)
			assert.Equal(t, int64(20), next.Offset)
		})

		t.Run("Last", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 90}
			next, err := p.Next(100)
			require.NoError(t, err)

			// the pagination should not change
			assert.Equal(t, p, next)
		})

		t.Run("PartialLast", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 5, Offset: 22}
			next, err := p.Next(30)
			require.NoError(t, err)

			assert.Equal(t, int64(22+5), next.Offset)
			assert.Equal(t, int64(5), next.Size)
		})
	})

	t.Run("PageSizeNumber", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 3}
			next, err := p.Next(50)
			require.NoError(t, err)

			assert.NotEqual(t, p, next)

			if assert.Equal(t, PageNumberPagination, next.Type) {
				number, size := next.GetNumberSize()
				assert.Equal(t, int64(10), size)
				assert.Equal(t, int64(4), number)
			}
		})

		t.Run("Last", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 4}
			next, err := p.Next(50)
			require.NoError(t, err)

			assert.Equal(t, p, next)
		})

		t.Run("Partial", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 4}
			next, err := p.Next(48)
			require.NoError(t, err)

			assert.Equal(t, p, next)
		})
	})
}

// TestPaginationPrevious tests the pagination 'Previous' function.
func TestPaginationPrevious(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple#01", func(t *testing.T) {
			p := Pagination{Type: LimitOffsetPagination, Offset: 20, Size: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			limit, offset := prev.GetLimitOffset()
			assert.Equal(t, int64(10), limit)
			assert.Equal(t, int64(10), offset)
		})

		t.Run("Simple#02", func(t *testing.T) {
			p := Pagination{Type: LimitOffsetPagination, Offset: 10, Size: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			limit, offset := prev.GetLimitOffset()
			assert.Equal(t, int64(10), limit)
			assert.Equal(t, int64(0), offset)
		})

		t.Run("CurrentFirst", func(t *testing.T) {
			p := Pagination{Type: LimitOffsetPagination, Offset: 0, Size: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, &p, prev)
		})

		t.Run("Partial", func(t *testing.T) {
			p := Pagination{Type: LimitOffsetPagination, Offset: 5, Size: 10}
			prev, err := p.Previous()
			require.NoError(t, err)

			limit, offset := prev.GetLimitOffset()
			assert.Equal(t, int64(5), limit)
			assert.Equal(t, int64(0), offset)
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Type: PageNumberPagination, Size: -10, Offset: 1}
		_, err := p.Previous()
		assert.Error(t, err)
	})

	t.Run("PageNumberSize", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 4}
			prev, err := p.Previous()
			require.NoError(t, err)

			number, size := prev.GetNumberSize()
			assert.Equal(t, int64(3), number)
			assert.Equal(t, int64(10), size)
		})

		t.Run("First", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 1}
			prev, err := p.Previous()
			require.NoError(t, err)

			assert.Equal(t, p, prev)
		})
	})
}

// TestPaginationLast tests the pagination 'Last' function.
func TestPaginationLast(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 50}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				limit, offset := last.GetLimitOffset()
				assert.Equal(t, int64(90), offset)
				assert.Equal(t, int64(10), limit)
			}
		})

		t.Run("CurrentLast", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 90}
			last, err := p.Last(100)
			require.NoError(t, err)

			assert.Equal(t, p, last)
		})

		t.Run("Partial#01", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 55}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				limit, offset := last.GetLimitOffset()
				assert.Equal(t, int64(95), offset)
				// page size doesn't change even though there is no more values
				assert.Equal(t, int64(10), limit)
			}
		})
		t.Run("Partial#02", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 57}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				limit, offset := last.GetLimitOffset()
				assert.Equal(t, int64(97), offset)
				// page size doesn't change even though there is no more values
				assert.Equal(t, int64(10), limit)
			}
		})

		t.Run("TotalZero", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 10}
			last, err := p.Last(0)
			require.NoError(t, err)

			limit, offset := last.GetLimitOffset()
			assert.Equal(t, int64(10), limit)
			assert.Equal(t, int64(0), offset)
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Type: LimitOffsetPagination, Size: -10, Offset: 0}
		_, err := p.Last(10)
		assert.Error(t, err)
	})

	t.Run("InvalidTotal", func(t *testing.T) {
		p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 0}
		_, err := p.Last(-1)
		assert.Error(t, err)
	})

	t.Run("PageNumberSize", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 4}
			last, err := p.Last(100)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				number, size := last.GetNumberSize()
				assert.Equal(t, int64(10), number)
				assert.Equal(t, int64(10), size)
			}
		})

		t.Run("CurrentLast", func(t *testing.T) {
			// PageSize: 10, PageNumber: 10 (instances 91-100)
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 10}
			last, err := p.Last(100)
			require.NoError(t, err)

			assert.Equal(t, p, last)
		})

		t.Run("Partial", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 4}
			last, err := p.Last(95)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				number, size := last.GetNumberSize()
				assert.Equal(t, int64(10), number)
				assert.Equal(t, int64(10), size)
			}
		})

		t.Run("First", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 5}
			last, err := p.Last(5)
			require.NoError(t, err)

			if assert.NotEqual(t, p, last) {
				number, size := last.GetNumberSize()
				assert.Equal(t, int64(1), number)
				assert.Equal(t, int64(10), size)
			}
		})
	})
}

// TestPaginationFirst tests the pagination 'First' function.
func TestPaginationFirst(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 200}
			first, err := p.First()
			require.NoError(t, err)

			limit, offset := first.GetLimitOffset()
			assert.Equal(t, int64(10), limit)
			assert.Equal(t, int64(0), offset)
		})

		t.Run("Current", func(t *testing.T) {
			p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: 0}
			first, err := p.First()
			require.NoError(t, err)

			assert.Equal(t, p, first)
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		p := &Pagination{Type: LimitOffsetPagination, Size: 10, Offset: -1}
		_, err := p.First()
		assert.Error(t, err)
	})

	t.Run("PageNumberSize", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 10}
			first, err := p.First()
			require.NoError(t, err)

			number, size := first.GetNumberSize()
			assert.Equal(t, int64(10), size)
			assert.Equal(t, int64(1), number)
		})

		t.Run("Current", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 1}
			first, err := p.First()
			require.NoError(t, err)

			assert.Equal(t, p, first)
		})

		t.Run("Invalid", func(t *testing.T) {
			p := &Pagination{Type: PageNumberPagination, Size: 10, Offset: 0}
			_, err := p.First()
			assert.Error(t, err)
		})
	})
}

// TestPaginationNew tests the pagination creator functions.
func TestPaginationNew(t *testing.T) {
	t.Run("LimitOffset", func(t *testing.T) {
		// test the NewPaginationLimitOffset function.
		p, err := NewPaginationLimitOffset(100, 10)
		require.NoError(t, err)

		limit, offset := p.GetLimitOffset()
		assert.Equal(t, int64(100), limit)
		assert.Equal(t, int64(10), offset)

		t.Run("Invalid", func(t *testing.T) {
			_, err := NewPaginationLimitOffset(100, -10)
			assert.Error(t, err)
		})
	})

	// test the NewPaginationNumberSize function.
	t.Run("PageNumberSize", func(t *testing.T) {
		p, err := NewPaginationNumberSize(10, 100)
		require.NoError(t, err)

		number, size := p.GetNumberSize()
		assert.Equal(t, int64(10), number)
		assert.Equal(t, int64(100), size)

		t.Run("Invalid", func(t *testing.T) {
			// Page number can't be 0.
			_, err := NewPaginationNumberSize(0, 10)
			assert.Error(t, err)
		})
	})
}

// TestPaginationString tests the String function of the pagination.
func TestPaginationString(t *testing.T) {
	p, err := NewPaginationNumberSize(10, 10)
	require.NoError(t, err)

	s := p.String()
	assert.Contains(t, s, "PageNumber: 10")
	assert.Contains(t, s, "PageSize: 10")

	p, err = NewPaginationLimitOffset(10, 10)
	require.NoError(t, err)

	s = p.String()

	assert.Contains(t, s, "Limit: 10")
	assert.Contains(t, s, "Offset: 10")

}
