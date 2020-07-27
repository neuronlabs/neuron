package mapping

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

// AddIndex adds the model index to provided fieldset mapping. If the fieldSet was not stored yet it would be added
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
