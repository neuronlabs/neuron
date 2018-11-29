package jsonapi

import (
	"fmt"
	"github.com/kucjac/jsonapi/flags"
	"golang.org/x/text/language"
	"reflect"
	"strconv"
	"strings"
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

	OpIsNull
	OpNotNull
	OpExists
	OpNotExists
)

const (
	highestOperator = OpNotExists
)

var (
	operatorsValue = map[string]FilterOperator{
		operatorEqual:        OpEqual,
		operatorIn:           OpIn,
		operatorNotEqual:     OpNotEqual,
		operatorNotIn:        OpNotIn,
		operatorGreaterThan:  OpGreaterThan,
		operatorGreaterEqual: OpGreaterEqual,
		operatorLessThan:     OpLessThan,
		operatorLessEqual:    OpLessEqual,
		// stronly
		operatorContains:   OpContains,
		operatorStartsWith: OpStartsWith,
		operatorEndsWith:   OpEndsWith,
		// introduced with maps
		operatorIsNull:    OpIsNull,
		operatorNotNull:   OpNotNull,
		operatorExists:    OpExists,
		operatorNotExists: OpNotExists,
	}
	operatorsStr = []string{
		operatorEqual,
		operatorIn,
		operatorNotEqual,
		operatorNotIn,
		operatorGreaterThan,
		operatorGreaterEqual,
		operatorLessThan,
		operatorLessEqual,
		operatorContains,
		operatorStartsWith,
		operatorEndsWith,
		operatorIsNull,
		operatorNotNull,
		operatorExists,
		operatorNotExists,
	}
)

func (f FilterOperator) isBasic() bool {
	return f == OpEqual || f == OpNotEqual
}

func (f FilterOperator) isRangable() bool {
	return f >= OpGreaterThan && f <= OpLessEqual
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

	// Key is used for the map type filters as a nested argument
	Key string

	// AttrFilters are the filter values for given attribute FilterField
	Values []*FilterValues

	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	Nested []*FilterField
}

func (f *FilterField) copy() *FilterField {
	dst := &FilterField{StructField: f.StructField}

	if len(f.Values) != 0 {
		dst.Values = make([]*FilterValues, len(f.Values))
		for i, value := range f.Values {
			dst.Values[i] = value.copy()
		}
	}

	if len(f.Nested) != 0 {
		dst.Nested = make([]*FilterField, len(f.Nested))
		for i, value := range f.Nested {
			dst.Nested[i] = value.copy()
		}
	}
	return dst
}

// getNested gets the nested field by it's key or structfield
func (f *FilterField) getOrCreateNested(field interface{}) *FilterField {
	if f.isMap() {
		key := field.(string)
		for _, nested := range f.Nested {
			if nested.Key == key {
				return nested
			}
		}
		nested := &FilterField{Key: key}
		f.Nested = append(f.Nested, nested)
		return nested
	} else {
		sField := field.(*StructField)
		for _, nested := range f.Nested {
			if nested.StructField == sField {
				return nested
			}
		}
		nested := &FilterField{StructField: sField}
		f.Nested = append(f.Nested, nested)
		return nested
	}
	return nil
}

