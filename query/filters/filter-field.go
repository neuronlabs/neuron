package filters

import (
	"fmt"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	icontroller "github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"net/url"
	"strings"
)

// QueryValuer is the interface that allows parsing the structures into golang filters string values
type QueryValuer interface {
	QueryValue(sField *mapping.StructField) string
}

// FilterField is a struct that keeps information about given query filters
// It is based on the mapping.StructField.
type FilterField filters.FilterField

// NewFilter creates new filterfield for given field, operator and values
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

// NewRelationshipFilter creates new relationship filter for the 'relation' StructField
// It adds all the nested relation subfilters 'relFilters'
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

// NewStringFilter creates the filter field based on the provided filter, schemaName and values
// The filter should not contain the values. If the 'schemaName' is not provided the defaultSchema for given controller would be taken.
// Example:
//	- 'fitler': "filter[collection][fieldName][operator]"
//  - 'schema': "schemaName"
//  - 'values': 5, 13
func NewStringFilter(c *controller.Controller, filter string, schemaName string, values ...interface{}) (*FilterField, error) {

	if strings.HasPrefix(filter, "filter") {
		filter = filter[6:]
	}

	params, err := internal.SplitBracketParameter(filter)
	if err != nil {
		return nil, err
	}

	schemas := (*icontroller.Controller)(c).ModelSchemas()
	var schema *models.Schema
	if schemaName != "" {
		var ok bool
		schema, ok = schemas.Schema(schemaName)
		if !ok {
			return nil, fmt.Errorf("Schema: '%s' is not found within the controller", schemaName)
		}
	} else {
		schema = schemas.DefaultSchema()
	}

	mStruct := schema.ModelByCollection(params[0])
	if mStruct == nil {
		return nil, fmt.Errorf("Model: '%s' is not found within the schema: '%s'", params[0], schema.Name)
	}

	var (
		f  *filters.FilterField
		op *filters.Operator
		ok bool
	)

	findOperator := func(index int) error {
		op, ok = filters.Operators.Get(params[index])
		if !ok {
			return fmt.Errorf("Operator not found: %s", params[index])
		}
		return nil
	}
	if len(params) <= 1 || len(params) > 4 {
		return nil, fmt.Errorf("Invalid filter format: %s", filter)
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

		if params[1] == mStruct.PrimaryField().ApiName() {
			field = mStruct.PrimaryField()
		} else if field, ok = mStruct.Attribute(params[1]); !ok {
			if field, ok = mStruct.ForeignKey(params[1]); !ok {
				if field, ok = mStruct.FilterKey(params[1]); !ok {
					return nil, fmt.Errorf("Field: '%s' not found within the Model: '%s'", params[1], mStruct.Collection())
				}
			}
		}

		f = filters.NewFilter(field)
		f.AddValues(filters.NewOpValuePair(op, values...))
	case 4:
		// [collection][relationship][field][op]
		log.Debugf("4")
		var (
			field *models.StructField
		)

		if err = findOperator(3); err != nil {
			return nil, err
		}

		if rel, ok := mStruct.RelationshipField(params[1]); ok {
			f = filters.NewFilter(rel)
			relStruct := rel.Struct()
			if params[2] == "id" {
				field = relStruct.PrimaryField()
			} else if field, ok = relStruct.Attribute(params[2]); !ok {
				if field, ok = relStruct.ForeignKey(params[2]); !ok {
					if field, ok = relStruct.FilterKey(params[2]); !ok {
						return nil, fmt.Errorf("Field: '%s' not found within the relation ModelStruct: '%s'", params[2], relStruct.Collection())
					}
				}
			}
			nested := filters.NewFilter(field, filters.NewOpValuePair(op, values...))
			f.AddNestedField(nested)

		} else if attr, ok := mStruct.Attribute(params[1]); ok {
			// TODO: support nested struct filtering

			return nil, fmt.Errorf("Nested struct's: '%s' filtering not implemented yet.", attr.Name())
		} else {
			return nil, fmt.Errorf("Field: '%s' not found for the Model: '%s'", params[1], mStruct.Collection())
		}
	}
	return (*FilterField)(f), nil
}

// NestedFilters returns the nested filters for given filter fields
// Nested filters are the filters used for relationship filters
func (f *FilterField) NestedFilters() []*FilterField {

	var nesteds []*FilterField
	for _, n := range (*filters.FilterField)(f).NestedFields() {
		nesteds = append(nesteds, (*FilterField)(n))
	}

	return nesteds

}

// FormatQuery formats the filter field into url.Value format
// If the 'q' optional parameter is set, then the function would add
// the values into the provided argument url.Values. Otherwise it
// creates new url.Values.
// Returns updated (new) url.Values
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

// StructField returns the structfield related with the filter field
func (f *FilterField) StructField() *mapping.StructField {
	var field *models.StructField
	if field = (*filters.FilterField)(f).StructField(); field == nil {
		return nil
	}
	return (*mapping.StructField)(field)
}

// Values returns OperatorValuesPair for given filterField
func (f *FilterField) Values() (values []*OperatorValuePair) {

	v := (*filters.FilterField)(f).Values()
	for _, sv := range v {
		values = append(values, (*OperatorValuePair)(sv))
	}
	return values
}

// formatQuery parses the into url.Values
func (f *FilterField) formatQuery(q url.Values, relName ...string) {

	// parse the internal value
	for _, fv := range (*filters.FilterField)(f).Values() {
		var fk string
		if len(relName) > 0 {
			fk = fmt.Sprintf("[%s][%s][%s]", relName[0], f.StructField().ApiName(), fv.Operator().Raw)
		} else if (*models.StructField)(f.StructField()).IsLanguage() {
			fk = internal.QueryParamLanguage
		} else {
			fk = fmt.Sprintf("[%s][%s]", f.StructField().ApiName(), fv.Operator().Raw)
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
			(*FilterField)(nested).formatQuery(q, f.StructField().ApiName())
		}
	}
}

// OperatorValuePair is a struct that holds the Operator information with the
type OperatorValuePair filters.OpValuePair

// Operator returns the operator for given pair
func (o *OperatorValuePair) Operator() *Operator {
	return (*Operator)((*filters.OpValuePair)(o).Operator())
}

// SetOperator sets the operator
func (o *OperatorValuePair) SetOperator(op *Operator) {
	(*filters.OpValuePair)(o).SetOperator((*filters.Operator)(op))
}

// OperatorContainer is a container for the provided operators.
// It allows registering new and getting already registered operators.
type OperatorContainer struct {
	c *filters.OperatorContainer
}

// NewContainer creates new operator container
func NewContainer() *OperatorContainer {
	c := &OperatorContainer{c: filters.NewOpContainer()}
	return c
}

// Register creates and registers new operator with 'raw' and 'name' values.
func (c *OperatorContainer) Register(raw, name string) (*Operator, error) {
	o := &filters.Operator{Raw: raw, Name: name}
	err := c.c.RegisterOperators(o)
	if err != nil {
		return nil, err
	}

	return (*Operator)(o), nil
}

// Get returns the operator for given raw string
func (c *OperatorContainer) Get(raw string) (*Operator, bool) {
	op, ok := c.c.Get(raw)
	if ok {
		return (*Operator)(op), ok
	}
	return nil, ok
}
