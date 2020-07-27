package query

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// Sort is an interface used by the queries.
type Sort interface {
	Copy() Sort
	Order() SortOrder
	Field() *mapping.StructField
}

// ParamSort is the url query parameter name for the sorting fields.
const ParamSort = "sort"

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
	if s.SortOrder == DescendingOrder {
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
func NewSortFields(m *mapping.ModelStruct, sortFields ...string) ([]Sort, error) {
	return newUniqueSortFields(m, sortFields...)
}

// NewSort creates new 'sort' field for given model 'm'. If the 'disallowFK' is set to true
// the function would not allow to create OrderBy field of foreign key field.
func NewSort(m *mapping.ModelStruct, sort string, order ...SortOrder) (Sort, error) {
	// Get the order of sort
	var o SortOrder

	if len(order) > 0 {
		o = order[0]
	} else if sort[0] == '-' {
		o = DescendingOrder
		sort = sort[1:]
	}
	return newStringSortField(m, sort, o)
}

func newUniqueSortFields(m *mapping.ModelStruct, sorts ...string) ([]Sort, error) {
	fields := make(map[string]int)
	// If the number of sort fields is too long then do not allow
	if len(sorts) > m.SortScopeCount() {
		return nil, errors.NewDet(ClassInvalidSort, "too many sort fields provided for given model").
			WithDetailf("There are too many sort parameters for the '%v' collection.", m.Collection())
	}

	var sortFields []Sort
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
			return nil, errors.NewDet(ClassInvalidSort, "duplicated sort field provided").WithDetailf("OrderBy parameter: %v used more than once.", sort)
		}

		sortField, err := newStringSortField(m, sort, order)
		if err != nil {
			return nil, err
		}
		sortFields = append(sortFields, sortField)
	}
	return sortFields, nil
}

// newStringSortField creates and returns new sort field for given model 'm', with sort field value 'sort'
// and a flag if foreign key should be disallowed - 'disallowFK'.
func newStringSortField(m *mapping.ModelStruct, sort string, order SortOrder) (Sort, error) {
	split := strings.Split(sort, mapping.AnnotationNestedSeparator)
	l := len(split)
	switch {
	case l == 1:
		// for length == 1 the sort must be an attribute, primary or a foreign key field
		if sort == m.Primary().Name() || sort == m.Primary().NeuronName() {
			return SortField{StructField: m.Primary(), SortOrder: order}, nil
		}

		// check attributes
		sField, ok := m.Attribute(sort)
		if ok {
			return SortField{StructField: sField, SortOrder: order}, nil
		}

		// check foreign key
		sField, ok = m.ForeignKey(sort)
		if !ok {
			// field not found for the model.
			return nil, errors.NewDetf(ClassInvalidSort, "sort field: '%s' not found", sort).
				WithDetailf("OrderBy: field '%s' not found in the model: '%s'", sort, m.Collection())
		}
		return SortField{StructField: sField, SortOrder: order}, nil
	case l <= 2:
		// for split length greater than 1 it must be a relationship
		sField, ok := m.RelationByName(split[0])
		if !ok {
			return nil, errors.NewDet(ClassInvalidSort, "sort field not found").
				WithDetailf("OrderBy: field '%s' not found in the model: '%s'", sort, m.Collection())
		}

		subFields, err := getSortRelationSubfield(sField, split[1:])
		if err != nil {
			return nil, err
		}
		return RelationSort{StructField: sField, RelationFields: subFields, SortOrder: order}, nil
	default:
		return nil, errors.NewDet(ClassInvalidSort, "sort field nested level too deep").
			WithDetailf("OrderBy: field '%s' nested level is too deep: '%d'", sort, l)
	}
}

func getSortRelationSubfield(relationField *mapping.StructField, sortSplit []string) ([]*mapping.StructField, error) {
	// Subfields are available only for the relationships
	if !relationField.IsRelationship() {
		return nil, errors.NewDet(ClassInvalidSort, "given sub sort field is not a relationship").
			WithDetailf("OrderBy: field '%s' is not a relationship in the model: '%s'", relationField.NeuronName(), relationField.Struct().Collection())
	}

	switch len(sortSplit) {
	case 0:
		log.Debug2("No sort field found")
		return nil, errors.NewDet(ClassInternal, "setting sub sort field failed with 0 length")
	case 1:
		// if len is equal to one then it should be primary or attribute field
		relatedModel := relationField.Relationship().Struct()
		sort := sortSplit[0]

		if sort == relatedModel.Primary().Name() || sort == relatedModel.Primary().NeuronName() {
			return []*mapping.StructField{relatedModel.Primary()}, nil
		}

		var ok bool
		// check if the 'sort' is an attribute
		relationField, ok = relatedModel.Attribute(sort)
		if ok {
			return []*mapping.StructField{relationField}, nil
		}

		// if the foreign key sorting is allowed check if given foreign key exists
		relationField, ok = relatedModel.ForeignKey(sort)
		if !ok {
			// no 'sort' field found.
			return nil, errors.NewDet(ClassInvalidSort, "sort field not found").
				WithDetailf("OrderBy: field '%s' not found in the model: '%s'", sort, relatedModel.Collection())
		}
		return []*mapping.StructField{relationField}, nil
	default:
		// if length is more than one -> there is a relationship
		relatedModel := relationField.Relationship().Struct()
		log.Debug2f("More sort fields: '%v'", sortSplit)

		sField, ok := relatedModel.RelationByName(sortSplit[0])
		if !ok {
			return nil, errors.NewDet(ClassInvalidSort, "sort field not found").
				WithDetailf("OrderBy: field '%s' not found in the model: '%s'", sortSplit[0], relatedModel.Collection())
		}

		// set the subfield of the field's subfield.
		subFields, err := getSortRelationSubfield(sField, sortSplit[1:])
		if err != nil {
			return nil, err
		}
		// if subfield found keep it in subfields.
		return append([]*mapping.StructField{sField}, subFields...), nil
	}
}
