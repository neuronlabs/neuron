package controller

import (
	"github.com/jinzhu/inflection"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
	"time"
	"unicode"
)

func (c *Controller) buildModelStruct(model interface{}, modelMap *ModelMap) error {
	var err error

	var modelType reflect.Type
	var modelValue reflect.Value
	modelValue = reflect.ValueOf(model)
	modelType = modelValue.Type()

	if modelType.Kind() == reflect.Ptr {
		modelValue = reflect.ValueOf(model).Elem()
		modelType = modelValue.Type()
	} else {
		return fmt.Errorf("Provide addressable models i.e.: &Model{} in order to precompute it. Invalid model: %v", modelType)
	}

	if modelType.Kind() != reflect.Struct {
		err = fmt.Errorf(`Provided model in invalid format. 
			The model must be of struct or ptr type, but is: %v`, modelType)
		return err
	}

	var collection string

	collectioner, ok := model.(Collectioner)
	if ok {
		collection = collectioner.CollectionName()
	} else {
		collection = c.NamerFunc(inflection.Plural(modelType.Name()))
	}

	modelStruct := models.NewModelStruct(modelType, collection, c.Flags)

	var assignedFields int
	for i := 0; i < modelType.NumField(); i++ {
		// don't use private fields
		if !modelValue.Field(i).CanSet() {
			continue
		}

		tag, ok := modelType.Field(i).Tag.Lookup(internal.AnnotationJSONAPI)
		if !ok {
			continue
		}

		if tag == "-" {
			continue
		}

		var tagValues url.Values
		tField := modelType.Field(i)

		structField := models.NewStructField(tField, modelStruct)
		tagValues, err = models.FieldTagValues(structField, tag)
		if err != nil {
			return errors.Wrapf(err, "Getting tag values failed. Model: %s, SField: %s", modelStruct.modelType.Name(), tField.Name)
		}

		assignedFields++

		// Check if field contains the name
		var apiName string
		name := tagValues.Get(internal.AnnotationName)
		if name != "" {
			apiName = name
		} else {
			apiName = c.NamerFunc(tField.Name)
		}

		models.FieldsSetApiName(structField, apiName)

		// Set field type
		values := tagValues[internal.AnnotationFieldType]
		if len(values) == 0 {
			return errors.Errorf("StructField.annotationFieldType struct field tag cannot be empty. Model: %s, field: %s", modelType.Name(), tField.Name)
		} else {
			// Set field type
			value := values[0]
			switch value {
			case internal.AnnotationPrimary:
				models.FieldSetFieldKind(structField, models.Primary)
				models.StructSetPrimary(modelStruct, structField)
				models.StructAppendField(modelStruct, structField)

			case internal.AnnotationRelation:
				models.StructAppendField(modelStruct, structField)
				err = models.FieldSetRelatedType(structField)
				if err != nil {
					return errors.Wrap(err, "FieldSetRelatedType failed")
				}
				_, ok := models.StructRelField(modelStruct, apiName)
				if ok {
					err = errors.Errorf("Duplicated jsonapi relationship field name: '%s' for model: '%v'.", apiName, modelType.Name())
					return err
				}
				models.StructSetRelField(modelStruct, structField)
			case internal.AnnotationAttribute:
				models.FieldSetFieldKind(structField, models.Attribute)
				// check if no duplicates
				_, ok := models.StructAttr(modelStruct, apiName)
				if ok {
					err = errors.Errorf("Duplicated jsonapi attribute name: '%s' for model: '%v'.",
						structField.jsonAPIName, modelStruct.modelType.Name())
					return err
				}

				models.StructAppendField(modelStruct, structField)

				t := structField.ReflectField().Type
				if t.Kind() == reflect.Ptr {
					models.FieldSetFlag(structField, models.FPtr)
					t = t.Elem()
				}

				switch t.Kind() {
				case reflect.Struct:
					if t == reflect.TypeOf(time.Time{}) {
						models.FieldSetFlag(structField, models.FTime)
					} else {
						// this case it must be a nested struct field
						models.FieldSetFlag(structField, models.FNestedStruct)

						nStruct, err := c.getNestedStruct(t, structField)
						if err != nil {
							c.log().Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
							return err
						}
						models.FieldSetNested(structField, nStruct)
					}

					if models.FieldIsPtr(structField) {
						models.FieldSetFlag(structField, models.FBasePtr)
					}

				case reflect.Map:
					models.FieldSetFlag(structField, models.FMap)

					mapElem := t.Elem()

					// isPtr is a bool that defines if the given slice Elem is a pointer type
					// flags are not set now due to the slice sliceElem possiblities
					var isPtr bool

					// map type must have a key of type string and value of basic type
					if mapElem.Kind() == reflect.Ptr {
						isPtr = true

						mapElem = mapElem.Elem()
					}

					// Check the map 'value' kind
					switch mapElem.Kind() {
					// struct may be time or nested struct
					case reflect.Struct:
						// check if it is time
						if mapElem == reflect.TypeOf(time.Time{}) {
							models.FieldSetFlag(structField, models.FTime)
							// otherwise it must be a nested struct
						} else {
							models.FieldSetFlag(structField, models.FNestedStruct)

							nStruct, err := c.getNestedStruct(mapElem, structField)
							if err != nil {
								c.log().Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
								return err
							}
							models.FieldSetNested(structField, nStruct)
						}
						// if the value is pointer add the base flag
						if isPtr {
							models.FieldSetFlag(structField, models.FBasePtr)
						}

						// map 'value' may be a slice or array
					case reflect.Slice, reflect.Array:
						// TO DO:
						// Support map of slices
						// add flag is slice?
						mapElem = mapElem.Elem()
						for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
							mapElem = mapElem.Elem()
						}

						if mapElem.Kind() == reflect.Ptr {
							models.FieldSetFlag(structField, models.FBasePtr)
							mapElem = mapElem.Elem()
						}

						switch mapElem.Kind() {
						case reflect.Struct:
							// check if it is time
							if mapElem == reflect.TypeOf(time.Time{}) {
								models.FieldSetFlag(structField, models.FTime)
								// otherwise it must be a nested struct
							} else {
								models.FieldSetFlag(structField, models.FNestedStruct)
								nStruct, err := c.getNestedStruct(mapElem, structField)
								if err != nil {
									c.log().Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
									return err
								}
								models.FieldSetNested(structField, nStruct)
							}
						case reflect.Slice, reflect.Array, reflect.Map:
							// disallow nested map, arrs, maps in ptr type slices
							err = errors.Errorf("StructField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map.", structField.Name())
							return err
						default:
						}
					default:
						if isPtr {
							models.FieldSetFlag(structField, models.FBasePtr)
						}

					}
				case reflect.Slice, reflect.Array:
					if t.Kind() == reflect.Slice {
						models.FieldSetFlag(structField, models.FSlice)
					} else {
						models.FieldSetFlag(structField, models.FArray)
					}

					// dereference the slice
					// check the field base type
					sliceElem := t
					for sliceElem.Kind() == reflect.Slice || sliceElem.Kind() == reflect.Array {
						sliceElem = sliceElem.Elem()
					}

					if sliceElem.Kind() == reflect.Ptr {
						// add maybe slice Field Ptr
						models.FieldSetFlag(structField, models.FBasePtr)
						sliceElem = sliceElem.Elem()
					}

					switch sliceElem.Kind() {
					case reflect.Struct:
						// check if time
						if sliceElem == reflect.TypeOf(time.Time{}) {
							models.FieldSetFlag(structField, models.FTime)
						} else {
							// this should be the nested struct
							models.FieldSetFlag(structField, models.FNestedStruct)
							nStruct, err := c.getNestedStruct(sliceElem, structField)
							if err != nil {
								c.log().Debugf("structField: %v getNestedStruct failed. %v", structField.Name(), err)
								return err
							}
							structField.nested = nStruct
						}
					case reflect.Map:
						// map should not be allow as slice nested field
						err = errors.Errorf("Map can't be a base of the slice. Field: '%s'", structField.Name())
						return err
					case reflect.Array, reflect.Slice:
						// cannot use slice of ptr slices
						err = errors.Errorf("Ptr slice can't be the base of the Slice field. Field: '%s'", structField.Name())
						return err
					default:
					}
				default:
					if models.FieldIsPtr(structField) {
						models.FieldSetFlag(structField, models.FBasePtr)
					}

				}
				models.StructSetAttr(modelStruct, structField)
			case internal.AnnotationForeignKey:
				models.FieldSetFieldKind(structField, models.ForeignKey)
				// Check if already exists
				_, ok := models.StructForeignKeyField(modelStruct, structField.ApiName())
				if ok {
					err = errors.Errorf("Duplicated jsonapi foreign key name: '%s' for model: '%v'", structField.ApiName(), modelStruct.Type().Name())
					return err
				}
				models.StructAppendField(modelStruct, structField)
				models.StructSetForeignkey(modelStruct, structField)

			case internal.AnnotationFilterKey:
				models.FieldSetFieldKind(structField, models.FilterKey)
				_, ok := models.StructFilterKeyField(modelStruct, structField.ApiName())
				if ok {
					err = errors.Errorf("Duplicated jsonapi filter key name: '%s' for model: '%v'", structField.ApiName(), modelStruct.Type().Name())
					return err
				}
				// modelStruct.fields = append(modelStruct.fields, structField)
				models.StructSetFilterKey(modelStruct, structField)
			default:
				return errors.Errorf("Unknown field type: %s. Model: %s, field: %s", value, modelStruct.Type().Name(), tField.Name)
			}
		}

		// iterate over structfield tags
		for key, values := range tagValues {
			switch key {
			case internal.AnnotationFieldType, internal.AnnotationName:
				continue
			case internal.AnnotationFlags:
				for _, value := range values {

					switch value {
					case internal.AnnotationClientID:
						models.FieldSetFlag(structField, models.FClientID)
					case internal.AnnotationNoFilter:
						models.FieldSetFlag(structField, models.FNoFilter)
					case internal.AnnotationHidden:
						models.FieldSetFlag(structField, models.FHidden)
					case internal.AnnotationNotSortable:
						models.FieldSetFlag(structField, models.FSortable)
					case internal.AnnotationISO8601:
						models.FieldSetFlag(structField, models.FIso8601)
					case internal.AnnotationOmitEmpty:
						models.FieldSetFlag(structField, models.FOmitempty)
					case internal.AnnotationI18n:
						models.FieldSetFlag(structField, models.FI18n)
						models.StructAppendI18n(modelStruct, structField)
					case internal.AnnotationLanguage:
						models.StructSetLanguage(modelStruct, structField)

					}
				}
			case internal.AnnotationRelation:
				// if relationship match the type e.t.c

				r := models.FieldRelationship(structField)
				if r == nil {
					r := &models.Relationship{}
					models.FieldSetRelationship(structField, r)
				}

				for _, value := range values {
					switch value {
					case internal.AnnotationRelationNoSync:
						b := false
						r.SetSync(&b)
					case internal.AnnotationManyToMany:
						r.SetKind(models.RelMany2Many)
					case internal.AnnotaitonRelationSync:
						b := true
						r.SetSync(&b)
					default:
						c.log().Debugf("Backreference field tag for relation: %s in model: %s. Value: %s", modelStruct.Type().Name(), structField.Name(), value)
						models.RelationshipSetBackrefFieldName(r, value)
					}
				}
				// if field is foreign key match with relationship
			case internal.AnnotationForeignKey:

			}
		}
	}

	if assignedFields == 0 {
		err = fmt.Errorf("Model has no correct jsonapi fields: %v", modelType)
		return err
	}

	if models.StructPrimary(modelStruct) == nil {
		err = fmt.Errorf("Model: %v must have a correct primary field.", modelType)
		return err
	}

	modelMap.Set(modelType, &mapping.ModelStruct{modelStruct})
	return err
}

