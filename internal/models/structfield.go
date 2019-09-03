package models

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute').
type FieldKind int

// Enums for the field kind.
const (
	// UnknownType is the unsupported unknown type of the struct field.
	UnknownType FieldKind = iota
	// KindPrimary is a 'primary' field.
	KindPrimary
	// KindAttribute is an 'attribute' field.
	KindAttribute
	// KindRelationshipSingle is a 'relationship' with single object.
	KindRelationshipSingle
	// KindRelationshipMultiple is a 'relationship' with multiple objects.
	KindRelationshipMultiple
	// KindForeignKey is the field type that is responsible for the relationships.
	KindForeignKey
	// KindFilterKey is the field that is used only for special case filtering.
	KindFilterKey
	// KindNested is the field that is nested within another structfield.
	KindNested
)

// String implements fmt.Stringer interface.
func (f FieldKind) String() string {
	switch f {
	case KindPrimary:
		return "Primary"
	case KindAttribute:
		return "Attribute"
	case KindRelationshipSingle, KindRelationshipMultiple:
		return "Relationship"
	case KindForeignKey:
		return "ForeignKey"
	case KindFilterKey:
		return "FilterKey"
	case KindNested:
		return "Nested"
	}
	return "Unknown"
}

type fieldFlag int

// field flags
const (
	// FDefault is a default flag value
	FDefault fieldFlag = iota
	// FOmitEmpty is a field flag for omitting empty value.
	FOmitempty fieldFlag = 1 << (iota - 1)
	// FISO8601 is a time field flag marking it usable with IS08601 formatting.
	FISO8601
	// FI18n is the i18n field flag.
	FI18n
	// FNoFilter is the 'no filter' field flag.
	FNoFilter
	// FLanguage is the language field flag.
	FLanguage
	// FHidden is a flag for hidden field.
	FHidden
	// FSortable is a flag used for sortable fields.
	FSortable
	// FClientID is flag used to mark field as allowable to set ClientID.
	FClientID
	// FTime  is a flag used to mark field type as a Time.
	FTime
	// FMap is a flag used to mark field as a map.
	FMap
	// FPtr is a flag used to mark field as a pointer.
	FPtr
	// FArray is a flag used to mark field as an array.
	FArray
	// FSlice is a flag used to mark field as a slice.
	FSlice
	// FBasePtr is flag used to mark field as a based pointer.
	FBasePtr
	// FNestedStruct is a flag used to mark field as a nested structure.
	FNestedStruct
	// FNested is a flag used to mark field as nested.
	FNestedField
)

// StructField is a struct that contains all neuron specific parameters.
type StructField struct {
	// model is the model struct that this field is part of.
	mStruct *ModelStruct

	// NeuronName is neuron field name.
	neuronName string

	// fieldKind
	fieldKind FieldKind

	reflectField reflect.StructField
	relationship *Relationship

	// nested is the NestedStruct definition
	nested     *NestedStruct
	fieldFlags fieldFlag

	fieldIndex []int

	// store is the key value store used for the local usage
	store map[string]interface{}
}

// BaseType returns the base 'reflect.Type' for the provided field.
// The base is the lowest possible dereference of the field's type.
func (s *StructField) BaseType() reflect.Type {
	return s.baseFieldType()
}

// CanBeSorted returns if the struct field can be sorted.
func (s *StructField) CanBeSorted() bool {
	return s.canBeSorted()
}

// FieldIndex - gets the field index in the given model.
func (s *StructField) FieldIndex() []int {
	return s.getFieldIndex()
}

// FieldKind returns structFields kind.
func (s *StructField) FieldKind() FieldKind {
	return s.fieldKind
}

// FieldName returns struct fields name.
func (s *StructField) FieldName() string {
	return s.reflectField.Name
}

// FieldType returns field's reflect.Type.
func (s *StructField) FieldType() reflect.Type {
	return s.reflectField.Type
}

// GetDereferencedType returns structField dereferenced type.
func (s *StructField) GetDereferencedType() reflect.Type {
	return s.getDereferencedType()
}

// IsArray checks if the field is an array.
func (s *StructField) IsArray() bool {
	return s.isArray()
}

