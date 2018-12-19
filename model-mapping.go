package jsonapi

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
	"sync"
	"time"
	"unicode"
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

func newModelMap() *ModelMap {
	var modelMap *ModelMap = &ModelMap{
		models:      make(map[reflect.Type]*ModelStruct),
		collections: make(map[string]reflect.Type),
	}

	return modelMap
}

// Set is concurrent safe setter for model structs
func (m *ModelMap) Set(key reflect.Type, value *ModelStruct) {
	m.Lock()
	defer m.Unlock()
	m.models[key] = value
}

// Get is concurrent safe getter of model structs.
func (m *ModelMap) Get(key reflect.Type) *ModelStruct {
	m.RLock()
	defer m.RUnlock()
	return m.models[key]
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

func (m *ModelMap) getSimilarCollections(collection string) (simillar []string) {
	/**

	TO IMPLEMENT:

	find closest match collection

	*/
	return []string{}
}

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

	modelStruct := new(ModelStruct)
	modelStruct.modelType = modelType
	modelStruct.ctrl = c

	collectioner, ok := model.(Collectioner)
	if ok {
		modelStruct.collectionType = collectioner.CollectionName()
	} else {
		modelStruct.collectionType = c.NamerFunc(inflection.Plural(modelType.Name()))
	}

	modelStruct.attributes = make(map[string]*StructField)
	modelStruct.relationships = make(map[string]*StructField)
	modelStruct.foreignKeys = make(map[string]*StructField)
	modelStruct.filterKeys = make(map[string]*StructField)

	modelStruct.collectionURLIndex = -1

	var assignedFields int
	for i := 0; i < modelType.NumField(); i++ {
		// don't use private fields
		if !modelValue.Field(i).CanSet() {
			continue
		}

		tag, ok := modelType.Field(i).Tag.Lookup(annotationJSONAPI)
		if !ok {
			continue
		}

		if tag == "-" {
			continue
		}

		var tagValues url.Values
		tField := modelType.Field(i)

		structField := new(StructField)
		tagValues, err = structField.getTagValues(tag)
		if err != nil {
			return errors.Wrapf(err, "Getting tag values failed. Model: %s, SField: %s", modelStruct.modelType.Name(), tField.Name)
		}

		structField.refStruct = tField
		structField.mStruct = modelStruct
		assignedFields++

		// Check if field contains the name
		name := tagValues.Get(annotationName)
		if name != "" {
			structField.jsonAPIName = name
		} else {
			structField.jsonAPIName = c.NamerFunc(tField.Name)
		}

		// Set field type
		values := tagValues[annotationFieldType]
		if len(values) == 0 {
			return errors.Errorf("StructField.annotationFieldType struct field tag cannot be empty. Model: %s, field: %s", modelStruct.modelType.Name(), tField.Name)
		} else {
			// Set field type
			value := values[0]
			switch value {
			case annotationPrimary:
				structField.fieldType = Primary
				modelStruct.primary = structField
				modelStruct.fields = append(modelStruct.fields, structField)

			case annotationRelation:
				modelStruct.fields = append(modelStruct.fields, structField)
				err = setRelatedType(structField)
				_, ok := modelStruct.relationships[structField.jsonAPIName]
				if ok {
					err = errors.Errorf("Duplicated jsonapi relationship field name: '%s' for model: '%v'.",
						structField.jsonAPIName, modelStruct.modelType.Name())
					return err
				}
				modelStruct.relationships[structField.jsonAPIName] = structField
			case annotationAttribute:
				structField.fieldType = Attribute
				// check if no duplicates
				_, ok := modelStruct.attributes[structField.jsonAPIName]
				if ok {
					err = errors.Errorf("Duplicated jsonapi attribute name: '%s' for model: '%v'.",
						structField.jsonAPIName, modelStruct.modelType.Name())
					return err
				}
				modelStruct.fields = append(modelStruct.fields, structField)

				t := structField.refStruct.Type
				if t.Kind() == reflect.Ptr {
					structField.fieldFlags = structField.fieldFlags | fPtr
					t = t.Elem()
				}

				switch t.Kind() {
				case reflect.Struct:
					if t == reflect.TypeOf(time.Time{}) {
						structField.fieldFlags = structField.fieldFlags | fTime
					} else {
						// this case it must be a nested struct field
						structField.fieldFlags = structField.fieldFlags | fNestedStruct

						nStruct, err := c.getNestedStruct(t, structField)
						if err != nil {
							c.log().Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
							return err
						}
						structField.nested = nStruct
					}

					if structField.isPtr() {
						structField.fieldFlags = structField.fieldFlags | fBasePtr
					}

				case reflect.Map:
					structField.fieldFlags = structField.fieldFlags | fMap
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
							structField.fieldFlags = structField.fieldFlags | fTime
							// otherwise it must be a nested struct
						} else {
							structField.fieldFlags = structField.fieldFlags | fNestedStruct

							nStruct, err := c.getNestedStruct(mapElem, structField)
							if err != nil {
								c.log().Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
								return err
							}
							structField.nested = nStruct
						}
						// if the value is pointer add the base flag
						if isPtr {
							structField.fieldFlags = structField.fieldFlags | fBasePtr
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
							structField.fieldFlags = structField.fieldFlags | fBasePtr
							mapElem = mapElem.Elem()
						}

						switch mapElem.Kind() {
						case reflect.Struct:
							// check if it is time
							if mapElem == reflect.TypeOf(time.Time{}) {
								structField.fieldFlags = structField.fieldFlags | fTime
								// otherwise it must be a nested struct
							} else {
								structField.fieldFlags = structField.fieldFlags | fNestedStruct

								nStruct, err := c.getNestedStruct(mapElem, structField)
								if err != nil {
									c.log().Debugf("structField: %v Map field getNestedStruct failed. %v", structField.fieldName(), err)
									return err
								}
								structField.nested = nStruct
							}
						case reflect.Slice, reflect.Array, reflect.Map:
							// disallow nested map, arrs, maps in ptr type slices
							err = errors.Errorf("StructField: '%s' nested type is invalid. The model doesn't allow one of slices to ptr of slices or map.", structField.Name())
							return err
						default:
						}
					default:
						if isPtr {
							structField.fieldFlags = structField.fieldFlags | fBasePtr
						}

					}
				case reflect.Slice, reflect.Array:
					if t.Kind() == reflect.Slice {
						structField.fieldFlags = structField.fieldFlags | fSlice
					} else {
						structField.fieldFlags = structField.fieldFlags | fArray
					}

					// dereference the slice
					// check the field base type
					sliceElem := t
					for sliceElem.Kind() == reflect.Slice || sliceElem.Kind() == reflect.Array {
						sliceElem = sliceElem.Elem()
					}

					if sliceElem.Kind() == reflect.Ptr {
						// add maybe slice Field Ptr
						structField.fieldFlags = structField.fieldFlags | fBasePtr
						sliceElem = sliceElem.Elem()
					}

					switch sliceElem.Kind() {
					case reflect.Struct:
						// check if time
						if sliceElem == reflect.TypeOf(time.Time{}) {
							structField.fieldFlags = structField.fieldFlags | fTime
						} else {
							// this should be the nested struct
							structField.fieldFlags = structField.fieldFlags | fNestedStruct
							nStruct, err := c.getNestedStruct(sliceElem, structField)
							if err != nil {
								c.log().Debugf("structField: %v getNestedStruct failed. %v", structField.fieldName(), err)
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
					if structField.isPtr() {
						structField.fieldFlags = structField.fieldFlags | fBasePtr
					}

				}

				modelStruct.attributes[structField.jsonAPIName] = structField
			case annotationForeignKey:
				structField.fieldType = ForeignKey
				_, ok := modelStruct.foreignKeys[structField.jsonAPIName]
				if ok {
					err = errors.Errorf("Duplicated jsonapi foreign key name: '%s' for model: '%v'", structField.jsonAPIName, modelStruct.modelType.Name())
					return err
				}
				modelStruct.fields = append(modelStruct.fields, structField)
				modelStruct.foreignKeys[structField.jsonAPIName] = structField
			case annotationFilterKey:
				structField.fieldType = FilterKey
				_, ok := modelStruct.filterKeys[structField.jsonAPIName]
				if ok {
					err = errors.Errorf("Duplicated jsonapi filter key name: '%s' for model: '%v'", structField.jsonAPIName, modelStruct.modelType.Name())
					return err
				}
				// modelStruct.fields = append(modelStruct.fields, structField)
				modelStruct.filterKeys[structField.jsonAPIName] = structField
			default:
				return errors.Errorf("Unknown field type: %s. Model: %s, field: %s", value, modelStruct.modelType.Name(), tField.Name)
			}
		}

		// iterate over structfield tags
		for key, values := range tagValues {
			switch key {
			case annotationFieldType, annotationName:
				continue
			case annotationFlags:
				for _, value := range values {

					switch value {
					case annotationClientID:
						structField.fieldFlags = structField.fieldFlags | fClientID
					case annotationNoFilter:
						structField.fieldFlags = structField.fieldFlags | fNoFilter
					case annotationHidden:
						structField.fieldFlags = structField.fieldFlags | fHidden
					case annotationNotSortable:
						structField.fieldFlags = structField.fieldFlags | fSortable
					case annotationISO8601:
						structField.fieldFlags = structField.fieldFlags | fIso8601
					case annotationOmitEmpty:
						structField.fieldFlags = structField.fieldFlags | fOmitempty
					case annotationI18n:
						structField.fieldFlags = structField.fieldFlags | fI18n
						if modelStruct.i18n == nil {
							modelStruct.i18n = make([]*StructField, 0)
						}
						modelStruct.i18n = append(modelStruct.i18n, structField)
					case annotationLanguage:
						structField.fieldFlags = structField.fieldFlags | fLanguage
						modelStruct.language = structField

					}
				}
			case annotationRelation:
				// if relationship match the type e.t.c
				if structField.relationship == nil {
					structField.relationship = &Relationship{}
				}

				r := structField.relationship
				for _, value := range values {
					switch value {
					case annotationRelationNoSync:
						b := false
						r.Sync = &b
					case annotationManyToMany:
						r.Kind = RelMany2Many
					case annotaitonRelationSync:
						b := true
						r.Sync = &b
					default:
						c.log().Debugf("Backreference field tag for relation: %s in model: %s. Value: %s", modelStruct.modelType.Name(), structField.fieldName, value)
						r.BackReferenceFieldname = value

					}
				}
				// if field is foreign key match with relationship
			case annotationForeignKey:

			}
		}
	}

	if assignedFields == 0 {
		err = fmt.Errorf("Model has no correct jsonapi fields: %v", modelType)
		return err
	}

	if modelStruct.primary == nil {
		err = fmt.Errorf("Model: %v must have a correct primary field.", modelType)
		return err
	}

	modelMap.Set(modelType, modelStruct)
	return err
}

func (c *Controller) getNestedStruct(
	t reflect.Type, sFielder structFielder,
) (*NestedStruct, error) {

	nestedStruct := &NestedStruct{
		modelType:   t,
		structField: sFielder,
		fields:      map[string]*NestedField{},
	}

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

		nestedField := &NestedField{
			StructField: &StructField{
				mStruct:    sFielder.self().mStruct,
				refStruct:  nField,
				fieldType:  FTNested,
				fieldFlags: fDefault | fNestedField,
			},
			root: nestedStruct,
		}

		tag, ok := nField.Tag.Lookup("jsonapi")
		if ok {
			if tag == "-" {
				marshalField.Tag = reflect.StructTag(`json:"-"`)
				marshalFields = append(marshalFields, marshalField)
				continue
			}

			tagValues, err := nestedField.getTagValues(tag)
			if err != nil {
				c.log().Debugf("nestedField: '%s', getTagValues failed. %v", nestedField.fieldName(), err)
				return nil, err
			}

			for tKey, tValue := range tagValues {
				switch tKey {
				case annotationName:
					nestedField.jsonAPIName = tValue[0]
				case annotationFieldType:
					if tValue[0] != annotationNestedField {
						c.log().Debugf("Invalid annotationNestedField value: '%s' for field: %s", tValue[0], nestedField.fieldName())
						err = errors.Errorf("Provided field type: '%s' is not allowed for the nested struct field: '%s'", nestedField.fieldName())
						return nil, err
					}
				case annotationFlags:
					for _, value := range tValue {
						switch value {
						case annotationNoFilter:
							nestedField.fieldFlags = nestedField.fieldFlags | fNoFilter
						case annotationHidden:
							nestedField.fieldFlags = nestedField.fieldFlags | fHidden
						case annotationNotSortable:
							nestedField.fieldFlags = nestedField.fieldFlags | fSortable
						case annotationISO8601:
							nestedField.fieldFlags = nestedField.fieldFlags | fIso8601
						case annotationOmitEmpty:
							nestedField.fieldFlags = nestedField.fieldFlags | fOmitempty
						}

					}
				}
			}
		}

		if nestedField.jsonAPIName == "" {
			nestedField.jsonAPIName = c.NamerFunc(nField.Name)
		}

		switch nestedField.jsonAPIName {
		case "relationships", "links":
			return nil, errors.Errorf("Nested field within: '%s' field in the model: '%s' has forbidden API name: '%s'. ", nestedStruct.attr().Name(), nestedStruct.attr().mStruct.modelType.Name(), nestedField.ApiName())
		default:
		}

		if _, ok = nestedStruct.fields[nestedField.jsonAPIName]; ok {
			return nil, errors.Errorf("NestedStruct: %v already has one nestedField: '%s'. The fields must be uniquely named", nestedStruct.modelType.Name(), nestedField.fieldName())
		}
		nestedStruct.fields[nestedField.jsonAPIName] = nestedField

		nFType := nField.Type
		if nFType.Kind() == reflect.Ptr {
			nFType = nFType.Elem()
			nestedField.fieldFlags = nestedField.fieldFlags | fPtr
		}

		switch nFType.Kind() {
		case reflect.Struct:
			if nFType == reflect.TypeOf(time.Time{}) {
				nestedField.fieldFlags = nestedField.fieldFlags | fTime
			} else {
				// nested nested field
				nStruct, err := c.getNestedStruct(nFType, nestedField)
				if err != nil {
					c.log().Debug("NestedField: %s. getNestedStruct failed. %v", nField.Name, err)
					return nil, err
				}
				nestedField.nested = nStruct
				marshalField.Type = nStruct.marshalType
			}
			if nestedField.isPtr() {
				nestedField.fieldFlags = nestedField.fieldFlags | fBasePtr
			}
		case reflect.Map:
			nestedField.fieldFlags = nestedField.fieldFlags | fMap
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
					nestedField.fieldFlags = nestedField.fieldFlags | fTime
					// otherwise it must be a nested struct
				} else {
					nestedField.fieldFlags = nestedField.fieldFlags | fNestedStruct

					nStruct, err := c.getNestedStruct(mapElem, nestedField)
					if err != nil {
						c.log().Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}
					nestedField.nested = nStruct
				}
				// if the value is pointer add the base flag
				if isPtr {
					nestedField.fieldFlags = nestedField.fieldFlags | fBasePtr
				}
			case reflect.Slice, reflect.Array:
				mapElem = mapElem.Elem()
				for mapElem.Kind() == reflect.Slice || mapElem.Kind() == reflect.Array {
					mapElem = mapElem.Elem()
				}

				if mapElem.Kind() == reflect.Ptr {
					nestedField.fieldFlags = nestedField.fieldFlags | fBasePtr
					mapElem = mapElem.Elem()
				}
				switch mapElem.Kind() {
				case reflect.Struct:
					// check if it is time
					if mapElem == reflect.TypeOf(time.Time{}) {
						nestedField.fieldFlags = nestedField.fieldFlags | fTime
						// otherwise it must be a nested struct
					} else {
						nestedField.fieldFlags = nestedField.fieldFlags | fNestedStruct

						nStruct, err := c.getNestedStruct(mapElem, nestedField)
						if err != nil {
							c.log().Debugf("nestedField: %v Map field getNestedStruct failed. %v", nestedField.fieldName(), err)
							return nil, err
						}
						nestedField.nested = nStruct
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
				nestedField.fieldFlags = nestedField.fieldFlags | fSlice
			} else {
				nestedField.fieldFlags = nestedField.fieldFlags | fArray
			}
			for nFType.Kind() == reflect.Slice || nFType.Kind() == reflect.Array {
				nFType = nFType.Elem()
			}

			if nFType.Kind() == reflect.Ptr {
				nestedField.fieldFlags = nestedField.fieldFlags | fBasePtr
				nFType = nFType.Elem()
			}

			switch nFType.Kind() {
			case reflect.Struct:
				// check if time
				if nFType == reflect.TypeOf(time.Time{}) {
					nestedField.fieldFlags = nestedField.fieldFlags | fTime
				} else {
					// this should be the nested struct
					nestedField.fieldFlags = nestedField.fieldFlags | fNestedStruct
					nStruct, err := c.getNestedStruct(nFType, nestedField)
					if err != nil {
						c.log().Debugf("nestedField: %v getNestedStruct failed. %v", nestedField.fieldName(), err)
						return nil, err
					}
					nestedField.nested = nStruct
				}
			case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Array:
				err := errors.Errorf("Nested field can't be a slice of pointer to slices|map|arrays. NestedField: '%s' within NestedStruct:'%s'", nestedField.Name(), nestedStruct.modelType.Name())
				return nil, err
			}

		default:
			// basic type (ints, uints, string, floats)
			// do nothing
			if nestedField.isPtr() {
				nestedField.fieldFlags = nestedField.fieldFlags | fBasePtr
			}
		}

		var tagValue string = nestedField.jsonAPIName

		if nestedField.isOmitEmpty() {
			tagValue += ",omitempty"
		}

		marshalField.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, tagValue))
		marshalFields = append(marshalFields, marshalField)
	}

	nestedStruct.marshalType = reflect.StructOf(marshalFields)

	return nestedStruct, nil
}

