package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// OrderBy adds the sort fields into given scope.
// If the scope already have sorted fields the function appends newly created sort fields.
// If the fields are duplicated returns error.
func (s *Scope) OrderBy(fields ...string) error {
	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f(s.logFormat("Sorting by fields: %v "), fields)
	}
	if len(fields) == 0 {
		log.Debug(s.logFormat("OrderBy - provided no fields"))
		return nil
	}
	if len(s.SortingOrder) > 0 {
		sortFields, err := s.createSortFields(fields...)
		if err != nil {
			return err
		}
		s.SortingOrder = append(s.SortingOrder, sortFields...)
		return nil
	}
	sortFields, err := newUniqueSortFields(s.ModelStruct, fields...)
	if err != nil {
		return err
	}
	s.SortingOrder = append(s.SortingOrder, sortFields...)
	return nil
}

/**

Private scope sort methods

*/

func (s *Scope) createSortFields(sortFields ...string) ([]Sort, error) {
	var (
		order            SortOrder
		fields           = make(map[string]int)
		sortStructFields []Sort
	)

	for _, sort := range sortFields {
		if sort[0] == '-' {
			order = DescendingOrder
			sort = sort[1:]
		} else {
			order = AscendingOrder
		}

		// Check if no duplicates provided.
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			return nil, errors.NewDet(ClassInvalidSort, "duplicated sort field").
				WithDetailf("OrderBy parameter: %v used more than once.", sort)
		}

		sortField, err := newStringSortField(s.ModelStruct, sort, order)
		if err != nil {
			return nil, err
		}
		sortStructFields = append(sortStructFields, sortField)
	}
	return sortStructFields, nil
}
