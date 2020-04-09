package query

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/mapping"
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
	// Key is used for the map type filters as a nested argument.
	Key string
	// Values are the filter values for given attribute FilterField
	Values []*OperatorValues
	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	Nested []*FilterField
}

// Copy returns the copy of the filter field.
func (f *FilterField) Copy() *FilterField {
	return f.copy()
}

// NestedFilters returns the nested filters for given filter fields.
// Nested filters are the filters used for relationship or composite attribute filters.
func (f *FilterField) NestedFilters() []*FilterField {
	return f.Nested
}

// FormatQuery formats the filter field into url.Values.
// If the 'q' optional parameter is set, then the function would add
// the values into the provided argument 'q' url.Values. Otherwise it
// creates new url.Values.
// Returns updated (new) url.Values.
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
		query.Add(k, strings.Join(val, annotation.Separator))
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
	for _, fv := range f.Values {
		if *filtersAdded != 0 {
			sb.WriteRune('&')
		}

		if len(relName) > 0 {
			sb.WriteString(fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField.NeuronName(), fv.Operator.Raw))
		} else {
			sb.WriteString(fmt.Sprintf("[%s][%s]", f.StructField.NeuronName(), fv.Operator.Raw))
		}

		var vals []string
		if len(fv.Values) > 0 {
			sb.WriteRune('=')
		}
		for _, val := range fv.Values {
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
	cp := &FilterField{
		StructField: f.StructField,
		Key:         f.Key,
	}

	for _, ov := range f.Values {
		cp.Values = append(cp.Values, ov.copy())
	}
	for _, nested := range f.Nested {
		cp.Nested = append(cp.Nested, nested.copy())
	}
	return cp
}

// formatQuery parses the into url.Values.
func (f *FilterField) formatQuery(q url.Values, relName ...string) {
	// parse the internal value
	for _, fv := range f.Values {
		var fk string
		switch {
		case len(relName) > 0:
			fk = fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField.NeuronName(), fv.Operator.Raw)
		case f.StructField.IsLanguage():
			fk = ParamLanguage
		default:
			fk = fmt.Sprintf("[%s][%s]", f.StructField.NeuronName(), fv.Operator.Raw)
		}

		// [fieldName][operator]
		var vals []string
		for _, val := range fv.Values {
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

// NewFilter creates new filterfield for given field, operator and values.
func NewFilter(field *mapping.StructField, op *Operator, values ...interface{}) *FilterField {
	return &FilterField{StructField: field, Values: []*OperatorValues{{values, op}}}
}

// NewRelationshipFilter creates new relationship filter for the 'relation' StructField.
// It adds all the nested relation subfilters 'relFilters'.
func NewRelationshipFilter(relation *mapping.StructField, relFilters ...*FilterField) *FilterField {
	return &FilterField{StructField: relation, Nested: relFilters}
}

// NewStringFilter creates the filter field based on the provided 'filter' and 'values'.
// Example:
//	- 'fitler': "filter[collection][fieldName][operator]"
//  - 'values': 5, 13
// This function doesn't allow to filter over foreign keys.
func NewStringFilter(c *controller.Controller, filter string, values ...interface{}) (*FilterField, error) {
	return newStringFilter(c, filter, false, values...)
}

// NewStringFilterWithForeignKey creates the filter field based on the provided filter, schemaName and values.
// Example:
//	- 'fitler': "filter[collection][fieldName][operator]"
//  - 'schema': "schemaName"
//  - 'values': 5, 13
// This function allow to filter over the foreign key fields.
func NewStringFilterWithForeignKey(c *controller.Controller, filter string, values ...interface{}) (*FilterField, error) {
	return newStringFilter(c, filter, true, values...)
}

func newStringFilter(c *controller.Controller, filter string, foreignKeyAllowed bool, values ...interface{}) (*FilterField, error) {
	filter = strings.TrimPrefix(filter, ParamFilter)

	params, err := SplitBracketParameter(filter)
	if err != nil {
		return nil, err
	}

	mStruct := c.ModelMap.GetByCollection(params[0])
	if mStruct == nil {
		err := errors.NewDet(class.QueryFilterUnknownCollection, "provided filter collection not found")
		err.SetDetailsf("Filter model: '%s' not found", params[0])
		return nil, err
	}

	var (
		f     *FilterField
		op    *Operator
		ok    bool
		field *mapping.StructField
	)
	findOperator := func(index int) error {
		op, ok = FilterOperators.Get(params[index])
		if !ok {
			err := errors.NewDet(class.QueryFilterUnknownOperator, "operator not found")
			err.SetDetailsf("Unknown filte operator not found: %s", params[index])
			return err
		}
		return nil
	}
	if len(params) <= 1 || len(params) > 4 {
		return nil, errors.NewDet(class.QueryFilterInvalidFormat, "invalid filter format")
	}

	switch len(params) {
	case 2, 3:
		// [collection][field][op]
		if len(params) == 3 {
			if err := findOperator(2); err != nil {
				return nil, err
			}
		} else {
			op = OpEqual
		}

		if params[1] == mStruct.Primary().NeuronName() {
			field = mStruct.Primary()
		} else if field, ok = mStruct.Attribute(params[1]); !ok {
			if foreignKeyAllowed {
				field, ok = mStruct.ForeignKey(params[1])
			}
			if !ok {
				field, ok = mStruct.FilterKey(params[1])
			}
			if !ok {
				err := errors.NewDet(class.QueryFilterUnknownField, "field not found")
				err.SetDetailsf("Field: '%s' not found for the Model: '%s'", params[1], mStruct.Collection())
				return nil, err
			}
		}
		f = &FilterField{StructField: field, Values: []*OperatorValues{{values, op}}}
	case 4:
		// [collection][relationship][field][op]
		if err = findOperator(3); err != nil {
			return nil, err
		}

		if rel, ok := mStruct.RelationField(params[1]); ok {
			f = &FilterField{StructField: rel}
			relStruct := rel.Relationship().Struct()
			if params[2] == "id" {
				field = relStruct.Primary()
			} else if field, ok = relStruct.Attribute(params[2]); !ok {
				if foreignKeyAllowed {
					field, ok = relStruct.ForeignKey(params[2])
				}
				if !ok {
					field, ok = relStruct.FilterKey(params[2])
				}
				if !ok {
					err := errors.NewDet(class.QueryFilterInvalidField, "field not found")
					err.SetDetailsf("Field: '%s' not found within the relation ModelStruct: '%s'", params[2], relStruct.Collection())
					return nil, err
				}
			}

			nested := NewFilter(field, op, values...)
			f.Nested = append(f.Nested, nested)
		} else if attr, ok := mStruct.Attribute(params[1]); ok {
			// TODO: support nested struct filtering
			err := errors.NewDet(class.QueryFilterUnsupportedField, "nested fields filter is not supported")
			err.SetDetailsf("Field: '%s' is a composite field. Filtering nested attribute fields is not supported.", attr.NeuronName())
			return nil, err
		} else {
			err := errors.NewDet(class.QueryFilterUnknownField, "field not found")
			err.SetDetailsf("Field: '%s' not found for the Model: '%s'", params[1], mStruct.Collection())
			return nil, err
		}
	}
	return f, nil
}
