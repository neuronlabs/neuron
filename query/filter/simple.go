package filter

import (
	"fmt"
	"strings"

	"github.com/neuronlabs/neuron/mapping"
)

// Simple is a struct that keeps information about given query filters.
// It is based on the mapping.StructField.
type Simple struct {
	StructField *mapping.StructField
	Operator    *Operator
	Values      []interface{}
}

// Copy returns the copy of the filter field.
func (f Simple) Copy() Filter {
	return f.copy()
}

// String implements fmt.Stringer interface.
func (f Simple) String() string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s %s", f.StructField.NeuronName(), f.Operator.URLAlias))
	return sb.String()
}

func (f Simple) copy() Simple {
	cp := Simple{StructField: f.StructField, Operator: f.Operator}
	if len(f.Values) > 0 {
		cp.Values = make([]interface{}, len(f.Values))
		copy(cp.Values, f.Values)
	}
	return cp
}

// New creates new filterField for given 'field', operator and 'values'.
func New(field *mapping.StructField, op *Operator, values ...interface{}) Simple {
	return Simple{StructField: field, Operator: op, Values: values}
}
