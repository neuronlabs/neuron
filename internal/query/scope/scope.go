package scope

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/query/filters"
	"github.com/neuronlabs/neuron-core/internal/query/paginations"
	"github.com/neuronlabs/neuron-core/internal/query/sorts"
	"github.com/neuronlabs/neuron-core/internal/safemap"
)

// MaxPermissibleDuplicates is the maximum permissible dupliactes value used for errors
// TODO: get the value from config
var MaxPermissibleDuplicates = 3

// New creates new scope for provided model.
func New(model *models.ModelStruct) *Scope {
	scope := newScope(model)
	return scope
}

// NewRootScope creates new root scope for provided model.
func NewRootScope(modelStruct *models.ModelStruct) *Scope {
	scope := newScope(modelStruct)
	scope.collectionScope = scope
	return scope
}

// Scope contains information about given query for specific collection.
// If the query defines the different collection (included) than the main scope, then
// every detail about querying (fieldset, filters, sorts) are within new scopes
// kept in the Subscopes.
type Scope struct {
	id uuid.UUID

	// Value is the values or / value of the queried object / objects
	Value interface{}

	// Struct is a modelStruct this scope is based on
	mStruct *models.ModelStruct

	// selectedFields are the fields that were updated
	selectedFields []*models.StructField

	// store stores the scope's related key values
	store map[interface{}]interface{}

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

	// processor set for given query
	processor Processor

	isMany bool

	totalIncludeCount int
	kind              Kind

	// CollectionScope is a pointer to the scope containing the collection root
	collectionScope *Scope

	// rootScope is the root of all scopes where the query begins
	rootScope *Scope

	currentIncludedFieldIndex int

	// used within the root scope as a language tag for whole query.
	queryLanguage language.Tag

	hasFieldNotInFieldset bool

	// subscopesChain is the array of the scope's used for committing or rolling back the transaction.
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
		return nil, errors.New(class.InternalQueryIncluded, "provided invalid included fields for given scope.")
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
func (s *Scope) PreparePaginatedValue(key, value string, index paginations.Parameter) *errors.Error {
	val, err := strconv.Atoi(value)
	if err != nil {
		err := errors.New(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetailf("Provided query parameter: %v, contains invalid value: %v. Positive integer value is required.", key, value)
		return err
	}

	if s.pagination == nil {
		// if paginatio nis not already created initialize it
		s.pagination = &paginations.Pagination{}
	}

	switch index {
	case 0:
		if s.pagination.Type() == paginations.TpPage {
			return errors.New(class.QueryPaginationType, "multiple pagination types").SetDetail("Multiple pagination types in the query")
		}
		s.pagination.SetValue(val, index)
		s.pagination.SetType(paginations.TpOffset)
	case 1:
		if s.pagination.Type() == paginations.TpPage {
			return errors.New(class.QueryPaginationType, "multiple pagination types").SetDetail("Multiple pagination types in the query")
		}
		s.pagination.SetValue(val, index)
		s.pagination.SetType(paginations.TpOffset)
	case 2:
		if s.pagination.Type() == paginations.TpOffset && !s.pagination.IsZero() {
			return errors.New(class.QueryPaginationType, "multiple pagination types").SetDetail("Multiple pagination types in the query")
		}
		s.pagination.SetValue(val, index)
		s.pagination.SetType(paginations.TpPage)
	case 3:
		if s.pagination.Type() == paginations.TpOffset && !s.pagination.IsZero() {
			return errors.New(class.QueryPaginationType, "multiple pagination types").SetDetail("Multiple pagination types in the query")
		}
		s.pagination.SetValue(val, index)
		s.pagination.SetType(paginations.TpPage)
	}
	return nil
}

// Processor returns the scope's processor
func (s *Scope) Processor() Processor {
	return s.processor
}

// QueryLanguage gets the QueryLanguage tag
func (s *Scope) QueryLanguage() language.Tag {
	return s.queryLanguage
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

	primIndex := s.mStruct.PrimaryField().FieldIndex()

	setValueToCollection := func(value reflect.Value) {
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

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return errors.New(class.QueryValueType, "invalid scope value type")
	}
	if v.IsNil() {
		return nil
	}
	tempV := v.Elem()

	switch tempV.Kind() {
	case reflect.Slice:
		for i := 0; i < tempV.Len(); i++ {
			elem := tempV.Index(i)
			if elem.Type().Kind() != reflect.Ptr {
				return errors.New(class.QueryValueType, "invalid scope value type")
			}
			if !elem.IsNil() {
				setValueToCollection(elem)
			}
		}
	case reflect.Struct:
		log.Debugf("Struct setValueToCollection")
		setValueToCollection(v)
	default:
		err := errors.New(class.QueryValueType, "invalid scope value type")
		return err
	}

	return nil
}

// SetPaginationNoCheck sets the pagination without check
func (s *Scope) SetPaginationNoCheck(p *paginations.Pagination) {
	s.pagination = p
}

// SetPagination sets the pagination with checking if the pagination is without errors.
func (s *Scope) SetPagination(p *paginations.Pagination) error {
	if err := p.Check(); err != nil {
		return err
	}
	s.pagination = p
	return nil
}

// SetProcessor sets the processor for given scope
func (s *Scope) SetProcessor(p Processor) {
	s.processor = p
}

// SetQueryLanguage sets the query language tag
func (s *Scope) SetQueryLanguage(tag language.Tag) {
	s.queryLanguage = tag
}

// Struct returns scope's model struct
func (s *Scope) Struct() *models.ModelStruct {
	return s.mStruct
}

// StoreSet sets the store value with the key and it's value
func (s *Scope) StoreSet(key, value interface{}) {
	s.store[key] = value
}

// StoreGet gets the store value and checks if exists such key in the store
func (s *Scope) StoreGet(key interface{}) (value interface{}, ok bool) {
	value, ok = s.store[key]
	return
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
func (s *Scope) GetCollectionScope() *Scope {
	return s.collectionScope
}

// GetFieldValue

// isRoot checks if given scope is a root scope of the query
func (s *Scope) isRoot() bool {
	return s.kind == RootKind
}

// setLangtagValue sets the langtag to the scope's value.
// returns an error
//		- if the Value is of invalid type or if the
//		- if the model does not support i18n
//		- if the scope's Value is nil pointer
func (s *Scope) setLangtagValue(langtag string) error {

	langField := s.Struct().LanguageField()
	if langField == nil {
		return errors.New(class.QueryFilterLanguage, "no language field found for the model")
	}

	v := reflect.ValueOf(s.Value)
	if v.IsNil() {
		return errors.New(class.QueryNoValue, "no scope's value provided")
	}

	if v.Kind() != reflect.Ptr {
		return errors.New(class.QueryValueType, "invalid query value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Struct:
		fieldValue := v.FieldByIndex(langField.ReflectField().Index)
		fieldValue.SetString(langtag)
		return nil
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.IsNil() {
				continue
			}
			elem.Elem().FieldByIndex(langField.ReflectField().Index).SetString(langtag)
		}
		return nil
	}

	return errors.New(class.QueryValueType, "invalid query value")
}

// GetPrimaryFieldValues - gets the primary field values from the scope.
// Returns the values within the []interface{} form.
func (s *Scope) GetPrimaryFieldValues() ([]interface{}, error) {
	if s.Value == nil {
		return nil, errors.New(class.QueryNoValue, "no scope value provided")
	}

	primaryIndex := s.mStruct.PrimaryField().FieldIndex()
	values := []interface{}{}

	addPrimaryValue := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(primaryIndex)
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New(class.QueryValueType, "invalid query value type")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)
			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one fo the slice values is of invalid type")
			}

			if single.IsNil() {
				continue
			}

			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one fo the slice values is of invalid type")
			}
			addPrimaryValue(single)
		}
	case reflect.Struct:
		addPrimaryValue(v)
	default:
		return nil, errors.New(class.QueryValueType, "one query value is of invalid type")
	}

	return values, nil
}

