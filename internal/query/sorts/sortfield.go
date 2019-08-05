package sorts

import (
	"strings"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
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

// SubFields returns sortFields nested sort fields.
func (s *SortField) SubFields() []*SortField {
	return s.subFields
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

// NewUniques creates new unique sort fiellds for provided model 'm'.
func NewUniques(m *models.ModelStruct, disallowFK bool, sorts ...string) ([]*SortField, error) {
	var (
		err  errors.DetailedError
		errs errors.MultiError
	)

	fields := make(map[string]int)
	// If the number of sort fields is too long then do not allow
	if len(sorts) > m.SortScopeCount() {
		err = errors.NewDet(class.QuerySortTooManyFields, "too many sort fields provided for given model")
		err.SetDetailsf("There are too many sort parameters for the '%v' collection.", m.Collection())
		errs = append(errs, err)
		return nil, errs
	}

	var sortFields []*SortField

	for _, sort := range sorts {
		var order Order
		if sort[0] == '-' {
			order = DescendingOrder
			sort = sort[1:]
		}

		// check if no dups provided
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			if count == 2 {
				err = errors.NewDet(class.QuerySortField, "duplicated sort field provided")
				err.SetDetailsf("Sort parameter: %v used more than once.", sort)
				errs = append(errs, err)
				continue
			} else if count > 2 {
				break
			}
		}

		sortField, err := newStringSortField(m, sort, order, disallowFK)
		if err != nil {
			errs = append(errs, err.(errors.ClassError))
			continue
		}

		sortFields = append(sortFields, sortField)
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return sortFields, nil
}

// New creates new SortField based on the provided model 'm', string 'sort', flag 'disallowFK' - which doesn't allow to create foreign keys,
// and optional order.
func New(m *models.ModelStruct, sort string, disallowFK bool, order ...Order) (*SortField, error) {
	// Get the order of sort
	var o Order
	if len(order) > 0 {
		o = order[0]
	} else {
		if sort[0] == '-' {
			o = DescendingOrder
			sort = sort[1:]
		}
	}
	return newStringSortField(m, sort, o, disallowFK)
}

// newStringSortField creates and returns new sort field for given model 'm', with sort field value 'sort'
// and a flag if foreign key should be disallowed - 'disallowFK'.
func newStringSortField(m *models.ModelStruct, sort string, order Order, disallowFK bool) (*SortField, errors.DetailedError) {
	var (
		sField    *models.StructField
		sortField *SortField
		ok        bool
		err       errors.DetailedError
	)

	splitted := strings.Split(sort, annotation.NestedSeparator)
	l := len(splitted)
	switch {
	case l == 1:
		// for length == 1 the sort must be an attribute, primary or a foreign key field
		if sort == annotation.ID {
			sField = m.PrimaryField()
			sortField = newSortField(sField, order)
			return sortField, nil
		}

		// check attributes
		sField, ok = m.Attribute(sort)
		if ok {
			sortField = newSortField(sField, order)
			return sortField, nil
		}

		if disallowFK {
			// field not found for the model.
			err = errors.NewDetf(class.QuerySortField, "sort field: '%s' not found", sort)
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}

		// check foreign key
		sField, ok = m.ForeignKey(sort)
		if !ok {
			// field not found for the model.
			err = errors.NewDetf(class.QuerySortField, "sort field: '%s' not found", sort)
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}
		sortField = newSortField(sField, order)
		return sortField, nil
	case l <= (MaxNestedRelLevel + 1):
		// for splitted length greater than 1 it must be a relationship
		sField, ok = m.RelationshipField(splitted[0])
		if !ok {
			err = errors.NewDet(class.QuerySortField, "sort field not found")
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}

		sortField = newSortField(sField, order)
		err := sortField.SetSubfield(splitted[1:], order, disallowFK)
		if err != nil {
			return nil, err
		}
		return sortField, nil
	default:
		err = errors.NewDet(class.QuerySortField, "sort field nested level too deep")
		err.SetDetailsf("Sort: field '%s' nested level is too deep: '%d'", sort, l)
		return nil, err
	}
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

// SetSubfield sets the subfield for given sortfield.
func (s *SortField) SetSubfield(sortSplitted []string, order Order, disallowFK bool) errors.DetailedError {
	return s.setSubfield(sortSplitted, order, disallowFK)
}

// setSubfield sets sortfield for subfield of given relationship field.
func (s *SortField) setSubfield(sortSplitted []string, order Order, disallowFK bool) errors.DetailedError {
	var (
		subField *SortField
		sField   *models.StructField
	)

	// Subfields are available only for the relationships
	if !s.structField.IsRelationship() {
		err := errors.NewDet(class.QuerySortRelatedFields, "given sub sortfield is not a relationship")
		err.SetDetailsf("Sort: field '%s' is not a relationship in the model: '%s'", s.structField.NeuronName(), s.structField.Struct().Collection())
		return err
	}

	// sort splitted is splitted sort query entry
	// i.e. a sort query for
	switch len(sortSplitted) {
	case 0:
		log.Debug2("No sort field found")
		return errors.NewDet(class.InternalQuerySort, "setting sub sortfield failed with 0 length")
	case 1:
		// if len is equal to one then it should be primary or attribute field
		relatedModel := s.structField.Relationship().Struct()
		sort := sortSplitted[0]

		if sort == annotation.ID {
			sField = relatedModel.PrimaryField()

			s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
			return nil
		}

		var ok bool
		// check if the 'sort' is an attribute
		sField, ok = relatedModel.Attribute(sort)
		if ok {
			s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
			return nil
		}

		if disallowFK {
			// if the 'sort' is not an attribute nor primary key and the foreign keys are not allowed to sort
			// return error.
			err := errors.NewDet(class.QuerySortField, "sort field not found")
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
			return err
		}

		// if the foreign key sorting is allowed check if given foreign key exists
		sField, ok = relatedModel.ForeignKey(sort)
		if !ok {
			// no 'sort' field found.
			err := errors.NewDet(class.QuerySortField, "sort field not found")
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
			return err
		}
		s.subFields = append(s.subFields, &SortField{structField: sField, order: order})
		return nil
	default:
		// if length is more than one -> there is a relationship
		relatedModel := s.structField.Relationship().Struct()
		var ok bool

		log.Debug2f("More sort fields: '%v'", sortSplitted)

		sField, ok = relatedModel.RelationshipField(sortSplitted[0])
		if !ok {
			err := errors.NewDet(class.QuerySortField, "sort field not found")
			err.SetDetailsf("Sort: field '%s' not found in the model: '%s'", sortSplitted[0], relatedModel.Collection())
			return err
		}

		// search for the subfields if already created
		for i := range s.subFields {
			if s.subFields[i].structField == sField {
				subField = s.subFields[i]
				break
			}
		}

		// if none found create new subfield.
		if subField == nil {
			subField = &SortField{structField: sField, order: order}
		}

		// set the subfield of the field's subfield.
		if err := subField.setSubfield(sortSplitted[1:], order, disallowFK); err != nil {
			return err
		}

		// if subfield found keep it in subfields.
		s.subFields = append(s.subFields, subField)
		return nil
	}
}

func newSortField(sField *models.StructField, o Order, subs ...*SortField) *SortField {
	sort := &SortField{structField: sField, order: o}
	sort.subFields = append(sort.subFields, subs...)

	return sort
}