// IsBasePtr checks if the field has a pointer type in the base.
func (s *StructField) IsBasePtr() bool {
	return s.isBasePtr()
}

// IsHidden checks if the field is hidden for marshaling processes.
func (s *StructField) IsHidden() bool {
	return s.isHidden()
}

// IsI18n returns flag if the struct fields is an i18n.
func (s *StructField) IsI18n() bool {
	return s.isI18n()
}

// IsISO8601 checks if it is a time field with ISO8601 formatting.
func (s *StructField) IsISO8601() bool {
	return s.isISO8601()
}

// IsLanguage checks if the field is a language type.
func (s *StructField) IsLanguage() bool {
	return s.isLanguage()
}

// IsMap checks if the field is of map type.
func (s *StructField) IsMap() bool {
	return s.isMap()
}

// IsNestedField checks if the field is not defined within ModelStruct.
func (s *StructField) IsNestedField() bool {
	return s.isNestedField()
}

//IsNestedStruct checks if the field is a nested structure.
func (s *StructField) IsNestedStruct() bool {
	return s.nested != nil
}

// IsNoFilter checks wether the field uses no filter flag.
func (s *StructField) IsNoFilter() bool {
	return s.isNoFilter()
}

// IsOmitEmpty checks if the given field has a omitempty flag.
func (s *StructField) IsOmitEmpty() bool {
	return s.isOmitEmpty()
}

// IsPrimary checks if the field is the primary field type.
func (s *StructField) IsPrimary() bool {
	return s.fieldKind == KindPrimary
}

// IsPtr checks if the field is a pointer.
func (s *StructField) IsPtr() bool {
	return s.isPtr()
}

// IsPtrTime checks wether the field is a base ptr time flag.
func (s *StructField) IsPtrTime() bool {
	return s.isPtrTime()
}

// IsRelationship checks if given field is a relationship.
func (s *StructField) IsRelationship() bool {
	return s.isRelationship()
}

// IsSlice checks if the field is a slice based.
func (s *StructField) IsSlice() bool {
	return s.isSlice()
}

// IsSortable checks if the field has a sortable flag.
func (s *StructField) IsSortable() bool {
	return s.isSortable()
}

// IsTime checks wether the field uses time flag.
func (s *StructField) IsTime() bool {
	return s.isTime()
}

// IsZeroValue checks if the provided field has Zero value.
func (s *StructField) IsZeroValue(fieldValue interface{}) bool {
	return reflect.DeepEqual(fieldValue, reflect.Zero(s.reflectField.Type).Interface())
}

// Name returns the reflect.StructField Name.
func (s *StructField) Name() string {
	return s.reflectField.Name
}

// Nested returns nested field's structure.
func (s *StructField) Nested() *NestedStruct {
	return s.nested
}

// NeuronName returns the NeuronName.
func (s *StructField) NeuronName() string {
	return s.neuronName
}

// ReflectField returns structs reflect.StructField.
func (s *StructField) ReflectField() reflect.StructField {
	return s.reflectField
}

// Relationship returns field's Relationships.
func (s *StructField) Relationship() *Relationship {
	return s.relationship
}

// RelatedModelType gets the relationship's model type.
// Returns nil if the field is not a relationship.
func (s *StructField) RelatedModelType() reflect.Type {
	return s.getRelatedModelType()
}

// RelatedModelStruct gets the ModelStruct of the field Type.
// Returns nil if the field is not a relationship.
func (s *StructField) RelatedModelStruct() *ModelStruct {
	if s.relationship == nil {
		return nil
	}
	return s.relationship.mStruct
}

// Self returns itself. Used in the nested fields.
// Implements Structfielder interface.
func (s *StructField) Self() *StructField {
	return s
}

// SetRelatedModel sets the related model for the given struct field.
func (s *StructField) SetRelatedModel(relModel *ModelStruct) {
	if s.relationship == nil {
		s.relationship = &Relationship{
			modelType: relModel.Type(),
		}
	}

	s.relationship.mStruct = relModel
}

// SetRelationship sets the relationship value for the struct field.
func (s *StructField) SetRelationship(rel *Relationship) {
	s.relationship = rel
}

