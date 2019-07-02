package filters

import (
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/query/filters"
)

var (
	// OpEqual is the standard filter operator.
	OpEqual = (*Operator)(filters.OpEqual)

	// OpIn is the standard filter operator.
	OpIn = (*Operator)(filters.OpIn)

	// OpNotEqual is the standard filter operator.
	OpNotEqual = (*Operator)(filters.OpNotEqual)

	// OpNotIn is the standard filter operator.
	OpNotIn = (*Operator)(filters.OpNotIn)

	// OpGreaterThan is the standard filter operator.
	OpGreaterThan = (*Operator)(filters.OpGreaterThan)

	// OpGreaterEqual is the standard filter operator.
	OpGreaterEqual = (*Operator)(filters.OpGreaterEqual)

	// OpLessThan is the standard filter operator.
	OpLessThan = (*Operator)(filters.OpLessThan)

	// OpLessEqual is the standard filter operator.
	OpLessEqual = (*Operator)(filters.OpLessEqual)

	// OpContains is the standard filter operator.
	OpContains = (*Operator)(filters.OpContains)

	// OpStartsWith is the standard filter operator.
	OpStartsWith = (*Operator)(filters.OpStartsWith)

	// OpEndsWith is the standard filter operator.
	OpEndsWith = (*Operator)(filters.OpEndsWith)

	// OpIsNull is the standard filter operator.
	OpIsNull = (*Operator)(filters.OpIsNull)

	// OpNotNull is the standard filter operator.
	OpNotNull = (*Operator)(filters.OpNotNull)

	// OpExists is the standard filter operator.
	OpExists = (*Operator)(filters.OpExists)

	// OpNotExists is the standard filter operator.
	OpNotExists = (*Operator)(filters.OpNotExists)
)

// Operator is a query filter operator.
type Operator filters.Operator

// IsStandard defines if the operator is a standard filter operator.
func (o *Operator) IsStandard() bool {
	return (*filters.Operator)(o).IsStandard()
}

// NewOperator creates new operator with the 'name' and 'queryRaw' provided in the arguments.
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
	log.Infof("Registered operator: %s with id: %d", o.ID)
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
