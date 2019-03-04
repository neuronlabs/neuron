package scope

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	aerrors "github.com/kucjac/jsonapi/pkg/errors"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/log"

	"github.com/kucjac/jsonapi/pkg/internal/namer/dialect"
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
	"github.com/kucjac/jsonapi/pkg/internal/query/paginations"
	"github.com/kucjac/jsonapi/pkg/internal/query/sorts"
	"github.com/kucjac/jsonapi/pkg/safemap"
	"github.com/pkg/errors"
	"golang.org/x/text/language"

	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNoParamsInContext = errors.New("No parameters in the request Context.")
	IErrNoValue          = errors.New("No value provided within the scope.")
)

var (
	// used for errors
	MaxPermissibleDuplicates = 3
)

type ScopeKind int

const (
	RootKind ScopeKind = iota
	IncludedKind
	RelationshipKind
	RelatedKind
)

// Scope contains information about given query for specific collection
// if the query defines the different collection than the main scope, then
// every detail about querying (fieldset, filters, sorts) are within new scopes
// kept in the Subscopes
type Scope struct {
	// Struct is a modelStruct this scope is based on
	mStruct *models.ModelStruct

	// Value is the values or / value of the queried object / objects
	Value interface{}

	// selectedFields are the fields that were updated
	selectedFields []*models.StructField

	// CollectionScopes contains filters, fieldsets and values for included collections
	// every collection that is inclued would contain it's subscope
	// if filters, fieldsets are set for non-included scope error should occur
	includedScopes map[*models.ModelStruct]*Scope

	// includedFields contain fields to include. If the included field is a relationship type, then
	// specific includefield contains information about it
	includedFields []*IncludeField

	// IncludeValues contain unique values for given include fields
	// the key is the - primary key value
	// the value is the single object value for given ID
	includedValues *safemap.SafeHashMap

	// PrimaryFilters contain filter for the primary field
	primaryFilters []*filters.FilterField

	// RelationshipFilters contain relationship field filters
	relationshipFilters []*filters.FilterField

	// AttributeFilters contain filter for the attribute fields
	attributeFilters []*filters.FilterField

	foreignFilters []*filters.FilterField

	// FilterKeyFilters are the for the 'FilterKey' field type
	keyFilters []*filters.FilterField

	// LanguageFilters contain information about language filters
	languageFilters *filters.FilterField

	// Fields represents fieldset used for this scope - jsonapi 'fields[collection]'
	fieldset map[string]*models.StructField

	// SortFields
	sortFields []*sorts.SortField

	// Pagination
	pagination *paginations.Pagination

	isMany bool

	// Flags is the container for all flag variablesF
	fContainer *flags.Container

	count int

	errorLimit        int
	maxNestedLevel    int
	currentErrorCount int
	totalIncludeCount int
	kind              ScopeKind

	// CollectionScope is a pointer to the scope containing the collection root
	collectionScope *Scope

	// rootScope is the root of all scopes where the query begins
	rootScope *Scope

	currentIncludedFieldIndex int
	isRelationship            bool

	// used within the root scope as a language tag for whole query.
	queryLanguage language.Tag

	hasFieldNotInFieldset bool

	ctx context.Context
}

// AddselectedFields adds provided fields into given Scope's selectedFields Container
func AddselectedFields(s *Scope, fields ...string) error {
	for _, addField := range fields {

		field := models.StructFieldByName(s.mStruct, addField)
		if field == nil {
			return errors.Errorf("Field: '%s' not found within model: %s", addField, s.mStruct.Collection())
		}
		s.selectedFields = append(s.selectedFields, field)

	}
	return nil
}

// DeleteselectedFields deletes the models.StructFields from the given scope Fieldset
func DeleteselectedFields(s *Scope, fields ...*models.StructField) error {

	erease := func(sFields *[]*models.StructField, i int) {
		if i < len(*sFields)-1 {
			(*sFields)[i] = (*sFields)[len(*sFields)-1]
		}
		(*sFields) = (*sFields)[:len(*sFields)-1]
		return
	}

ScopeFields:
	for i := len(s.selectedFields) - 1; i >= 0; i-- {
		if len(fields) == 0 {
			break ScopeFields
		}

		for j, field := range fields {
			if s.selectedFields[i] == field {
				// found the field
				// erease from fields
				erease(&fields, j)

				// Erease from Selected fields
				s.selectedFields = append(s.selectedFields[:i], s.selectedFields[i+1:]...)
				continue ScopeFields
			}
		}
	}

	if len(fields) > 0 {
		var notEreased string
		for _, field := range fields {
			notEreased += field.ApiName() + " "
		}
		return errors.Errorf("The following fields were not in the Selected fields scope: '%v'", notEreased)
	}

	return nil
}

// AddToSelectedFields adds the fields to the scope's selected fields
func (s *Scope) AddToSelectedFields(fields ...interface{}) error {
	return s.addToSelectedFields(fields...)
}

func (s *Scope) addToSelectedFields(fields ...interface{}) error {
	var selectedFields map[*models.StructField]struct{} = make(map[*models.StructField]struct{})

	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:

			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField.ApiName() == f || sField.Name() == f {
					selectedFields[sField] = struct{}{}
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
				return internal.IErrFieldNotFound
			}

		case *models.StructField:
			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField == f {
					found = true
					selectedFields[sField] = struct{}{}
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return internal.IErrFieldNotFound
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return internal.IErrInvalidType
		}
	}

	// check if fields were not already selected
	for _, alreadySelected := range s.selectedFields {
		_, ok := selectedFields[alreadySelected]
		if ok {
			log.Errorf("Field: %s already set for the given scope.", alreadySelected.Name())
			return internal.IErrFieldAlreadySelected
		}
	}

	// add all fields to scope's selected fields
	for f := range selectedFields {
		s.selectedFields = append(s.selectedFields, f)
	}

	return nil
}

func (s *Scope) AddSelectedField(field *models.StructField) {
	s.selectedFields = append(s.selectedFields, field)
}

// AddFilterField adds the filter to the given scope
func (s *Scope) AddFilterField(filter *filters.FilterField) error {
	return s.addFilterField(filter)
}

// Fieldset returns current scope fieldset
func (s *Scope) Fieldset() []*models.StructField {
	var fs []*models.StructField
	for _, field := range s.fieldset {
		fs = append(fs, field)
	}
	return fs
}

func (s *Scope) InFieldset(field string) (*models.StructField, bool) {
	f, ok := s.fieldset[field]
	return f, ok
}

func (s *Scope) Flags() *flags.Container {
	return s.getFlags()
}

// setFlags sets the flags for the given scope
func (s *Scope) SetFlags(c *flags.Container) {
	s.fContainer = c
}

// SetPaginationNoCheck sets the pagination without check
func (s *Scope) SetPaginationNoCheck(p *paginations.Pagination) {
	s.pagination = p
}

// IsRoot checks if the scope is root kind
func (s *Scope) IsRoot() bool {
	return s.isRoot()
}

// FillFieldsetIfNotSet sets the fieldset to full if the fieldset is not set
func (s *Scope) FillFieldsetIfNotSet() {
	if s.fieldset == nil || (s.fieldset != nil && len(s.fieldset) == 0) {
		s.setAllFields()
	}
}

// SetAllFields sets all fields in the fieldset
func (s *Scope) SetAllFields() {
	s.setAllFields()
}

