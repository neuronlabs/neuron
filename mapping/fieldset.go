package mapping

import (
	"sort"
)

// FieldSet is a slice of fields, with some basic search functions.
type FieldSet []*StructField

// Contains checks if given fieldset contains given 'sField'.
func (f FieldSet) Contains(sField *StructField) bool {
	for _, field := range f {
		if field == sField {
			return true
		}
	}
	return false
}

// ContainsFieldName checks if a field with 'fieldName' exists in given set.
func (f FieldSet) ContainsFieldName(fieldName string) bool {
	for _, field := range f {
		if field.neuronName == fieldName || field.Name() == fieldName {
			return true
		}
	}
	return false
}

// Copy creates a copy of the fieldset.
func (f FieldSet) Copy() FieldSet {
	cp := make(FieldSet, len(f))
	copy(cp, f)
	return cp
}

// Hash returns the map entry
func (f FieldSet) Hash() (hash string) {
	for _, field := range f {
		hash += field.neuronName
	}
	return hash
}

// Sort sorts given fieldset by fields indices.
func (f FieldSet) Sort() {
	sort.Sort(f)
}

// Len implements sort.Interface interface.
func (f FieldSet) Len() int {
	return len(f)
}

// Less implements sort.Interface interface.
func (f FieldSet) Less(i, j int) bool {
	return f.less(f[i], f[j])
}

// Swap implements sort.Interface interface.
func (f FieldSet) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FieldSet) less(first, second *StructField) bool {
	var result bool
	for k := 0; k < len(first.Index); k++ {
		if first.Index[k] != second.Index[k] {
			result = first.Index[k] < second.Index[k]
			break
		}
	}
	return result
}

// BulkFieldSet is the bulk query fieldset container. It stores unique
type BulkFieldSet struct {
	FieldSets []FieldSet       `json:"fieldSets"`
	Indices   map[string][]int `json:"indices"`
}

// CheckFieldset checks if the fieldset exists in given bulk fieldset.
func (b *BulkFieldSet) CheckFieldset(fieldSet FieldSet) bool {
	_, ok := b.Indices[fieldSet.Hash()]
	return ok
}

// GetIndicesByFieldset gets indices by provided fieldset.
func (b *BulkFieldSet) GetIndicesByFieldset(fieldSet FieldSet) []int {
	return b.Indices[fieldSet.Hash()]
}

// Add adds the model index to provided fieldset mapping. If the fieldSet was not stored yet it would be added
// to the slice of field sets.
func (b *BulkFieldSet) Add(fieldSet FieldSet, index int) {
	hash := fieldSet.Hash()
	if b.Indices == nil {
		b.Indices = map[string][]int{}
	}
	indices, ok := b.Indices[hash]
	if !ok {
		b.FieldSets = append(b.FieldSets, fieldSet)
	}
	b.Indices[hash] = append(indices, index)
}
