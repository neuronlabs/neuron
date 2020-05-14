package query

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// ParamSort is the url query parameter name for the sorting fields.
const ParamSort = "sort"

// SortField is a field that contains sorting information.
type SortField struct {
	StructField *mapping.StructField
	// Order defines if the sorting order (ascending or descending)
	Order SortOrder
	// SubFields are the relationship sub field sorts
	SubFields []*SortField
}

// Copy creates a copy of the sort field.
func (s *SortField) Copy() *SortField {
	return s.copy()
}

func (s *SortField) String() string {
	var v string
	if s.Order == DescendingOrder {
		v = "-"
	}
	v += s.StructField.NeuronName()
	return v
}

// FormatQuery returns the sort field formatted for url query.
// If the optional argument 'q' is provided the format would be set into the provided url.Models.
// Otherwise it creates new url.Models instance.
// Returns modified url.Models
func (s *SortField) FormatQuery(q ...url.Values) url.Values {
	var query url.Values
	if len(q) > 0 {
		query = q[0]
	}

	if query == nil {
		query = url.Values{}
	}

	var sign string
	if s.Order == DescendingOrder {
		sign = "-"
	}

	var v string
	if values, ok := query[ParamSort]; ok {
		if len(values) > 0 {
			v = values[0]
		}

		if len(v) > 0 {
			v += ","
		}
	}
	v += fmt.Sprintf("%s%s", sign, s.StructField.NeuronName())
	query.Set(ParamSort, v)

	return query
}

// SortOrder is the enum used as the sorting values order.
type SortOrder int

const (
	// AscendingOrder defines the sorting ascending order.
	AscendingOrder SortOrder = iota
	// DescendingOrder defines the sorting descending order.
	DescendingOrder
)

// String implements fmt.Stringer interface.
func (o SortOrder) String() string {
	if o == AscendingOrder {
		return "ascending"
	}
	return "descending"
}

// NewSortFields creates new 'sortFields' for given model 'm'. If the 'disallowFK' is set to true
// the function would not allow to create foreign key sort field.
// The function throws errors on duplicated field values.
func NewSortFields(m *mapping.ModelStruct, disallowFK bool, sortFields ...string) ([]*SortField, error) {
	return newUniqueSortFields(m, disallowFK, sortFields...)
}

// NewSort creates new 'sort' field for given model 'm'. If the 'disallowFK' is set to true
// the function would not allow to create OrderBy field of foreign key field.
func NewSort(m *mapping.ModelStruct, sort string, disallowFK bool, order ...SortOrder) (*SortField, error) {
	// Get the order of sort
	var o SortOrder

	if len(order) > 0 {
		o = order[0]
	} else if sort[0] == '-' {
		o = DescendingOrder
		sort = sort[1:]
	}
	return newStringSortField(m, sort, o, disallowFK)
}

