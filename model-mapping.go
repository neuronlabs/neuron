package jsonapi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
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

func buildModelStruct(model interface{}, modelMap *ModelMap) error {
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

	modelStruct.attributes = make(map[string]*StructField)
	modelStruct.relationships = make(map[string]*StructField)
	modelStruct.collectionURLIndex = -1

	var assignedFields int
	for i := 0; i < modelType.NumField(); i++ {
		// don't use private fields
		if !modelValue.Field(i).CanSet() {
			continue
		}

		tag := modelType.Field(i).Tag.Get(annotationJSONAPI)
		if tag == "" {
			continue
		}

		args := strings.Split(tag, annotationSeperator)
		if len(args) == 1 {
			if args[0] != annotationClientID && args[0] != "-" && args[0] != annotationLanguage {
				err = fmt.Errorf("Bad JSONAPI struct tag format: %s for model: %v", tag, modelType)
				break
			}
		}

		var resName string
		if len(args) > 1 {
			resName = args[1]
		}

		tField := modelType.Field(i)

		structField := new(StructField)
		structField.refStruct = tField
		structField.fieldName = tField.Name
		structField.mStruct = modelStruct
		assignedFields++

		switch kind := args[0]; kind {
		case annotationPrimary:
			// Primary is not a part of fields
			structField.jsonAPIName = "id"
			structField.fieldType = Primary
			modelStruct.collectionType = resName
			modelStruct.primary = structField
			if len(args) > 2 {
				for _, arg := range args[2:] {
					switch arg {
					case annotationNoFilter:
						structField.fieldFlags = structField.fieldFlags | fNoFilter
					case annotationHidden:
						structField.fieldFlags = structField.fieldFlags | fHidden
					case annotationSortable:
						structField.fieldFlags = structField.fieldFlags | fSortable
					}
				}
			}

		case annotationClientID:
			// ClientID is not a part of fields also
			structField.jsonAPIName = "id"
			structField.fieldType = ClientID
			modelStruct.clientID = structField
		case annotationAttribute:
			structField.jsonAPIName = resName
			structField.fieldType = Attribute
			modelStruct.fields = append(modelStruct.fields, structField)

			// check if no duplicates
			_, ok := modelStruct.attributes[resName]
			if ok {
				err = fmt.Errorf("Duplicated json:api field name: '%s' for model: '%v'.",
					resName, modelStruct.modelType.Name())
				return err
			}
			modelStruct.attributes[resName] = structField

			// additional options for given tag
			if len(args) > 2 {
				for _, arg := range args[2:] {
					switch arg {
					case annotationISO8601:
						structField.fieldFlags = structField.fieldFlags | fIso8601
					case annotationOmitEmpty:
						structField.fieldFlags = structField.fieldFlags | fOmitempty
					case annotationNoFilter:
						structField.fieldFlags = structField.fieldFlags | fNoFilter
					case annotationI18n:
						structField.fieldFlags = structField.fieldFlags | fI18n
						if modelStruct.i18n == nil {
							modelStruct.i18n = make([]*StructField, 0)
						}
						modelStruct.i18n = append(modelStruct.i18n, structField)
					case annotationLanguage:
						structField.fieldFlags = structField.fieldFlags | fLanguage
						modelStruct.language = structField
					case annotationHidden:
						structField.fieldFlags = structField.fieldFlags | fHidden
					case annotationSortable:
						structField.fieldFlags = structField.fieldFlags | fSortable
					}

				}
			}
			if tField.Type == reflect.TypeOf(time.Time{}) {
				structField.fieldFlags = structField.fieldFlags | fTime
			} else if tField.Type == reflect.TypeOf(new(time.Time)) {
				structField.fieldFlags = structField.fieldFlags | fPtrTime
			}

		case annotationRelation:
			structField.jsonAPIName = resName
			modelStruct.fields = append(modelStruct.fields, structField)
			err = setRelatedType(structField)

			_, ok := modelStruct.relationships[resName]
			if ok {
				err = fmt.Errorf("Duplicated json:api field name: '%s' for model '%v'.",
					resName, modelStruct.modelType.Name())
				return err
			}
			modelStruct.relationships[resName] = structField

			if len(args) > 2 {
				for _, arg := range args[2:] {
					switch arg {
					case annotationNoFilter:
						structField.fieldFlags = structField.fieldFlags | fNoFilter
					case annotationOmitEmpty:
						structField.fieldFlags = structField.fieldFlags | fOmitempty
					case annotationHidden:
						structField.fieldFlags = structField.fieldFlags | fHidden
					case annotationSortable:
						structField.fieldFlags = structField.fieldFlags | fSortable
					}
				}
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

// func checkModelRelationships(model *ModelStruct) (err error) {
// 	for _, rel := range model.relationships {
// 		val := cacheModelMap.Get(rel.relatedModelType)
// 		if val == nil {
// 			err = fmt.Errorf("Model: %v, not precalculated but is used in relationships for: %v field in %v model.", rel.relatedModelType, rel.fieldName, model.modelType.Name())
// 			return err
// 		}
// 		rel.relatedStruct = val
// 	}
// 	return
// }

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