// StoreDelete deletes the store value at 'key'.
func (s *StructField) StoreDelete(key string) {
	if s.store == nil {
		return
	}
	delete(s.store, key)
	log.Debug2f("[STORE][%s][%s] deleting key: '%s'", s.mStruct.collectionType, s.neuronName, key)
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *StructField) StoreGet(key string) (interface{}, bool) {
	if s.store == nil {
		s.store = make(map[string]interface{})
		return nil, false
	}
	v, ok := s.store[key]
	return v, ok
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (s *StructField) StoreSet(key string, value interface{}) {
	if s.store == nil {
		s.store = make(map[string]interface{})
	}
	s.store[key] = value
	log.Debug2f("[STORE][%s][%s] Set Key: %s, Value: %v", s.mStruct.collectionType, s.NeuronName(), key, value)
}

func (s *StructField) String() string {
	return s.NeuronName()
}

// Struct returns fields modelstruct
func (s *StructField) Struct() *ModelStruct {
	return s.mStruct
}

// ValueFromString gets the string 'value' of the provided struct field.
func (s *StructField) ValueFromString(value string) (interface{}, error) {
	fieldValue := reflect.New(s.reflectField.Type).Elem()
	var err error
	switch s.FieldKind() {
	case KindPrimary:
		err = setPrimaryField(value, fieldValue)
	case KindAttribute:
		err = setAttributeField(value, fieldValue)
	case KindForeignKey:
		err = setPrimaryField(value, fieldValue)
	case KindFilterKey:
		err = setAttributeField(value, fieldValue)
	default:
		err = errors.NewDetf(class.ModelFieldType, "invalid field kind: '%s' for getting value", s.FieldKind().String())
	}
	if err != nil {
		return nil, err
	}
	return fieldValue.Interface(), nil
}

// TagValues returns the url.Values for the specific tag.
func (s *StructField) TagValues(tag string) url.Values {
	return s.getTagValues(tag)
}

/**

PRIVATES

*/

func (s *StructField) allowClientID() bool {
	return s.fieldFlags&FClientID != 0
}

// baseFieldType is the field's base dereferenced type
func (s *StructField) baseFieldType() reflect.Type {
	elem := s.reflectField.Type
	for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Slice ||
		elem.Kind() == reflect.Array || elem.Kind() == reflect.Map {

		elem = elem.Elem()
	}
	return elem
}

func (s *StructField) canBeSorted() bool {
	switch s.fieldKind {
	case KindRelationshipSingle, KindRelationshipMultiple, KindAttribute:
		return true
	}
	return false
}

func (s *StructField) fieldName() string {
	return s.reflectField.Name
}

func (s *StructField) fieldSetRelatedType() error {
	modelType := s.reflectField.Type
	// get error function
	getError := func() error {
		return errors.NewDetf(class.ModelRelationshipType, "incorrect relationship type provided. The Only allowable types are structs, pointers or slices. This type is: %v", modelType)
	}

	switch modelType.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Struct:
	default:
		err := getError()
		return err
	}
	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		if modelType.Kind() == reflect.Slice {
			s.fieldKind = KindRelationshipMultiple
		}
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err := getError()
		return err
	}

	if s.fieldKind == UnknownType {
		s.fieldKind = KindRelationshipSingle
	}
	if s.relationship == nil {
		s.relationship = &Relationship{}
	}

	s.relationship.modelType = modelType

	return nil
}

