package mapping

import (
	"net/url"
	"reflect"
	"sort"
	"strings"

	"github.com/neuronlabs/neuron/annotation"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// FieldSet is a slice of fields, with some basic search functions.
type FieldSet []*StructField

// Contains checks if given fieldset contains given 'sField'.
func (f FieldSet) Contains(sField *StructField) bool {
	for _, field := range f {
		if field == sField {
			return true
		}
	}
	return false
}

// ContainsFieldName checks if a field with 'fieldName' exists in given set.
func (f FieldSet) ContainsFieldName(fieldName string) bool {
	for _, field := range f {
		if field.neuronName == fieldName || field.Name() == fieldName {
			return true
		}
	}
	return false
}

// StructField represents a field structure with its json api parameters.
// and model relationships.
type StructField struct {
	// model is the model struct that this field is part of.
	mStruct *ModelStruct
	// NeuronName is neuron field name.
	neuronName string
	// kind
	kind FieldKind

	reflectField reflect.StructField
	relationship *Relationship
	// nested is the NestedStruct definition
	nested     *NestedStruct
	fieldFlags fieldFlag

	// store is the key value store used for the local usage
	store map[string]interface{}
	Index []int
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

// Kind returns struct fields kind.
func (s *StructField) Kind() FieldKind {
	return s.kind
}

// GetDereferencedType returns structField dereferenced type.
func (s *StructField) GetDereferencedType() reflect.Type {
	return s.getDeferenceType()
}

// IsArray checks if the field is an array.
func (s *StructField) IsArray() bool {
	return s.isArray()
}

// IsBasePtr checks if the field has a pointer type in the base.
func (s *StructField) IsBasePtr() bool {
	return s.isBasePtr()
}

// IsCreatedAt returns the boolean if the field is a 'CreatedAt' field.
func (s *StructField) IsCreatedAt() bool {
	return s.isCreatedAt()
}

// IsDeletedAt returns the boolean if the field is a 'DeletedAt' field.
func (s *StructField) IsDeletedAt() bool {
	return s.isDeletedAt()
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

// IsNestedStruct checks if the field is a nested structure.
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
	return s.kind == KindPrimary
}

// IsPtr checks if the field is a pointer.
func (s *StructField) IsPtr() bool {
	return s.isPtr()
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

// IsField checks if given struct field is a primary key, attribute or foreign key field.
func (s *StructField) IsField() bool {
	switch s.kind {
	case KindPrimary, KindAttribute, KindForeignKey:
		return true
	default:
		return false
	}
}

// IsTime checks whether the field uses time flag.
func (s *StructField) IsTime() bool {
	return s.isTime()
}

// IsTimePointer checks if the field's type is a *time.time.
func (s *StructField) IsTimePointer() bool {
	return s.isPtrTime()
}

// IsUpdatedAt returns the boolean if the field is a 'UpdatedAt' field.
func (s *StructField) IsUpdatedAt() bool {
	return s.isUpdatedAt()
}

// IsZeroValue checks if the provided field has Zero value.
func (s *StructField) IsZeroValue(fieldValue interface{}) bool {
	return reflect.DeepEqual(fieldValue, reflect.Zero(s.reflectField.Type).Interface())
}

// ModelStruct returns field's model struct.
func (s *StructField) ModelStruct() *ModelStruct {
	return s.mStruct
}

// Name returns field's 'golang' name.
func (s *StructField) Name() string {
	return s.reflectField.Name
}

// Nested returns the nested structure.
func (s *StructField) Nested() *NestedStruct {
	return s.nested
}

// NeuronName returns the field's 'api' name.
func (s *StructField) NeuronName() string {
	return s.neuronName
}

// ReflectField returns reflect.StructField related with this StructField.
func (s *StructField) ReflectField() reflect.StructField {
	return s.reflectField
}

// Relationship returns relationship for provided field.
func (s *StructField) Relationship() *Relationship {
	return s.relationship
}

// String implements fmt.Stringer interface.
func (s *StructField) String() string {
	return s.NeuronName()
}

// StoreDelete deletes the store value at 'key'.
func (s *StructField) StoreDelete(key string) {
	if s.store == nil {
		return
	}
	delete(s.store, key)
	log.Debug2f("[STORE][%s][%s] deleting key: '%s'", s.mStruct.collection, s.neuronName, key)
}

// StoreGet gets the value from the store at the key: 'key'..
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
	log.Debug2f("[STORE][%s][%s] AddModel Key: %s, Models: %v", s.mStruct.collection, s.NeuronName(), key, value)
}

// Struct returns fields Model Structure.
func (s *StructField) Struct() *ModelStruct {
	return s.mStruct
}

// TagValues returns the url.Models for the specific tag.
func (s *StructField) TagValues(tag string) url.Values {
	return s.getTagValues(tag)
}

/**

Privates

*/

func (s *StructField) allowClientID() bool {
	return s.fieldFlags&fClientID != 0
}

// baseFieldType is the field's base dereference type
func (s *StructField) baseFieldType() reflect.Type {
	elem := s.reflectField.Type
	for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Slice ||
		elem.Kind() == reflect.Array || elem.Kind() == reflect.Map {
		elem = elem.Elem()
	}
	return elem
}

func (s *StructField) canBeSorted() bool {
	switch s.kind {
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
		return errors.NewDetf(ClassModelDefinition, "incorrect relationship type provided. The Only allowable types are structs, pointers or slices. This type is: %v", modelType)
	}

	switch modelType.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Struct:
	default:
		err := getError()
		return err
	}
	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		if modelType.Kind() == reflect.Slice {
			s.kind = KindRelationshipMultiple
		}
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err := getError()
		return err
	}

	if s.kind == KindUnknown {
		s.kind = KindRelationshipSingle
	}
	if s.relationship == nil {
		s.relationship = &Relationship{}
	}

	s.relationship.modelType = modelType

	return nil
}