// GetForeignKeyValues gets the values of the foreign key struct field
func (s *Scope) GetForeignKeyValues(foreign *models.StructField) ([]interface{}, error) {
	if s.mStruct != foreign.Struct() {
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.New(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	} else if foreign.FieldKind() != models.KindForeignKey {
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.FieldKind())
		return nil, errors.New(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	} else if s.Value == nil {
		return nil, errors.New(class.QueryNoValue, "provided nil scope value")
	}

	// initialize the array
	values := []interface{}{}

	// set the adding functino
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New(class.QueryNoValue, "provided no scope value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)

			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			if single.IsNil() {
				continue
			}
			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			addForeignKey(single)
		}
	case reflect.Struct:
		addForeignKey(v)
	default:
		return nil, errors.New(class.QueryValueType, "invalid query value type")
	}

	return values, nil
}

// GetUniqueForeignKeyValues gets the unique values of the foreign key struct field
func (s *Scope) GetUniqueForeignKeyValues(foreign *models.StructField) ([]interface{}, error) {
	if s.mStruct != foreign.Struct() {
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.New(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	} else if foreign.FieldKind() != models.KindForeignKey {
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.FieldKind())
		return nil, errors.New(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	} else if s.Value == nil {
		return nil, errors.New(class.QueryNoValue, "provided nil scope value")
	}

	// initialize the array
	foreigns := make(map[interface{}]struct{})
	values := []interface{}{}

	// set the adding functino
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		foreigns[primaryValue.Interface()] = struct{}{}
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New(class.QueryNoValue, "provided no scope value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)

			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			if single.IsNil() {
				continue
			}
			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.New(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			addForeignKey(single)
		}
	case reflect.Struct:
		addForeignKey(v)
	default:
		return nil, errors.New(class.QueryValueType, "invalid query value type")
	}

	for foreign := range foreigns {
		values = append(values, foreign)
	}

	return values, nil
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *models.ModelStruct) *Scope {
	scope := &Scope{
		// TODO: set the scope's id based on the domain
		id:                        uuid.New(),
		mStruct:                   modelStruct,
		fieldset:                  make(map[string]*models.StructField),
		currentIncludedFieldIndex: -1,
		store:                     map[interface{}]interface{}{},
	}

	// set all fields
	for _, field := range modelStruct.Fields() {
		scope.fieldset[field.NeuronName()] = field
	}

	if log.Level() <= log.LDEBUG2 {
		log.Debug2f("[SCOPE][%s] query new scope", scope.id.String())
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
	scope.store[internal.ControllerStoreKey] = s.store[internal.ControllerStoreKey]

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
		return errors.New(class.QueryNoValue, "provided query with no value")
	}
	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return errors.New(class.QueryValueUnaddressable, "provided unadressable query value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Struct:
		if t := v.Type(); t != s.mStruct.Type() {
			return errors.New(class.QueryValueType, "query value type doesn't match it's model struct")
		}
		includeField.setRelationshipValue(v)
	// case reflect.Slice:
	// TODO: set relationship value for slice value scope
	// for i := 0; i < v.Len(); i++ {
	// 	elem := v.Index(i)
	// 	if elem.Kind() != reflect.Ptr {
	// 		return errors.New(class.QueryValueUnaddressable, "one of the query values in slice is unadressable")
	// 	}

	// 	if elem.IsNil() {
	// 		continue
	// 	}

	// 	elem = elem.Elem()
	// 	include
	// }
	default:
		return errors.New(class.QueryValueType, "query value type is of invalid ")
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

func (s *Scope) getFieldValue(sField *models.StructField) (reflect.Value, error) {
	return modelValueByStructField(s.Value, sField)
}

func (s *Scope) checkField(field string) (*models.StructField, *errors.Error) {
	sField, err := s.mStruct.CheckField(field)
	if err != nil {
		return nil, err
	}
	return sField, nil
}

// modelValueByStrucfField gets the value by the provided StructField
func modelValueByStructField(model interface{}, sField *models.StructField) (reflect.Value, error) {
	if model == nil {
		return reflect.Value{}, errors.New(class.QueryNoValue, "empty value provided").SetOperation("modelValueByStructField")
	}

	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return v, errors.New(class.QueryValueUnaddressable, "non addressable value provided")
	}
	v = v.Elem()

	if v.Kind() == reflect.Slice {
		return reflect.Value{}, errors.New(class.QueryValueType, "invalid value type provided for the function").SetOperation("modelValueByStructField")
	}

	return v.FieldByIndex(sField.ReflectField().Index), nil
}
