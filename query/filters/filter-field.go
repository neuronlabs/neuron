package filters

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal"
	internalController "github.com/neuronlabs/neuron-core/internal/controller"
	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/query/filters"
)

// QueryValuer is the interface that allows parsing the structures into golang filters string values.
type QueryValuer interface {
	QueryValue(sField *mapping.StructField) string
}

// FilterField is a struct that keeps information about given query filters.
// It is based on the mapping.StructField.
type FilterField filters.FilterField

// NewFilter creates new filterfield for given field, operator and values.
func NewFilter(
	field *mapping.StructField,
	op *Operator,
	values ...interface{},
) (f *FilterField) {
	// Create operator values
	ov := filters.NewOpValuePair((*filters.Operator)(op), values...)

	// Create filter
	f = (*FilterField)(filters.NewFilter((*models.StructField)(field), ov))
	return
}

// NewRelationshipFilter creates new relationship filter for the 'relation' StructField.
// It adds all the nested relation subfilters 'relFilters'.
func NewRelationshipFilter(
	relation *mapping.StructField,
	relFilters ...*FilterField,
) *FilterField {
	f := filters.NewFilter((*models.StructField)(relation))
	for _, relFilter := range relFilters {
		f.AddNestedField((*filters.FilterField)(relFilter))
	}
	return (*FilterField)(f)
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
	if strings.HasPrefix(filter, "filter") {
		filter = filter[6:]
	}

	params, err := common.SplitBracketParameter(filter)
	if err != nil {
		return nil, err
	}

	mStruct := (*internalController.Controller)(c).ModelMap().GetByCollection(params[0])
	if mStruct == nil {
		err := errors.New(class.QueryFilterUnknownCollection, "provided filter collection not found")
		err.SetDetailf("Filter model: '%s' not found", params[0])
		return nil, err
	}

	var (
		f  *filters.FilterField
		op *filters.Operator
		ok bool
	)

	findOperator := func(index int) error {
		op, ok = filters.Operators.Get(params[index])
		if !ok {
			err := errors.New(class.QueryFilterUnknownOperator, "operator not found")
			err.SetDetailf("Unknown filte operator not found: %s", params[index])
			return err
		}
		return nil
	}
	if len(params) <= 1 || len(params) > 4 {
		return nil, errors.New(class.QueryFilterInvalidFormat, "invalid filter format")
	}

	switch len(params) {
	case 2, 3:
		// [collection][field][op]
		if len(params) == 3 {
			if err := findOperator(2); err != nil {
				return nil, err
			}
		} else {
			op = filters.OpEqual
		}

		var field *models.StructField

		if params[1] == mStruct.PrimaryField().NeuronName() {
			field = mStruct.PrimaryField()
		} else if field, ok = mStruct.Attribute(params[1]); !ok {
			if foreignKeyAllowed {
				field, ok = mStruct.ForeignKey(params[1])
			}
			if !ok {
				field, ok = mStruct.FilterKey(params[1])
			}
			if !ok {
				log.Debugf("Filtering over unknown field: '%s'.", params[1])

				err := errors.New(class.QueryFilterUnknownField, "field not found")
				err.SetDetailf("Field: '%s' not found for the Model: '%s'", params[1], mStruct.Collection())

				return nil, err
			}
		}

		f = filters.NewFilter(field)
		f.AddValues(filters.NewOpValuePair(op, values...))
	case 4:
		// [collection][relationship][field][op]
		var (
			field *models.StructField
		)

		if err = findOperator(3); err != nil {
			return nil, err
		}

		if rel, ok := mStruct.RelationshipField(params[1]); ok {
			f = filters.NewFilter(rel)
			relStruct := rel.Relationship().Struct()
			if params[2] == "id" {
				field = relStruct.PrimaryField()
			} else if field, ok = relStruct.Attribute(params[2]); !ok {
				if foreignKeyAllowed {
					field, ok = relStruct.ForeignKey(params[2])
				}
				if !ok {
					field, ok = relStruct.FilterKey(params[2])
				}
				if !ok {
					err := errors.New(class.QueryFilterInvalidField, "field not found")
					err.SetDetailf("Field: '%s' not found within the relation ModelStruct: '%s'", params[2], relStruct.Collection())
					return nil, err
				}
			}

			nested := filters.NewFilter(field, filters.NewOpValuePair(op, values...))
			f.AddNestedField(nested)
		} else if attr, ok := mStruct.Attribute(params[1]); ok {
			// TODO: support nested struct filtering
			err := errors.New(class.QueryFilterUnsupportedField, "nested fields filter is not supported")
			err.SetDetailf("Field: '%s' is a composite field. Filtering nested attribute fields is not supported.", attr.NeuronName())
			return nil, err
		} else {
			log.Debugf("Filter over unknown field: '%s'.", params[1])
			err := errors.New(class.QueryFilterUnknownField, "field not found")
			err.SetDetailf("Field: '%s' not found for the Model: '%s'", params[1], mStruct.Collection())
			return nil, err
		}
	}
	return (*FilterField)(f), nil
}

