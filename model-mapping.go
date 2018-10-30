package jsonapi

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
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

	collectioner, ok := model.(Collectioner)
	if ok {
		modelStruct.collectionType = collectioner.CollectionName()
	} else {
		modelStruct.collectionType = c.NamerFunc(inflection.Plural(modelType.Name()))
	}

	modelStruct.attributes = make(map[string]*StructField)
	modelStruct.relationships = make(map[string]*StructField)
	modelStruct.foreignKeys = make(map[string]*StructField)
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
		structField.fieldName = tField.Name
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

				switch structField.refStruct.Type {
				case reflect.TypeOf(time.Time{}):
					structField.fieldFlags = structField.fieldFlags | fTime
				case reflect.TypeOf(&time.Time{}):
					structField.fieldFlags = structField.fieldFlags | fPtrTime
				default:

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