func (s *StructField) getDereferencedType() reflect.Type {
	t := s.reflectField.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (s *StructField) getFieldIndex() []int {
	return s.fieldIndex
}

func (s *StructField) getRelatedModelType() reflect.Type {
	if s.relationship == nil {
		return nil
	}
	return s.relationship.modelType
}

func (s *StructField) getTagValues(tag string) url.Values {
	mp := url.Values{}
	if tag == "" {
		return mp
	}

	separated := strings.Split(tag, annotation.TagSeparator)
	for _, option := range separated {
		i := strings.IndexRune(option, annotation.TagEqual)
		var values []string
		if i == -1 {
			mp[option] = values
			continue
		}
		key := option[:i]
		if i != len(option)-1 {
			values = strings.Split(option[i+1:], annotation.Separator)
		}

		mp[key] = values
	}
	return mp
}

func (s *StructField) initCheckFieldType() error {
	fieldType := s.reflectField.Type
	switch s.fieldKind {
	case KindPrimary:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		default:
			return errors.NewDetf(class.ModelFieldType, "invalid primary field type: %s for the field: %s in model: %s", fieldType, s.fieldName(), s.mStruct.modelType.Name())
		}
	case KindAttribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			return errors.NewDetf(class.ModelFieldType, "invalid attribute field type: %v for field: %s in model: %s", fieldType, s.Name(), s.mStruct.modelType.Name())
		}
		if s.isLanguage() {
			if fieldType.Kind() != reflect.String {
				return errors.NewDetf(class.ModelFieldType, "incorrect field type: %v for language field. The langtag field must be a string. Model: '%v'", fieldType, s.mStruct.modelType.Name())
			}
		}

	case KindRelationshipSingle, KindRelationshipMultiple:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.Struct:
		case reflect.Slice:
			fieldType = fieldType.Elem()
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.Struct {
				return errors.NewDetf(class.ModelRelationshipType, "invalid slice type: %v, for the relationship: %v", fieldType, s.neuronName)
			}
		default:
			return errors.NewDetf(class.ModelRelationshipType, "invalid field type: %v, for the relationship: %v", fieldType, s.neuronName)
		}
	}
	return nil
}

func (s *StructField) isArray() bool {
	return s.fieldFlags&FArray != 0
}

func (s *StructField) isBasePtr() bool {
	return s.fieldFlags&FBasePtr != 0
}

func (s *StructField) isHidden() bool {
	return s.fieldFlags&FHidden != 0
}

func (s *StructField) isI18n() bool {
	return s.fieldFlags&FI18n != 0
}

func (s *StructField) isISO8601() bool {
	return s.fieldFlags&FISO8601 != 0
}

func (s *StructField) isLanguage() bool {
	return s.fieldFlags&FLanguage != 0
}

func (s *StructField) isMap() bool {
	return s.fieldFlags&FMap != 0
}

// isNested means that the struct field is not defined within the ModelStruct
func (s *StructField) isNestedField() bool {
	return s.fieldFlags&FNestedField != 0
}

func (s *StructField) isNoFilter() bool {
	return s.fieldFlags&FNoFilter != 0
}

func (s *StructField) isOmitEmpty() bool {
	return s.fieldFlags&FOmitempty != 0
}

func (s *StructField) isPtr() bool {
	return s.fieldFlags&FPtr != 0
}

func (s *StructField) isPtrTime() bool {
	return (s.fieldFlags&FTime != 0) && (s.fieldFlags&FPtr != 0)
}

func (s *StructField) isRelationship() bool {
	return s.fieldKind == KindRelationshipMultiple || s.fieldKind == KindRelationshipSingle
}

func (s *StructField) isSlice() bool {
	return s.fieldFlags&FSlice != 0
}

func (s *StructField) isSortable() bool {
	return s.fieldFlags&FSortable != 0
}

func (s *StructField) isTime() bool {
	return s.fieldFlags&FTime != 0
}

func (s *StructField) setTagValues() error {
	tag, hasTag := s.reflectField.Tag.Lookup(annotation.Neuron)
	if !hasTag {
		return nil
	}

	var multiError errors.MultiError

	tagValues := s.TagValues(tag)
	// iterate over structfield additional tags
	for key, values := range tagValues {
		switch key {
		case annotation.FieldType, annotation.Name, annotation.ForeignKey, annotation.Relation:
			continue
		case annotation.Flags:
			s.setFlags(values...)
			continue
		}

		if !s.isRelationship() {
			log.Debugf("Field: %s tagged with: %s is not a relationship.", s.reflectField.Name, annotation.ManyToMany)
			return errors.NewDetf(class.ModelFieldTag, "%s tag on non relationship field", key)
		}

		r := s.relationship
		if r == nil {
			r = &Relationship{}
			s.relationship = r
		}

		if key == annotation.ManyToMany {
			r.kind = RelMany2Many
			// first value is join model
			// the second is the backreference field
			switch len(values) {
			case 1:
				if values[0] != "_" {
					r.joinModelName = values[0]
				}
			case 0:
			default:
				err := errors.NewDet(class.ModelFieldTag, "relationship many2many tag has too many values")
				multiError = append(multiError, err)
			}
			continue
		}

		kv := make(map[string]string)
		initLength := len(multiError)
		for _, value := range values {
			i := strings.IndexRune(value, '=')
			if i == -1 {
				err := errors.NewDetf(class.ModelFieldTag, "model: '%s' field: '%s' tag: '%s' doesn't have 'equal' sign in key=value pair: '%s'", s.Struct().Type().Name(), s.Name(), key, value)
				multiError = append(multiError, err)
				continue
			}
			kv[value[:i]] = value[i+1:]
		}
		if initLength != len(multiError) {
			continue
		}

		var errs errors.MultiError
		//`neuron:"on_delete=order=1;on_error=fail;on_change=restrict"`
		switch key {
		case annotation.OnDelete:
			errs = r.onDelete.parse(kv)
		case annotation.OnPatch:
			errs = r.onPatch.parse(kv)
		case annotation.OnCreate:
			errs = r.onCreate.parse(kv)
		default:
			errs = append(errs, errors.NewDetf(class.ModelFieldTag, "unknown relationship field tag: '%s'", key))
		}

		if len(errs) > 0 {
			multiError = append(multiError, errs...)
		}
	}
	if len(multiError) > 0 {
		return multiError
	}
	return nil
}

