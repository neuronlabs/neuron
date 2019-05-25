package scope

import (
	"fmt"
	"github.com/google/uuid"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/paginations"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"github.com/neuronlabs/neuron/internal/safemap"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"sync"

	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Errors used in the scope
var (
	ErrNoParamsInContext = errors.New("No parameters in the request Context.")
	ErrNoValue           = errors.New("No value provided within the scope.")
)

var (
	// MaxPermissibleDuplicates is the maximum permissible dupliactes value used for errors
	MaxPermissibleDuplicates = 3
)

// Scope contains information about given query for specific collection
// if the query defines the different collection than the main scope, then
// every detail about querying (fieldset, filters, sorts) are within new scopes
// kept in the Subscopes
type Scope struct {
	id uuid.UUID

	// Value is the values or / value of the queried object / objects
	Value interface{}

	// Store stores the scope's related key values
	Store map[interface{}]interface{}

	// Struct is a modelStruct this scope is based on
	mStruct *models.ModelStruct

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
	kind              Kind

	// CollectionScope is a pointer to the scope containing the collection root
	collectionScope *Scope

	// rootScope is the root of all scopes where the query begins
	rootScope *Scope

	currentIncludedFieldIndex int
	isRelationship            bool

	// used within the root scope as a language tag for whole query.
	queryLanguage language.Tag

	hasFieldNotInFieldset bool

	// subscopeChain is the array of the scope's used for commiting or rolling back the transaction
	subscopesChain []*Scope

	filterLock sync.Mutex
}

// AddChainSubscope adds the subscope to the 's' subscopes
func (s *Scope) AddChainSubscope(sub *Scope) {
	s.subscopesChain = append(s.subscopesChain, sub)
}

// Chain returns the scope's subscopes chain
func (s *Scope) Chain() []*Scope {
	return s.subscopesChain
}

// CreateModelsRootScope creates scope for given model (mStruct) and
// stores it within the rootScope.includedScopes.
// Used for collection unique root scopes
// (filters, fieldsets etc. for given collection scope)
func (s *Scope) CreateModelsRootScope(mStruct *models.ModelStruct) *Scope {
	return s.createModelsRootScope(mStruct)
}

// CurrentErrorCount returns current error count number
func (s *Scope) CurrentErrorCount() int {
	return s.currentErrorCount
}

// Flags gets the scope flags
func (s *Scope) Flags() *flags.Container {
	return s.getFlags()
}

// GetCollection Returns the collection name for given scope
func (s *Scope) GetCollection() string {
	return s.mStruct.Collection()
}

// GetFieldValue gets the value of the provided field
func (s *Scope) GetFieldValue(sField *models.StructField) (reflect.Value, error) {
	return s.getFieldValue(sField)
}

// GetModelsRootScope returns the scope for given model that is stored within
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

// GetPrimaryFieldValue gets the primary field reflect.Value
func (s *Scope) GetPrimaryFieldValue() (reflect.Value, error) {
	return s.getFieldValue(s.Struct().PrimaryField())
}

// GetScopeValueString gets the scope's value string
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

// ID gets the scope's ID
func (s *Scope) ID() uuid.UUID {
	return s.id
}

// IncreaseErrorCount adds another error count for the given scope
func (s *Scope) IncreaseErrorCount(count int) {
	s.currentErrorCount += count
}

// IsMany checks if the value is a slice
func (s *Scope) IsMany() bool {
	return s.isMany
}

// IsRoot checks if the scope is root kind
func (s *Scope) IsRoot() bool {
	return s.isRoot()
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

// Pagination returns scope's pagination
func (s *Scope) Pagination() *paginations.Pagination {
	return s.pagination
}

// PreparePaginatedValue prepares paginated value for given key, value and index
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

// QueryLanguage gets the QueryLanguage tag
func (s *Scope) QueryLanguage() language.Tag {
	return s.queryLanguage
}

// SetFlags sets the flags for the given scope
func (s *Scope) SetFlags(c *flags.Container) {
	s.fContainer = c
}

// SetFlagsFrom sets the flags from the provided provided flags array
func (s *Scope) SetFlagsFrom(flgs ...*flags.Container) {
	for _, f := range internal.ScopeCtxFlags {
		s.Flags().SetFirst(f, flgs...)
	}
}

// SetIsMany sets the isMany variable from the provided argument
func (s *Scope) SetIsMany(isMany bool) {
	s.isMany = isMany
}

// SetCollectionScope sets the collection scope for given scope
func (s *Scope) SetCollectionScope(cs *Scope) {
	s.collectionScope = cs
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
				s.collectionScope.includedValues.UnsafeSet(primary, value.Interface())
				return
			}

			if insider == nil {
				// in order to prevent the nil values set within given key
				s.collectionScope.includedValues.UnsafeSet(primary, value.Interface())
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
		return internal.ErrUnexpectedType
	}
	if !v.IsNil() {

		tempV := v.Elem()
		switch tempV.Kind() {
		case reflect.Slice:
			for i := 0; i < tempV.Len(); i++ {
				elem := tempV.Index(i)
				if elem.Type().Kind() != reflect.Ptr {
					return internal.ErrUnexpectedType
				}
				if !elem.IsNil() {
					setValueToCollection(elem)
				}
			}
		case reflect.Struct:
			log.Debugf("Struct setValueToCollection")
			setValueToCollection(v)

		default:
			err := internal.ErrUnexpectedType
			return err
		}
	} else {
		log.Debugf("Nil field's value.")
	}

	return nil
}

// SetPaginationNoCheck sets the pagination without check
func (s *Scope) SetPaginationNoCheck(p *paginations.Pagination) {
	s.pagination = p
}

// SetQueryLanguage sets the query language tag
func (s *Scope) SetQueryLanguage(tag language.Tag) {
	s.queryLanguage = tag
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

// SetValueFromAddressable - lack of generic makes it hard for preparing addressable value.
// While getting the addressable value with GetValueAddress, this function makes use of it
// by setting the Value from addressable.
// // Returns an error if the addressable is nil.
// func (s *Scope) SetValueFromAddressable() error {
// 	return s.setValueFromAddressable()
// }

/**

PRIVATE METHODS

*/

// GetCollectionScope gets the collection root scope for given scope.
// Used for included Field scopes for getting their model root scope, that contains all
func (s *Scope) getCollectionScope() *Scope {
	return s.collectionScope
}

func (s *Scope) getFlags() *flags.Container {
	return s.fContainer
}

// GetFieldValue
func (s *Scope) getFieldValuePublic(field *models.StructField) (value interface{}, err error) {
	if s.Value == nil {
		err = internal.ErrNilValue
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
			return internal.ErrUnexpectedType
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

// GetPrimaryFieldValues - gets the primary field values from the scope.
// Returns the values within the []interface{} form
//			returns	- ErrNoValue if no value provided.
//					- ErrInvalidType if the scope's value is of invalid type
// 					- *reflect.ValueError if internal occurs.
func (s *Scope) GetPrimaryFieldValues() (values []interface{}, err error) {
	if s.Value == nil {
		err = ErrNoValue
		return
	}

	primaryIndex := models.StructPrimary(s.mStruct).FieldIndex()

	addPrimaryValue := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(primaryIndex)
		values = append(values, primaryValue.Interface())
	}
	values = []interface{}{}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		err = internal.ErrInvalidType
		return
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)
			if single.Kind() != reflect.Ptr {
				err = internal.ErrInvalidType
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return
			}

			if single.IsNil() {
				continue
			}

			single = single.Elem()
			if single.Kind() != reflect.Struct {
				err = internal.ErrInvalidType
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return
			}
			addPrimaryValue(single)
		}
	case reflect.Struct:
		addPrimaryValue(v)
	default:
		err = internal.ErrInvalidType
		return
	}
	return
}

