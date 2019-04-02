package filters

import (
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/log"
)

// Standard filter operators
var (
	// OpEqual is the standard filter operator
	OpEqual *Operator = (*Operator)(filters.OpEqual)
	// OpIn is the standard filter operator
	OpIn *Operator = (*Operator)(filters.OpIn)
	// OpNotEqual is the standard filter operator
	OpNotEqual *Operator = (*Operator)(filters.OpNotEqual)
	// OpNotIn is the standard filter operator
	OpNotIn *Operator = (*Operator)(filters.OpNotIn)
	// OpGreaterThan is the standard filter operator
	OpGreaterThan *Operator = (*Operator)(filters.OpGreaterThan)
	// OpGreaterEqual is the standard filter operator
	OpGreaterEqual *Operator = (*Operator)(filters.OpGreaterEqual)
	// OpLessThan is the standard filter operator
	OpLessThan *Operator = (*Operator)(filters.OpLessThan)
	// OpLessEqual is the standard filter operator
	OpLessEqual *Operator = (*Operator)(filters.OpLessEqual)
	// OpContains is the standard filter operator
	OpContains *Operator = (*Operator)(filters.OpContains)
	// OpStartsWith is the standard filter operator
	OpStartsWith *Operator = (*Operator)(filters.OpStartsWith)
	// OpEndsWith is the standard filter operator
	OpEndsWith *Operator = (*Operator)(filters.OpEndsWith)
	// OpIsNull is the standard filter operator
	OpIsNull *Operator = (*Operator)(filters.OpIsNull)
	// OpExists is the standard filter operator
	OpExists *Operator = (*Operator)(filters.OpExists)
	// OpNotExists is the standard filter operator
	OpNotExists *Operator = (*Operator)(filters.OpNotExists)
)

// Operator is a query filter operator.
// It is an alias to a pointer so it can be comparable with equal sign
type Operator filters.Operator

// IsStandard defines if the operator is a standard filter operator
func (o *Operator) IsStandard() bool {
	return (*filters.Operator)(o).IsStandard()
}

// NewOperator creates new operator with the 'name' and 'queryRaw' provided in the arguments
func NewOperator(queryRaw, name string) *Operator {
	o := &Operator{
		Raw:  queryRaw,
		Name: name,
	}
	return o
}

// RegisterOperator registers the operator in the provided container
func RegisterOperator(o *Operator) error {
	err := filters.Operators.RegisterOperator((*filters.Operator)(o))
	log.Infof("Registered operator: %s with id: %d", o.Id)
	return err
}

// RegisterMultipleOperators registers multiple operators at once
func RegisterMultipleOperators(operators ...*Operator) error {
	fOperators := []*filters.Operator{}
	for _, o := range operators {
		fOperators = append(fOperators, (*filters.Operator)(o))
	}
	return filters.Operators.RegisterOperators(fOperators...)
}
