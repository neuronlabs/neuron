package filter

import (
	"strings"

	"github.com/neuronlabs/neuron/mapping"
)

var _ Filter = Relation{}

// Relation is the relationship filter field.
type Relation struct {
	StructField *mapping.StructField
	// Nested are the relationship fields filters.
	Nested []Filter
}

// Copy implements Filter interface.
func (r Relation) Copy() Filter {
	cp := Relation{
		StructField: r.StructField,
	}
	if len(r.Nested) > 0 {
		cp.Nested = make([]Filter, len(r.Nested))
		for i, nested := range r.Nested {
			cp.Nested[i] = nested.Copy()
		}
	}
	return cp
}

// String implements fmt.Stringer interface.
func (r Relation) String() string {
	sb := strings.Builder{}
	sb.WriteString("Relation: ")
	sb.WriteString(r.StructField.Name())
	for i, nested := range r.Nested {
		sb.WriteString(nested.String())
		if i != len(r.Nested)-1 {
			sb.WriteRune(',')
		}
	}
	return sb.String()
}

// NewRelation creates new relationship filter for the 'relation' StructField.
// It adds all the nested relation sub filters 'relFilters'.
func NewRelation(relation *mapping.StructField, relFilters ...Filter) Relation {
	return Relation{StructField: relation, Nested: relFilters}
}
