package jsonapi

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
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
	OpIn
	OpNotEqual
	OpNotIn
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
		annotationIn:           OpIn,
		annotationNotEqual:     OpNotEqual,
		annotationNotIn:        OpNotIn,
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
		annotationIn,
		annotationNotEqual,
		annotationNotIn,
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
		return str
	}
	str = operatorsStr[int(f)]
	return str
}

type FilterValues struct {
	Values   []interface{}
	Operator FilterOperator
}

func (f *FilterValues) copy() *FilterValues {
	fv := &FilterValues{Operator: f.Operator}
	fv.Values = make([]interface{}, len(f.Values))
	copy(fv.Values, f.Values)
	return fv
}

// FilterField is a field that contains information about filters
type FilterField struct {
	*StructField

	// AttrFilters are the filter values for given attribute FilterField
	Values []*FilterValues

	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	Relationships []*FilterField
}

func (f *FilterField) copy() *FilterField {
	dst := &FilterField{StructField: f.StructField}

	if len(f.Values) != 0 {
		dst.Values = make([]*FilterValues, len(f.Values))
		for i, value := range f.Values {
			dst.Values[i] = value.copy()
		}
	}

	if len(f.Relationships) != 0 {
		dst.Relationships = make([]*FilterField, len(f.Relationships))
		for i, value := range f.Relationships {
			dst.Relationships[i] = value.copy()
		}
	}
	return dst
}

// setValues set the string type values to the related field values
func (f *FilterField) setValues(collection string, values []string, op FilterOperator,
) (errs []*ErrorObject) {
	// var errObj *ErrorObject
	var (
		er     error
		errObj *ErrorObject

		opInvalid = func() {
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.jsonAPIName)
			errs = append(errs, errObj)
		}
	)

	if op > OpEndsWith {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported by the server.", op)
	}

	t := f.getDereferencedType()
	// create new FilterValue
	fv := new(FilterValues)
	fv.Operator = op

	// Add and check all values for given field type
	switch f.fieldType {
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
			er = setPrimaryField(value, fieldValue)
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for primary field in collection: '%s'. %s. ", collection, er)
				errs = append(errs, errObj)
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.Values = append(f.Values, fv)

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
			er = setAttributeField(value, fieldValue)
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.jsonAPIName, collection, er)
				errs = append(errs, errObj)
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.Values = append(f.Values, fv)

	}
	return
}

func (f *FilterField) addSubfieldFilter(appendFilter *FilterField) {
	for _, rel := range f.Relationships {
		if rel.getFieldIndex() == appendFilter.getFieldIndex() {
			rel.Values = append(rel.Values, appendFilter.Values...)
			return
		}
	}

	// If there is no relationships already add whole filter
	f.Relationships = append(f.Relationships, appendFilter)

	return
}

// Values are the values for given subfield
// Splitted should contain = [relationship's subfield name][operator]
func (f *FilterField) buildSubfieldFilter(values []string, splitted ...string) *ErrorObject {
	var (
		subfield *StructField
		op       FilterOperator
		ok       bool
	)

	getSubfield := func() *ErrorObject {
		if splitted[0] == "id" {
			subfield = f.relatedStruct.primary
		} else {
			subfield, ok = f.relatedStruct.attributes[splitted[0]]
			if !ok {
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The subfield name: '%s' is invalid. Collection: '%s', Field: '%s'", splitted[0], f.jsonAPIName, f.mStruct.collectionType)
				return errObj
			}
		}
		return nil
	}

	switch len(splitted) {
	case 1:
		// Only the relationships fieldname
		if err := getSubfield(); err != nil {
			return err
		}
		if len(values) > 1 {
			op = OpIn
		} else {
			op = OpEqual
		}
	case 2:
		// relationships field name and operator
		if err := getSubfield(); err != nil {
			return err
		}
		operator := splitted[1]
		op, ok = operatorsValue[operator]
		if !ok {
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' on collection:'%s'.", operator, f.mStruct.collectionType)
			return errObj
		}
	}

	filter := &FilterField{StructField: subfield}
	errs := filter.setValues(filter.mStruct.collectionType, values, op)
	if len(errs) > 0 {
		return errs[0]
	}
	f.addSubfieldFilter(filter)
	return nil

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
			values = append(values, bracketed[startIndex+1:endIndex])
		}
	}
	if (startIndex != -1 && endIndex == -1) || startIndex > endIndex {
		err = doubleOpen()
		return
	}
	return
}

func setPrimaryField(value string, fieldValue reflect.Value) (err error) {
	// if the id field is of string type set it to the strValue
	t := fieldValue.Type()

	switch t.Kind() {
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
		/**

		TO DO:

		Panic - recover
		for internals

		*/
		err = IErrInvalidType
		// err = fmt.Errorf("Internal error. Invalid model primary field format: %v", t)
	}
	return
}

func setAttributeField(value string, fieldValue reflect.Value) (err error) {
	// the attribute can be:
	t := fieldValue.Type()
	switch t.Kind() {
	case reflect.Int:
		err = setIntField(value, fieldValue, 64)
	case reflect.Int8:
		err = setIntField(value, fieldValue, 8)
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
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Bool:
		err = setBoolField(value, fieldValue)
	case reflect.Float32:
		err = setFloatField(value, fieldValue, 32)
	case reflect.Float64:
		err = setFloatField(value, fieldValue, 64)
	case reflect.Struct:
		// check if it is time

		if _, ok := fieldValue.Elem().Interface().(time.Time); ok {
			// it is time
		} else {
			// structs are not allowed as attribute
			err = fmt.Errorf("The struct is not allowed as an attribute. FieldName: '%s'",
				t.Name())
		}
	default:
		// unknown field
		err = fmt.Errorf("Unsupported field type as an attribute: '%s'.", t.Name())
	}
	return
}

func setTimeField(value string, fieldValue reflect.Value) (err error) {
	return
}

func setUintField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var uintValue uint64

	// Parse unsigned int
	uintValue, err = strconv.ParseUint(value, 10, bitSize)

	if err != nil {
		return err
	}

	// Set uint
	fieldValue.SetUint(uintValue)
	return nil
}

func setIntField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var intValue int64
	intValue, err = strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return err
	}

	// Set value if no error
	fieldValue.SetInt(intValue)
	return nil
}

func setFloatField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var floatValue float64

	// Parse float
	floatValue, err = strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}
	fieldValue.SetFloat(floatValue)
	return nil
}

func setBoolField(value string, fieldValue reflect.Value) (err error) {
	var boolValue bool
	// set default if empty
	if value == "" {
		value = "false"
	}
	boolValue, err = strconv.ParseBool(value)
	if err != nil {
		return err
	}
	fieldValue.SetBool(boolValue)
	return nil
}
