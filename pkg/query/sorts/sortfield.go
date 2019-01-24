package sorts

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/sorts"
	"github.com/kucjac/jsonapi/pkg/mapping"
)

// SortField is a field that contains sorting information
type SortField sorts.SortField

// StructField returns sortfield's structure
func (s *SortField) StructField() *mapping.StructField {
	sField := (*sorts.SortField)(s).StructField()

	return (*mapping.StructField)(sField)
}

// Order returns sortfield's order
func (s *SortField) Order() Order {
	return Order((*sorts.SortField)(s).Order())
}
