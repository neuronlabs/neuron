package scope

import (
	"strings"

	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"

	"github.com/neuronlabs/neuron-core/internal/models"
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
func (s *Scope) BuildSortFields(sortFields ...string) []*errors.Error {
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
				errs = append(errs, errors.New(class.QuerySortField, "duplicated sort field").SetDetailf("Sort parameter: %v used more than once.", sort))
				continue
			} else if count > 2 {
				break
			}
		}

		sortField, err := s.newSortField(sort, order, disallowFK)
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
func (s *Scope) buildSortFields(disallowFK bool, sortFields ...string) []*errors.Error {
	var (
		err   *errors.Error
		errs  []*errors.Error
		order sorts.Order
	)

	fields := make(map[string]int)

	// define the common error adder
	badField := func(fieldName string) {
		err = errors.New(class.QuerySortField, "invalid sort field provided")
		err = err.SetDetailf("Provided sort parameter: '%v' is not valid for '%v' collection.", fieldName, s.mStruct.Collection())
		errs = append(errs, err)
	}

	// If the number of sort fields is too long then do not allow
	if len(sortFields) > s.mStruct.SortScopeCount() {
		err = errors.New(class.QuerySortTooManyFields, "too many sort fields provided for given model")
		err = err.SetDetailf("There are too many sort parameters for the '%v' collection.", s.mStruct.Collection())
		errs = append(errs, err)
		return errs
	}

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
				err = errors.New(class.QuerySortField, "duplicated sort field provided")
				err = err.SetDetailf("Sort parameter: %v used more than once.", sort)
				errs = append(errs, err)
				continue
			} else if count > 2 {
				break
			}
		}

		sortField, err := s.newSortField(sort, order, disallowFK)
		if err != nil {
			badField(sort)
			continue
		}

		s.sortFields = append(s.sortFields, sortField)
	}

	return errs
}

func (s *Scope) newSortField(sort string, order sorts.Order, disallowFK bool) (*sorts.SortField, *errors.Error) {
	var (
		sField    *models.StructField
		sortField *sorts.SortField
		err       *errors.Error
		ok        bool
	)

	splitted := strings.Split(sort, common.AnnotationNestedSeparator)
	l := len(splitted)
	switch {
	// for length == 1 the sort must be an attribute, primary or a foreign key field
	case l == 1:
		if sort == common.AnnotationID {
			sField = s.mStruct.PrimaryField()
		} else {
			sField, ok = s.mStruct.Attribute(sort)
			if !ok {
				if disallowFK {
					err = errors.New(class.QuerySortField, "sort field not found")
					err = err.SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection())
					return nil, err
				}

				sField, ok = s.mStruct.ForeignKey(sort)
				if !ok {
					err = errors.New(class.QuerySortField, "sort field not found")
					err = err.SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection())
					return nil, err
				}
			}
		}

		sortField = sorts.NewSortField(sField, order)
	case l <= (sorts.MaxNestedRelLevel + 1):
		sField, ok = s.mStruct.RelationshipField(splitted[0])
		if !ok {
			err = errors.New(class.QuerySortField, "sort field not found")
			err = err.SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection())
			return nil, err
		}

		sortField = sorts.NewSortField(sField, order)
		err := sortField.SetSubfield(splitted[1:], order, disallowFK)
		if err != nil {
			return nil, err
		}
		s.sortFields = append(s.sortFields, sortField)
	default:
		err = errors.New(class.QuerySortField, "sort field not found")
		err = err.SetDetailf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection())
		return nil, err
	}

	return sortField, nil
}