func (s *Scope) SetFlagsFrom(flgs ...*flags.Container) {
	for _, f := range internal.ScopeCtxFlags {
		s.Flags().SetFirst(f, flgs...)
	}
}

// WithContext sets the query Scope context
func (s *Scope) WithContext(ctx context.Context) {
	// overwrite the scopIDCtx with it's own value
	ctx = context.WithValue(ctx, internal.ScopeIDCtxKey, s.ctx.Value(internal.ScopeIDCtxKey))

	ctrl := s.ctx.Value(internal.ControllerIDCtxKey)
	if ctrl != nil {
		ctx = context.WithValue(ctx, internal.ControllerIDCtxKey, ctrl)
	}

	s.ctx = ctx
}

// SetQueryLanguage sets the query language tag
func (s *Scope) SetQueryLanguage(tag language.Tag) {
	s.queryLanguage = tag
}

// QueryLanguage gets the QueryLanguage tag
func (s *Scope) QueryLanguage() language.Tag {
	return s.queryLanguage
}

// LanguageFilter return language filters for given scope
func (s *Scope) LanguageFilter() *filters.FilterField {
	return s.languageFilters
}

/**

TO DO:

Set the flags

*/
func (s *Scope) getFlags() *flags.Container {
	// if s.fContainer == nil {
	// 	// s.fContainer = flags.New()
	// 	// for _, sf := range scopeCtxFlags {
	// 	// 	s.fContainer.SetFrom(sf, s.mStruct.ctrl.fContainer)
	// 	// }
	// }

	return s.fContainer
}

func (s *Scope) deleteSelectedField(index int) error {
	if index > len(s.selectedFields)-1 {
		return errors.Errorf("Index out of range: %v", index)
	}

	// copy the last element
	if index < len(s.selectedFields)-1 {
		s.selectedFields[index] = s.selectedFields[len(s.selectedFields)-1]
		s.selectedFields[len(s.selectedFields)-1] = nil
	}
	s.selectedFields = s.selectedFields[:len(s.selectedFields)-1]
	return nil
}

func (s *Scope) addFilterField(filter *filters.FilterField) error {
	if models.FieldsStruct(filter.StructField()).ID() != s.mStruct.ID() {
		err := fmt.Errorf("Filter Struct does not match with the scope. Model: %v, filterField: %v", s.mStruct.Type().Name(), filter.StructField().Name())
		return err
	}
	switch filter.StructField().FieldKind() {
	case models.KindPrimary:
		s.primaryFilters = append(s.primaryFilters, filter)
	case models.KindAttribute:
		s.attributeFilters = append(s.attributeFilters, filter)
	case models.KindForeignKey:
		s.foreignFilters = append(s.foreignFilters, filter)
	case models.KindRelationshipMultiple, models.KindRelationshipSingle:
		s.relationshipFilters = append(s.relationshipFilters, filter)
	case models.KindFilterKey:
		s.keyFilters = append(s.keyFilters, filter)

	default:
		err := fmt.Errorf("Provided filter field of invalid kind. Model: %v. FilterField: %v", s.mStruct.Type().Name(), filter.StructField().Name())
		return err
	}
	return nil
}

// Context gets the context for given scope
func (s *Scope) Context() context.Context {
	return s.ctx
}

// Returns the collection name for given scope
func (s *Scope) GetCollection() string {
	return s.mStruct.Collection()
}

// Kind returns scope's kind
func (s *Scope) Kind() ScopeKind {
	return s.kind
}

// SetIsMany sets the isMany variable from the provided argument
func (s *Scope) SetIsMany(isMany bool) {
	s.isMany = isMany
}

// IsMany checks if the value is a slice
func (s *Scope) IsMany() bool {
	return s.isMany
}

// IncludedFields returns included fields slice
func (s *Scope) IncludedFields() []*IncludeField {
	return s.includedFields
}

// SortFields return current scope sort fields
func (s *Scope) SortFields() []*sorts.SortField {
	return s.sortFields
}

// GetCollectionScope gets the collection root scope for given scope.
// Used for included Field scopes for getting their model root scope, that contains all
func (s *Scope) getCollectionScope() *Scope {
	return s.collectionScope
}

// selectedFieldValues gets the values for given
func (s *Scope) selectedFieldValues(dialectNamer dialect.DialectFieldNamer) (
	values map[string]interface{}, err error,
) {
	if s.Value == nil {
		err = internal.IErrNilValue
		return
	}

	values = map[string]interface{}{}

	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, field := range s.selectedFields {
		fieldName := dialectNamer(field)
		// Skip empty fieldnames
		if fieldName == "" {
			continue
		}
		values[fieldName] = v.FieldByIndex(field.ReflectField().Index).Interface()
	}
	return
}

// FieldsetDialectNames gets the fieldset names, named using provided DialetFieldNamer
func (s *Scope) FieldsetDialectNames(dialectNamer dialect.DialectFieldNamer) []string {
	fieldNames := []string{}
	for _, field := range s.fieldset {
		dialectName := dialectNamer(field)
		if dialectName == "" {
			continue
		}
		fieldNames = append(fieldNames, dialectName)
	}
	return fieldNames
}

// GetFieldValue
func (s *Scope) getFieldValuePublic(field *models.StructField) (value interface{}, err error) {
	if s.Value == nil {
		err = internal.IErrNilValue
		return
	}

	if s.mStruct.ID() != models.FieldsStruct(field).ID() {
		err = errors.Errorf("Field: %s is not a Model: %v field.", field.Name(), s.mStruct.Type().Name())
		return
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	value = v.FieldByIndex(field.ReflectField().Index).Interface()
	return
}

func (s *Scope) setAllFields() {
	fieldset := map[string]*models.StructField{}

	for _, field := range models.StructAllFields(s.mStruct) {
		fieldset[field.ApiName()] = field
	}
	s.fieldset = fieldset
}

// SetFieldsetNoCheck adds fields to the scope without checking if the fields are correct.
func (s *Scope) SetFieldsetNoCheck(fields ...*models.StructField) {
	for _, field := range fields {
		s.fieldset[field.ApiName()] = field
	}
}

// SetNilFieldset sets the scope's fieldset to nil
func (s *Scope) SetEmptyFieldset() {
	s.fieldset = map[string]*models.StructField{}
}

// SetNilFieldset sets the fieldset to nil
func (s *Scope) SetNilFieldset() {
	s.fieldset = nil
}

// AddToFieldset adds the fields into the fieldset
func (s *Scope) AddToFieldset(fields ...interface{}) error {
	if s.fieldset == nil {
		s.fieldset = map[string]*models.StructField{}
	}

	return s.addToFieldset(fields...)
}

// SetFields sets the fieldset from the provided fields
func (s *Scope) SetFields(fields ...interface{}) error {
	s.fieldset = map[string]*models.StructField{}
	s.addToFieldset(fields...)
	return nil
}

func (s *Scope) addToFieldset(fields ...interface{}) error {
	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:

			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField.ApiName() == f || sField.Name() == f {
					s.fieldset[sField.ApiName()] = sField
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
				return internal.IErrFieldNotFound
			}

		case *models.StructField:
			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField == f {
					s.fieldset[sField.ApiName()] = f
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return internal.IErrFieldNotFound
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return internal.IErrInvalidType
		}
	}
	return nil
}

