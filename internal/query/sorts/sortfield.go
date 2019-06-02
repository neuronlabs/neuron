package sorts

import (
	"fmt"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
)

// SortField is a field that describes the sorting rules for given
type SortField struct {
	structField *models.StructField

	// Order defines if the sorting order (ascending or descending)
	order Order

	// SubFields is the relationship sub field sorts
	subFields []*SortField
}

// Order returns sortFields order
func (s *SortField) Order() Order {
	return s.order
}

// StructField returns sortField's structure
func (s *SortField) StructField() *models.StructField {
	return s.structField
}

// Copy copies provided SortField
func Copy(s *SortField) *SortField {
	return s.copy()
}

func (s *SortField) copy() *SortField {
	sort := &SortField{structField: s.structField, order: s.order}
	if len(s.subFields) != 0 {
		sort.subFields = make([]*SortField, len(s.subFields))
		for i, v := range s.subFields {
			sort.subFields[i] = v.copy()
		}
	}
	return sort

}

// SetSubfield sets the subfield for given sortfield
func (s *SortField) SetSubfield(sortSplitted []string, order Order, disallowFK bool) error {
	return s.setSubfield(sortSplitted, order, disallowFK)
}

// setSubfield sets sortfield for subfield of given relationship field.
func (s *SortField) setSubfield(sortSplitted []string, order Order, disallowFK bool) (err error) {
	var (
		subField *SortField
		sField   *models.StructField
	)

	// Subfields are available only for the relationships
	if !s.structField.IsRelationship() {
		err = errors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' is not a relationship in the model: '%s'", s.structField.ApiName(), s.structField.Struct().Collection()))
		return
	}

	// sort splitted is splitted sort query entry
	// i.e. a sort query for
	switch len(sortSplitted) {
	case 0:
		err = errors.ErrInternalError.Copy().WithDetail("Sort: setting sub sortfield failed")
		return
	case 1:
		// if len is equal to one then it should be primary or attribute field

		relatedModel := s.structField.Relationship().Struct()
		sort := sortSplitted[0]

		if sort == internal.AnnotationID {
			sField = relatedModel.PrimaryField()
		} else {
			var ok bool
			sField, ok = relatedModel.Attribute(sort)
			if !ok {
				if disallowFK {
					err = errors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection()))
					return
				}
				sField, ok = relatedModel.ForeignKey(sort)
				if !ok {
					err = errors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection()))
					return
				}
			}
		}

		s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
	default:
		// if length is more than one -> there is a relationship
		relatedModel := s.structField.Relationship().Struct()
		var ok bool

		sField, ok = relatedModel.RelationshipField(sortSplitted[0])
		if !ok {
			err = errors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sortSplitted[0], relatedModel.Collection()))
			return
		}

		// search for the subfields if already created
		for i := range s.subFields {
			if s.subFields[i].structField == sField {
				subField = s.subFields[i]
				break
			}
		}

		// if none found create new
		if subField == nil {
			subField = &SortField{structField: sField, order: order}
		}

		//
		err = subField.setSubfield(sortSplitted[1:], order, disallowFK)
		if err != nil {
			return
		}
		// if found keep the subfield in subfields
		s.subFields = append(s.subFields, subField)

	}
	return
}