// buildFilterField
func buildFilterField(
	s *Scope,
	collection string,
	values []string,
	c *Controller,
	m *ModelStruct,
	f *flags.Container,
	splitted ...string,
) (fField *FilterField, errs []*ErrorObject) {
	var (
		sField    *StructField
		op        FilterOperator
		ok        bool
		fieldName string

		errObj     *ErrorObject
		errObjects []*ErrorObject
		// private function for returning ErrObject
		invalidName = func(fieldName, collection string) {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter field name: '%s' for the '%s' collection.", fieldName, collection)
			errs = append(errs, errObj)
			return
		}
		invalidOperator = func(operator string) {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' for the '%s' field.", operator, fieldName)
			errs = append(errs, errObj)
			return
		}
	)

	// check if any parameters are set for filtering
	if len(splitted) == 0 {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprint("Too few filter parameters. Valid format is: filter[collection][field][subfield|operator]([operator])*.")
		errs = append(errs, errObj)
		return
	}

	// for all cases first value should be a fieldName
	fieldName = splitted[0]

	// check if there is filter value limit
	if fvLimit, ok := f.Get(flags.UseFilterValueLimit); ok && fvLimit {
		if len(values) > c.FilterValueLimit {
			errObj = ErrOutOfRangeQueryParameterValue.Copy()
			errObj.Detail = fmt.Sprintf("The number of values for the filter: %s within collection: %s exceeds the permissible length: '%d'", strings.Join(splitted, annotationSeperator), collection, c.FilterValueLimit)
			errs = append(errs, errObj)
			return
		}
	}

	switch len(splitted) {
	case 1, 2:
		// if there is only one argument it must be an attribute.

		// PrimaryKey
		if fieldName == "id" {
			fField = s.getOrCreateIDFilter()

			// Attributes
		} else if sField, ok = m.attributes[fieldName]; ok {

			// check if it is language field
			if sField.isLanguage() {

				fField = s.getOrCreateLangaugeFilter()

				// check if the field is of map type
			} else if sField.isMap() {
				if len(splitted) == 1 {
					// Map doesn't allow any default parameters
					// i.e. filter[collection][mapfield]=something
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = "Cannot filter field of 'Hashmap' type with no operator"
					errs = append(errs, errObj)
					return
				} else {

					// create the nested map field
					fField = s.getOrCreateAttributeFilter(sField)

					errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
					if errObj != nil {
						errs = append(errs, errObj)
					}
					return
				}

				// otherwise it is a normal attribute
			} else {
				fField = s.getOrCreateAttributeFilter(sField)
			}

			// FilterKey
		} else if sField, ok = m.filterKeys[fieldName]; ok {
			fField = s.getOrCreateFilterKeyFilter(sField)

			// If the value is the foreign key, check if it is allowed to use it for these flags
		} else if sField, ok = m.foreignKeys[fieldName]; ok {
			allowFk, ok := f.Get(flags.AllowForeignKeyFilter)
			if ok && allowFk {
				fField = s.getOrCreateForeignKeyFilter(sField)
			} else {
				invalidName(fieldName, m.collectionType)
				return
			}

			// If the field is a relationship check what additional arguments are provided
		} else if sField, ok = m.relationships[fieldName]; ok {

			// The relationships must contain subfield scope
			if len(splitted) == 1 {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided filter field: '%s' is a relationship. In order to filter a relationship specify the relationship's field. i.e. '/%s?filter[%s][id]=123'", fieldName, m.collectionType, fieldName)
				errs = append(errs, errObj)
				return
			} else {
				fField = s.getOrCreateRelationshipFilter(sField)

				errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
				if errObj != nil {
					errs = append(errs, errObj)
				}
				return
			}

			// if the value is still unknown, then it must be invalidName
		} else {
			invalidName(fieldName, m.collectionType)
			return
		}

		if len(splitted) == 1 {
			op = OpEqual
		} else {
			// it is an attribute filter
			op, ok = operatorsValue[splitted[1]]
			if !ok {
				invalidOperator(splitted[1])
				return
			}
		}
		errObjects = setFilterValues(fField, c, m.collectionType, values, op)
		errs = append(errs, errObjects...)

		// Case with len() = 3 is when the field is relationship or the attribute
		// with nested substruct
	case 3:

		// if the value is a relationship create the nested filter
		if sField, ok = m.relationships[fieldName]; ok {
			fField = s.getOrCreateRelationshipFilter(sField)

			errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
			if errObj != nil {
				errs = append(errs, errObj)
			}

			// if the filter field is an attribute check if it is allowed to
			// use nested fields for given attribute field
		} else if sField, ok = m.attributes[fieldName]; ok {
			// The attribute field must be an map with these many operators
			if sField.isMap() {

				fField = s.getOrCreateAttributeFilter(sField)
				errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
				if errObj != nil {
					errs = append(errs, errObj)
				}
			} else {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Too many parameters for the attribute field: '%s'.", fieldName)
				errs = append(errs, errObj)
				return
			}
		} else {
			invalidName(fieldName, m.collectionType)
			return
		}

	default:
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.
			Sprintf("Too many filter parameters for '%s' collection. ", collection)
		errs = append(errs, errObj)
		_, ok = m.attributes[fieldName]
		if !ok {
			_, ok = m.relationships[fieldName]
			if !ok {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.
					Sprintf("Invalid field name: '%s' for '%s' collection.", fieldName, collection)
				errs = append(errs, errObj)
			}
		}
	}
	return
}