// getLangtagValue returns the value of the langtag for given scope
// returns error if:
//		- the scope's model does not support i18n
//		- provided nil Value for the scope
//		- the scope's Value is of invalid type
func (s *Scope) getLangtagValue() (langtag string, err error) {
	var index []int
	if index, err = s.getLangtagIndex(); err != nil {
		return
	}

	v := reflect.ValueOf(s.Value)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			err = s.errNilValueProvided()
			return
		}

		if v.Elem().Type() != s.mStruct.Type() {
			err = s.errValueTypeDoesNotMatch(v.Elem().Type())
			return
		}

		langField := v.Elem().FieldByIndex(index)
		langtag = langField.String()
		return
	case reflect.Invalid:
		err = s.errInvalidValue()
	default:
		err = fmt.Errorf("The GetLangtagValue allows single pointer type value only. Value type:'%v'", v.Type())
	}
	return
}

// IsPrimaryFieldSelected checks if the Scopes primary field is selected
func (s *Scope) IsPrimaryFieldSelected() bool {
	for _, field := range s.selectedFields {
		if field == models.StructPrimary(s.mStruct) {
			return true
		}
	}
	return false
}

// isRoot checks if given scope is a root scope of the query
func (s *Scope) isRoot() bool {
	return s.kind == RootKind
}

// setLangtagValue sets the langtag to the scope's value.
// returns an error
//		- if the Value is of invalid type or if the
//		- if the model does not support i18n
//		- if the scope's Value is nil pointer
func (s *Scope) setLangtagValue(langtag string) (err error) {
	var index []int
	if index, err = s.getLangtagIndex(); err != nil {
		return
	}

	v := reflect.ValueOf(s.Value)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return s.errNilValueProvided()
		}

		if v.Elem().Type() != s.mStruct.Type() {
			return s.errValueTypeDoesNotMatch(v.Type())
		}
		v.Elem().FieldByIndex(index).SetString(langtag)
	case reflect.Slice:

		if v.Type().Elem().Kind() != reflect.Ptr {
			return internal.IErrUnexpectedType
		}
		if t := v.Type().Elem().Elem(); t != s.mStruct.Type() {
			return s.errValueTypeDoesNotMatch(t)
		}

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.IsNil() {
				continue
			}
			elem.Elem().FieldByIndex(index).SetString(langtag)
		}

	case reflect.Invalid:
		err = s.errInvalidValue()
		return
	default:
		err = errors.New("The SetLangtagValue allows single pointer or Slice of pointers as value type. Value type")
		return
	}
	return

}

// getValueAddress gets the address of the value for given scope
// in order to set it use the SetValueFromAddressable
// func (s *Scope) getValueAddress() interface{} {
// 	return s.valueAddress
// }

// getTotalIncludeFieldCount gets the count for all included Fields. May be used
// as a wait group counter.
func (s *Scope) getTotalIncludeFieldCount() int {
	return s.totalIncludeCount
}

// getRelatedScope gets the related scope with preset filter values.
// The filter values are being taken form the root 's' Scope relationship id's.
// Returns error if the scope was not build by controller BuildRelatedScope.
func (s *Scope) getRelatedScope() (relScope *Scope, err error) {
	if len(s.includedFields) < 1 {
		return nil, fmt.Errorf("The root scope of type: '%v' was not built using controller.BuildRelatedScope() method.", s.mStruct.Type())
	}
	relatedField := s.includedFields[0]
	relScope = relatedField.Scope
	if s.Value == nil {
		err = s.errNilValueProvided()
		return
	}
	relScope.ctx = s.Context()

	scopeValue := reflect.ValueOf(s.Value)
	if scopeValue.Type().Kind() != reflect.Ptr {
		err = s.errInvalidValue()
		return
	}

	fieldValue := scopeValue.Elem().FieldByIndex(relatedField.FieldIndex())
	var primaries reflect.Value

	// if no values are present within given field
	// return nil scope.Value
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return
	} else if fieldValue.Kind() == reflect.Slice && fieldValue.Len() == 0 {
		relScope.isMany = true
		return
	} else {

		primaries, err = models.FieldsRelatedModelStruct(relatedField.StructField).PrimaryValues(fieldValue)
		if err != nil {
			return
		}
		if primaries.Kind() == reflect.Slice {
			if primaries.Len() == 0 {
				return
			}
		}
	}

	filterValue := &filters.OpValuePair{}

	if relatedField.FieldKind() == models.KindRelationshipSingle {
		filterValue.SetOperator(filters.OpEqual)
		filterValue.Values = append(filterValue.Values, primaries.Interface())
		relScope.newValueSingle()
	} else {
		filterValue.SetOperator(filters.OpIn)
		var values []interface{}
		for i := 0; i < primaries.Len(); i++ {
			values = append(values, primaries.Index(i).Interface())
		}
		filterValue.Values = values
		relScope.isMany = true
		relScope.newValueMany()
	}
	primaryFilter := filters.NewFilter(relatedField.StructField, filterValue)

	relScope.primaryFilters = append(relScope.primaryFilters, primaryFilter)
	return
}

// getPrimaryFieldValues - gets the primary field values from the scope.
// Returns the values within the []interface{} form
//			returns	- IErrNoValue if no value provided.
//					- IErrInvalidType if the scope's value is of invalid type
// 					- *reflect.ValueError if internal occurs.
func (s *Scope) getPrimaryFieldValues() (values []interface{}, err error) {
	if s.Value == nil {
		err = IErrNoValue
		return
	}

	defer func() {
		if r := recover(); r != nil {
			switch vt := r.(type) {
			case *reflect.ValueError:
				err = vt
			default:
				err = fmt.Errorf("Internal error")
			}
		}
	}()

	primaryIndex := models.StructPrimary(s.mStruct).FieldIndex()

	addPrimaryValue := func(single reflect.Value) {
		primaryValue := single.Elem().FieldByIndex(primaryIndex)
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			addPrimaryValue(v.Index(i))
		}
	case reflect.Ptr:
		addPrimaryValue(v)
	default:
		err = internal.IErrInvalidType
		return
	}
	return
}

// GetRelationshipScope - for given root Scope 's' gets the value of the relationship
// used for given request and set it's value into relationshipScope.
// returns an error if the value is not set or there is no relationship includedField
// for given scope.
func (s *Scope) GetRelationshipScope() (relScope *Scope, err error) {
	if len(s.includedFields) != 1 {
		return nil, errors.New("Provided invalid includedFields for given scope.")
	}
	if err = s.setIncludedFieldValue(s.includedFields[0]); err != nil {
		return
	}
	relScope = s.includedFields[0].Scope
	return
}

// NewValueMany creates empty slice of ptr value for given scope
// value is of type []*models.ModelStruct.Type
func (s *Scope) NewValueMany() {
	s.newValueMany()
}

// NewValueSingle creates new value for given scope of a type *models.ModelStruct.Type
func (s *Scope) NewValueSingle() {
	s.newValueSingle()
}

