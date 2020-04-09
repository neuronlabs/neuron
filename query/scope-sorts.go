package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Sort adds the sort fields into given scope.
// If the scope already have sorted fields the function appends newly created sort fields.
// If the fields are duplicated returns error.
func (s *Scope) Sort(fields ...string) (err error) {
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("[SCOPE][%s] Sorting by fields: %v ", s.ID(), fields)
	}
	if len(fields) == 0 {
		log.Debug("[SCOPE][%s] - Sort - provided no fields")
		return nil
	}
	if len(s.SortFields) > 0 {
		sortFields, err := s.createSortFields(false, fields...)
		if err != nil {
			return err
		}
		s.SortFields = append(s.SortFields, sortFields...)
		return nil
	}
	sortFields, err := newUniqueSortFields(s.Struct(), false, fields...)
	if err != nil {
		return err
	}
	s.SortFields = append(s.SortFields, sortFields...)
	return nil

}

// SortField adds the sort 'field' to the scope.
func (s *Scope) SortField(field interface{}) error {
	var sortField *SortField

	switch ft := field.(type) {
	case string:
		sf, err := s.createSortFields(false, ft)
		if err != nil {
			return err
		}
		sortField = sf[0]
	case *SortField:
		sortField = ft
	default:
		return errors.NewDetf(class.QuerySortField, "invalid sort field type: %T", field)
	}
	s.SortFields = append(s.SortFields, sortField)
	return nil
}

/**

Private scope sort methods

*/

func (s *Scope) createSortFields(disallowFK bool, sortFields ...string) ([]*SortField, error) {
	var (
		order            SortOrder
		fields           = make(map[string]int)
		errs             errors.MultiError
		sortStructFields []*SortField
	)

	for _, sort := range sortFields {
		if sort[0] == '-' {
			order = DescendingOrder
			sort = sort[1:]

		} else {
			order = AscendingOrder
		}

		// check if no dups provided
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			if count == 2 {
				er := errors.NewDet(class.QuerySortField, "duplicated sort field")
				er.SetDetailsf("Sort parameter: %v used more than once.", sort)
				errs = append(errs, er)
				continue
			} else if count > 2 {
				break
			}
		}

		sortField, err := newStringSortField(s.mStruct, sort, order, disallowFK)
		if err != nil {
			return nil, err
		}
		sortStructFields = append(sortStructFields, sortField)
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return sortStructFields, nil
}
