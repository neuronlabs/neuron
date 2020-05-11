package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// ClearFilters clears all scope filters.
func (s *Scope) ClearFilters() {
	s.clearFilters()
}

// Where parses the filter into the  and adds it to the given scope.
// The 'filter' should be of form:
// 	- Field Operator 					'ID IN', 'Name CONTAINS', 'TransactionID in', 'name contains'
//	- Relationship.Field Operator		'Car.UserID IN', 'Car.Doors ==', 'car.user_id >=",
// The field might be a Golang model field name or the neuron name.
func (s *Scope) Where(filter string, values ...interface{}) error {
	filterField, err := newFilter(s.DB().Controller(), s.mStruct, filter, values...)
	if err != nil {
		log.Debugf("SCOPE[%s] Where '%s' with values: %v failed %v", filter, values, err)
		return err
	}
	return s.addFilterField(filterField)
}

// Filter adds the filter field to the given query.
func (s *Scope) Filter(filter *FilterField) error {
	return s.addFilterField(filter)
}

/**

Private filter methods and functions

*/

func (s *Scope) addFilterField(filter *FilterField) error {
	if filter.StructField.Struct() != s.mStruct {
		log.Debugf("Where's ModelStruct does not match scope's model. Scope's Model: %v, filterField: %v, filterModel: %v", s.mStruct.Type().Name(), filter.StructField.Name(), filter.StructField.Struct().Type().Name())
		err := errors.NewDet(ClassInvalidField, "provided filter field's model structure doesn't match scope's model")
		return err
	}
	for _, existingFilter := range s.Filters {
		if existingFilter.StructField == filter.StructField {
			existingFilter.Values = append(existingFilter.Values, filter.Values...)
			existingFilter.Nested = append(existingFilter.Nested, filter.Nested...)
			return nil
		}
	}
	s.Filters = append(s.Filters, filter)
	return nil
}

func (s *Scope) clearFilters() {
	s.Filters = Filters{}
}

func (s *Scope) getOrCreateFieldFilter(structField *mapping.StructField) *FilterField {
	for _, filter := range s.Filters {
		if filter.StructField == structField {
			return filter
		}
	}
	filter := &FilterField{StructField: structField}
	s.Filters = append(s.Filters, filter)
	return filter
}