// SetCollectionValues iterate over the scope's Value field and add it to the collection root
// scope.if the collection root scope contains value with given primary field it checks if given // scope containsincluded fields that are not within fieldset. If so it adds the included field
// value to the value that were inside the collection root scope.
func (s *Scope) SetCollectionValues() error {
	if s.collectionScope.includedValues == nil {
		s.collectionScope.includedValues = safemap.New()
	}
	s.collectionScope.includedValues.Lock()
	defer s.collectionScope.includedValues.Unlock()

	var (
		primIndex = s.mStruct.PrimaryField().FieldIndex()

		setValueToCollection = func(value reflect.Value) {

			primaryValue := value.Elem().FieldByIndex(primIndex)
			if !primaryValue.IsValid() {
				return
			}
			primary := primaryValue.Interface()
			insider, ok := s.collectionScope.includedValues.UnsafeGet(primary)
			if !ok {
				s.collectionScope.includedValues.UnsafeAdd(primary, value.Interface())
				return
			}

			if insider == nil {
				// in order to prevent the nil values set within given key
				s.collectionScope.includedValues.UnsafeAdd(primary, value.Interface())
			} else if s.hasFieldNotInFieldset {
				// this scopes value should have more fields
				insideValue := reflect.ValueOf(insider)

				for _, included := range s.includedFields {
					// only the fields that are not in the fieldset should be added
					if included.NotInFieldset {

						// get the included field index
						index := included.FieldIndex()

						// check if included field in the collection Values has this field
						if insideField := insideValue.Elem().FieldByIndex(index); !insideField.IsNil() {
							thisField := value.Elem().FieldByIndex(index)
							if thisField.IsNil() {
								thisField.Set(insideField)
							}
						}
					}
				}
			}
		}
	)

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return internal.IErrUnexpectedType
	}
	if !v.IsNil() {
		v = v.Elem()
		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Type().Kind() != reflect.Ptr {
					return internal.IErrUnexpectedType
				}
				if !elem.IsNil() {
					setValueToCollection(elem)
				}
			}
		case reflect.Struct:
			log.Debugf("Struct setValueToCollection")
			setValueToCollection(v)

		default:
			err := internal.IErrUnexpectedType
			return err
		}
	} else {
		log.Debugf("Nil field's value.")
	}

	return nil
}

// NextIncludedField allows iteration over the includedFields.
// If there is any included field it changes the current field index to the next available.
func (s *Scope) NextIncludedField() bool {
	if s.currentIncludedFieldIndex >= len(s.includedFields)-1 {
		return false
	}

	s.currentIncludedFieldIndex++
	return true
}

// CurrentIncludedField gets current included field, based on the index
func (s *Scope) CurrentIncludedField() (*IncludeField, error) {
	if s.currentIncludedFieldIndex == -1 || s.currentIncludedFieldIndex > len(s.includedFields)-1 {
		return nil, errors.New("Getting non-existing included field.")
	}

	return s.includedFields[s.currentIncludedFieldIndex], nil
}

// ResetIncludedField resets the current included field pointer
func (s *Scope) ResetIncludedField() {
	s.currentIncludedFieldIndex = -1
}

// SetIDFilters sets the ID Filter for given values.
func (s *Scope) SetIDFilters(idValues ...interface{}) {
	s.setIDFilterValues(idValues...)
}

// SetPrimaryFilters sets the primary filter for given values.
func (s *Scope) SetPrimaryFilters(values ...interface{}) {
	s.setIDFilterValues(values...)
}

// Sets the LanguageFilter for given scope.
// If the scope's model does not support i18n it does not create language filter, and ends fast.
func (s *Scope) SetLanguageFilter(languages ...interface{}) {
	s.setLanguageFilterValues(languages...)
}

// SetBelongsToForeignKeyFields
func (s *Scope) SetBelongsToForeignKeyFields() error {
	if s.Value == nil {
		return internal.IErrNilValue
	}

	setField := func(v reflect.Value) ([]*models.StructField, error) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Type() != s.Struct().Type() {
			return nil, internal.IErrInvalidType
		}

		var fks []*models.StructField
		for _, field := range s.selectedFields {
			relField, ok := s.mStruct.RelationshipField(field.ApiName())
			if ok {
				rel := relField.Relationship()
				if rel != nil && rel.Kind() == models.RelBelongsTo {
					relVal := v.FieldByIndex(relField.ReflectField().Index)

					// Check if the value is non zero
					if reflect.DeepEqual(
						relVal.Interface(),
						reflect.Zero(relVal.Type()).Interface(),
					) {
						// continue if non zero
						continue
					}

					if relVal.Kind() == reflect.Ptr {
						relVal = relVal.Elem()
					}

					fkVal := v.FieldByIndex(rel.ForeignKey().ReflectField().Index)

					relPrim := rel.Struct().PrimaryField()

					relPrimVal := relVal.FieldByIndex(relPrim.ReflectField().Index)
					fkVal.Set(relPrimVal)
					fks = append(fks, rel.ForeignKey())
				}

			}
		}
		return fks, nil
	}

	v := reflect.ValueOf(s.Value)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:

		fks, err := setField(v)
		if err != nil {
			return err
		}
		for _, fk := range fks {
			var found bool

		inner:
			for _, selected := range s.selectedFields {
				if fk == selected {
					found = true
					break inner
				}
			}
			if !found {
				s.selectedFields = append(s.selectedFields, fk)
			}
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			fks, err := setField(v)
			if err != nil {
				return errors.Wrapf(err, "At index: %d. Value: %v", i, elem.Interface())
			}
			for _, fk := range fks {
				var found bool
			innerSlice:
				for _, selected := range s.selectedFields {
					if fk == selected {
						found = true
						break innerSlice
					}
				}
				if !found {
					s.selectedFields = append(s.selectedFields, fk)
				}
			}
		}
	}
	return nil
}

// SelectedFields return fields that were selected during unmarshaling
func (s *Scope) SelectedFields() []*models.StructField {
	return s.selectedFields
}

// Struct returns scope's model struct
func (s *Scope) Struct() *models.ModelStruct {
	return s.mStruct
}

// UseI18n is a bool that defines if given scope uses the i18n field.
// I.e. it allows to predefine if model should set language filter.
func (s *Scope) UseI18n() bool {
	return s.mStruct.UseI18n()
}

func (s *Scope) GetScopeValueString() string {
	var value string
	v := reflect.ValueOf(s.Value)
	if v.IsNil() {
		return "No Value"
	}
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Elem()
			value += fmt.Sprintf("%+v;", elem.Interface())
		}
	case reflect.Ptr:
		elem := v.Elem()
		value = fmt.Sprintf("%+v;", elem.Interface())
	}
	return value
}

// SetValueFromAddressable - lack of generic makes it hard for preparing addressable value.
// While getting the addressable value with GetValueAddress, this function makes use of it
// by setting the Value from addressable.
// // Returns an error if the addressable is nil.
// func (s *Scope) SetValueFromAddressable() error {
// 	return s.setValueFromAddressable()
// }

// SetCollectionScope sets the collection scope for given scope
func (s *Scope) SetCollectionScope(cs *Scope) {
	s.collectionScope = cs
}

// CurrentErrorCount returns current error count number
func (s *Scope) CurrentErrorCount() int {
	return s.currentErrorCount
}

// IncreaseErrorCount adds another error count for the given scope
func (s *Scope) IncreaseErrorCount(count int) {
	s.currentErrorCount += count
}

func (s *Scope) InitializeIncluded(maxNestedLevel int) {
	s.includedScopes = make(map[*models.ModelStruct]*Scope)
	s.maxNestedLevel = maxNestedLevel
}

// PrimaryFilters returns scopes primary filter values
func (s *Scope) PrimaryFilters() []*filters.FilterField {
	return s.primaryFilters
}

// AttributeFilters returns scopes attribute filters
func (s *Scope) AttributeFilters() []*filters.FilterField {
	return s.attributeFilters
}

