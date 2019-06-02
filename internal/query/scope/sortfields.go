package scope

import (
	"fmt"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"strings"
)

/**

SORTS

*/

// AppendSortFields appends the sortfield to the given scope
func (s *Scope) AppendSortFields(fromStart bool, sortFields ...*sorts.SortField) {
	if fromStart {
		s.sortFields = append(sortFields, s.sortFields...)
	} else {
		s.sortFields = append(s.sortFields, sortFields...)
	}
}

// BuildSortFields sets the sort fields for given string array with the disallowed foreign keys.
func (s *Scope) BuildSortFields(sortFields ...string) (errs []*aerrors.ApiError) {
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

// SortFields return current scope sort fields
func (s *Scope) SortFields() []*sorts.SortField {
	return s.sortFields
}

func (s *Scope) createSortFields(disallowFK bool, sortFields ...string) ([]*sorts.SortField, error) {
	var (
		order            sorts.Order
		fields           = make(map[string]int)
		errs             aerrors.MultipleErrors
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
				errs = append(errs, aerrors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort parameter: %v used more than once.", sort)))
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
	return sortStructFields, nil
}

// setSortFields sets the sort fields for given string array.
func (s *Scope) buildSortFields(disallowFK bool, sortFields ...string) (errs []*aerrors.ApiError) {
	var (
		err      *aerrors.ApiError
		order    sorts.Order
		fields   = make(map[string]int)
		badField = func(fieldName string) {
			err = aerrors.ErrInvalidQueryParameter.Copy()
			err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", fieldName, s.mStruct.Collection())
			errs = append(errs, err)
		}
	)

	// If the number of sort fields is too long then do not allow
	if len(sortFields) > s.mStruct.SortScopeCount() {
		err = aerrors.ErrOutOfRangeQueryParameterValue.Copy()
		err.Detail = fmt.Sprintf("There are too many sort parameters for the '%v' collection.", s.mStruct.Collection())
		errs = append(errs, err)
		return
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
				err = aerrors.ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Sort parameter: %v used more than once.", sort)
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
	return
}

func (s *Scope) newSortField(sort string, order sorts.Order, disallowFK bool) (sortField *sorts.SortField, err error) {
	var (
		sField *models.StructField
		ok     bool
	)

	splitted := strings.Split(sort, internal.AnnotationNestedSeperator)
	l := len(splitted)
	switch {
	// for length == 1 the sort must be an attribute, primary or a foreign key field
	case l == 1:
		if sort == internal.AnnotationID {
			sField = models.StructPrimary(s.mStruct)
		} else {
			sField, ok = models.StructAttr(s.mStruct, sort)
			if !ok {
				if disallowFK {
					err = aerrors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection()))
					return
				}

				sField, ok = s.mStruct.ForeignKey(sort)
				if !ok {
					err = aerrors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection()))
					return
				}
			}
		}

		sortField = sorts.NewSortField(sField, order)

	case l <= (sorts.MaxNestedRelLevel + 1):
		sField, ok = models.StructRelField(s.mStruct, splitted[0])
		if !ok {
			err = aerrors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found in the model: '%s'", sort, s.mStruct.Collection()))
			return
		}

		// // if true then the nested should be an attribute for given
		// var found bool
		// for i := range s.sortFields {
		// 	if s.sortFields[i].StructField() == sField {
		// 		sortField = s.sortFields[i]
		// 		found = true
		// 		break
		// 	}
		// }
		// if !found {

		sortField = sorts.NewSortField(sField, order)
		// }

		err = sortField.SetSubfield(splitted[1:], order, disallowFK)
		if err != nil {
			return
		}
		s.sortFields = append(s.sortFields, sortField)

	default:
		err = aerrors.ErrInvalidQueryParameter.Copy().WithDetail(fmt.Sprintf("Sort: field '%s' not found", sort))
	}
	return
}
