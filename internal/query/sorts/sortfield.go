package sorts

import (
	"strings"

	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// MaxNestedRelLevel is a temporary maximum nested check while creating sort fields
// TODO: change the variable into config settable.
var MaxNestedRelLevel = 1

// SortField is a field that describes the sorting rules for given query.
type SortField struct {
	structField *models.StructField

	// Order defines if the sorting order (ascending or descending)
	order Order

	// SubFields is the relationship sub field sorts
	subFields []*SortField
}

// Order gets the sort field's order.
func (s *SortField) Order() Order {
	return s.order
}

// StructField returns sortField's model.StructField.
func (s *SortField) StructField() *models.StructField {
	return s.structField
}

// Copy copies provided SortField.
func Copy(s *SortField) *SortField {
	return s.copy()
}

// NewSortField creates new SortField with given models.StructField 'sField',
// order 'o' and sub sort fields: 'subs'.
func NewSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	return newSortField(sField, o, subs...)
}

// NewRawSortField creates and returns new sort field for given model 'm', with sort field value 'sort'
// and a flag if foreign key should be disallowed - 'disallowFK'
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

	splitted := strings.Split(sort, common.AnnotationNestedSeparator)
	l := len(splitted)

	switch {
	// for length == 1 the sort must be an attribute or a primary field
	case l == 1:
		if sort == common.AnnotationID {
			sField = m.PrimaryField()
		} else {
			sField, ok = m.Attribute(sort)
			if !ok {
				if disallowFK {
					return nil, errors.New(class.QuerySortField, "sort field not found")
				}
				sField, ok = m.ForeignKey(sort)
				if !ok {
					return nil, errors.New(class.QuerySortField, "sort field not found").SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection())
				}
			}
		}

		// create sortfield
		sortField = newSortField(sField, order)
	case l <= (MaxNestedRelLevel + 1):
		// Get Relationship
		sField, ok = m.RelationshipField(splitted[0])
		if !ok {
			return nil, errors.New(class.QuerySortRelatedFields, "relationship field not found")
		}

		sortField = newSortField(sField, AscendingOrder)

		err := sortField.setSubfield(splitted[1:], order, disallowFK)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(class.QuerySortField, "sort field not found").SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection())
	}

	return sortField, nil
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
func (s *SortField) SetSubfield(sortSplitted []string, order Order, disallowFK bool) *errors.Error {
	return s.setSubfield(sortSplitted, order, disallowFK)
}

// setSubfield sets sortfield for subfield of given relationship field.
func (s *SortField) setSubfield(sortSplitted []string, order Order, disallowFK bool) *errors.Error {
	var (
		subField *SortField
		sField   *models.StructField
	)

	// Subfields are available only for the relationships
	if !s.structField.IsRelationship() {
		err := errors.New(class.QuerySortRelatedFields, "given sub sortfield is not a relationship")
		err.SetDetailf("Sort: field '%s' is not a relationship in the model: '%s'", s.structField.NeuronName(), s.structField.Struct().Collection())
		return err
	}

	// sort splitted is splitted sort query entry
	// i.e. a sort query for
	switch len(sortSplitted) {
	case 0:
		log.Debug2("No sort field found")
		return errors.New(class.InternalQuerySort, "setting sub sortfield failed with 0 length")
	case 1:
		// if len is equal to one then it should be primary or attribute field
		relatedModel := s.structField.Relationship().Struct()
		sort := sortSplitted[0]

		if sort == common.AnnotationID {
			log.Debug2("Primary sort field")
			sField = relatedModel.PrimaryField()
		} else {
			log.Debug2f("One sort field: '%s'", sort)
			var ok bool
			sField, ok = relatedModel.Attribute(sort)
			if !ok {
				if disallowFK {
					return errors.New(class.QuerySortField, "sort field not found").SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
				}

				// if the foreign key sorting is allowed check if given foreign key exists
				sField, ok = relatedModel.ForeignKey(sort)
				if !ok {
					return errors.New(class.QuerySortField, "sort field not found").SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
				}
			}
		}

		s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
	default:
		// if length is more than one -> there is a relationship
		relatedModel := s.structField.Relationship().Struct()
		var ok bool

		log.Debug2f("More sort fields: '%v'", sortSplitted)

		sField, ok = relatedModel.RelationshipField(sortSplitted[0])
		if !ok {
			return errors.New(class.QuerySortField, "sort field not found").SetDetailf("Sort: field '%s' not found in the model: '%s'", sortSplitted[0], relatedModel.Collection())
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
		if err := subField.setSubfield(sortSplitted[1:], order, disallowFK); err != nil {
			return err
		}

		// if subfield found keep it in subfields.
		s.subFields = append(s.subFields, subField)
	}
	return nil
}

func newSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	sort := &SortField{structField: sField, order: o}
	sort.subFields = append(sort.subFields, subs...)

	return sort
}
