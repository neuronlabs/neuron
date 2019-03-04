package models

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/namer"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/pkg/errors"
	"net/url"
	"time"
	"unicode"

	"reflect"
	"sync"
)

var (
	errBadJSONAPIStructTag = errors.New("Bad jsonapi struct tag format")
	IErrModelNotMapped     = errors.New("Unmapped model provided.")
)

// ModelMap contains mapped models ( as reflect.Type ) to its ModelStruct representation.
// Allow concurrent safe gets and sets to the map.
type ModelMap struct {
	models      map[reflect.Type]*ModelStruct
	collections map[string]reflect.Type
	sync.RWMutex
}

// NewModelMap creates new model map
func NewModelMap() *ModelMap {
	var modelMap *ModelMap = &ModelMap{
		models:      make(map[reflect.Type]*ModelStruct),
		collections: make(map[string]reflect.Type),
	}

	return modelMap
}

// Set sets the modelstruct for given map
// If the model already exists the function returns an error
func (m *ModelMap) Set(value *ModelStruct) error {
	m.Lock()
	defer m.Unlock()

	_, ok := m.models[value.modelType]
	if ok {
		return errors.Errorf("Model: %s already registered", value.Type())
	}

	_, ok = m.collections[value.collectionType]
	if ok {
		return errors.Errorf("Model: %s already registered", value.Type())
	}

	m.models[value.modelType] = value
	m.collections[value.collectionType] = value.Type()

	return nil
}

// Get is concurrent safe getter of model structs.
func (m *ModelMap) Get(model reflect.Type) *ModelStruct {
	m.RLock()
	defer m.RUnlock()
	return m.models[model]
}

func (m *ModelMap) GetByCollection(collection string) *ModelStruct {
	m.RLock()
	defer m.RUnlock()
	t, ok := m.collections[collection]
	if !ok || t == nil {
		return nil
	}
	return m.models[t]
}

// Models returns all models set within given model map
func (m *ModelMap) Models() []*ModelStruct {
	structs := []*ModelStruct{}

	for _, model := range m.models {
		structs = append(structs, model)
	}
	return structs
}

// SetByCollection sets the model by it's collection
func (m *ModelMap) SetByCollection(ms *ModelStruct) {
	m.Lock()
	defer m.Unlock()

	m.collections[ms.Collection()] = ms.Type()
}

func (m *ModelMap) getSimilarCollections(collection string) (simillar []string) {
	/**

	TO IMPLEMENT:

	find closest match collection

	*/
	return []string{}
}

// // Register creates and maps the model into *ModelStruct.
// // Returns error if the model is not correctly defined.
// func (m *ModelMap) Register(
// 	model interface{},
// 	namerFunc namer.Namer,
// 	flgs *flags.Container,
// ) error {
// 	mStruct, err := buildModelStruct(model, namerFunc, flgs)
// 	if err != nil {
// 		return err
// 	}

// 	m.Set(mStruct.Type(), mStruct)
// 	return nil
// }

func getSliceElemType(modelType reflect.Type) (reflect.Type, error) {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Slice {
		err := fmt.Errorf("Invalid input for slice elem type: %v", modelType)
		return modelType, err
	}

	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}

	return modelType, nil
}