func buildNestedFilter(
	c *Controller,
	f *FilterField,
	values []string,
	splitted ...string,
) (errObj *ErrorObject) {

	// internal variables definitions
	var (
		subfield *StructField
		op       FilterOperator
		ok       bool
	)

	//	internal function definitions
	var (
		addDetails = func() {
			errObj.Detail += fmt.Sprintf("Collection: '%s', Field: '%s'", f.mStruct.collectionType, f.jsonAPIName)
		}

		getSubfield = func() (subfield *StructField, eObj *ErrorObject) {
			var ok bool
			if splitted[0] == "id" {
				subfield = f.relatedStruct.primary
				return
			}
			subfield, ok = f.relatedStruct.attributes[splitted[0]]
			if !ok {
				eObj = ErrInvalidQueryParameter.Copy()
				eObj.Detail = fmt.Sprintf("The subfield name: '%s' is invalid.", splitted[0])
				return
			}
			return
		}
	)

	// if the field is of relationship kind
	if f.IsRelationship() {
		// Get the subfield
		subfield, errObj = getSubfield()
		if errObj != nil {
			addDetails()
			return
		}

		filter := f.getOrCreateNested(subfield)

		// Check if the subfield is a map type
		if subfield.isMap() {

			// the splitted[0] is hash map and there is no operator
			if len(splitted) == 1 {

				// if no operators provided the value is an error

				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The filter subfield: '%s' is of HashMap type. This field type requires at least 'key' filter.", splitted[0])
				addDetails()
				return errObj

				// the splitted[0] is map, check the splitted[1] if it is an operator or key
			} else {

				// otherwise check if the filter contains map->key or the operator
				_, ok := operatorsValue[splitted[1]]
				if ok {

					// filter[colleciton][relationship][map][$eq]

					// the filter to the map cannot be compared with an operator (with no key)
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprint("The HashMap type field cannot be comparable using raw operators. Add 'key' filter to the query.")
					addDetails()
					return
				}

				// otherwise the value must be a map - key
				// filter[collection][relationship][map][attr]
				if len(values) > 1 {
					op = OpIn
				} else {
					op = OpEqual
				}

				// Get the Subfield Filtervalue
				fv := &FilterValues{Operator: op}
				for _, str := range values {
					fv.Values = append(fv.Values, str)
				}

				// getOrCreateNested field for given 'key'
				nested := filter.getOrCreateNested(splitted[1])

				// add the 'key' values
				nested.Values = append(nested.Values, fv)

			}

			// if the relatinoship subfield is not a map - create just raw relationship subfield
		} else {

			// when no operator is provided set based on the value length
			if len(splitted) == 1 {
				if len(values) > 1 {
					op = OpIn
				} else {
					op = OpEqual
				}

				// otherwise get it from the splitted[1]
			} else {
				operator := splitted[1]
				op, ok = operatorsValue[operator]
				if !ok {
					// invalid operator is provided or nested structure.
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' on collection:'%s'.", operator, f.mStruct.collectionType)
					return errObj
				}
			}

			// set filter values for the given filter field
			errs := setFilterValues(filter, c, f.mStruct.collectionType, values, op)
			if len(errs) > 0 {
				return errs[0]
			}
			// no need to add subfield filter, it is already set by the function getOrCreateNested
		}

		// if the filter field is a map type attribute, check the operators or keys
	} else if f.isMap() {

		// if the filter contains single parameter over the map
		if len(splitted) == 1 {
			// otherwise check if the filter contains map->key or the operator
			_, ok = operatorsValue[splitted[0]]
			if ok {

				// the filter to the map must be a
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprint("The HashMap type field cannot be comparable using raw operators. Add 'key' filter to the query.")
				addDetails()
				return
			}

			// otherwise the value must be a map - key
			if len(values) > 1 {
				op = OpIn
			} else {
				op = OpEqual
			}
			key := splitted[0]

			// get or create (if already exists) nested filter field
			filter := f.getOrCreateNested(key)

			// when there are more filter paramters over the map
			fv := &FilterValues{Operator: op}
			for _, v := range values {
				fv.Values = append(fv.Values, v)
			}

			filter.Values = append(filter.Values, fv)
			// the map filter should be done already

			// if the map filter contains the additional arguments they
			// should be of type: 'filter[collection][map][attr][$operator]'
		} else {

			// splitted[1] is a map -> key
			key := splitted[0]

			op, ok := operatorsValue[splitted[1]]
			if !ok {
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' on collection:'%s, field: %v, and key: %s.'.", splitted[1], f.mStruct.collectionType, f.jsonAPIName, key)
				return errObj
			}

			filter := f.getOrCreateNested(key)

			fv := &FilterValues{Operator: op}
			for _, v := range values {
				fv.Values = append(fv.Values, v)
			}

			// the filter
			filter.Values = append(filter.Values, fv)
		}

		// no other filter field type is known
	} else {
		// error
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The filter over field: %v contains too many arguments.", f.jsonAPIName)
		return errObj
	}

	return nil
}

