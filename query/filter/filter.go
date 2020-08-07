package filter

import (
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// Filters is the wrapper over the slice of filter fields.
type Filters []Filter

// Filter is the interface used a filter for the queries.
type Filter interface {
	Copy() Filter
	String() string
}

// String implements fmt.Stringer interface.
func (f Filters) String() string {
	sb := &strings.Builder{}
	for i, ff := range f {
		sb.WriteString(ff.String())
		if i != len(f)-1 {
			sb.WriteRune(',')
		}
	}
	return sb.String()
}

// NewFilter creates new filterField for the default controller, 'model', 'filter' query and 'values'.
// The 'filter' should be of form:
// 	- Simple Operator 					'ID IN', 'Name CONTAINS', 'id in', 'name contains'
//	- Relationship.Simple Operator		'Car.UserID IN', 'Car.Doors ==', 'car.user_id >=",
// The field might be a Golang model field name or the neuron name.
func NewFilter(model *mapping.ModelStruct, filter string, values ...interface{}) (Filter, error) {
	field, op, err := filterSplitOperator(filter)
	if err != nil {
		return nil, err
	}
	return newModelFilter(model, field, op, values...)
}

func newModelFilter(m *mapping.ModelStruct, field string, op *Operator, values ...interface{}) (Filter, error) {
	// check if the field is created for the
	dotIndex := strings.IndexRune(field, '.')
	if dotIndex != -1 {
		// the filter must be of relationship type
		relation, relationField := field[:dotIndex], field[dotIndex+1:]
		sField, ok := m.RelationByName(relation)
		if !ok {
			return nil, errors.NewDetf(ClassFilterField, "provided unknown field: '%s'", field)
		}
		subFilter, err := newModelFilter(sField.Relationship().RelatedModelStruct(), relationField, op, values...)
		if err != nil {
			return nil, err
		}
		return Relation{
			StructField: sField,
			Nested:      []Filter{subFilter},
		}, nil
	}
	sField, ok := m.FieldByName(field)
	if !ok {
		return nil, errors.NewDetf(ClassFilterField, "provided unknown field: '%s'", field)
	}
	return Simple{StructField: sField, Operator: op, Values: values}, nil
}

func filterSplitOperator(filter string) (string, *Operator, error) {
	// divide the query into field and operator
	filter = strings.TrimSpace(filter)
	spaceIndex := strings.IndexRune(filter, ' ')
	if spaceIndex == -1 {
		return "", nil, errors.NewDetf(ClassFilterFormat, "provided invalid filter format: '%s'", filter)
	}
	field, operator := filter[:spaceIndex], filter[spaceIndex+1:]
	if spaceIndex = strings.IndexRune(operator, ' '); spaceIndex != -1 {
		operator = operator[:spaceIndex]
	}
	op, ok := Operators.Get(strings.ToLower(operator))
	if !ok {
		return "", nil, errors.NewDetf(ClassFilterFormat, "provided unsupported operator: '%s'", operator)
	}
	return field, op, nil
}
