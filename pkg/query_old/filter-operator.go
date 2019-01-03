package query

import (
	"github.com/pkg/errors"
)

// Operator is the operator used while filtering the query
type Operator struct {
	// Id is the filter operator id used for comparing the operator type
	Id int

	// Raw is the raw value of the current operator
	Raw string

	// Name is the human readable filter operator string value
	Name string
}

// OperatorContainer is the container for the filter operators
// It registers new operators and checks if no operator with provided Raw value already
// exists inside.
type OperatorContainer map[string]*Operator

// NewOpContainer creates new container operator
func NewOpContainer() *OperatorContainer {
	o := &OperatorContainer{}
	err := o.registerManyOperators(defaultOperators...)
	if err != nil {
		panic(err)
	}

	return o
}

func (c *OperatorContainer) Get(raw string) (*Operator, bool) {
	op, ok := (*c)[raw]
	return op, ok
}

// RegisterOperators registers multiple operators
func (c *OperatorContainer) RegisterOperators(ops ...*Operator) error {
	return c.registerManyOperators(ops...)
}

func (c OperatorContainer) nextID() int {
	var nextId int

	// get the highest value
	for _, v := range c {
		if v.Id > nextId {
			nextId = v.Id
		}
	}
	return nextId
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

func (c *OperatorContainer) registerOperator(op *Operator, nextID int) error {
	if _, ok := (*c)[op.Raw]; ok {
		return errors.Errorf("Operator already registered. %+v", op)
	}
	op.Id = nextID
	(*c)[op.Raw] = op
	return nil
}

var (
	// Logical Operators
	OpEqual        *Operator = &Operator{Raw: operatorEqual, Name: "Equal"}
	OpIn           *Operator = &Operator{Raw: operatorIn, Name: "In"}
	OpNotEqual     *Operator = &Operator{Raw: operatorNotEqual, Name: "NotEqual"}
	OpNotIn        *Operator = &Operator{Raw: operatorNotIn, Name: "NotIn"}
	OpGreaterThan  *Operator = &Operator{Raw: operatorGreaterThan, Name: "GreaterThan"}
	OpGreaterEqual *Operator = &Operator{Raw: operatorGreaterEqual, Name: "GreaterThanOrEqualTo"}
	OpLessThan     *Operator = &Operator{Raw: operatorLessThan, Name: "LessThan"}
	OpLessEqual    *Operator = &Operator{Raw: operatorLessEqual, Name: "LessThanOrEqualTo"}

	// Strings Only operators
	OpContains   = &Operator{Raw: operatorContains, Name: "Contains"}
	OpStartsWith = &Operator{Raw: operatorStartsWith, Name: "StartsWith"}
	OpEndsWith   = &Operator{Raw: operatorEndsWith, Name: "EndsWith"}

	OpIsNull    *Operator = &Operator{Raw: operatorIsNull, Name: "IsNull"}
	OpNotNull   *Operator = &Operator{Raw: operatorNotNull, Name: "NotNull"}
	OpExists    *Operator = &Operator{Raw: operatorExists, Name: "Exists"}
	OpNotExists *Operator = &Operator{Raw: operatorNotExists, Name: "NotExists"}
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

func (f Operator) isBasic() bool {
	return f.Id == OpEqual.Id || f.Id == OpNotEqual.Id
}

func (f Operator) isRangable() bool {
	return f.Id >= OpGreaterThan.Id && f.Id <= OpLessEqual.Id
}

func (f Operator) isStringOnly() bool {
	return f.Id >= OpContains.Id && f.Id <= OpEndsWith.Id
}
