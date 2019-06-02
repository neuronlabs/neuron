package sorts

import (
	"fmt"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"strings"
)

var (
	MaxNestedRelLevel int = 1
)

type SortError struct {
	FieldName string
	Err       string
}

func (s *SortError) Error() string {
	return fmt.Sprintf("Field not found: '%v'. %v", s.FieldName, s.Err)
}

// NewSortField creates new sortField
func NewSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	return newSortField(sField, o, subs...)
}

func newSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	sort := &SortField{structField: sField, order: o}
	sort.subFields = append(sort.subFields, subs...)

	return sort
}

// NewRawSortField returns raw sortfield
func NewRawSortField(m *models.ModelStruct, sort string, disallowFK bool) (*SortField, error) {
	var (
		sField    *models.StructField
		sortField *SortField
		order     Order
		ok        bool
	)

	// Get the order of sort
	if sort[0] == '-' {
		order = DescendingOrder
		sort = sort[1:]
	} else {
		order = AscendingOrder
	}

	splitted := strings.Split(sort, internal.AnnotationNestedSeperator)
	l := len(splitted)

	switch {
	// for length == 1 the sort must be an attribute or a primary field
	case l == 1:
		if sort == internal.AnnotationID {
			sField = m.PrimaryField()
		} else {
			sField, ok = models.StructAttr(m, sort)
			if !ok {

				return nil, &SortError{FieldName: sort}
			}
		}

		// create sortfield
		sortField = newSortField(sField, order)
	case l <= (MaxNestedRelLevel + 1):

		// Get Relationship
		sField, ok = models.StructRelField(m, splitted[0])
		if !ok {
			return nil, &SortError{FieldName: sort, Err: "Relationship not found."}
		}

		sortField = newSortField(sField, AscendingOrder)

		err := sortField.setSubfield(splitted[1:], order, disallowFK)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection()))
	}

	return sortField, nil
}