func (c *Controller) getNestedStruct(
	t reflect.Type, sFielder models.StructFielder,
) (*models.NestedStruct, error) {
	nestedStruct := models.NewNestedStruct(t, sFielder)

	var marshalFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		nField := t.Field(i)

		marshalField := reflect.StructField{
			Name: nField.Name,
			Type: nField.Type,
		}

		if unicode.IsLower(rune(nField.Name[0])) {
			marshalField.Tag = `json:"-"`
			marshalFields = append(marshalFields, marshalField)
			continue
		}

		nestedField := models.NewNestedField(root, sFielder, nField)

		tag, ok := nField.Tag.Lookup("jsonapi")
		if ok {
			if tag == "-" {
				marshalField.Tag = reflect.StructTag(`json:"-"`)
				marshalFields = append(marshalFields, marshalField)
				continue
			}

			tagValues, err := models.FieldTagValues(nestedField.StructField, tag)
			if err != nil {
				c.log().Debugf("nestedField: '%s', getTagValues failed. %v", nestedField.Name())
				return nil, err
			}

			for tKey, tValue := range tagValues {
				switch tKey {
				case internal.AnnotationName:
					models.FieldsSetApiName(nestedField.StructField, tValue[0])
				case internal.AnnotationFieldType:
					if tValue[0] != internal.AnnotationNestedField {
						c.log().Debugf("Invalid annotationNestedField value: '%s' for field: %s", tValue[0], nestedField.Name())
						err = errors.Errorf("Provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.Name())
						return nil, err
					}
				case internal.AnnotationFlags:
					for _, value := range tValue {
						switch value {
						case internal.AnnotationNoFilter:
							models.FieldSetFlag(nestedField.StructField, models.FNoFilter)
						case internal.AnnotationHidden:
							models.FieldSetFlag(nestedField.StructField, models.FHidden)
						case internal.AnnotationNotSortable:
							models.FieldSetFlag(nestedField.StructField, models.FSortable)
						case annotationISO8601:
							models.FieldSetFlag(nestedField.StructField, models.FIso8601)
						case annotationOmitEmpty:
							models.FieldSetFlag(nestedField.StructField, models.FOmitempty)
						}

					}
				}
			}
		}

		if nestedField.ApiName() == "" {
			models.FieldsSetApiName(nestedField.StructField, c.NamerFunc(nField.Name))
		}

		switch nestedField.ApiName() {
		case "relationships", "links":
			return nil, errors.Errorf("Nested field within: '%s' field in the model: '%s' has forbidden API name: '%s'. ",
				models.NestedStructAttr(nestedStruct).Name(),
				models.FieldsStruct(models.NestedStructAttr(nestedStruct)).Type().Name(),
				nestedField.ApiName(),
			)
		default:
		}

		if _, ok = models.NestedStructSubField(nestedStruct, nestedField.ApiName()); ok {
			return nil, errors.Errorf("NestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.Type().Name(), nestedField.Name())
		}
		models.NestedStructSetSubfield(nestedStruct, nestedField)

		nFType := nField.Type
		if nFType.Kind() == reflect.Ptr {
			nFType = nFType.Elem()
			models.FieldSetFlag(nestedField.StructField, models.FPtr)
		}

		switch nFType.Kind() {
		case reflect.Struct:
			if nFType == reflect.TypeOf(time.Time{}) {
				models.FieldSetFlag(nestedField.StructField, models.FTime)
			} else {
				// nested nested field
				nStruct, err := c.getNestedStruct(nFType, nestedField)
				if err != nil {
					c.log().Debug("NestedField: %s. getNestedStruct failed. %v", nField.Name, err)
					return nil, err
				}

				models.FieldSetNested(nestedField.StructField, nStruct)

				marshalField.Type = nStruct.marshalType
			}

			if models.FieldIsPtr(nestedField.StructField) {
				models.FieldSetFlag(nestedField.StructField, models.FBasePtr)
			}
		case reflect.Map:
			models.FieldSetFlag(nestedField.StructField, models.FMap)
			// should it be disallowed?
			// check the inner kind
			mapElem := nFType.Elem()

			var isPtr bool
			if mapElem.Kind() == reflect.Ptr {
				isPtr = true
				mapElem = mapElem.Elem()
			}

			switch mapElem.Kind() {
			case reflect.Struct:

				// check if it is time
				if mapElem == reflect.TypeOf(time.Time{}) {
					models.FieldSetFlag(nestedField.StructField, models.FTime)
					// otherwise it must be a nested struct
				} else {
					models.FieldSetFlag(nestedField.StructField, models.FNestedStruct)

					nStruct, err := c.getNestedStruct(mapElem, nestedField)
					if err != nil {
						c.log().Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}

					models.FieldSetNested(nestedField.StructField, nStruct)
				}
				// if the value is pointer add the base flag
				if isPtr {
					models.FieldSetFlag(nestedField.StructField, models.FBasePtr)
				}
			case reflect.Slice, reflect.Array:
				mapElem = mapElem.Elem()
				for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
					mapElem = mapElem.Elem()
				}

				if mapElem.Kind() == reflect.Ptr {
					models.FieldSetFlag(nestedField.StructField, models.FBasePtr)
					mapElem = mapElem.Elem()
				}

				switch mapElem.Kind() {
				case reflect.Struct:
					// check if it is time
					if mapElem == reflect.TypeOf(time.Time{}) {
						models.FieldSetFlag(nestedField.StructField, models.FTime)
						// otherwise it must be a nested struct
					} else {
						models.FieldSetFlag(nestedField.StructField, models.FNestedStruct)

						nStruct, err := c.getNestedStruct(mapElem, nestedField)
						if err != nil {
							c.log().Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
							return nil, err
						}
						models.FieldSetNested(nestedField.StructField, nStruct)
					}
				case reflect.Slice, reflect.Array, reflect.Map:
					// disallow nested map, arrs, maps in ptr type slices
					err := errors.Errorf("StructField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map.", nestedField.Name())
					return nil, err
				default:
				}
			}

		case reflect.Slice, reflect.Array:
			if nFType.Kind() == reflect.Slice {
				models.FieldSetFlag(nestedField.StructField, models.FSlice)
			} else {
				models.FieldSetFlag(nestedField.StructField, models.FArray)
			}
			for nFType.Kind() == reflect.Slice || nFType.Kind() == reflect.Array {
				nFType = nFType.Elem()
			}

			if nFType.Kind() == reflect.Ptr {
				models.FieldSetFlag(nestedField.StructField, models.FBasePtr)
				nFType = nFType.Elem()
			}

			switch nFType.Kind() {
			case reflect.Struct:
				// check if time
				if nFType == reflect.TypeOf(time.Time{}) {
					models.FieldSetFlag(nestedField.StructField, models.FTime)
				} else {
					// this should be the nested struct
					models.FieldSetFlag(nestedField.StructField, models.FNestedStruct)
					nStruct, err := c.getNestedStruct(nFType, nestedField)
					if err != nil {
						c.log().Debugf("nestedField: %v getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}
					models.FieldSetNested(nestedField.StructField, nStruct)
				}
			case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Array:
				err := errors.Errorf("Nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.Name(), nestedStruct.modelType.Name())
				return nil, err
			}

		default:
			// basic type (ints, uints, string, floats)
			// do nothing

			if models.FieldIsPtr(nestedField.StructField) {
				models.FieldSetFlag(nestedField.StructField, models.FBasePtr)
			}
		}

		var tagValue string = nestedField.ApiName()

		if models.FieldIsOmitEmpty(nestedField.StructField) {
			tagValue += ",omitempty"
		}

		marshalField.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, tagValue))
		marshalFields = append(marshalFields, marshalField)
	}

	models.NestedStructSetMarshalType(nestedStruct, reflect.StructOf(marshalFields))

	return nestedStruct, nil
}
