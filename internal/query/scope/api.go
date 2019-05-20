package scope

import (
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/namer/dialect"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/paginations"
	"github.com/neuronlabs/neuron/internal/query/sorts"
)

// AddFilterField adds the filter field for given scope
func AddFilterField(s *Scope, filter *filters.FilterField) error {
	return s.addFilterField(filter)
}

// AppendSortFields appends the sortfield to the given scope
func (s *Scope) AppendSortFields(fromStart bool, sortFields ...*sorts.SortField) {
	if fromStart {
		s.sortFields = append(sortFields, s.sortFields...)
	} else {
		s.sortFields = append(s.sortFields, sortFields...)
	}
}

// CopyScope copies provided scope and sets its root
func CopyScope(s *Scope, root *Scope, isRoot bool) *Scope {
	return s.copy(isRoot, root)
}

// Fieldset returns given scope fieldset
func Fieldset(s *Scope) map[string]*models.StructField {
	return s.fieldset
}

// FiltersPrimary returns scope's primary filters
func FiltersPrimary(s *Scope) []*filters.FilterField {
	return s.primaryFilters
}

// FiltersAttributes returns scope's attribute filters
func FiltersAttributes(s *Scope) []*filters.FilterField {
	return s.attributeFilters
}

// FiltersRelationFields returns scope's relationship filters
func FiltersRelationFields(s *Scope) []*filters.FilterField {
	return s.relationshipFilters
}

// FiltersForeigns returns all foreign key filter
func FiltersForeigns(s *Scope) []*filters.FilterField {
	return s.foreignFilters
}

// FiltersKeys return all FilterKey filters
func FiltersKeys(s *Scope) []*filters.FilterField {
	return s.keyFilters
}

// GetCollectionScope gets the collection root scope for given scope.
// Used for included Field scopes for getting their model root scope, that contains all
func GetCollectionScope(s *Scope) *Scope {
	return s.getCollectionScope()
}

// GetFieldValue gets the scope's field value
func GetFieldValue(s *Scope, sField *models.StructField) (interface{}, error) {
	return s.getFieldValuePublic(sField)
}

// GetLangtagValue returns the value of the langtag for given scope
// returns error if:
//		- the scope's model does not support i18n
//		- provided nil Value for the scope
//		- the scope's Value is of invalid type
func GetLangtagValue(s *Scope) (string, error) {
	return s.getLangtagValue()
}

// GetPrimaryFieldValues - gets the primary field values from the scope.
// Returns the values within the []interface{} form
//			returns	- ErrNoValue if no value provided.
//					- ErrInvalidType if the scope's value is of invalid type
// 					- *reflect.ValueError if internal occurs.
func GetPrimaryFieldValues(s *Scope) ([]interface{}, error) {
	return s.getPrimaryFieldValues()
}

// GetRelatedScope gets the related scope with preset filter values.
// The filter values are being taken form the root 's' Scope relationship id's.
// Returns error if the scope was not build by controller BuildRelatedScope.
func GetRelatedScope(s *Scope) (*Scope, error) {
	return s.getRelatedScope()
}

// GetTotalIncludeFieldCount gets the count for all included Fields. May be used
// as a wait group counter.
func GetTotalIncludeFieldCount(s *Scope) int {
	return s.getTotalIncludeFieldCount()
}

// GetValueAddress gets the address of the value for given scope
// in order to set it use the SetValueFromAddressable
// func GetValueAddress(s *Scope) interface{} {
// 	return s.getValueAddress()
// }

// IsRoot checks if given scope is a root scope of the query
func IsRoot(s *Scope) bool {
	return s.isRoot()
}

// New creates new scope for provided model
func New(model *models.ModelStruct) *Scope {
	scope := newScope(model)

	return scope
}

// NewRootScope creates new root scope for provided model
func NewRootScope(modelStruct *models.ModelStruct) *Scope {
	scope := newScope(modelStruct)
	scope.collectionScope = scope

	return scope
}

// SelectedFieldValues gets the scopes field values with provided dialectNamer
func SelectedFieldValues(s *Scope, dialectNamer dialect.FieldNamer) (map[string]interface{}, error) {
	return s.selectedFieldValues(dialectNamer)
}

// SetAllFields sets the fieldset to all possible fields
func SetAllFields(s *Scope) {
	s.setAllFields()
}

// SetFields the fieldset for given scope
func SetFields(s *Scope, fields ...interface{}) error {
	return s.SetFields(fields...)
}

// SetLangTagValue sets the langtag to the scope's value.
// returns an error
//		- if the Value is of invalid type or if the
//		- if the model does not support i18n
//		- if the scope's Value is nil pointer
func SetLangTagValue(s *Scope, langtag string) error {
	return s.setLangtagValue(langtag)
}

// SetPagination sets the Pagination for the query scope
func SetPagination(s *Scope, p *paginations.Pagination) error {
	if err := paginations.CheckPagination(p); err != nil {
		return err
	}
	s.pagination = p
	return nil
}
