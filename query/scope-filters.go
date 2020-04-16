package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

// ClearFilters clears all scope filters.
func (s *Scope) ClearFilters() {
	s.clearFilters()
}

// Filter parses the filter into the  and adds it to the given scope.
// The 'filter' should be of form:
// 	- Field Operator 					'ID IN', 'Name CONTAINS', 'id in', 'name contains'
//	- Relationship.Field Operator		'Car.UserID IN', 'Car.Doors ==', 'car.user_id >=",
// The field might be a Golang model field name or the neuron name.
func (s *Scope) Filter(filter string, values ...interface{}) error {
	filterField, err := newFilter(s.c, s.mStruct, filter, values...)
	if err != nil {
		log.Debugf("SCOPE[%s] Filter '%s' with values: %v failed %v", filter, values, err)
		return err
	}
	return s.addFilterField(filterField)
}

// FilterField adds the filter field to the given query.
func (s *Scope) FilterField(filter *FilterField) error {
	return s.addFilterField(filter)
}

/**

Private filter methods and functions

*/

func (s *Scope) addFilterField(filter *FilterField) error {
	if filter.StructField.Struct() != s.mStruct {
		log.Debugf("Filter's ModelStruct does not match scope's model. Scope's Model: %v, filterField: %v, filterModel: %v", s.mStruct.Type().Name(), filter.StructField.Name(), filter.StructField.Struct().Type().Name())
		err := errors.NewDet(class.QueryFitlerNonMatched, "provided filter field's model structure doesn't match scope's model")
		return err
	}
	switch filter.StructField.Kind() {
	case mapping.KindPrimary:
		s.PrimaryFilters = append(s.PrimaryFilters, filter)
	case mapping.KindAttribute:
		s.AttributeFilters = append(s.AttributeFilters, filter)
	case mapping.KindForeignKey:
		s.ForeignFilters = append(s.ForeignFilters, filter)
	case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
		s.RelationFilters = append(s.RelationFilters, filter)
	default:
		err := errors.NewDetf(class.QueryFilterFieldKind, "unknown field kind: %v", filter.StructField.Kind())
		return err
	}
	return nil
}

func (s *Scope) clearFilters() {
	s.AttributeFilters = Filters{}
	s.ForeignFilters = Filters{}
	s.PrimaryFilters = Filters{}
	s.RelationFilters = Filters{}
}

func (s *Scope) getOrCreatePrimaryFilter() *FilterField {
	if s.PrimaryFilters == nil {
		s.PrimaryFilters = Filters{}
	}

	for _, pf := range s.PrimaryFilters {
		if pf.StructField == s.Struct().Primary() {
			return pf
		}
	}

	// if not found within primary filters
	filter := &FilterField{StructField: s.Struct().Primary()}
	s.PrimaryFilters = append(s.PrimaryFilters, filter)
	return filter
}

func (s *Scope) getOrCreateAttributeFilter(sField *mapping.StructField) *FilterField {
	if s.AttributeFilters == nil {
		s.AttributeFilters = Filters{}
	}

	for _, attrFilter := range s.AttributeFilters {
		if attrFilter.StructField == sField {
			return attrFilter
		}
	}
	filter := &FilterField{StructField: sField}
	s.AttributeFilters = append(s.AttributeFilters, filter)
	return filter
}

// getOrCreateForeignKeyFilter gets the filter field for given StructField
// If the filterField already exists for given scope, the function returns the existing one.
// Otherwise it creates new filter field and returns it.
func (s *Scope) getOrCreateForeignKeyFilter(sField *mapping.StructField) *FilterField {
	if s.ForeignFilters == nil {
		s.ForeignFilters = Filters{}
	}

	for _, fkFilter := range s.ForeignFilters {
		if fkFilter.StructField == sField {
			return fkFilter
		}
	}
	filter := &FilterField{StructField: sField}
	s.ForeignFilters = append(s.ForeignFilters, filter)
	return filter
}

func (s *Scope) getOrCreateRelationshipFilter(sField *mapping.StructField) *FilterField {
	// Create if empty
	if s.RelationFilters == nil {
		s.RelationFilters = Filters{}
	}

	// Check if no relationship filter already exists
	for _, relFilter := range s.RelationFilters {
		if relFilter.StructField == sField {
			return relFilter
		}
	}

	filter := &FilterField{StructField: sField}
	s.RelationFilters = append(s.RelationFilters, filter)
	return filter
}

// setBelongsToForeignKeyFields sets the foreign key fields for the 'belongs to' relationships.
func (s *Scope) setBelongsToForeignKeyFields() error {
	if s.Value == nil {
		return errors.NewDet(class.QueryNoValue, "nil query scope value provided")
	}

	setField := func(v reflect.Value) ([]*mapping.StructField, error) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Type() != s.Struct().Type() {
			return nil, errors.NewDet(class.QueryValueType, "model's struct mismatch")
		}

		var fks []*mapping.StructField
		for _, field := range s.Fieldset {
			relField, ok := s.mStruct.RelationField(field.NeuronName())
			if ok {
				rel := relField.Relationship()
				if rel != nil && rel.Kind() == mapping.RelBelongsTo {
					relVal := v.FieldByIndex(relField.ReflectField().Index)

					// Check if the value is non zero
					if reflect.DeepEqual(
						relVal.Interface(),
						reflect.Zero(relVal.Type()).Interface(),
					) {
						// continue if non zero
						continue
					}

					if relVal.Kind() == reflect.Ptr {
						relVal = relVal.Elem()
					}

					fkVal := v.FieldByIndex(rel.ForeignKey().ReflectField().Index)
					relPrim := rel.Struct().Primary()

					relPrimVal := relVal.FieldByIndex(relPrim.ReflectField().Index)
					fkVal.Set(relPrimVal)
					fks = append(fks, rel.ForeignKey())
				}
			}
		}
		return fks, nil
	}

	v := reflect.ValueOf(s.Value)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		fks, err := setField(v)
		if err != nil {
			return err
		}
		for _, fk := range fks {
			if _, found := s.Fieldset[fk.NeuronName()]; !found {
				s.setFieldsetNoCheck(fk)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.IsNil() {
				continue
			}
			fks, err := setField(elem)
			if err != nil {
				return err
			}
			for _, fk := range fks {
				if _, found := s.Fieldset[fk.NeuronName()]; !found {
					s.setFieldsetNoCheck(fk)
				}
			}
		}
	}
	return nil
}

// setFiltersTo set the filters to the scope with the same model struct.
func (s *Scope) setFiltersTo(to *Scope) error {
	if s.mStruct != to.mStruct {
		log.Errorf("SetFiltersTo mismatch scope's struct. Is: '%s' should be: '%s'", to.mStruct.Collection(), s.mStruct.Collection())
		return errors.NewDet(class.InternalQueryModelMismatch, "scope's model mismatch")
	}

	to.PrimaryFilters = s.PrimaryFilters
	to.AttributeFilters = s.AttributeFilters
	to.RelationFilters = s.RelationFilters
	to.ForeignFilters = s.ForeignFilters

	return nil
}

func (s *Scope) setPrimaryFilterValues(values ...interface{}) {
	filter := s.getOrCreatePrimaryFilter()
	filter.Values = append(filter.Values, OperatorValues{Operator: OpIn, Values: values})
}