func (s *StructField) getDeferenceType() reflect.Type {
	t := s.reflectField.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
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
	switch s.kind {
	case KindPrimary:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		case reflect.Array, reflect.Slice:
			fieldType = fieldType.Elem()
			if fieldType.Kind() != reflect.Uint8 {
				return errors.NewDetf(ClassModelDefinition, "invalid primary field type: %s for the field: %s in model: %s", fieldType, s.fieldName(), s.mStruct.modelType.Name())
			}
		default:
			return errors.NewDetf(ClassModelDefinition, "invalid primary field type: %s for the field: %s in model: %s", fieldType, s.fieldName(), s.mStruct.modelType.Name())
		}
	case KindAttribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			return errors.NewDetf(ClassModelDefinition, "invalid attribute field type: %v for field: %s in model: %s", fieldType, s.Name(), s.mStruct.modelType.Name())
		}
		if s.isLanguage() {
			if fieldType.Kind() != reflect.String {
				return errors.NewDetf(ClassModelDefinition, "incorrect field type: %v for language field. The langtag field must be a string. Model: '%v'", fieldType, s.mStruct.modelType.Name())
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
				return errors.NewDetf(ClassModelDefinition, "invalid slice type: %v, for the relationship: %v", fieldType, s.neuronName)
			}
		default:
			return errors.NewDetf(ClassModelDefinition, "invalid field type: %v, for the relationship: %v", fieldType, s.neuronName)
		}
	}
	return nil
}

func (s *StructField) isArray() bool {
	return s.fieldFlags&fArray != 0
}

func (s *StructField) isBasePtr() bool {
	return s.fieldFlags&fBasePtr != 0
}

func (s *StructField) isCreatedAt() bool {
	return s.fieldFlags&fCreatedAt != 0
}

func (s *StructField) isDeletedAt() bool {
	return s.fieldFlags&fDeletedAt != 0
}

func (s *StructField) isHidden() bool {
	return s.fieldFlags&fHidden != 0
}

func (s *StructField) isI18n() bool {
	return s.fieldFlags&fI18n != 0
}

func (s *StructField) isISO8601() bool {
	return s.fieldFlags&fISO8601 != 0
}

func (s *StructField) isLanguage() bool {
	return s.fieldFlags&fLanguage != 0
}

func (s *StructField) isMap() bool {
	return s.fieldFlags&fMap != 0
}

// isNested means that the struct field is not defined within the ModelStruct
func (s *StructField) isNestedField() bool {
	return s.fieldFlags&fNestedField != 0
}

func (s *StructField) isNoFilter() bool {
	return s.fieldFlags&fNoFilter != 0
}

func (s *StructField) isOmitEmpty() bool {
	return s.fieldFlags&fOmitempty != 0
}

func (s *StructField) isPtr() bool {
	return s.fieldFlags&fPtr != 0
}

func (s *StructField) isPtrTime() bool {
	return (s.fieldFlags&fTime != 0) && (s.fieldFlags&fPtr != 0)
}

func (s *StructField) isRelationship() bool {
	return s.kind == KindRelationshipMultiple || s.kind == KindRelationshipSingle
}

func (s *StructField) isSlice() bool {
	return s.fieldFlags&fSlice != 0
}