func (s *StructField) setFlags(flags ...string) {
	for _, single := range flags {
		switch single {
		case annotation.ClientID:
			s.setFlag(FClientID)
		case annotation.NoFilter:
			s.setFlag(FNoFilter)
		case annotation.Hidden:
			s.setFlag(FHidden)
		case annotation.NotSortable:
			s.setFlag(FSortable)
		case annotation.ISO8601:
			s.setFlag(FISO8601)
		case annotation.OmitEmpty:
			s.setFlag(FOmitempty)
		case annotation.I18n:
			s.setFlag(FI18n)
			s.mStruct.i18n = append(s.mStruct.i18n, s)
		case annotation.Language:
			s.mStruct.setLanguage(s)
		default:
			log.Debugf("Unknown field's: '%s' flag tag: '%s'", s.Name(), single)
		}
	}
}

func (s *StructField) setFlag(flag fieldFlag) {
	s.fieldFlags = s.fieldFlags | flag
}

// newStructField is the creator function for the struct field
func newStructField(refField reflect.StructField, mStruct *ModelStruct) *StructField {
	return &StructField{reflectField: refField, mStruct: mStruct}
}

func setPrimaryField(value string, fieldValue reflect.Value) errors.DetailedError {
	var err errors.DetailedError
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
		err = errors.NewDetf(class.InternalQueryFilter, "model primary invalid format type: '%s'", t.Name())
	}
	return err
}

func setAttributeField(value string, fieldValue reflect.Value) errors.DetailedError {
	var err errors.DetailedError
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
			// TODO: set the time field
			err = setTimeField(value, fieldValue)
		} else {
			// TODO: set the nested attribute struct
			err = errors.NewDet(class.QueryFilterValue, "filtering over nested structure is not supported yet")
			err.SetDetails("Filtering over nested structures is not supported yet.")
		}
	default:
		log.Debug("Filtering over unsupported type: '%s'", t.Name())

		err = errors.NewDet(class.QueryFilterValue, "Unsupported field value")
	}
	return err
}

func setTimeField(value string, fieldValue reflect.Value) errors.DetailedError {
	err := errors.NewDet(class.QueryFilterValue, "filtering over time field is not supported yet")
	err.SetDetails("Filtering over time fields is not supported yet.")
	return err
}

func setUintField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	// Parse unsigned int
	uintValue, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid unsigned integer value.")
		return err
	}

	// Set uint
	fieldValue.SetUint(uintValue)
	return nil
}

func setIntField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	intValue, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid integer value.")
		return err
	}

	// Set value if no error
	fieldValue.SetInt(intValue)
	return nil
}

func setFloatField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	// Parse float
	floatValue, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid float value.")
		return err
	}
	fieldValue.SetFloat(floatValue)

	return nil
}

func setBoolField(value string, fieldValue reflect.Value) errors.DetailedError {
	// set default if empty
	if value == "" {
		return nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid boolean value.")
		return err
	}

	fieldValue.SetBool(boolValue)
	return nil
}