// NestedFilters returns the nested filters for given filter fields.
// Nested filters are the filters used for relationship or composite attribute filters.
func (f *FilterField) NestedFilters() []*FilterField {
	var nesteds []*FilterField
	for _, n := range (*filters.FilterField)(f).NestedFields() {
		nesteds = append(nesteds, (*FilterField)(n))
	}

	return nesteds

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

	collection := f.StructField().ModelStruct().Collection()

	for k, vals := range temp {
		if k[0] == '[' {
			k = fmt.Sprintf("filter[%s]%s", collection, k)
		}
		query.Add(k, strings.Join(vals, internal.AnnotationSeperator))
	}
	return query

}

// StructField returns the structfield related with the filter field.
func (f *FilterField) StructField() *mapping.StructField {
	var field *models.StructField
	if field = (*filters.FilterField)(f).StructField(); field == nil {
		return nil
	}
	return (*mapping.StructField)(field)
}

// Values returns OperatorValuesPair.
func (f *FilterField) Values() (values []*OperatorValuePair) {
	v := (*filters.FilterField)(f).Values()
	for _, sv := range v {
		values = append(values, (*OperatorValuePair)(sv))
	}
	return values
}

// formatQuery parses the into url.Values.
func (f *FilterField) formatQuery(q url.Values, relName ...string) {
	// parse the internal value
	for _, fv := range (*filters.FilterField)(f).Values() {
		var fk string

		if len(relName) > 0 {
			fk = fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField().NeuronName(), fv.Operator().Raw)
		} else if (*filters.FilterField)(f).StructField().IsLanguage() {
			fk = common.QueryParamLanguage
		} else {
			fk = fmt.Sprintf("[%s][%s]", f.StructField().NeuronName(), fv.Operator().Raw)
		}

		// [fieldname][operator]
		var vals []string
		for _, val := range fv.Values {
			stringValue(f.StructField(), val, &vals)
		}

		for _, v := range vals {
			q.Add(fk, v)
		}
	}

	if len((*filters.FilterField)(f).NestedFields()) > 0 {
		for _, nested := range (*filters.FilterField)(f).NestedFields() {
			(*FilterField)(nested).formatQuery(q, f.StructField().NeuronName())
		}
	}
}

// OperatorValuePair is a struct that holds the Operator information with
// the related filter values.
type OperatorValuePair filters.OpValuePair

// Operator returns the operator.
func (o *OperatorValuePair) Operator() *Operator {
	return (*Operator)((*filters.OpValuePair)(o).Operator())
}

// SetOperator sets the operator in the operator value pair.
func (o *OperatorValuePair) SetOperator(op *Operator) {
	(*filters.OpValuePair)(o).SetOperator((*filters.Operator)(op))
}
