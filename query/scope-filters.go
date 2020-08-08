package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/filter"
)

// ClearFilters clears all scope filters.
func (s *Scope) ClearFilters() {
	s.Filters = filter.Filters{}
}

// GetOrCreateRelationFilter gets or creates new filter field for given structField within the scope filters.
func (s *Scope) GetOrCreateRelationFilter(structField *mapping.StructField) filter.Relation {
	for _, f := range s.Filters {
		if relation, ok := f.(filter.Relation); ok && relation.StructField == structField {
			return relation
		}
	}
	f := filter.Relation{StructField: structField}
	s.Filters = append(s.Filters, f)
	return f
}

// Where parses the filter into the  and adds it to the given scope.
// The 'filter' should be of form:
// 	- Field Operator 					'ID IN', 'Name CONTAINS', 'id in', 'name contains'
//	- Relationship.Field Operator		'Car.UserID IN', 'Car.Doors ==', 'car.user_id >=",
// The field might be a Golang model field name or the neuron name.
func (s *Scope) Where(where string, values ...interface{}) error {
	f, err := filter.NewFilter(s.ModelStruct, where, values...)
	if err != nil {
		log.Debug2f("Where '%s' with values: %v failed %v", where, values, err)
		return err
	}
	s.Filters = append(s.Filters, f)
	return nil
}

// Filter adds the filter field to the given query.
func (s *Scope) Filter(f filter.Filter) {
	s.Filters = append(s.Filters, f)
}

func (s *Scope) WhereOr(filters ...filter.Simple) error {
	for i := range filters {
		if filters[i].StructField.ModelStruct() != s.ModelStruct {
			return errors.Wrap(filter.ErrFilterCollection, "Or filter elements have different root model")
		}
	}
	s.Filters = append(s.Filters, filter.OrGroup(filters))
	return nil
}