func BuildModelStruct(
	model interface{},
	namerFunc namer.Namer,
	flgs *flags.Container,
) (*ModelStruct, error) {
	var err error

	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()

	if modelType.Kind() == reflect.Ptr {
		modelValue = reflect.ValueOf(model).Elem()
		modelType = modelValue.Type()
	} else {
		return nil, fmt.Errorf("Provide addressable models i.e.: &Model{} in order to precompute it. Invalid model: %v", modelType)
	}

	if modelType.Kind() != reflect.Struct {
		err = fmt.Errorf(`Provided model in invalid format. 
			The model must be of struct or ptr type, but is: %v`, modelType)
		return nil, err
	}

	var collection string

	collectioner, ok := model.(Collectioner)
	if ok {
		collection = collectioner.CollectionName()
	} else {
		collection = namerFunc(inflection.Plural(modelType.Name()))
	}

	modelStruct := NewModelStruct(modelType, collection, flgs)

	// Define the function definition
	var mapFields func(modelType reflect.Type, modelValue reflect.Value, index []int) error

	var (
		assignedFields int
		init           bool = true
	)

	// assign the function to it
	mapFields = func(modelType reflect.Type, modelValue reflect.Value, index []int) error {

		for i := 0; i < modelType.NumField(); i++ {

			var fieldIndex []int

			// check if field is embedded
			tField := modelType.Field(i)

			if init {
				fieldIndex = []int{i}
				init = false
			} else {
				fieldIndex = append(index, i)
			}

			if tField.Anonymous {
				// the field is embedded struct or ptr to struct
				nestedModelType := tField.Type
				var nestedModelValue reflect.Value

				if nestedModelType.Kind() == reflect.Ptr {
					nestedModelType = nestedModelType.Elem()
				}

				nestedModelValue = reflect.New(nestedModelType).Elem()

				if err := mapFields(nestedModelType, nestedModelValue, fieldIndex); err != nil {
					log.Debugf("Mapping embedded field: %s failed: %v", tField.Name, err)
					return err
				}
				continue
			}

			// don't use private fields
			if !modelValue.Field(i).CanSet() {
				log.Debugf("Field not settable: %s", modelType.Field(i).Name)
				continue
			}

			tag, ok := tField.Tag.Lookup(internal.AnnotationJSONAPI)
			if !ok {
				continue
			}

			if tag == "-" {
				continue
			}

			var tagValues url.Values

			structField := NewStructField(tField, modelStruct)
			tagValues, err = FieldTagValues(structField, tag)
			if err != nil {
				return errors.Wrapf(err, "Getting tag values failed. Model: %s, SField: %s", modelStruct.modelType.Name(), tField.Name)
			}
			structField.fieldIndex = make([]int, len(fieldIndex))
			copy(structField.fieldIndex, fieldIndex)

			assignedFields++

			// Check if field contains the name
			var apiName string
			name := tagValues.Get(internal.AnnotationName)
			if name != "" {
				apiName = name
			} else {
				apiName = namerFunc(tField.Name)
			}

			FieldsSetApiName(structField, apiName)

			// Set field type
			values := tagValues[internal.AnnotationFieldType]
			if len(values) == 0 {
				return errors.Errorf("StructField.annotationFieldType struct field tag cannot be empty. Model: %s, field: %s", modelType.Name(), tField.Name)
			} else {
				// Set field type
				value := values[0]
				switch value {
				case internal.AnnotationPrimary:
					FieldSetFieldKind(structField, KindPrimary)
					StructSetPrimary(modelStruct, structField)
					StructAppendField(modelStruct, structField)

				case internal.AnnotationRelation:
					// add relationField to fields
					StructAppendField(modelStruct, structField)

					// set related type
					err = FieldSetRelatedType(structField)
					if err != nil {
						return errors.Wrap(err, "FieldSetRelatedType failed")
					}

					// check duplicates
					_, ok := StructRelField(modelStruct, apiName)
					if ok {
						err = errors.Errorf("Duplicated jsonapi relationship field name: '%s' for model: '%v'.", apiName, modelType.Name())
						return err
					}

					// set relationship field
					StructSetRelField(modelStruct, structField)
				case internal.AnnotationAttribute:
					FieldSetFieldKind(structField, KindAttribute)
					// check if no duplicates
					_, ok := StructAttr(modelStruct, apiName)
					if ok {
						err = errors.Errorf("Duplicated jsonapi attribute name: '%s' for model: '%v'.",
							structField.apiName, modelStruct.modelType.Name())
						return err
					}

					StructAppendField(modelStruct, structField)

					t := structField.ReflectField().Type
					if t.Kind() == reflect.Ptr {
						FieldSetFlag(structField, FPtr)
						t = t.Elem()
					}

					switch t.Kind() {
					case reflect.Struct:
						if t == reflect.TypeOf(time.Time{}) {
							FieldSetFlag(structField, FTime)
						} else {
							// this case it must be a nested struct field
							FieldSetFlag(structField, FNestedStruct)

							nStruct, err := getNestedStruct(t, structField, namerFunc)
							if err != nil {
								log.Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
								return err
							}
							FieldSetNested(structField, nStruct)
						}

						if FieldIsPtr(structField) {
							FieldSetFlag(structField, FBasePtr)
						}

					case reflect.Map:
						FieldSetFlag(structField, FMap)

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
								FieldSetFlag(structField, FTime)
								// otherwise it must be a nested struct
							} else {
								FieldSetFlag(structField, FNestedStruct)

								nStruct, err := getNestedStruct(mapElem, structField, namerFunc)
								if err != nil {
									log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
									return err
								}
								FieldSetNested(structField, nStruct)
							}
							// if the value is pointer add the base flag
							if isPtr {
								FieldSetFlag(structField, FBasePtr)
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
								FieldSetFlag(structField, FBasePtr)
								mapElem = mapElem.Elem()
							}

							switch mapElem.Kind() {
							case reflect.Struct:
								// check if it is time
								if mapElem == reflect.TypeOf(time.Time{}) {
									FieldSetFlag(structField, FTime)
									// otherwise it must be a nested struct
								} else {
									FieldSetFlag(structField, FNestedStruct)
									nStruct, err := getNestedStruct(mapElem, structField, namerFunc)
									if err != nil {
										log.Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
										return err
									}
									FieldSetNested(structField, nStruct)
								}
							case reflect.Slice, reflect.Array, reflect.Map:
								// disallow nested map, arrs, maps in ptr type slices
								err = errors.Errorf("StructField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map.", structField.Name())
								return err
							default:
							}
						default:
							if isPtr {
								FieldSetFlag(structField, FBasePtr)
							}

						}
					case reflect.Slice, reflect.Array:
						if t.Kind() == reflect.Slice {
							FieldSetFlag(structField, FSlice)
						} else {
							FieldSetFlag(structField, FArray)
						}

						// dereference the slice
						// check the field base type
						sliceElem := t
						for sliceElem.Kind() == reflect.Slice || sliceElem.Kind() == reflect.Array {
							sliceElem = sliceElem.Elem()
						}

						if sliceElem.Kind() == reflect.Ptr {
							// add maybe slice Field Ptr
							FieldSetFlag(structField, FBasePtr)
							sliceElem = sliceElem.Elem()
						}

						switch sliceElem.Kind() {
						case reflect.Struct:
							// check if time
							if sliceElem == reflect.TypeOf(time.Time{}) {
								FieldSetFlag(structField, FTime)
							} else {
								// this should be the nested struct
								FieldSetFlag(structField, FNestedStruct)
								nStruct, err := getNestedStruct(sliceElem, structField, namerFunc)
								if err != nil {
									log.Debugf("structField: %v getNestedStruct failed. %v", structField.Name(), err)
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
						if FieldIsPtr(structField) {
							FieldSetFlag(structField, FBasePtr)
						}

					}
					StructSetAttr(modelStruct, structField)
				case internal.AnnotationForeignKey:
					FieldSetFieldKind(structField, KindForeignKey)
					// Check if already exists
					_, ok := StructForeignKeyField(modelStruct, structField.ApiName())
					if ok {
						err = errors.Errorf("Duplicated jsonapi foreign key name: '%s' for model: '%v'", structField.ApiName(), modelStruct.Type().Name())
						return err
					}
					StructAppendField(modelStruct, structField)
					StructSetForeignKey(modelStruct, structField)

				case internal.AnnotationFilterKey:
					FieldSetFieldKind(structField, KindFilterKey)
					_, ok := StructFilterKeyField(modelStruct, structField.ApiName())
					if ok {
						err = errors.Errorf("Duplicated jsonapi filter key name: '%s' for model: '%v'", structField.ApiName(), modelStruct.Type().Name())
						return err
					}
					// modelStruct.fields = append(modelStruct.fields, structField)
					StructSetFilterKey(modelStruct, structField)
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
							FieldSetFlag(structField, FClientID)
						case internal.AnnotationNoFilter:
							FieldSetFlag(structField, FNoFilter)
						case internal.AnnotationHidden:
							FieldSetFlag(structField, FHidden)
						case internal.AnnotationNotSortable:
							FieldSetFlag(structField, FSortable)
						case internal.AnnotationISO8601:
							FieldSetFlag(structField, FIso8601)
						case internal.AnnotationOmitEmpty:
							FieldSetFlag(structField, FOmitempty)
						case internal.AnnotationI18n:
							FieldSetFlag(structField, FI18n)
							StructAppendI18n(modelStruct, structField)
						case internal.AnnotationLanguage:
							StructSetLanguage(modelStruct, structField)

						}
					}
				case internal.AnnotationRelation:
					// if relationship match the type e.t.c

					r := FieldRelationship(structField)
					if r == nil {
						r := &Relationship{}
						FieldSetRelationship(structField, r)
					}

					for _, value := range values {
						switch value {
						case internal.AnnotationRelationNoSync:
							b := false
							r.SetSync(&b)
						case internal.AnnotationManyToMany:
							r.SetKind(RelMany2Many)
						case internal.AnnotationRelationSync:
							b := true
							r.SetSync(&b)
						default:
							log.Debugf("Backreference field tag for relation: %s in model: %s. Value: %s", modelStruct.Type().Name(), structField.Name(), value)
							RelationshipSetBackrefFieldName(r, value)
						}
					}
					// if field is foreign key match with relationship
				case internal.AnnotationForeignKey:

				}
			}
		}
		return nil
	}

	// map fields
	if err := mapFields(modelType, modelValue, nil); err != nil {
		return nil, err
	}
	if assignedFields == 0 {
		err = fmt.Errorf("Model has no correct jsonapi fields: %v", modelType)
		return nil, err
	}

	if StructPrimary(modelStruct) == nil {
		err = fmt.Errorf("Model: %v must have a correct primary field.", modelType)
		return nil, err
	}

	return modelStruct, nil
}