// RelationshipFilters returns scopes relationship filters
func (s *Scope) RelationshipFilters() []*filters.FilterField {
	return s.relationshipFilters
}

// FilterKeyFilters return key filters for the scope
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	return s.keyFilters
}

// SetKind sets the scope's kind
func (s *Scope) SetKind(kind ScopeKind) {
	s.kind = kind
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *models.ModelStruct) *Scope {

	scope := &Scope{
		mStruct:                   modelStruct,
		fieldset:                  make(map[string]*models.StructField),
		currentIncludedFieldIndex: -1,
		fContainer:                flags.New(),
	}

	for _, sf := range internal.ScopeCtxFlags {
		scope.fContainer.SetFrom(sf, models.StructFlags(modelStruct))
	}

	// set all fields
	for _, field := range models.StructAllFields(modelStruct) {
		scope.fieldset[field.ApiName()] = field
	}

	return scope
}

/**

FIELDSET

*/

// fields[collection] = field1, field2
func (s *Scope) BuildFieldset(fields ...string) (errs []*aerrors.ApiError) {
	var (
		errObj *aerrors.ApiError
	)

	if len(fields) > models.StructWorkingFieldCount(s.mStruct) {
		errObj = aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Too many fields to set.")
		errs = append(errs, errObj)
		return
	}

	prim := models.StructPrimary(s.mStruct)
	s.fieldset = map[string]*models.StructField{
		prim.Name(): prim,
	}

	for _, field := range fields {
		if field == "" {
			continue
		}

		sField, err := s.checkField(field)
		if err != nil {
			if field == "id" {
				err = aerrors.ErrInvalidQueryParameter.Copy()
				err.Detail = "Invalid fields parameter. 'id' is not a field - it is primary key."
			}
			errs = append(errs, err)
			continue
		}

		_, ok := s.fieldset[sField.ApiName()]
		if ok {
			// duplicate
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Duplicated fieldset parameter: '%s' for: '%s' collection.", field, s.mStruct.Collection())
			errs = append(errs, errObj)
			if len(errs) > MaxPermissibleDuplicates {
				return
			}
			continue
		}
		s.fieldset[sField.ApiName()] = sField

		if sField.IsRelationship() {

			r := models.FieldRelationship(sField)
			if r != nil && models.RelationshipGetKind(r) == models.RelBelongsTo {
				if fk := models.RelationshipForeignKey(r); fk != nil {
					s.fieldset[fk.ApiName()] = fk
				}
			}
		}
	}

	return

}

/**

FILTERS

*/

func (s *Scope) setIDFilterValues(values ...interface{}) {
	s.setPrimaryFilterValues(models.StructPrimary(s.mStruct), values...)
	return
}

func (s *Scope) setLanguageFilterValues(values ...interface{}) {
	filter := s.getOrCreateLangaugeFilter()
	if filter == nil {
		return
	}

	fv := &filters.OpValuePair{}
	fv.SetOperator(filters.OpIn)

	fv.Values = append(fv.Values, values...)
	filters.FilterAppendValues(filter, fv)

	return
}

func (s *Scope) setPrimaryFilterValues(primField *models.StructField, values ...interface{}) {
	filter := s.getOrCreatePrimaryFilter(primField)

	filters.FilterAppendValues(filter, filters.NewOpValuePair(filters.OpIn, values...))
}

func (s *Scope) getOrCreatePrimaryFilter(primField *models.StructField) (filter *filters.FilterField) {
	if s.primaryFilters == nil {
		s.primaryFilters = []*filters.FilterField{}
	}

	for _, pf := range s.primaryFilters {

		if pf.StructField() == primField {
			filter = pf
			return
		}
	}

	// if not found within primary filters
	filter = filters.NewFilter(primField)

	s.primaryFilters = append(s.primaryFilters, filter)

	return filter
}

// GetOrCreateIDFilter gets or creates new filterField
func (s *Scope) GetOrCreateIDFilter() *filters.FilterField {
	return s.getOrCreateIDFilter()
}

func (s *Scope) getOrCreateIDFilter() (filter *filters.FilterField) {
	return s.getOrCreatePrimaryFilter(models.StructPrimary(s.mStruct))
}

func (s *Scope) GetOrCreateLanguageFilter() (filter *filters.FilterField) {
	return s.getOrCreateLangaugeFilter()
}

func (s *Scope) getOrCreateLangaugeFilter() (filter *filters.FilterField) {
	if !s.mStruct.UseI18n() {
		return nil
	}

	if s.languageFilters != nil {
		return s.languageFilters
	}

	filter = filters.NewFilter(models.StructLanguage(s.mStruct))
	s.languageFilters = filter
	return

}

// GetOrCreateAttributeFilter creates or gets existing attribute filter for given sField
func (s *Scope) GetOrCreateAttributeFilter(
	sField *models.StructField,
) (filter *filters.FilterField) {
	return s.getOrCreateAttributeFilter(sField)
}

func (s *Scope) getOrCreateAttributeFilter(
	sField *models.StructField,
) (filter *filters.FilterField) {

	if s.attributeFilters == nil {
		s.attributeFilters = []*filters.FilterField{}
	}

	for _, attrFilter := range s.attributeFilters {
		if attrFilter.StructField() == sField {
			filter = attrFilter
			return
		}
	}
	filter = filters.NewFilter(sField)
	s.attributeFilters = append(s.attributeFilters, filter)

	return filter
}

//GetOrCreateFilterKeyFilter creates or get an existing filter field
func (s *Scope) GetOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateFilterKeyFilter(sField)
}

func (s *Scope) getOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	if s.keyFilters == nil {
		s.keyFilters = []*filters.FilterField{}
	}

	for _, fkFilter := range s.keyFilters {
		if fkFilter.StructField() == sField {
			filter = fkFilter
			return
		}
	}
	filter = filters.NewFilter(sField)
	s.keyFilters = append(s.keyFilters, filter)
	return filter
}

// getOrCreateForeignKeyFilter gets the filter field for given StructField
// If the filterField already exists for given scope, the function returns the existing one.
// Otherwise it craetes new filter field and returns it.
func (s *Scope) getOrCreateForeignKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	if s.foreignFilters == nil {
		s.foreignFilters = []*filters.FilterField{}
	}

	for _, fkFilter := range s.foreignFilters {
		if fkFilter.StructField() == sField {
			filter = fkFilter
			return
		}
	}
	filter = filters.NewFilter(sField)
	s.foreignFilters = append(s.foreignFilters, filter)
	return filter
}

// GetOrCreateRelationshipFilter creates or gets existing fitler field for given struct field.
func (s *Scope) GetOrCreateRelationshipFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateRelationshipFilter(sField)
}

func (s *Scope) getOrCreateRelationshipFilter(sField *models.StructField) (filter *filters.FilterField) {
	// Create if empty
	if s.relationshipFilters == nil {
		s.relationshipFilters = []*filters.FilterField{}
	}

	// Check if no relationship filter already exists
	for _, relFilter := range s.relationshipFilters {
		if relFilter.StructField() == sField {
			filter = relFilter

			return
		}
	}

	filter = filters.NewFilter(sField)
	s.relationshipFilters = append(s.relationshipFilters, filter)
	return filter
}

/**

INCLUDES

*/

// IncludedScopes returns included scopes
func (s *Scope) IncludedScopes() []*Scope {
	if len(s.includedScopes) == 0 {
		return nil
	}

	scopes := []*Scope{}
	for _, included := range s.includedScopes {
		scopes = append(scopes, included)
	}
	return scopes
}

