package query

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
	"github.com/kucjac/jsonapi/pkg/mapping"
)

// FilterField is a struct that keeps information about given query filters
// It is based on the mapping.StructField.
type FilterField filters.FilterField

// NewFilter creates new filterfield for given field, operator and values
func NewFilter(
	field *mapping.StructField,
	op Operator,
	values ...interface{},
) (f *FilterField) {
	// Create operator values
	ov := filters.NewOpValuePair(op, values...)

	// Create filter
	f = (*FilterField)(filters.NewFilter((*models.StructField)(field), ov))
	return
}

// func (f *FilterField) OpValues()

// StructField returns the filterfield struct
func (f *FilterField) StructField() *mapping.StructField {

	return (*mapping.StructField)(f.StructField())
}

// OpValuePair is a struct that holds the Operator information with the
type OperatorValuePair filters.OpValuePair

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
func (c *OperatorContainer) Register(raw, name string) Operator {
	o := &filters.Operator{Raw: raw, Name: name}
	c.c.RegisterOperators(o)
	return Operator(o)
}

// Get returns the operator for given raw string
func (c *OperatorContainer) Get(raw string) (Operator, bool) {
	op, ok := c.c.Get(raw)
	if ok {
		return Operator(op), ok
	}
	return nil, ok
}

// Operator is a query filter operator.
// It is an alias to a pointer so it can be comparable with equal sign
type Operator *filters.Operator
