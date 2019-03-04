package sorts

import (
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/models"
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
func SetSubfield(s *SortField, sortSplitted []string, order Order) (invalidField bool) {
	return s.setSubfield(sortSplitted, order)
}

// setSubfield sets sortfield for subfield of given relationship field.
func (s *SortField) setSubfield(sortSplitted []string, order Order) (invalidField bool) {
	var (
		subField *SortField
		sField   *models.StructField
	)

	// Subfields are available only for the relationships
	if !s.structField.IsRelationship() {
		invalidField = true
		return
	}

	// sort splitted is splitted sort query entry
	// i.e. a sort query for
	switch len(sortSplitted) {
	case 0:
		invalidField = true
		return
	case 1:
		// if len is equal to one then it should be primary or attribute field
		sort := sortSplitted[0]
		if sort == internal.AnnotationID {
			sField = models.StructPrimary(models.FieldsRelatedModelStruct(s.structField))
		} else {
			var ok bool
			sField, ok = models.StructAttr(models.FieldsRelatedModelStruct(s.structField), sortSplitted[0])
			if !ok {
				invalidField = true
				return
			}
		}

		s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
	default:
		// if length is more than one -> there is a relationship
		var ok bool
		sField, ok := models.StructRelField(models.FieldsRelatedModelStruct(s.structField), sortSplitted[0])
		if !ok {
			invalidField = true
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
			subField = &SortField{structField: sField}
		}

		//
		invalidField = subField.setSubfield(sortSplitted[1:], order)
		if !invalidField {
			// if found keep the subfield in subfields
			s.subFields = append(s.subFields, subField)
		}
	}
	return
}
