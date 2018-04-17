package jsonapi

import (
	"fmt"
	"reflect"
)

/**

TO IMPLEMENT:

- checking filter field for possibility of filter operator
- checking for the negating fields i.e.
- checking for the shortcuts i.e. [not][gt] = le
- disabling / controlling filtering in jsonapi struct tags
- design filtering policy.

*/

type FilterOperator int

const (
	// Logical Operators
	OpEqual FilterOperator = iota
	OpNotEqual
	OpGreaterThan
	OpGreaterEqual
	OpLessThan
	OpLessEqual

	// Strings Only operators
	OpContains
	OpStartsWith
	OpEndsWith
)

var (
	operatorsValue = map[string]FilterOperator{
		annotationEqual:        OpEqual,
		annotationNotEqual:     OpNotEqual,
		annotationGreaterThan:  OpGreaterThan,
		annotationGreaterEqual: OpGreaterEqual,
		annotationLessThan:     OpLessThan,
		annotationLessEqual:    OpLessEqual,
		// stronly
		annotationContains:   OpContains,
		annotationStartsWith: OpStartsWith,
		annotationEndsWith:   OpEndsWith,
	}
	operatorsStr = []string{
		annotationEqual,
		annotationNotEqual,
		annotationGreaterThan,
		annotationGreaterEqual,
		annotationLessThan,
		annotationLessEqual,
		annotationContains,
		annotationStartsWith,
		annotationEndsWith,
	}
)

func (f FilterOperator) isBasic() bool {
	return f == OpEqual || f == OpNotEqual
}

func (f FilterOperator) isRangable() bool {
	return f > OpNotEqual && f < OpContains
}

func (f FilterOperator) isStringOnly() bool {
	return f >= OpContains && f <= OpEndsWith
}

func (f FilterOperator) String() string {
	var str string
	if int(f) > len(operatorsStr)-1 || int(f) < 0 {
		str = "unknown operator"
	}
	str = operatorsStr[int(f)]
	return str
}

type FilterValues struct {
	Values   []interface{}
	Operator FilterOperator
}

type FilterField struct {
	*StructField

	// PrimFilters are the filter values for the primary field
	PrimFilters []*FilterValues

	// AttrFilters are the filter values for given attribute FilterField
	AttrFilters []*FilterValues

	// RelFilters are the filter values for given relationship FilterField
	RelFilters []*FilterField
}

func (f *FilterField) checkValues() (errs []*ErrorObject) {
	// check the length of the values provided
	return
}

// setValues set the string type values to the related field values
func (f *FilterField) setValues(collection string, values []string, op FilterOperator,
) (errs []*ErrorObject, err error) {
	// var errObj *ErrorObject
	var (
		internal bool
		er       error
		errObj   *ErrorObject

		opInvalid = func() {
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.jsonAPIName)
			errs = append(errs, errObj)
		}
	)

	if op > OpEndsWith {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprint("The filter operator: '%s' is not supported by the server.", op)
	}

	t := f.getDereferencedType()
	// create new FilterValue
	fv := new(FilterValues)
	fv.Operator = op

	// Add and check all values for given field type
	switch f.jsonAPIType {
	case Primary:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if op > OpLessEqual {
				opInvalid()
			}
		case reflect.String:
			if !op.isBasic() {
				opInvalid()
			}
		}
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er, internal = setPrimaryField(value, fieldValue)
			if internal {
				err = er
				return
			}
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for primary field for collection: '%s'. %s. ", collection, er)
				errs = append(errs, errObj)
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.PrimFilters = append(f.PrimFilters, fv)

		// if it is of integer type check which kind of it
	case Attribute:
		switch t.Kind() {
		case reflect.String:
		default:
			if op.isStringOnly() {
				opInvalid()
			}
		}
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er, internal = setAttributeField(value, fieldValue)
			if internal {
				err = er
				return
			}
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.jsonAPIName, collection, er)
				errs = append(errs, errObj)
			}
			fv.Values = append(fv.Values, fieldValue)
		}

		f.AttrFilters = append(f.AttrFilters, fv)
	case RelationshipSingle, RelationshipMultiple:
		errObj = ErrInternalError.Copy()
		errs = append(errs, errObj)
		err = fmt.Errorf("Setting values for the relationship field directly!: FieldName: %s, Collection: '%s'", f.fieldName, collection)
		return
	default:
		errObj = ErrInternalError.Copy()
		errs = append(errs, errObj)
		err = fmt.Errorf("JSONAPIType not set for this field -  index:%d, name: %s", f.getFieldIndex(), f.fieldName)
	}
	return
}

func (f *FilterField) appendRelFilter(appendFilter *FilterField) {
	var found bool
	for _, rel := range f.RelFilters {
		if rel.getFieldIndex() == appendFilter.getFieldIndex() {
			found = true
			if l := len(appendFilter.PrimFilters); l > 0 {
				rel.PrimFilters = append(rel.PrimFilters, appendFilter.PrimFilters[l-1])
			}

			if l := len(appendFilter.AttrFilters); l > 0 {
				rel.AttrFilters = append(rel.AttrFilters, appendFilter.AttrFilters[l-1])
			}
		}
	}
	if !found {
		f.RelFilters = append(f.RelFilters, appendFilter)
	}
	return
}

func splitBracketParameter(bracketed string) (values []string, err error) {
	// look for values in
	doubleOpen := func() error {
		return fmt.Errorf("Open square bracket '[' found, without closing ']' in: '%s'.",
			bracketed)
	}

	var startIndex int = -1
	var endIndex int = -1
	for i := 0; i < len(bracketed); i++ {
		c := bracketed[i]
		switch c {
		case annotationOpenedBracket:
			if startIndex > endIndex {
				err = doubleOpen()
				return
			}
			startIndex = i
		case annotationClosedBracket:
			// if opening bracket not set or in case of more than one brackets
			// if start was not set before this endIndex
			if startIndex == -1 || startIndex < endIndex {
				err = fmt.Errorf("Close square bracket ']' found, without opening '[' in '%s'.", bracketed)
				return
			}
			endIndex = i
			values = append(values, bracketed[startIndex:endIndex])
		}
	}
	if (startIndex != -1 && endIndex == -1) || startIndex > endIndex {
		err = doubleOpen()
		return
	}
	return
}