func newUniqueSortFields(m *mapping.ModelStruct, disallowFK bool, sorts ...string) ([]*SortField, error) {
	var (
		err  errors.DetailedError
		errs errors.MultiError
	)
	fields := make(map[string]int)
	// If the number of sort fields is too long then do not allow
	if len(sorts) > m.SortScopeCount() {
		err = errors.NewDet(ClassInvalidSort, "too many sort fields provided for given model")
		err.SetDetailsf("There are too many sort parameters for the '%v' collection.", m.Collection())
		errs = append(errs, err)
		return nil, errs
	}
	var sortFields []*SortField

	for _, sort := range sorts {
		var order SortOrder
		if sort[0] == '-' {
			order = DescendingOrder
			sort = sort[1:]
		}

		// check if no duplicates provided
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			if count == 2 {
				err = errors.NewDet(ClassInvalidSort, "duplicated sort field provided")
				err.SetDetailsf("OrderBy parameter: %v used more than once.", sort)
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

// newStringSortField creates and returns new sort field for given model 'm', with sort field value 'sort'
// and a flag if foreign key should be disallowed - 'disallowFK'.
func newStringSortField(m *mapping.ModelStruct, sort string, order SortOrder, disallowFK bool) (*SortField, errors.DetailedError) {
	var (
		sField    *mapping.StructField
		sortField *SortField
		ok        bool
		err       errors.DetailedError
	)

	split := strings.Split(sort, mapping.AnnotationNestedSeparator)
	l := len(split)
	switch {
	case l == 1:
		// for length == 1 the sort must be an attribute, primary or a foreign key field
		if sort == mapping.AnnotationID {
			sField = m.Primary()

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
			err = errors.NewDetf(ClassInvalidSort, "sort field: '%s' not found", sort)
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}

		// check foreign key
		sField, ok = m.ForeignKey(sort)
		if !ok {
			// field not found for the model.
			err = errors.NewDetf(ClassInvalidSort, "sort field: '%s' not found", sort)
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}
		sortField = newSortField(sField, order)
		return sortField, nil
	case l <= 2:
		// for split length greater than 1 it must be a relationship
		sField, ok = m.RelationByName(split[0])
		if !ok {
			err = errors.NewDet(ClassInvalidSort, "sort field not found")
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sort, m.Collection())
			return nil, err
		}

		sortField = newSortField(sField, order)
		if err = sortField.setSubfield(split[1:], order, disallowFK); err != nil {
			return nil, err
		}
		return sortField, nil
	default:
		err = errors.NewDet(ClassInvalidSort, "sort field nested level too deep")
		err.SetDetailsf("OrderBy: field '%s' nested level is too deep: '%d'", sort, l)
		return nil, err
	}
}

func (s *SortField) copy() *SortField {
	sort := &SortField{StructField: s.StructField, Order: s.Order}
	if len(s.SubFields) != 0 {
		sort.SubFields = make([]*SortField, len(s.SubFields))
		for i, v := range s.SubFields {
			sort.SubFields[i] = v.copy()
		}
	}
	return sort
}

func (s *SortField) setSubfield(sortSplit []string, order SortOrder, disallowFK bool) errors.DetailedError {
	var (
		subField *SortField
		sField   *mapping.StructField
	)

	// Subfields are available only for the relationships
	if !s.StructField.IsRelationship() {
		err := errors.NewDet(ClassInvalidSort, "given sub sort field is not a relationship")
		err.SetDetailsf("OrderBy: field '%s' is not a relationship in the model: '%s'", s.StructField.NeuronName(), s.StructField.Struct().Collection())
		return err
	}

	switch len(sortSplit) {
	case 0:
		log.Debug2("No sort field found")
		return errors.NewDet(ClassInternal, "setting sub sort field failed with 0 length")
	case 1:
		// if len is equal to one then it should be primary or attribute field
		relatedModel := s.StructField.Relationship().Struct()
		sort := sortSplit[0]

		if sort == mapping.AnnotationID {
			sField = relatedModel.Primary()

			s.SubFields = append(s.SubFields, &SortField{StructField: sField, Order: order})
			return nil
		}

		var ok bool
		// check if the 'sort' is an attribute
		sField, ok = relatedModel.Attribute(sort)
		if ok {
			s.SubFields = append(s.SubFields, &SortField{StructField: sField, Order: order})
			return nil
		}

		if disallowFK {
			// if the 'sort' is not an attribute nor primary key and the foreign keys are not allowed to sort
			// return error.
			err := errors.NewDet(ClassInvalidSort, "sort field not found")
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
			return err
		}

		// if the foreign key sorting is allowed check if given foreign key exists
		sField, ok = relatedModel.ForeignKey(sort)
		if !ok {
			// no 'sort' field found.
			err := errors.NewDet(ClassInvalidSort, "sort field not found")
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
			return err
		}
		s.SubFields = append(s.SubFields, &SortField{StructField: sField, Order: order})
		return nil
	default:
		// if length is more than one -> there is a relationship
		relatedModel := s.StructField.Relationship().Struct()
		var ok bool

		log.Debug2f("More sort fields: '%v'", sortSplit)

		sField, ok = relatedModel.RelationByName(sortSplit[0])
		if !ok {
			err := errors.NewDet(ClassInvalidSort, "sort field not found")
			err.SetDetailsf("OrderBy: field '%s' not found in the model: '%s'", sortSplit[0], relatedModel.Collection())
			return err
		}

		// search for the subfields if already created
		for i := range s.SubFields {
			if s.SubFields[i].StructField == sField {
				subField = s.SubFields[i]
				break
			}
		}

		// if none found create new subfield.
		if subField == nil {
			subField = &SortField{StructField: sField, Order: order}
		}

		// set the subfield of the field's subfield.
		if err := subField.setSubfield(sortSplit[1:], order, disallowFK); err != nil {
			return err
		}

		// if subfield found keep it in subfields.
		s.SubFields = append(s.SubFields, subField)
		return nil
	}
}

//nolint:unparam
func newSortField(sField *mapping.StructField, o SortOrder, subs ...*SortField) *SortField {
	sort := &SortField{StructField: sField, Order: o}
	sort.SubFields = append(sort.SubFields, subs...)
	return sort
}