func errNoRelationship(jsonapiType, included string) *ErrorObject {
	err := ErrInvalidResourceName.Copy()
	err.Detail = fmt.Sprintf("Object: '%v', has no relationship named: '%v'.",
		jsonapiType, included)
	return err
}

func errNoModelMappedForRel(model, relatedTo reflect.Type, fieldName string) error {
	err := fmt.Errorf("Model '%v', not mapped! Relationship to '%s' is set for '%s' field.",
		model, relatedTo, fieldName,
	)
	return err
}

func errNoRelationshipInModel(sFieldType, modelType reflect.Type, relationship string) error {
	err := fmt.Errorf("Structfield of type: '%s' has no relationship within model: '%s', in relationship named: %v", sFieldType, modelType, relationship)
	return err
}

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

func setRelatedType(sField *StructField) error {

	modelType := sField.refStruct.Type
	// get error function
	getError := func() error {
		return fmt.Errorf("Incorrect relationship type provided. The Only allowable types are structs, pointers or slices. This type is: %v", modelType)
	}

	switch modelType.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Struct:
	default:
		err := getError()
		return err
	}
	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		if modelType.Kind() == reflect.Slice {
			sField.fieldType = RelationshipMultiple
		}
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err := getError()
		return err
	}

	if sField.fieldType == UnknownType {
		sField.fieldType = RelationshipSingle
	}
	sField.relatedModelType = modelType
	return nil
}
