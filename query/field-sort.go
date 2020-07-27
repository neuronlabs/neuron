package query

import (
	"github.com/neuronlabs/neuron/mapping"
)

// SortField is a field that contains sorting information.
type SortField struct {
	StructField *mapping.StructField
	// Order defines if the sorting order (ascending or descending)
	SortOrder SortOrder
}

// Order implements Sort interface.
func (s SortField) Order() SortOrder {
	return s.SortOrder
}

// Field implements Sort interface.
func (s SortField) Field() *mapping.StructField {
	return s.StructField
}

// Copy creates a copy of the sort field.
func (s SortField) Copy() Sort {
	return SortField{StructField: s.StructField, SortOrder: s.SortOrder}
}

func (s SortField) String() string {
	var v string
	if s.SortOrder == DescendingOrder {
		v = "-"
	}
	v += s.StructField.NeuronName()
	return v
}

// RelationSort is an
type RelationSort struct {
	StructField    *mapping.StructField
	SortOrder      SortOrder
	RelationFields []*mapping.StructField
}

// Order implements Sort interface.
func (r RelationSort) Order() SortOrder {
	return r.SortOrder
}

// Field implements Sort interface.
func (r RelationSort) Field() *mapping.StructField {
	return r.StructField
}

// Copy implements Sort interface.
func (r RelationSort) Copy() Sort {
	cp := RelationSort{StructField: r.StructField, SortOrder: r.Order(), RelationFields: make([]*mapping.StructField, len(r.RelationFields))}
	copy(cp.RelationFields, r.RelationFields)
	return cp
}
