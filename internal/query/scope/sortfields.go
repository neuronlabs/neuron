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

// BuildSortFields sets the sort fields for given string array.
func (s *Scope) BuildSortFields(sortFields ...string) (errs []*aerrors.ApiError) {
	return s.buildSortFields(sortFields...)
}

// SortFields return current scope sort fields
func (s *Scope) SortFields() []*sorts.SortField {
	return s.sortFields
}

// setSortFields sets the sort fields for given string array.
func (s *Scope) buildSortFields(sortFields ...string) (errs []*aerrors.ApiError) {
	var (
		err      *aerrors.ApiError
		order    sorts.Order
		fields   = make(map[string]int)
		badField = func(fieldName string) {
			err = aerrors.ErrInvalidQueryParameter.Copy()
			err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", fieldName, s.mStruct.Collection())
			errs = append(errs, err)
		}
		invalidField bool
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

		invalidField = newSortField(sort, order, s)
		if invalidField {
			badField(sort)
			continue
		}

	}
	return
}

func newSortField(sort string, order sorts.Order, scope *Scope) (invalidField bool) {
	var (
		sField    *models.StructField
		ok        bool
		sortField *sorts.SortField
	)

	splitted := strings.Split(sort, internal.AnnotationNestedSeperator)
	l := len(splitted)
	switch {
	// for length == 1 the sort must be an attribute or a primary field
	case l == 1:
		if sort == internal.AnnotationID {
			sField = models.StructPrimary(scope.mStruct)
		} else {
			sField, ok = models.StructAttr(scope.mStruct, sort)
			if !ok {
				invalidField = true
				return
			}
		}

		sortField = sorts.NewSortField(sField, order)
		scope.sortFields = append(scope.sortFields, sortField)
	case l <= (sorts.MaxNestedRelLevel + 1):
		sField, ok = models.StructRelField(scope.mStruct, splitted[0])
		if !ok {

			invalidField = true
			return
		}
		// if true then the nested should be an attribute for given
		var found bool
		for i := range scope.sortFields {
			if scope.sortFields[i].StructField() == sField {
				sortField = scope.sortFields[i]
				found = true
				break
			}
		}
		if !found {
			sortField = sorts.NewSortField(sField, sorts.AscendingOrder)
		}

		invalidField = sorts.SetSubfield(sortField, splitted[1:], order)
		if !found && !invalidField {
			scope.sortFields = append(scope.sortFields, sortField)
		}
	default:
		invalidField = true
	}
	return
}
