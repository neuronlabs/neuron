package controller

import (
	aerrors "github.com/kucjac/jsonapi/pkg/errors"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/kucjac/jsonapi/pkg/query"
	"reflect"
)

// buildFilterField builds the filter based on the provided query
func buildField(
	s *query.Scope,
	collection string,
	values []string,
	c *Controller,
	m *mapping.ModelStruct,
	f *flags.Container,
	splitted ...string,
) (fField *query.FilterField, errs []*aerrors.ApiError) {
	var (
		sField    *mapping.StructField
		op        query.Operator
		ok        bool
		fieldName string

		errObj     *aerrors.ApiError
		errObjects []*aerrors.ApiError
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
	f *query.FilterField,
	values []string,
	splitted ...string,
) (errObj *aerrors.ApiError) {

	// internal variables definitions
	var (
		subfield *mapping.StructField
		op       query.Operator
		ok       bool
	)

	//	internal function definitions
	var (
		addDetails = func() {
			errObj.Detail += fmt.Sprintf("Collection: '%s', FilterField: '%s'", f.mStruct.collectionType, f.jsonAPIName)
		}

		getSubfield = func() (subfield *mapping.StructField, eObj *aerrors.ApiError) {
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
	f *query.FilterField,
	c *Controller,
	collection string,
	values []string,
	op query.Operator,
) (errs []*aerrors.ApiError) {
	var (
		er     error
		errObj *aerrors.ApiError

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
