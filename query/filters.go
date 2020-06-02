package query

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// Filters is the wrapper over the slice of filter fields.
type Filters []*FilterField

// String implements fmt.Stringer interface.
func (f Filters) String() string {
	sb := &strings.Builder{}
	var filtersAdded int
	for _, ff := range f {
		ff.buildString(sb, &filtersAdded)
	}
	return sb.String()
}

// FilterField is a struct that keeps information about given query filters.
// It is based on the mapping.StructField.
type FilterField struct {
	StructField *mapping.StructField
	// Models are the filter values for given attribute Filter
	Values []OperatorValues
	// Nested are the relationship fields filters.
	Nested []*FilterField
}

// Copy returns the copy of the filter field.
func (f *FilterField) Copy() *FilterField {
	return f.copy()
}

// FormatQuery formats the filter field into url.Models.
// If the 'q' optional parameter is set, then the function would add
// the values into the provided argument 'q' url.Models. Otherwise it
// creates new url.Models.
// Returns updated (new) url.Models.
func (f *FilterField) FormatQuery(q ...url.Values) url.Values {
	var query url.Values
	if len(q) != 0 {
		query = q[0]
	}

	if query == nil {
		query = url.Values{}
	}

	temp := url.Values{}
	f.formatQuery(temp)

	collection := f.StructField.ModelStruct().Collection()

	for k, val := range temp {
		if k[0] == '[' {
			k = fmt.Sprintf("filter[%s]%s", collection, k)
		}
		query.Add(k, strings.Join(val, mapping.AnnotationSeparator))
	}
	return query
}

// String implements fmt.Stringer interface.
func (f *FilterField) String() string {
	sb := &strings.Builder{}
	var filtersAdded int
	f.buildString(sb, &filtersAdded)
	return sb.String()
}

func (f *FilterField) buildString(sb *strings.Builder, filtersAdded *int, relName ...string) {
	for i := range f.Values {
		if *filtersAdded != 0 {
			sb.WriteRune('&')
		}

		if len(relName) > 0 {
			sb.WriteString(fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField.NeuronName(), f.Values[i].Operator.URLAlias))
		} else {
			sb.WriteString(fmt.Sprintf("[%s][%s]", f.StructField.NeuronName(), f.Values[i].Operator.URLAlias))
		}

		var vals []string
		if len(f.Values[i].Values) > 0 {
			sb.WriteRune('=')
		}
		for _, val := range f.Values[i].Values {
			mapping.StringValues(val, &vals)
		}

		for i, v := range vals {
			sb.WriteString(v)
			if i != len(vals)-1 {
				sb.WriteRune(',')
			}
		}
		*filtersAdded++
	}
	for _, nested := range f.Nested {
		nested.buildString(sb, filtersAdded, f.StructField.NeuronName())
	}
}

func (f *FilterField) copy() *FilterField {
	cp := &FilterField{StructField: f.StructField}
	if len(f.Values) > 0 {
		cp.Values = make([]OperatorValues, len(f.Values))
		for i := range f.Values {
			cp.Values[i] = f.Values[i].copy()
		}
	}
	if len(f.Nested) > 0 {
		cp.Nested = make([]*FilterField, len(f.Nested))
		for i, nested := range f.Nested {
			cp.Nested[i] = nested.copy()
		}
	}
	return cp
}

// formatQuery parses the into url.Models.
func (f *FilterField) formatQuery(q url.Values, relName ...string) {
	// parse the internal value
	for i := range f.Values {
		var fk string
		switch {
		case len(relName) > 0:
			fk = fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField.NeuronName(), f.Values[i].Operator.URLAlias)
		case f.StructField.IsLanguage():
			fk = ParamLanguage
		default:
			fk = fmt.Sprintf("[%s][%s]", f.StructField.NeuronName(), f.Values[i].Operator.URLAlias)
		}

		// [fieldName][operator]
		var vals []string
		for _, val := range f.Values[i].Values {
			mapping.StringValues(val, &vals)
		}

		for _, v := range vals {
			q.Add(fk, v)
		}
	}
	for _, nested := range f.Nested {
		nested.formatQuery(q, f.StructField.NeuronName())
	}
}

// NewFilterField creates new filterField for given 'field', operator and 'values'.
func NewFilterField(field *mapping.StructField, op *Operator, values ...interface{}) *FilterField {
	return &FilterField{StructField: field, Values: []OperatorValues{{values, op}}}
}

// NewFilter creates new filterField for the default controller, 'model', 'filter' query and 'values'.
// The 'filter' should be of form:
// 	- Field Operator 					'ID IN', 'Name CONTAINS', 'id in', 'name contains'
//	- Relationship.Field Operator		'Car.UserID IN', 'Car.Doors ==', 'car.user_id >=",
// The field might be a Golang model field name or the neuron name.
func NewFilter(model *mapping.ModelStruct, filter string, values ...interface{}) (*FilterField, error) {
	return newFilter(model, filter, values...)
}

// newRelationshipFilter creates new relationship filter for the 'relation' StructField.
// It adds all the nested relation sub filters 'relFilters'.
func newRelationshipFilter(relation *mapping.StructField, relFilters ...*FilterField) *FilterField {
	return &FilterField{StructField: relation, Nested: relFilters}
}

func newFilter(mStruct *mapping.ModelStruct, filter string, values ...interface{}) (*FilterField, error) {
	field, op, err := filterSplitOperator(filter)
	if err != nil {
		return nil, err
	}
	return newModelFilter(mStruct, field, op, values...)
}

func newModelFilter(m *mapping.ModelStruct, field string, op *Operator, values ...interface{}) (*FilterField, error) {
	// check if the field is created for the
	dotIndex := strings.IndexRune(field, '.')
	if dotIndex != -1 {
		// the filter must be of relationship type
		relation, relationField := field[:dotIndex], field[dotIndex+1:]
		sField, ok := m.RelationByName(relation)
		if !ok {
			return nil, errors.NewDetf(ClassFilterField, "provided unknown field: '%s'", field)
		}
		subFilter, err := newModelFilter(sField.Relationship().Struct(), relationField, op, values...)
		if err != nil {
			return nil, err
		}
		return &FilterField{
			StructField: sField,
			Nested:      []*FilterField{subFilter},
		}, nil
	}
	sField, ok := m.FieldByName(field)
	if !ok {
		return nil, errors.NewDetf(ClassFilterField, "provided unknown field: '%s'", field)
	}
	return &FilterField{
		StructField: sField,
		Values:      []OperatorValues{{Operator: op, Values: values}},
	}, nil
}

func filterSplitOperator(filter string) (string, *Operator, error) {
	// divide the query into field and operator
	filter = strings.TrimSpace(filter)
	spaceIndex := strings.IndexRune(filter, ' ')
	if spaceIndex == -1 {
		return "", nil, errors.NewDetf(ClassFilterFormat, "provided invalid filter format: '%s'", filter)
	}
	field, operator := filter[:spaceIndex], filter[spaceIndex+1:]
	op, ok := FilterOperators.Get(strings.ToLower(operator))
	if !ok {
		return "", nil, errors.NewDetf(ClassFilterFormat, "provided unsupported operator: '%s'", operator)
	}
	return field, op, nil
}
