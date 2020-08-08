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
		return errors.Wrapf(ErrInvalidFieldSet, "provided no fields")
	}

	currentFieldset, hasCommon := s.CommonFieldSet()
	if hasCommon || len(s.FieldSets) > 1 {
		return errors.WrapDetf(ErrInvalidFieldSet, "cannot select fields for multiple field sets")
	}

	for _, field := range fields {
		if field.Struct() != s.ModelStruct {
			return errors.Wrapf(ErrInvalidField, "provided field: '%s' does not belong to model: '%s'", field, s.ModelStruct)
		}
		if currentFieldset.Contains(field) {
			log.Debugf("Field: '%s' is already included in the scope's fieldset", field)
			continue
		}
		currentFieldset = append(currentFieldset, field)
	}
	s.FieldSets = append(s.FieldSets, currentFieldset)
	return nil
}

// CommonFieldSet gets the common fieldset for all models. CommonField
func (s *Scope) CommonFieldSet() (mapping.FieldSet, bool) {
	if len(s.FieldSets) != 1 {
		return nil, false
	}
	return s.FieldSets[0], true
}
