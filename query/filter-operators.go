package query

import (
	"sync"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// FilterOperators is the container that stores all
// query filter operators.
var FilterOperators = newOpContainer()

// Operator definitions variables.
var (
	// Logical Operators
	OpEqual        = &Operator{Raw: operatorEqualRaw, Name: "Equal"}
	OpIn           = &Operator{Raw: operatorInRaw, Name: "In"}
	OpNotEqual     = &Operator{Raw: operatorNotEqualRaw, Name: "NotEqual"}
	OpNotIn        = &Operator{Raw: operatorNotInRaw, Name: "NotIn"}
	OpGreaterThan  = &Operator{Raw: operatorGreaterThanRaw, Name: "GreaterThan"}
	OpGreaterEqual = &Operator{Raw: operatorGreaterEqualRaw, Name: "GreaterThanOrEqualTo"}
	OpLessThan     = &Operator{Raw: operatorLessThanRaw, Name: "LessThan"}
	OpLessEqual    = &Operator{Raw: operatorLessEqualRaw, Name: "LessThanOrEqualTo"}

	// Strings Only operators.
	OpContains   = &Operator{Raw: operatorContainsRaw, Name: "Contains"}
	OpStartsWith = &Operator{Raw: operatorStartsWithRaw, Name: "StartsWith"}
	OpEndsWith   = &Operator{Raw: operatorEndsWithRaw, Name: "EndsWith"}

	// Null and Existence operators.
	OpIsNull    = &Operator{Raw: operatorIsNullRaw, Name: "IsNull"}
	OpNotNull   = &Operator{Raw: operatorNotNullRaw, Name: "NotNull"}
	OpExists    = &Operator{Raw: operatorExistsRaw, Name: "Exists"}
	OpNotExists = &Operator{Raw: operatorNotExistsRaw, Name: "NotExists"}
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
	OpExists,
	OpNotExists,
}

// Operator is the operator used for filtering the query.
type Operator struct {
	// ID is the filter operator id used for comparing the operator type
	ID uint16
	// Raw is the filtering string value of the current operator
	Raw string
	// Name is the human readable filter operator name
	Name string
}

// IsStandard checks if the operator is standard
func (f Operator) IsStandard() bool {
	return f.ID <= FilterOperators.lastStandardID
}

// IsBasic checks if the operator is 'OpEqual' or OpNotEqual
func (f Operator) IsBasic() bool {
	return f.isBasic()
}

// IsRangeable checks if the operator allows to have value ranges.
func (f Operator) IsRangeable() bool {
	return f.isRangable()
}

// IsStringOnly checks if the operator is 'string only'.
func (f Operator) IsStringOnly() bool {
	return f.isStringOnly()
}

// String implements Stringer interface.
func (f Operator) String() string {
	return f.Name
}

func (f Operator) isBasic() bool {
	return f.ID == OpEqual.ID || f.ID == OpNotEqual.ID
}

func (f Operator) isRangable() bool {
	return f.ID >= OpGreaterThan.ID && f.ID <= OpLessEqual.ID
}

func (f Operator) isStringOnly() bool {
	return f.ID >= OpContains.ID && f.ID <= OpEndsWith.ID
}

/**

Operator Values

*/

// OperatorValues is a struct that holds the Operator with
// the filter values.
type OperatorValues struct {
	Values   []interface{}
	Operator *Operator
}

func (o *OperatorValues) copy() *OperatorValues {
	ov := &OperatorValues{
		Operator: o.Operator,
		Values:   make([]interface{}, len(o.Values)),
	}
	copy(ov.Values, o.Values)
	return ov
}

// RegisterOperator registers the operator in the provided container
func RegisterOperator(o *Operator) error {
	err := FilterOperators.registerOperator(o)
	log.Infof("Registered operator: %s with id: %d", o.ID)
	return err
}

// RegisterMultipleOperators registers multiple operators at once
func RegisterMultipleOperators(operators ...*Operator) error {
	return FilterOperators.registerOperators(operators...)
}

/**

Operator Container

*/

// operatorContainer is the container for the filter operators.
// It registers new operators and checks if no operator with
// provided Raw value already exists inside.
type operatorContainer struct {
	operators map[string]*Operator
	*sync.Mutex
	lastID         uint16
	lastStandardID uint16
}

// newOpContainer creates new container operator.
func newOpContainer() *operatorContainer {
	o := &operatorContainer{
		operators: make(map[string]*Operator),
		Mutex:     &sync.Mutex{},
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
	c.Lock()
	defer c.Unlock()

	for _, op := range ops {
		nextID := c.nextID()
		err := c.registerOperatorNonLocked(op, nextID)
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

func (c *operatorContainer) registerOperatorNonLocked(op *Operator, nextID uint16) error {
	for _, o := range c.operators {
		if o.Name == op.Name {
			return errors.NewDetf(class.InternalQueryFilter, "Operator with the name: %s and value: %s already registered.", op.Name, op.Raw)
		}
		if o.Raw == op.Raw {
			return errors.NewDetf(class.InternalQueryFilter, "Operator already registered. %+v", op)
		}
	}
	op.ID = nextID

	c.operators[op.Raw] = op
	return nil
}

// RegisterOperator registers single operator.
func (c *operatorContainer) registerOperator(op *Operator) error {
	c.Lock()
	defer c.Unlock()
	id := c.nextID()

	return c.registerOperatorNonLocked(op, id)
}