// GetForeignKeyValues gets the values of the foreign key struct field
func (s *Scope) GetForeignKeyValues(foreign *models.StructField) (values []interface{}, err error) {
	if s.mStruct != foreign.Struct() {
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.New("foreign key mismatched ModelStruct")
	} else if foreign.FieldKind() != models.KindForeignKey {
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.FieldKind())
		return nil, errors.New("foreign key is not a valid ForeignKey")
	} else if s.Value == nil {
		return nil, internal.ErrNilValue
	}

	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		values = append(values, primaryValue.Interface())
	}

	// initialize the array
	values = []interface{}{}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		err = internal.ErrInvalidType
		return
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)

			if single.Kind() != reflect.Ptr {
				err = internal.ErrInvalidType
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return
			}

			if single.IsNil() {
				continue
			}
			single = single.Elem()
			if single.Kind() != reflect.Struct {
				err = internal.ErrInvalidType
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return
			}

			addForeignKey(single)
		}
	case reflect.Struct:
		addForeignKey(v)
	default:
		err = internal.ErrInvalidType
		return
	}
	return
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *models.ModelStruct) *Scope {

	scope := &Scope{
		// TODO: set the scope's id based on the domain
		id:                        uuid.New(),
		mStruct:                   modelStruct,
		fieldset:                  make(map[string]*models.StructField),
		currentIncludedFieldIndex: -1,
		fContainer:                flags.New(),
		Store:                     map[interface{}]interface{}{},
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
	scope.Store[internal.ControllerCtxKey] = s.Store[internal.ControllerCtxKey]

	if s.rootScope == nil {
		scope.rootScope = s
	} else {
		scope.rootScope = s.rootScope
	}

	return scope
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
	var collectionIndex = -1
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
			scope.includedValues.UnsafeSet(k, v)
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
