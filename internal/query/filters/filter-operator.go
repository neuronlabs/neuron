package filters

import (
	"github.com/pkg/errors"
	"sync"
)

// Operators contains all the registered operators
var Operators *OperatorContainer = NewOpContainer()

// Operator is the operator used while filtering the query
type Operator struct {
	// Id is the filter operator id used for comparing the operator type
	Id uint16

	// Raw is the raw value of the current operator
	Raw string

	// Name is the human readable filter operator string value
	Name string
}

// OperatorContainer is the container for the filter operators
// It registers new operators and checks if no operator with provided Raw value already
// exists inside.
type OperatorContainer struct {
	operators map[string]*Operator
	*sync.Mutex
	lastID uint16

	lastStandardID uint16
}

// NewOpContainer creates new container operator
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

// Get gets the operator on the base of the raw value
func (c *OperatorContainer) Get(raw string) (*Operator, bool) {
	op, ok := c.operators[raw]
	return op, ok
}

// GetByName gets the operator by the name
func (c *OperatorContainer) GetByName(name string) (*Operator, bool) {
	for _, o := range c.operators {
		if o.Name == name {
			return o, true
		}
	}
	return nil, false
}

// RegisterOperators registers multiple operators
func (c *OperatorContainer) RegisterOperators(ops ...*Operator) error {
	return c.registerManyOperators(ops...)
}

// nextID generates next operator ID
func (c *OperatorContainer) nextID() uint16 {
	c.Lock()
	defer c.Unlock()

	c.lastID += 1

	return c.lastID
}

// RegisterOperator registers single operator
func (c *OperatorContainer) RegisterOperator(op *Operator) error {
	id := c.nextID()

	return c.registerOperator(op, id)
}

func (c *OperatorContainer) registerManyOperators(ops ...*Operator) error {
	nextID := c.nextID()
	for _, op := range ops {
		err := c.registerOperator(op, nextID)
		if err != nil {
			return err
		}
		nextID += 1
	}
	return nil
}

func (c *OperatorContainer) registerOperator(op *Operator, nextID uint16) error {
	for _, o := range c.operators {
		if o.Name == op.Name {
			return errors.Errorf("Operator with the name: %s and value: %s already registered.", op.Name, op.Raw)
		}
		if o.Raw == op.Raw {
			return errors.Errorf("Operator already registered. %+v", op)
		}
	}

	op.Id = nextID

	c.operators[op.Raw] = op
	return nil
}

var (
	// Logical Operators
	OpEqual        *Operator = &Operator{Raw: AnnotationOperatorEqual, Name: "Equal"}
	OpIn           *Operator = &Operator{Raw: AnnotationOperatorIn, Name: "In"}
	OpNotEqual     *Operator = &Operator{Raw: AnnotationOperatorNotEqual, Name: "NotEqual"}
	OpNotIn        *Operator = &Operator{Raw: AnnotationOperatorNotIn, Name: "NotIn"}
	OpGreaterThan  *Operator = &Operator{Raw: AnnotationOperatorGreaterThan, Name: "GreaterThan"}
	OpGreaterEqual *Operator = &Operator{Raw: AnnotationOperatorGreaterEqual, Name: "GreaterThanOrEqualTo"}
	OpLessThan     *Operator = &Operator{Raw: AnnotationOperatorLessThan, Name: "LessThan"}
	OpLessEqual    *Operator = &Operator{Raw: AnnotationOperatorLessEqual, Name: "LessThanOrEqualTo"}

	// Strings Only operators
	OpContains   = &Operator{Raw: AnnotationOperatorContains, Name: "Contains"}
	OpStartsWith = &Operator{Raw: AnnotationOperatorStartsWith, Name: "StartsWith"}
	OpEndsWith   = &Operator{Raw: AnnotationOperatorEndsWith, Name: "EndsWith"}

	OpIsNull    *Operator = &Operator{Raw: AnnotationOperatorIsNull, Name: "IsNull"}
	OpNotNull   *Operator = &Operator{Raw: AnnotationOperatorNotNull, Name: "NotNull"}
	OpExists    *Operator = &Operator{Raw: AnnotationOperatorExists, Name: "Exists"}
	OpNotExists *Operator = &Operator{Raw: AnnotationOperatorNotExists, Name: "NotExists"}
)

var defaultOperators []*Operator = []*Operator{
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
	OpExists,
	OpNotExists,
}

// String implements Stringer interface
func (f Operator) String() string {
	return f.Name
}

// IsStandard checks if the operator is standard
func (f Operator) IsStandard() bool {
	return f.Id <= Operators.lastStandardID
}

func (f Operator) isBasic() bool {
	return f.Id == OpEqual.Id || f.Id == OpNotEqual.Id
}

func (f Operator) isRangable() bool {
	return f.Id >= OpGreaterThan.Id && f.Id <= OpLessEqual.Id
}

func (f Operator) isStringOnly() bool {
	return f.Id >= OpContains.Id && f.Id <= OpEndsWith.Id
}
