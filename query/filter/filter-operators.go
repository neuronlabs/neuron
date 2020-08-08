package filter

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// FilterOperators is the container that stores all
// query filter operators.
var Operators = newOpContainer()

// Logical Operators
var (
	OpEqual        = &Operator{Value: "=", URLAlias: "$eq", Name: "Equal", Aliases: []string{"=="}}
	OpIn           = &Operator{Value: "in", URLAlias: "$in", Name: "In"}
	OpNotEqual     = &Operator{Value: "!=", URLAlias: "$ne", Name: "NotEqual", Aliases: []string{"<>"}}
	OpNotIn        = &Operator{Value: "not in", URLAlias: "$not_in", Name: "NotIn"}
	OpGreaterThan  = &Operator{Value: ">", URLAlias: "$gt", Name: "GreaterThan"}
	OpGreaterEqual = &Operator{Value: ">=", URLAlias: "$ge", Name: "GreaterThanOrEqualTo"}
	OpLessThan     = &Operator{Value: "<", URLAlias: "$lt", Name: "LessThan"}
	OpLessEqual    = &Operator{Value: "<=", URLAlias: "$le", Name: "LessThanOrEqualTo"}
)

// Strings Only operators.
var (
	OpContains   = &Operator{Value: "contains", URLAlias: "$contains", Name: "Contains"}
	OpStartsWith = &Operator{Value: "starts with", URLAlias: "$starts_with", Name: "StartsWith"}
	OpEndsWith   = &Operator{Value: "ends with", URLAlias: "$ends_with", Name: "EndsWith"}
)

// Null and Existence operators.
var (
	OpIsNull  = &Operator{Value: "is null", URLAlias: "$is_null", Name: "IsNull"}
	OpNotNull = &Operator{Value: "not null", URLAlias: "$not_null", Name: "NotNull"}
)

var defaultOperators = []*Operator{
	OpEqual,
	OpIn,
	OpNotEqual,
	OpNotIn,
	OpGreaterThan,
	OpGreaterEqual,
	OpLessThan,
	OpLessEqual,
	OpContains,
	OpStartsWith,
	OpEndsWith,
	OpIsNull,
	OpNotNull,
}

// Operator is the operator used for filtering the query.
type Operator struct {
	// ID is the filter operator id used for comparing the operator type.
	ID uint16
	// Models is the operator query value
	Value string
	// Name is the human readable filter operator name.
	Name string
	// URLAlias is the alias value for the url parsable value operator.
	URLAlias string
	// Aliases is the alias for the operator raw value.
	Aliases []string
}

// IsStandard checks if the operator is standard.
func (f *Operator) IsStandard() bool {
	return f.ID <= Operators.lastStandardID
}

// IsBasic checks if the operator is 'OpEqual' or OpNotEqual.
func (f *Operator) IsBasic() bool {
	return f.isBasic()
}

// IsRangeable checks if the operator allows to have value ranges.
func (f *Operator) IsRangeable() bool {
	return f.isRangeable()
}

// IsStringOnly checks if the operator is 'string only'.
func (f *Operator) IsStringOnly() bool {
	return f.isStringOnly()
}

// String implements Stringer interface.
func (f *Operator) String() string {
	return f.Name
}

func (f *Operator) isBasic() bool {
	return f.ID == OpEqual.ID || f.ID == OpNotEqual.ID
}

func (f *Operator) isRangeable() bool {
	return f.ID >= OpGreaterThan.ID && f.ID <= OpLessEqual.ID
}

func (f *Operator) isStringOnly() bool {
	return f.ID >= OpContains.ID && f.ID <= OpEndsWith.ID
}

/**

Operator Models

*/

// RegisterOperator registers the operator in the provided container
func RegisterOperator(o *Operator) error {
	err := Operators.registerOperator(o)
	log.Infof("Registered operator: %s with id: %d", o.ID)
	return err
}

// RegisterMultipleOperators registers multiple operators at once
func RegisterMultipleOperators(operators ...*Operator) error {
	return Operators.registerOperators(operators...)
}

/**

Operator Container

*/

// operatorContainer is the container for the filter operators.
// It registers new operators and checks if no operator with
// provided Raw value already exists inside.
type operatorContainer struct {
	operators      map[string]*Operator
	lastID         uint16
	lastStandardID uint16
}

// newOpContainer creates new container operator.
func newOpContainer() *operatorContainer {
	o := &operatorContainer{
		operators: make(map[string]*Operator),
	}

	err := o.registerManyOperators(defaultOperators...)
	if err != nil {
		panic(err)
	}

	o.lastStandardID = o.lastID
	return o
}

// Get gets the operator on the base of the raw value.
func (c *operatorContainer) Get(raw string) (*Operator, bool) {
	op, ok := c.operators[raw]
	return op, ok
}

// GetByName gets the operator by it's name.
func (c *operatorContainer) GetByName(name string) (*Operator, bool) {
	for _, o := range c.operators {
		if o.Name == name {
			return o, true
		}
	}
	return nil, false
}

// nextID generates next operator ID.
func (c *operatorContainer) nextID() uint16 {
	c.lastID++
	return c.lastID
}

func (c *operatorContainer) registerManyOperators(ops ...*Operator) error {
	for _, op := range ops {
		err := c.registerOperator(op)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterOperators registers multiple operators.
func (c *operatorContainer) registerOperators(ops ...*Operator) error {
	return c.registerManyOperators(ops...)
}

func (c *operatorContainer) registerOperator(op *Operator) error {
	op.ID = c.nextID()
	for _, o := range c.operators {
		if o.Name == op.Name {
			return errors.WrapDetf(errors.ErrInternal, "operator with the name: %s and value: %s already registered.", op.Name, op.URLAlias)
		}
		if (op.URLAlias != "" && (op.URLAlias == o.URLAlias)) || op.Value == o.Value {
			return errors.WrapDetf(errors.ErrInternal, "operator already registered. %+v", op)
		}
	}

	c.operators[op.Value] = op
	// addModel any possible alias for given operator.
	if op.URLAlias != "" {
		c.operators[op.URLAlias] = op
	}
	for _, alias := range op.Aliases {
		_, ok := c.operators[alias]
		if ok {
			return errors.WrapDetf(errors.ErrInternal, "operator alias: '%s' already registered. Operator: '%s'", alias, op.Name)
		}
		c.operators[alias] = op
	}
	return nil
}
