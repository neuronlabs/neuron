package query

import (
	"sort"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// AutoSelectedFields checks if the scope fieldset was set automatically.
// This function returns false if a user had defined any field in the Fieldset.
func (s *Scope) AutoSelectedFields() bool {
	return s.autoSelectedFields
}

// OrderedFieldset gets the fieldset fields sorted by the struct field's index.
func (s *Scope) OrderedFieldset() (ordered mapping.OrderedFieldset) {
	ordered = mapping.OrderedFieldset(s.Fieldset)
	sort.Sort(ordered)
	return ordered
}

// Select adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (s *Scope) Select(fields ...*mapping.StructField) error {
	if len(fields) == 0 {
		return errors.Newf(class.QueryInvalidField, "provided no fields")
	}
	for _, field := range fields {
		if field.Struct() != s.mStruct {
			return errors.Newf(class.QueryInvalidField, "provided field: '%s' does not belong to model: '%s'", field, s.mStruct)
		}
		if s.Fieldset.Contains(field) {
			log.Debugf("Field: '%s' is already included in the scope's fieldset", field)
			continue
		}
		s.Fieldset = append(s.Fieldset, field)
	}
	return nil
}
