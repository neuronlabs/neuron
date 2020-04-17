package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/mapping"
)

// Include includes 'relation' field in the scope's query results.
func (s *Scope) Include(relation string, relationFieldset ...string) error {
	structField, ok := s.mStruct.RelationField(relation)
	if !ok {
		return errors.NewDetf(class.QueryRelationNotFound,
			"included relation: '%s' is not found for the model: '%s'", relation, s.mStruct.String())
	}

	includedField := &IncludedField{StructField: structField}
	if len(relationFieldset) == 0 {
		fieldset := map[string]*mapping.StructField{}
		for _, field := range structField.Relationship().Struct().Fields() {
			fieldset[field.NeuronName()] = field
		}
		includedField.Fieldset = fieldset
	} else {
		if err := includedField.setFieldset(relationFieldset...); err != nil {
			return err
		}
	}
	s.IncludedFields = append(s.IncludedFields, includedField)
	return nil
}