// IncludeScopeByStruct returns the included scope by model struct
func (s *Scope) IncludeScopeByStruct(mStruct *models.ModelStruct) (*Scope, bool) {
	scope, ok := s.includedScopes[mStruct]
	return scope, ok
}

// IncludedScopes returns included scopes
func (s *Scope) IncludedValues() *safemap.SafeHashMap {
	return s.includedValues
}

// BuildIncludeList provide fast checks for the includedList
// if given include passes use buildInclude method on it.
func (s *Scope) BuildIncludeList(includedList ...string,
) (errs []*aerrors.ApiError) {
	var errorObjects []*aerrors.ApiError
	var errObj *aerrors.ApiError

	// check if the number of included fields is possible
	if len(includedList) > models.StructMaxIncludedCount(s.mStruct) {
		errObj = aerrors.ErrOutOfRangeQueryParameterValue.Copy()
		errObj.Detail = fmt.Sprintf("Too many included parameter values for: '%s' collection.",
			s.mStruct.Collection())
		errs = append(errs, errObj)
		return
	}

	// includedScopes for root are always set
	s.includedScopes = make(map[*models.ModelStruct]*Scope)

	var includedMap map[string]int

	// many includes flag if there is more than one include
	var manyIncludes bool = len(includedList) > 1

	if manyIncludes {
		includedMap = make(map[string]int)
	}

	// having multiple included in the query
	for _, included := range includedList {

		// check the nested level of every included
		annotCount := strings.Count(included, internal.AnnotationNestedSeperator)
		if annotCount > s.maxNestedLevel {
			log.Debugf("AnnotCount: %v MaxNestedLevel: %v", annotCount, s.maxNestedLevel)
			errs = append(errs, aerrors.ErrTooManyNestedRelationships(included))
			continue
		}

		// if there are more than one include
		if manyIncludes {

			// assert no duplicates are provided in the include list
			includedCount := includedMap[included]
			includedCount++
			includedMap[included] = includedCount
			if annotCount == 0 && includedCount > 1 {
				if includedCount == 2 {
					errObj = aerrors.ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Included parameter '%s' used more than once.", included)
					errs = append(errs, errObj)
					continue
				} else if includedCount >= MaxPermissibleDuplicates {
					break
				}
			}
		}
		errorObjects = s.buildInclude(included)
		errs = append(errs, errorObjects...)
	}
	return
}

// buildInclude searches for the relationship field within given scope
// if not found, then tries to seperate the 'included' argument
// by the 'annotationNestedSeperator'. If seperated correctly
// it tries to create nested fields.
// adds IncludeScope for given field.
func (s *Scope) buildInclude(included string) (errs []*aerrors.ApiError) {
	var includedField *IncludeField
	// search for the 'included' in the model's
	relationField, ok := models.StructRelField(s.mStruct, included)
	if !ok {
		// no relationship found check nesteds
		index := strings.Index(included, internal.AnnotationNestedSeperator)
		if index == -1 {
			errs = append(errs, errNoRelationship(s.mStruct.Collection(), included))
			return
		}

		// root part of included (root.subfield)
		field := included[:index]
		relationField, ok := models.StructRelField(s.mStruct, field)
		if !ok {
			errs = append(errs, errNoRelationship(s.mStruct.Collection(), field))
			return
		}

		// create new included field
		includedField = s.getOrCreateIncludeField(relationField)

		errs = includedField.Scope.buildInclude(included[index+1:])
		if errs != nil {
			return
		}

	} else {
		// create new includedField if the field was not already created during nested process.
		includedField = s.getOrCreateIncludeField(relationField)
	}

	includedField.Scope.kind = IncludedKind
	return
}

// CopyIncludedBoundaries copies all included data from scope's included fields
// Into it's included scopes.
func (s *Scope) CopyIncludedBoundaries() {
	s.copyIncludedBoundaries()
}

// copies the and fieldset for given include and it's nested fields.
func (s *Scope) copyIncludedBoundaries() {
	for _, includedField := range s.includedFields {
		includedField.copyScopeBoundaries()
	}
}

// CreateModelsRootScope creates scope for given model (mStruct) and
// stores it within the rootScope.includedScopes.
// Used for collection unique root scopes
// (filters, fieldsets etc. for given collection scope)
func (s *Scope) CreateModelsRootScope(mStruct *models.ModelStruct) *Scope {
	return s.createModelsRootScope(mStruct)
}

// createModelsRootScope creates scope for given model (mStruct) and
// stores it within the rootScope.includedScopes.
// Used for collection unique root scopes
// (filters, fieldsets etc. for given collection scope)
func (s *Scope) createModelsRootScope(mStruct *models.ModelStruct) *Scope {
	scope := s.createModelsScope(mStruct)
	scope.rootScope.includedScopes[mStruct] = scope
	scope.includedValues = safemap.New()

	*scope.fContainer = *scope.rootScope.fContainer.Copy()

	return scope
}

// getModelsRootScope returns the scope for given model that is stored within
// the rootScope
func (s *Scope) GetModelsRootScope(mStruct *models.ModelStruct) (collRootScope *Scope) {
	if s.rootScope == nil {
		// if 's' is root scope and is related to model that is looking for
		if s.mStruct == mStruct {
			return s
		}

		return s.includedScopes[mStruct]
	}

	return s.rootScope.includedScopes[mStruct]
}

// getOrCreateModelsRootScope gets ModelsRootScope and if it is null it creates new.
func (s *Scope) getOrCreateModelsRootScope(mStruct *models.ModelStruct) *Scope {
	scope := s.GetModelsRootScope(mStruct)
	if scope == nil {
		scope = s.createModelsRootScope(mStruct)
	}
	return scope
}

// NonRootScope creates non root scope
func (s *Scope) NonRootScope(mStruct *models.ModelStruct) *Scope {
	return s.createModelsScope(mStruct)
}

// createsModelsScope
func (s *Scope) createModelsScope(mStruct *models.ModelStruct) *Scope {
	scope := newScope(mStruct)
	if s.rootScope == nil {
		scope.rootScope = s
	} else {
		scope.rootScope = s.rootScope
	}

	scope.ctx = context.WithValue(s.Context(), internal.ScopeIDCtxKey, uuid.New())

	return scope
}

// GetOrCreateIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field
func (s *Scope) GetOrCreateIncludeField(field *models.StructField,
) (includeField *IncludeField) {
	return s.getOrCreateIncludeField(field)
}

// createOrGetIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field
func (s *Scope) getOrCreateIncludeField(
	field *models.StructField,
) (includeField *IncludeField) {
	for _, included := range s.includedFields {
		if included.StructField == field {
			return included
		}
	}

	return s.createIncludedField(field)
}

func (s *Scope) createIncludedField(
	field *models.StructField,
) (includeField *IncludeField) {
	includeField = newIncludeField(field, s)
	if s.includedFields == nil {
		s.includedFields = make([]*IncludeField, 0)
	}

	s.includedFields = append(s.includedFields, includeField)
	return
}

// setIncludedFieldValue - used while getting the Relationship Scope,
// and the 's' has the 'includedField' value in it's value.
func (s *Scope) setIncludedFieldValue(includeField *IncludeField) error {
	if s.Value == nil {
		return s.errNilValueProvided()
	}
	v := reflect.ValueOf(s.Value)

	switch v.Kind() {
	case reflect.Ptr:
		if t := v.Elem().Type(); t != s.mStruct.Type() {
			return s.errValueTypeDoesNotMatch(t)
		}
		includeField.setRelationshipValue(v.Elem())
	default:
		return fmt.Errorf("Scope has invalid value type: %s", v.Type())
	}
	return nil
}

