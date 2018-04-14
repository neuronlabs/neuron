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
	OpNot
	OpOr
	OpAnd

	// Strings Only operators
	OpContains
	OpStartsWith
	OpEndsWith
)

var operatorsValue = map[string]FilterOperator{
	annotationEqual:        OpEqual,
	annotationNotEqual:     OpNotEqual,
	annotationGreaterThan:  OpGreaterThan,
	annotationGreaterEqual: OpGreaterEqual,
	annotationLessThan:     OpLessThan,
	annotationLessEqual:    OpLessEqual,
	annotationNot:          OpNot,
	annotationOr:           OpOr,
	annotationAnd:          OpAnd,
	annotationContains:     OpContains,
	annotationStartsWith:   OpStartsWith,
	annotationEndsWith:     OpEndsWith,
}

type FilterValue struct {
	Value          []interface{}
	ValueOperators []FilterOperator
}

type FilterField struct {
	*StructField

	// Values are the filter values already checked for the type correction.
	Values []*FilterValue
	values []string
}

func (f *FilterField) checkValues() (errs []*ErrorObject) {
	// check the length of the values provided
	return
}

// setValues set the string type values to the related field values
func (f *FilterField) setValues() (errs []*ErrorObject) {

	// check if the field is relation
	for _, value := range f.values {
		// check if correct and append to values

		// check if value is coorect
		if value == "" {
			// otherwise set an error
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Invalid filter parameter value: %s, for field: %s", value, f.jsonAPIName)
			errs = append(errs, errObj)
			continue
		}
		changedValue := value
		f.Values = append(f.Values, changedValue)
	}
	return
}

func (f *FilterField) setValue(value string) (errs []*ErrorObject, er error) {
	switch f.jsonAPIResKind {
	case annotationPrimary:
		// set id field
		fieldType := f.GetFieldType()

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		fieldValue := reflect.New(fieldType)

		var err error
		// if the id field is of string type set it to the value
		switch fieldType.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int:
			err = setIntField(value, fieldValue, 64)
		case reflect.Int16:
			err = setIntField(value, fieldValue, 16)
		case reflect.Int32:
			err = setIntField(value, fieldValue, 32)
		case reflect.Int64:
			err = setIntField(value, fieldValue, 64)
		case reflect.Uint:
			err = setUintField(value, fieldValue, 64)
		case reflect.Uint16:
			err = setUintField(value, fieldValue, 16)
		case reflect.Uint32:
			err = setUintField(value, fieldValue, 32)
		case reflect.Uint64:
			err = setUintField(value, fieldValue, 64)
		default:
			// should never happen - model checked at precomputation.
			er = fmt.Errorf("Internal error. Invalid model primary field format: %v",
				fieldType)
			return
		}
		fmt.Println(err)
		// if it is of integer type check which kind of it
	case annotationAttribute:
	case annotationRelation:
	}
	return
}

func splitBrackets(bracketed string, parameter string) (values []string, err *ErrorObject) {
	// look for values in
	var startIndex int = -1
	var endIndex int = -1
	for i := 0; i < len(bracketed); i++ {
		c := bracketed[i]
		switch c {
		case annotationOpenedBracket:
			startIndex = i
		case annotationClosedBracket:
			// if opening bracket not set or in case of more than one brackets
			// if start was not set before this endIndex
			if startIndex == -1 || startIndex < endIndex {
				err = ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Invalid '%s' parameter. Close square bracket ']' found, befwithout opening ('[') in '%s'.", parameter, bracketed)
				return
			}
			endIndex = i
			values = append(values, bracketed[startIndex:endIndex])
		}
	}
	if (startIndex != -1 && endIndex == -1) || startIndex > endIndex {
		err = ErrInvalidQueryParameter.Copy()
		err.Detail = fmt.Sprintf("Invalid '%s' parameter. Open square bracket '[' found, without closing (']') in: '%s'.", parameter, bracketed)
		return
	}
}
