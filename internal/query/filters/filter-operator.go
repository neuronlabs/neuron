package filters

import (
	"sync"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
)

// Operators contains all the registered operators.
var Operators = NewOpContainer()

// Operator definitions variables.
var (
	// Logical Operators
	OpEqual        = &Operator{Raw: AnnotationOperatorEqual, Name: "Equal"}
	OpIn           = &Operator{Raw: AnnotationOperatorIn, Name: "In"}
	OpNotEqual     = &Operator{Raw: AnnotationOperatorNotEqual, Name: "NotEqual"}
	OpNotIn        = &Operator{Raw: AnnotationOperatorNotIn, Name: "NotIn"}
	OpGreaterThan  = &Operator{Raw: AnnotationOperatorGreaterThan, Name: "GreaterThan"}
	OpGreaterEqual = &Operator{Raw: AnnotationOperatorGreaterEqual, Name: "GreaterThanOrEqualTo"}
	OpLessThan     = &Operator{Raw: AnnotationOperatorLessThan, Name: "LessThan"}
	OpLessEqual    = &Operator{Raw: AnnotationOperatorLessEqual, Name: "LessThanOrEqualTo"}

	// Strings Only operators
	OpContains   = &Operator{Raw: AnnotationOperatorContains, Name: "Contains"}
	OpStartsWith = &Operator{Raw: AnnotationOperatorStartsWith, Name: "StartsWith"}
	OpEndsWith   = &Operator{Raw: AnnotationOperatorEndsWith, Name: "EndsWith"}

	OpIsNull    = &Operator{Raw: AnnotationOperatorIsNull, Name: "IsNull"}
	OpNotNull   = &Operator{Raw: AnnotationOperatorNotNull, Name: "NotNull"}
	OpExists    = &Operator{Raw: AnnotationOperatorExists, Name: "Exists"}
	OpNotExists = &Operator{Raw: AnnotationOperatorNotExists, Name: "NotExists"}
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

// OperatorContainer is the container for the filter operators.
// It registers new operators and checks if no operator with
// provided Raw value already exists inside.
type OperatorContainer struct {
	operators map[string]*Operator
	*sync.Mutex
	lastID uint16

	lastStandardID uint16
}

// NewOpContainer creates new container operator.
func NewOpContainer() *OperatorContainer {
	o := &OperatorContainer{
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
func (c *OperatorContainer) Get(raw string) (*Operator, bool) {
	op, ok := c.operators[raw]
	return op, ok
}

// GetByName gets the operator by it's name.
func (c *OperatorContainer) GetByName(name string) (*Operator, bool) {
	for _, o := range c.operators {
		if o.Name == name {
			return o, true
		}
	}
	return nil, false
}

// IsStandard checks if the operator is standard
func (f Operator) IsStandard() bool {
	return f.ID <= Operators.lastStandardID
}

// RegisterOperators registers multiple operators.
func (c *OperatorContainer) RegisterOperators(ops ...*Operator) error {
	return c.registerManyOperators(ops...)
}

// RegisterOperator registers single operator.
func (c *OperatorContainer) RegisterOperator(op *Operator) error {
	c.Lock()
	defer c.Unlock()
	id := c.nextID()

	return c.registerOperator(op, id)
}

// String implements Stringer interface.
func (f Operator) String() string {
	return f.Name
}

// nextID generates next operator ID.
func (c *OperatorContainer) nextID() uint16 {
	c.lastID++
	return c.lastID
}

func (c *OperatorContainer) registerManyOperators(ops ...*Operator) error {
	c.Lock()
	defer c.Unlock()

	for _, op := range ops {
		nextID := c.nextID()
		err := c.registerOperator(op, nextID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *OperatorContainer) registerOperator(op *Operator, nextID uint16) error {
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

func (f Operator) isBasic() bool {
	return f.ID == OpEqual.ID || f.ID == OpNotEqual.ID
}

func (f Operator) isRangable() bool {
	return f.ID >= OpGreaterThan.ID && f.ID <= OpLessEqual.ID
}

func (f Operator) isStringOnly() bool {
	return f.ID >= OpContains.ID && f.ID <= OpEndsWith.ID
}