/**

PAGINATION

*/

// Pagination returns scope's pagination
func (s *Scope) Pagination() *paginations.Pagination {
	return s.pagination
}

// PreparepaginatedValue prepares paginated value for given key, value and index
func (s *Scope) PreparePaginatedValue(key, value string, index paginations.Parameter) *aerrors.ApiError {
	val, err := strconv.Atoi(value)
	if err != nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided query parameter: %v, contains invalid value: %v. Positive integer value is required.", key, value)
		return errObj
	}

	if s.pagination == nil {
		s.pagination = &paginations.Pagination{}
	}
	switch index {
	case 0:
		s.pagination.Limit = val
		s.pagination.SetType(paginations.TpOffset)
	case 1:
		s.pagination.Offset = val
		s.pagination.SetType(paginations.TpOffset)
	case 2:
		s.pagination.PageNumber = val
		s.pagination.SetType(paginations.TpPage)
	case 3:
		s.pagination.PageSize = val
		s.pagination.SetType(paginations.TpPage)
	}
	return nil
}

/**

SORTS

*/
// BuildSortFields sets the sort fields for given string array.
func (s *Scope) BuildSortFields(sortFields ...string) (errs []*aerrors.ApiError) {
	return s.buildSortFields(sortFields...)
}

// setSortFields sets the sort fields for given string array.
func (s *Scope) buildSortFields(sortFields ...string) (errs []*aerrors.ApiError) {
	var (
		err      *aerrors.ApiError
		order    sorts.Order
		fields   map[string]int = make(map[string]int)
		badField                = func(fieldName string) {
			err = aerrors.ErrInvalidQueryParameter.Copy()
			err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", fieldName, s.mStruct.Collection())
			errs = append(errs, err)
		}
		invalidField bool
	)

	// If the number of sort fields is too long then do not allow
	if len(sortFields) > models.StructSortScopeCount(s.mStruct) {
		err = aerrors.ErrOutOfRangeQueryParameterValue.Copy()
		err.Detail = fmt.Sprintf("There are too many sort parameters for the '%v' collection.", s.mStruct.Collection())
		errs = append(errs, err)
		return
	}

	for _, sort := range sortFields {
		if sort[0] == '-' {
			order = sorts.DescendingOrder
			sort = sort[1:]

		} else {
			order = sorts.AscendingOrder
		}

		// check if no dups provided
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			if count == 2 {
				err = aerrors.ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Sort parameter: %v used more than once.", sort)
				errs = append(errs, err)
				continue
			} else if count > 2 {
				break
			}
		}

		invalidField = newSortField(sort, order, s)
		if invalidField {
			badField(sort)
			continue
		}

	}
	return
}

// VALUES

func (s *Scope) newValueSingle() {
	s.Value = reflect.New(s.mStruct.Type()).Interface()
}

func (s *Scope) newValueMany() {
	s.Value = reflect.New(reflect.SliceOf(reflect.New(s.mStruct.Type()).Type())).Interface()
	s.isMany = true
}

// func (s *Scope) setValueFromAddressable() error {
// 	if s.valueAddress != nil && reflect.TypeOf(s.valueAddress).Kind() == reflect.Ptr {
// 		s.Value = reflect.ValueOf(s.valueAddress).Elem().Interface()
// 		return nil
// 	}
// 	return fmt.Errorf("Provided invalid valueAddress for scope of type: %v. ValueAddress: %v", s.mStruct.Type(), s.valueAddress)
// }

// GetPrimaryFieldValue
func (s *Scope) GetPrimaryFieldValue() (reflect.Value, error) {
	return s.getFieldValue(s.Struct().PrimaryField())
}

// GetFieldValue gets the value of the provided field
func (s *Scope) GetFieldValue(sField *models.StructField) (reflect.Value, error) {
	return s.getFieldValue(sField)
}

func (s *Scope) getFieldValue(sField *models.StructField) (reflect.Value, error) {
	return modelValueByStructField(s.Value, sField)
}

func (s *Scope) setBelongsToForeignKey() error {
	if s.Value == nil {
		return errors.Errorf("Nil value provided. %#v", s)
	}
	v := reflect.ValueOf(s.Value)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("Provided Scope Value is not addressable.")
	}

	switch v.Elem().Kind() {
	case reflect.Struct:
		err := models.StructSetBelongsToForeigns(s.mStruct, v)
		if err != nil {
			return err
		}

	case reflect.Slice:
		v = v.Elem()
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			err := models.StructSetBelongsToForeigns(s.mStruct, elem)
			if err != nil {
				return errors.Wrapf(err, "At index: %d. Value: %v", i, elem.Interface())
			}
		}
	}
	return nil
}

func (s *Scope) checkField(field string) (*models.StructField, *aerrors.ApiError) {
	sField, err := models.StructCheckField(s.mStruct, field)
	if err != nil {
		return nil, err
	}
	return sField, nil
}

// func (m *models.ModelStruct) setBelongsToForeignsWithFields(
// 	v reflect.Value, scope *Scope,
// ) ([]*models.StructField, error) {
// 	if v.Kind() == reflect.Ptr {
// 		v = v.Elem()
// 	}

// 	if v.Type() != m.modelType {
// 		return nil, errors.Errorf("Invalid model type. Wanted: %v. Actual: %v", m.modelType.Name(), v.Type().Name())
// 	}
// 	fks := []*models.StructField{}
// 	for _, field := range scope.selectedFields {
// 		rel, ok := m.relationships[field.jsonAPIName]
// 		if ok &&
// 			rel.relationship != nil &&
// 			rel.relationship.Kind == RelBelongsTo {
// 			relVal := v.FieldByIndex(rel.refStruct.Index)
// 			if reflect.DeepEqual(relVal.Interface(), reflect.Zero(relVal.Type()).Interface()) {
// 				continue
// 			}
// 			if relVal.Kind() == reflect.Ptr {
// 				relVal = relVal.Elem()
// 			}
// 			fkVal := v.FieldByIndex(rel.relationship.ForeignKey.refStruct.Index)
// 			relPrim := rel.relatedStruct.primary
// 			relPrimVal := relVal.FieldByIndex(relPrim.refStruct.Index)
// 			fkVal.Set(relPrimVal)
// 			fks = append(fks, rel.relationship.ForeignKey)
// 		}
// 	}
// 	return fks, nil
// }

func (s *Scope) setBelongsToRelationWithFields(fields ...*models.StructField) error {
	if s.Value == nil {
		return errors.Errorf("Nil value provided. %#v", s)
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return errors.Errorf("Provided scope value is not a pointer.")
	}

	switch v.Elem().Kind() {
	case reflect.Struct:
		err := models.StructSetBelongsToRelationWithFields(s.mStruct, v, fields...)
		if err != nil {
			return err
		}

	case reflect.Slice:
		v = v.Elem()
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			err := models.StructSetBelongsToRelationWithFields(s.mStruct, elem, fields...)
			if err != nil {
				return errors.Wrapf(err, "At index: %d. Value: %v", i, elem.Interface())
			}
		}
	}
	return nil
}

