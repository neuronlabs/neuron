package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// IncludedRelation is the includes information scope
// it contains the field to include from the root scope
// related subScope, and subfields to include.
type IncludedRelation struct {
	StructField       *mapping.StructField
	Fieldset          mapping.FieldSet
	IncludedRelations []*IncludedRelation
}

func (i *IncludedRelation) copy() *IncludedRelation {
	copiedIncludedField := &IncludedRelation{StructField: i.StructField}
	if i.Fieldset != nil {
		copiedIncludedField.Fieldset = make([]*mapping.StructField, len(i.Fieldset))
		for index, v := range i.Fieldset {
			copiedIncludedField.Fieldset[index] = v
		}
	}
	if i.IncludedRelations != nil {
		copiedIncludedField.IncludedRelations = make([]*IncludedRelation, len(i.IncludedRelations))
		for index, v := range i.IncludedRelations {
			copiedIncludedField.IncludedRelations[index] = v.copy()
		}
	}
	return copiedIncludedField
}

// Select sets the fieldset for given included
func (i *IncludedRelation) SetFieldset(fields ...*mapping.StructField) error {
	model := i.StructField.Relationship().RelatedModelStruct()
	for _, field := range fields {
		// Check if the field belongs to the relationship's model.
		if field.Struct() != model {
			return errors.WrapDetf(ErrInvalidField, "provided field: '%s' doesn't belong to the relationship model: '%s'", field, model)
		}
		// Check if provided field is not a relationship.
		if !field.IsField() {
			return errors.WrapDetf(ErrInvalidField,
				"provided invalid field: '%s' in the fieldset of included field: '%s'", field, i.StructField.Name())
		}

		// Check if given fieldset doesn't contain this field already.
		if i.Fieldset.Contains(field) {
			return errors.WrapDetf(ErrInvalidField,
				"provided field: '%s' in the fieldset of included field: '%s' is already in the included fieldset",
				field, i.StructField.Name())
		}
		i.Fieldset = append(i.Fieldset, field)
	}
	return nil
}

// Include includes 'relation' field in the scope's query results.
func (s *Scope) Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	if !relation.IsRelationship() {
		return errors.WrapDetf(ErrInvalidField,
			"included relation: '%s' is not found for the model: '%s'", relation, s.ModelStruct.String())
	}

	includedField := &IncludedRelation{StructField: relation}
	if len(relationFieldset) == 0 {
		includedField.Fieldset = append(mapping.FieldSet{}, relation.Relationship().RelatedModelStruct().Fields()...)
	} else if err := includedField.SetFieldset(relationFieldset...); err != nil {
		return err
	}
	s.IncludedRelations = append(s.IncludedRelations, includedField)
	return nil
}
