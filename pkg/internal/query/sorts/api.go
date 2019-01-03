package sorts

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
)

var MaxNestedRelLevel int = 1

// NewSortField creates new sortField
func NewSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	sort := &SortField{StructField: sField, order: o}
	sort.subFields = append(sort.subFields, subs...)

	return sort
}
