package scope

import (
	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"

	"github.com/neuronlabs/neuron-core/internal/query/sorts"
)

// AppendSortFields appends the sortfield to the given scope.
func (s *Scope) AppendSortFields(fromStart bool, sortFields ...*sorts.SortField) {
	if fromStart {
		s.sortFields = append(sortFields, s.sortFields...)
	} else {
		s.sortFields = append(s.sortFields, sortFields...)
	}
}

// BuildSortFields sets the sort fields for given string array with the disallowed foreign keys.
func (s *Scope) BuildSortFields(sortFields ...string) error {
	return s.buildSortFields(true, sortFields...)
}

// CreateSortFields creates the sortfields provided as arguments.
// The 'disallowFK' is a flag that checks if sorting by Foreign Keys are allowed.
func (s *Scope) CreateSortFields(disallowFK bool, sortFields ...string) ([]*sorts.SortField, error) {
	return s.createSortFields(disallowFK, sortFields...)
}

// HaveSortFields boolean that defines if the scope has any sort fields.
func (s *Scope) HaveSortFields() bool {
	return len(s.sortFields) > 0
}

// SortFields return current scope sort fields.
func (s *Scope) SortFields() []*sorts.SortField {
	return s.sortFields
}

func (s *Scope) createSortFields(disallowFK bool, sortFields ...string) ([]*sorts.SortField, error) {
	var (
		order            sorts.Order
		fields           = make(map[string]int)
		errs             errors.MultiError
		sortStructFields []*sorts.SortField
	)

	for _, sort := range sortFields {
		if sort[0] == '-' {
			order = sorts.DescendingOrder
			sort = sort[1:]

		} else {
			order = sorts.AscendingOrder
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

		sortField, err := sorts.New(s.mStruct, sort, disallowFK, order)
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

// setSortFields sets the sort fields for given string array.
func (s *Scope) buildSortFields(disallowFK bool, sortFields ...string) error {
	uniqueSortFields, err := sorts.NewUniques(s.mStruct, disallowFK, sortFields...)
	if err != nil {
		return err
	}
	s.sortFields = uniqueSortFields
	return nil
}
