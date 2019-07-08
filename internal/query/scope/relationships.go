package scope

import (
	"reflect"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/models"
)

func setBelongsToForeigns(m *models.ModelStruct, v reflect.Value) error {
	// dereference the pointer
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// check if belongs to model match the scope's value
	if v.Type() != m.Type() {
		return errors.Newf(class.QueryValueType, "invalid model value type. Wanted: %v. Actual: %v", m.Type().Name(), v.Type().Name())
	}

	for _, rel := range m.RelationshipFields() {
		// check if the relationship is not a RelBelongsTo
		if rel.Relationship().Kind() != models.RelBelongsTo {
			continue
		}

		// check if value is a reflect zero
		relVal := v.FieldByIndex(rel.ReflectField().Index)
		if reflect.DeepEqual(relVal.Interface(), reflect.Zero(relVal.Type()).Interface()) {
			continue
		}

		//
		if relVal.Kind() == reflect.Ptr {
			relVal = relVal.Elem()
		}

		// get the foreign key value
		fkVal := v.FieldByIndex(rel.Relationship().ForeignKey().ReflectField().Index)

		// set the relation primary value to the foreign key
		relationPrimary := rel.Relationship().Struct().PrimaryField()
		relPrimVal := relVal.FieldByIndex(relationPrimary.ReflectField().Index)
		fkVal.Set(relPrimVal)
	}
	return nil
}

func setBelongsToRelationWithFields(m *models.ModelStruct, v reflect.Value, fields ...*models.StructField) error {
	// dereference the pointer value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// check if the model's type matches the
	if v.Type() != m.Type() {
		return errors.Newf(class.QueryValueType, "invalid query value type. Wanted: %v. Actual: %v", m.Type().Name(), v.Type().Name())
	}

	// iterate over provided fields
	for _, field := range fields {
		// check if the field is a relationship
		rel, ok := m.RelationshipField(field.NeuronName())
		if !ok {
			continue
		}

		if rel.Relationship() == nil {
			log.Errorf("Relationship field without Relationship: %v", field)
			continue
		}

		if rel.Relationship().Kind() == models.RelBelongsTo {
			// get foreign key value from 'v' value
			fkVal := v.FieldByIndex(rel.Relationship().ForeignKey().ReflectField().Index)

			// get the relationship field
			relVal := v.FieldByIndex(rel.ReflectField().Index)

			// check if the field is a pointer
			relType := relVal.Type()
			if relType.Kind() == reflect.Ptr {
				relType = relType.Elem()
			}

			// if the value is nil create new of given type
			if relVal.IsNil() {
				relVal.Set(reflect.New(relType))
			}

			if relVal.Kind() == reflect.Ptr {
				relVal = relVal.Elem()
			}

			relPrim := relVal.FieldByIndex(rel.Relationship().Struct().PrimaryField().ReflectField().Index)
			relPrim.Set(fkVal)
		}
	}
	return nil
}
