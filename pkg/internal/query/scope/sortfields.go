package scope

import (
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/internal/query/sorts"
	"strings"
)

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
			if scope.sortFields[i].StructField().FieldIndex() == sField.FieldIndex() {
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