func getNestedStruct(
	t reflect.Type, sFielder StructFielder, namerFunc namer.Namer,
) (*NestedStruct, error) {
	nestedStruct := NewNestedStruct(t, sFielder)

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

		nestedField := NewNestedField(nestedStruct, sFielder, nField)

		tag, ok := nField.Tag.Lookup("jsonapi")
		if ok {
			if tag == "-" {
				marshalField.Tag = reflect.StructTag(`json:"-"`)
				marshalFields = append(marshalFields, marshalField)
				continue
			}

			tagValues, err := FieldTagValues(nestedField.StructField, tag)
			if err != nil {
				log.Debugf("nestedField: '%s', getTagValues failed. %v", nestedField.Name())
				return nil, err
			}

			for tKey, tValue := range tagValues {
				switch tKey {
				case internal.AnnotationName:
					FieldsSetApiName(nestedField.StructField, tValue[0])
				case internal.AnnotationFieldType:
					if tValue[0] != internal.AnnotationNestedField {
						log.Debugf("Invalid annotationNestedField value: '%s' for field: %s", tValue[0], nestedField.Name())
						err = errors.Errorf("Provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.Name())
						return nil, err
					}
				case internal.AnnotationFlags:
					for _, value := range tValue {
						switch value {
						case internal.AnnotationNoFilter:
							FieldSetFlag(nestedField.StructField, FNoFilter)
						case internal.AnnotationHidden:
							FieldSetFlag(nestedField.StructField, FHidden)
						case internal.AnnotationNotSortable:
							FieldSetFlag(nestedField.StructField, FSortable)
						case internal.AnnotationISO8601:
							FieldSetFlag(nestedField.StructField, FIso8601)
						case internal.AnnotationOmitEmpty:
							FieldSetFlag(nestedField.StructField, FOmitempty)
						}

					}
				}
			}
		}

		if nestedField.ApiName() == "" {
			FieldsSetApiName(nestedField.StructField, namerFunc(nField.Name))
		}

		switch nestedField.ApiName() {
		case "relationships", "links":
			return nil, errors.Errorf("Nested field within: '%s' field in the model: '%s' has forbidden API name: '%s'. ",
				NestedStructAttr(nestedStruct).Name(),
				FieldsStruct(NestedStructAttr(nestedStruct)).Type().Name(),
				nestedField.ApiName(),
			)
		default:
		}

		if _, ok = NestedStructSubField(nestedStruct, nestedField.ApiName()); ok {
			return nil, errors.Errorf("NestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.Type().Name(), nestedField.Name())
		}
		NestedStructSetSubfield(nestedStruct, nestedField)

		nFType := nField.Type
		if nFType.Kind() == reflect.Ptr {
			nFType = nFType.Elem()
			FieldSetFlag(nestedField.StructField, FPtr)
		}

		switch nFType.Kind() {
		case reflect.Struct:
			if nFType == reflect.TypeOf(time.Time{}) {
				FieldSetFlag(nestedField.StructField, FTime)
			} else {
				// nested nested field
				nStruct, err := getNestedStruct(nFType, nestedField, namerFunc)
				if err != nil {
					log.Debug("NestedField: %s. getNestedStruct failed. %v", nField.Name, err)
					return nil, err
				}

				FieldSetNested(nestedField.StructField, nStruct)

				marshalField.Type = nStruct.marshalType
			}

			if FieldIsPtr(nestedField.StructField) {
				FieldSetFlag(nestedField.StructField, FBasePtr)
			}
		case reflect.Map:
			FieldSetFlag(nestedField.StructField, FMap)
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
					FieldSetFlag(nestedField.StructField, FTime)
					// otherwise it must be a nested struct
				} else {
					FieldSetFlag(nestedField.StructField, FNestedStruct)

					nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}

					FieldSetNested(nestedField.StructField, nStruct)
				}
				// if the value is pointer add the base flag
				if isPtr {
					FieldSetFlag(nestedField.StructField, FBasePtr)
				}
			case reflect.Slice, reflect.Array:
				mapElem = mapElem.Elem()
				for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
					mapElem = mapElem.Elem()
				}

				if mapElem.Kind() == reflect.Ptr {
					FieldSetFlag(nestedField.StructField, FBasePtr)
					mapElem = mapElem.Elem()
				}

				switch mapElem.Kind() {
				case reflect.Struct:
					// check if it is time
					if mapElem == reflect.TypeOf(time.Time{}) {
						FieldSetFlag(nestedField.StructField, FTime)
						// otherwise it must be a nested struct
					} else {
						FieldSetFlag(nestedField.StructField, FNestedStruct)

						nStruct, err := getNestedStruct(mapElem, nestedField, namerFunc)
						if err != nil {
							log.Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
							return nil, err
						}
						FieldSetNested(nestedField.StructField, nStruct)
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
				FieldSetFlag(nestedField.StructField, FSlice)
			} else {
				FieldSetFlag(nestedField.StructField, FArray)
			}
			for nFType.Kind() == reflect.Slice || nFType.Kind() == reflect.Array {
				nFType = nFType.Elem()
			}

			if nFType.Kind() == reflect.Ptr {
				FieldSetFlag(nestedField.StructField, FBasePtr)
				nFType = nFType.Elem()
			}

			switch nFType.Kind() {
			case reflect.Struct:
				// check if time
				if nFType == reflect.TypeOf(time.Time{}) {
					FieldSetFlag(nestedField.StructField, FTime)
				} else {
					// this should be the nested struct
					FieldSetFlag(nestedField.StructField, FNestedStruct)
					nStruct, err := getNestedStruct(nFType, nestedField, namerFunc)
					if err != nil {
						log.Debugf("nestedField: %v getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}
					FieldSetNested(nestedField.StructField, nStruct)
				}
			case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Array:
				err := errors.Errorf("Nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.Name(), nestedStruct.modelType.Name())
				return nil, err
			}

		default:
			// basic type (ints, uints, string, floats)
			// do nothing

			if FieldIsPtr(nestedField.StructField) {
				FieldSetFlag(nestedField.StructField, FBasePtr)
			}
		}

		var tagValue string = nestedField.ApiName()

		if FieldIsOmitEmpty(nestedField.StructField) {
			tagValue += ",omitempty"
		}

		marshalField.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, tagValue))
		marshalFields = append(marshalFields, marshalField)
	}

	NestedStructSetMarshalType(nestedStruct, reflect.StructOf(marshalFields))

	return nestedStruct, nil
}
