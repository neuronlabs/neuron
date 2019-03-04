package filters

import (
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/mapping"
)

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

// NestedFilters returns the nested filters for given filter fields
// Nested filters are the filters used for relationship filters
func (f *FilterField) NestedFilters() []*FilterField {

	var nesteds []*FilterField
	for _, n := range (*filters.FilterField)(f).NestedFields() {
		nesteds = append(nesteds, (*FilterField)(n))
	}

	return nesteds

}

// StructField returns the filterfield struct
func (f *FilterField) StructField() *mapping.StructField {
	return (*mapping.StructField)((*filters.FilterField)(f).StructField())
}

// Values returns OperatorValuesPair for given filterField
func (f *FilterField) Values() (values []*OperatorValuePair) {

	v := (*filters.FilterField)(f).Values()
	for _, sv := range v {
		values = append(values, (*OperatorValuePair)(sv))
	}
	return values
}

// OpValuePair is a struct that holds the Operator information with the
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

// Operator is a query filter operator.
// It is an alias to a pointer so it can be comparable with equal sign
type Operator filters.Operator