func setFilterValues(
	f *FilterField,
	c *Controller,
	collection string,
	values []string,
	op FilterOperator,
) (errs []*ErrorObject) {
	var (
		er     error
		errObj *ErrorObject

		opInvalid = func() {
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.jsonAPIName)
			errs = append(errs, errObj)
		}
	)

	if op > highestOperator {
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
		if f.isLanguage() {
			switch op {
			case OpIn, OpEqual, OpNotIn, OpNotEqual:
				for i, value := range values {
					tag, err := language.Parse(value)
					if err != nil {
						switch v := err.(type) {
						case language.ValueError:
							errObj := ErrLanguageNotAcceptable.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s', is not a valid language. Cannot recognize subfield: '%s'.", value, f.GetJSONAPIName(),
								collection, v.Subtag())
							errs = append(errs, errObj)
							continue
						default:
							errObj := ErrInvalidQueryParameter.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%v' for the '%s' filter field within the collection '%s' is not syntetatically valid.", value, f.GetJSONAPIName(), collection)
							errs = append(errs, errObj)
							continue
						}
					}
					if op == OpEqual {
						var confidence language.Confidence
						tag, _, confidence = c.Matcher.Match(tag)
						if confidence <= language.Low {
							errObj := ErrLanguageNotAcceptable.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s' does not match any supported langauges. The server supports following langauges: %s", value, f.GetJSONAPIName(), collection,
								strings.Join(c.displaySupportedLanguages(), annotationSeperator),
							)
							errs = append(errs, errObj)
							return
						}
					}
					b, _ := tag.Base()
					values[i] = b.String()

				}
			default:
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided operator: '%s' for the language field is not acceptable", op.String())
				errs = append(errs, errObj)
				return
			}
		}

		// otherwise it should
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er = setAttributeField(value, fieldValue)
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.jsonAPIName, collection, er)
				errs = append(errs, errObj)
				continue
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.Values = append(f.Values, fv)

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