// func (s *Scope) setBelongsToForeignKeyWithFields() error {
// 	if s.Value == nil {
// 		return errors.Errorf("Nil value provided. %#v", s)
// 	}

// 	v := reflect.ValueOf(s.Value)
// 	switch v.Kind() {
// 	case reflect.Ptr:
// 		fks, err := s.mStruct.setBelongsToForeignsWithFields(v, s)
// 		if err != nil {
// 			return err
// 		}
// 		for _, fk := range fks {
// 			var found bool
// 		inner:
// 			for _, selected := range s.selectedFields {
// 				if fk == selected {
// 					found = true
// 					break inner
// 				}
// 			}
// 			if !found {
// 				s.selectedFields = append(s.selectedFields, fk)
// 			}
// 		}

// 	case reflect.Slice:
// 		for i := 0; i < v.Len(); i++ {
// 			elem := v.Index(i)
// 			fks, err := s.mStruct.setBelongsToForeignsWithFields(elem, s)
// 			if err != nil {
// 				return errors.Wrapf(err, "At index: %d. Value: %v", i, elem.Interface())
// 			}
// 			for _, fk := range fks {
// 				var found bool
// 			innerSlice:
// 				for _, selected := range s.selectedFields {
// 					if fk == selected {
// 						found = true
// 						break innerSlice
// 					}
// 				}
// 				if !found {
// 					s.selectedFields = append(s.selectedFields, fk)
// 				}
// 			}
// 		}
// 	}
// 	return nil

// }

// modelValueByStrucfField gets the value by the provided StructField
func modelValueByStructField(
	model interface{},
	sField *models.StructField,
) (reflect.Value, error) {
	if model == nil {
		return reflect.ValueOf(model), errors.New("Provided empty value.")
	}

	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return v, errors.New("The value must be a single, non nil pointer value.")
	}
	v = v.Elem()

	return v.FieldByIndex(sField.ReflectField().Index), nil
}

func getURLVariables(req *http.Request, mStruct *models.ModelStruct, indexFirst, indexSecond int,
) (valueFirst, valueSecond string, err error) {

	path := req.URL.Path
	var invalidURL = func() error {
		return fmt.Errorf("Provided url is invalid for getting url variables: '%s' with indexes: '%d'/ '%d'", path, indexFirst, indexSecond)
	}
	pathSplitted := strings.Split(path, "/")
	if indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	var collectionIndex int = -1
	if models.StructCollectionUrlIndex(mStruct) != -1 {
		collectionIndex = models.StructCollectionUrlIndex(mStruct)
	} else {
		for i, splitted := range pathSplitted {
			if splitted == mStruct.Collection() {
				collectionIndex = i
				break
			}
		}
		if collectionIndex == -1 {
			err = fmt.Errorf("The url for given request does not contain collection name: %s", mStruct.Collection())
			return
		}
	}

	if collectionIndex+indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	valueFirst = pathSplitted[collectionIndex+indexFirst]

	if indexSecond > 0 {
		if collectionIndex+indexSecond > len(pathSplitted)-1 {
			err = invalidURL()
			return
		}
		valueSecond = pathSplitted[collectionIndex+indexSecond]
	}
	return
}

func getID(req *http.Request, mStruct *models.ModelStruct) (id string, err error) {
	id, _, err = getURLVariables(req, mStruct, 1, -1)
	return
}

func getIDAndRelationship(req *http.Request, mStruct *models.ModelStruct,
) (id, relationship string, err error) {
	return getURLVariables(req, mStruct, 1, 3)

}

func getIDAndRelated(req *http.Request, mStruct *models.ModelStruct,
) (id, related string, err error) {
	return getURLVariables(req, mStruct, 1, 2)
}

/**
Language
*/

func (s *Scope) getLangtagIndex() (index []int, err error) {
	if models.StructLanguage(s.mStruct) == nil {
		err = fmt.Errorf("Model: '%v' does not support i18n langtags.", s.mStruct.Type())
		return
	}
	index = models.StructLanguage(s.mStruct).FieldIndex()
	return
}

/**

Preset

*/

func (s *Scope) copyPresetParameters() {
	for _, includedField := range s.includedFields {
		includedField.copyPresetFullParameters()
	}
}

func (s *Scope) copy(isRoot bool, root *Scope) *Scope {
	scope := *s

	if isRoot {
		scope.rootScope = nil
		scope.collectionScope = &scope
		root = &scope
	} else {
		if s.rootScope == nil {
			scope.rootScope = nil
		}
		scope.collectionScope = root.getOrCreateModelsRootScope(s.mStruct)
	}

	if s.fieldset != nil {
		scope.fieldset = make(map[string]*models.StructField)
		for k, v := range s.fieldset {
			scope.fieldset[k] = v
		}
	}

	if s.primaryFilters != nil {
		scope.primaryFilters = make([]*filters.FilterField, len(s.primaryFilters))
		for i, v := range s.primaryFilters {

			scope.primaryFilters[i] = filters.CopyFilter(v)
		}
	}

	if s.attributeFilters != nil {
		scope.attributeFilters = make([]*filters.FilterField, len(s.attributeFilters))
		for i, v := range s.attributeFilters {
			scope.attributeFilters[i] = filters.CopyFilter(v)
		}
	}

	if s.relationshipFilters != nil {
		for i, v := range s.relationshipFilters {
			scope.relationshipFilters[i] = filters.CopyFilter(v)
		}
	}

	if s.sortFields != nil {
		scope.sortFields = make([]*sorts.SortField, len(s.sortFields))
		for i, v := range s.sortFields {
			scope.sortFields[i] = sorts.Copy(v)
		}

	}

	if s.includedScopes != nil {
		scope.includedScopes = make(map[*models.ModelStruct]*Scope)
		for k, v := range s.includedScopes {
			scope.includedScopes[k] = v.copy(false, root)
		}
	}

	if s.includedFields != nil {
		scope.includedFields = make([]*IncludeField, len(s.includedFields))
		for i, v := range s.includedFields {
			scope.includedFields[i] = v.copy(&scope, root)
		}
	}

	if s.includedValues != nil {
		scope.includedValues = safemap.New()
		for k, v := range s.includedValues.Values() {
			scope.includedValues.UnsafeAdd(k, v)
		}
	}

	return &scope
}

// func (s *Scope) buildPresetFields(model *models.ModelStruct, presetFields ...string) error {
// 	l := len(presetFields)
// 	switch {
// 	case l < 1:
// 		return fmt.Errorf("No preset field provided. For model: %s.", model.modelType.Name(), presetFields)
// 	case l == 1:
// 		if presetCollection := presetFields[0]; presetCollection != model.collectionType {
// 			return fmt.Errorf("Invalid preset collection: '%s'. The collection does not match with provided model: '%s'.", presetCollection, model.modelType.Name())
// 		}
// 		return nil

// 	case l > 2:
// 		collection := presetFields[0]

// 	}
// 	return nil
// }

/**
Errors
*/

func (s *Scope) errNilValueProvided() error {
	return fmt.Errorf("Provided nil value for the scope of type: '%v'.", s.mStruct.Type())
}

func (s *Scope) errValueTypeDoesNotMatch(t reflect.Type) error {
	return fmt.Errorf("The scope's Value type: '%v' does not match it's model structure: '%v'", t, s.mStruct.Type())
}

func (s *Scope) errInvalidValue() error {
	return fmt.Errorf("Invalid value provided for the scope: '%v'.", s.mStruct.Type())
}
