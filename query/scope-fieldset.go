package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// Select adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (s *Scope) Select(fields ...*mapping.StructField) error {
	if len(fields) == 0 {
		return errors.Newf(ClassInvalidFieldSet, "provided no fields")
	}
	for _, field := range fields {
		if field.Struct() != s.ModelStruct {
			return errors.Newf(ClassInvalidField, "provided field: '%s' does not belong to model: '%s'", field, s.ModelStruct)
		}
		if s.FieldSet.Contains(field) {
			log.Debugf("Field: '%s' is already included in the scope's fieldset", field)
			continue
		}
		s.FieldSet = append(s.FieldSet, field)
	}
	return nil
}