func (s *StructField) isSortable() bool {
	return s.fieldFlags&fSortable != 0
}

func (s *StructField) isTime() bool {
	return s.fieldFlags&fTime != 0
}

func (s *StructField) isUpdatedAt() bool {
	return s.fieldFlags&fUpdatedAt != 0
}

// Self returns itself. Used in the nested fields.
// Implements structfielder interface.
func (s *StructField) Self() *StructField {
	return s
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
			if err := s.setFlags(values...); err != nil {
				return err
			}
			continue
		}

		if !s.isRelationship() {
			log.Debugf("Field: %s tagged with: %s is not a relationship.", s.reflectField.Name, annotation.ManyToMany)
			return errors.NewDetf(ClassModelDefinition, "%s tag on non relationship field", key)
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
				err := errors.NewDet(ClassModelDefinition, "relationship many2many tag has too many values")
				multiError = append(multiError, err)
			}
			continue
		}

		kv := make(map[string]string)
		initLength := len(multiError)
		for _, value := range values {
			i := strings.IndexRune(value, '=')
			if i == -1 {
				err := errors.NewDetf(ClassModelDefinition, "model: '%s' field: '%s' tag: '%s' doesn't have 'equal' sign in key=value pair: '%s'", s.Struct().Type().Name(), s.Name(), key, value)
				multiError = append(multiError, err)
				continue
			}
			kv[value[:i]] = value[i+1:]
		}
		if initLength != len(multiError) {
			continue
		}
	}
	if len(multiError) > 0 {
		return multiError
	}
	return nil
}

func (s *StructField) setFlags(flags ...string) error {
	var err error
	for _, single := range flags {
		switch single {
		case annotation.ClientID:
			s.setFlag(fClientID)
		case annotation.NoFilter:
			s.setFlag(fNoFilter)
		case annotation.Hidden:
			s.setFlag(fHidden)
		case annotation.NotSortable:
			s.setFlag(fSortable)
		case annotation.ISO8601:
			s.setFlag(fISO8601)
		case annotation.OmitEmpty:
			s.setFlag(fOmitempty)
		case annotation.I18n:
			s.setFlag(fI18n)
		case annotation.DeletedAt:
			err = s.mStruct.setTimeRelatedField(s, fDeletedAt)
		case annotation.CreatedAt:
			err = s.mStruct.setTimeRelatedField(s, fCreatedAt)
		case annotation.UpdatedAt:
			err = s.mStruct.setTimeRelatedField(s, fUpdatedAt)
		default:
			log.Debugf("Unknown field's: '%s' flag tag: '%s'", s.Name(), single)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StructField) setFlag(flag fieldFlag) {
	s.fieldFlags |= flag
}

// setRelatedModel sets the related model for the given struct field.
func (s *StructField) setRelatedModel(relModel *ModelStruct) {
	if s.relationship == nil {
		s.relationship = &Relationship{modelType: relModel.Type()}
	}
	s.relationship.mStruct = relModel
}

// newStructField is the creator function for the struct field
// nolint:gocritic
func newStructField(refField reflect.StructField, mStruct *ModelStruct) *StructField {
	return &StructField{reflectField: refField, mStruct: mStruct}
}

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute').
type FieldKind int

const (
	// KindUnknown is the undefined field kind.
	KindUnknown FieldKind = iota
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
	// KindNested is the field's type that is signed as Nested.
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
	case KindNested:
		return "Nested"
	}
	return "Unknown"
}

var _ sort.Interface = OrderedFieldset{}

// OrderedFieldset is the wrapper over the slice of struct fields that allows to keep the fields
// in an ordered sorting. The sorts is based on the fields index.
type OrderedFieldset []*StructField

// Insert inserts the field into an ordered fields slice. In order to insert the field
// a pointer to ordered fields must be used.
func (o *OrderedFieldset) Insert(field *StructField) {
	for i, in := range *o {
		if !o.less(in, field) {
			*o = append((*o)[:i], append([]*StructField{field}, (*o)[i:]...)...) // nolint
			break
		}
	}
}

// Len implements sort.Interface interface.
func (o OrderedFieldset) Len() int {
	return len(o)
}

// Less implements sort.Interface interface.
func (o OrderedFieldset) Less(i, j int) bool {
	return o.less(o[i], o[j])
}

// Swap implements sort.Interface interface.
func (o OrderedFieldset) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o OrderedFieldset) less(first, second *StructField) bool {
	var result bool
	for k := 0; k < len(first.Index); k++ {
		if first.Index[k] != second.Index[k] {
			result = first.Index[k] < second.Index[k]
			break
		}
	}
	return result
}
