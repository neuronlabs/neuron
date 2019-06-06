package scope

import (
	"fmt"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"github.com/neuronlabs/neuron/internal/safemap"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

/**

SCOPE INCLUDED FIELDS

*/

// BuildIncludeList provide fast checks for the includedList
// if given include passes use buildInclude method on it.
func (s *Scope) BuildIncludeList(includedList ...string,
) (errs []*aerrors.ApiError) {
	var errorObjects []*aerrors.ApiError
	var errObj *aerrors.ApiError

	// check if the number of included fields is possible
	if len(includedList) > s.mStruct.MaxIncludedCount() {
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
	var manyIncludes = len(includedList) > 1

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

// GetOrCreateIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field
func (s *Scope) GetOrCreateIncludeField(field *models.StructField,
) (includeField *IncludeField) {
	return s.getOrCreateIncludeField(field)
}

// IncludedFields returns included fields slice
func (s *Scope) IncludedFields() []*IncludeField {
	return s.includedFields
}

// IncludedFieldsChan generates an included field channel
func (s *Scope) IncludedFieldsChan() <-chan *IncludeField {
	fields := make(chan *IncludeField)

	go func() {
		defer close(fields)
		for _, includedField := range s.includedFields {
			fields <- includedField
		}
	}()
	return fields
}

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

// IncludedValues returns included scope values
func (s *Scope) IncludedValues() *safemap.SafeHashMap {
	return s.includedValues
}

// InitializeIncluded initializes the included scopes
func (s *Scope) InitializeIncluded(maxNestedLevel int) {
	s.includedScopes = make(map[*models.ModelStruct]*Scope)
	s.maxNestedLevel = maxNestedLevel
}

// CopyIncludedBoundaries copies all included data from scope's included fields
// Into it's included scopes.
func (s *Scope) CopyIncludedBoundaries() {
	s.copyIncludedBoundaries()
}

// CurrentIncludedField gets current included field, based on the index
func (s *Scope) CurrentIncludedField() (*IncludeField, error) {
	if s.currentIncludedFieldIndex == -1 || s.currentIncludedFieldIndex > len(s.includedFields)-1 {
		return nil, errors.New("Getting non-existing included field.")
	}

	return s.includedFields[s.currentIncludedFieldIndex], nil
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

// ResetIncludedField resets the current included field pointer
func (s *Scope) ResetIncludedField() {
	s.currentIncludedFieldIndex = -1
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

// copies the and fieldset for given include and it's nested fields.
func (s *Scope) copyIncludedBoundaries() {
	for _, includedField := range s.includedFields {
		includedField.copyScopeBoundaries()
	}
}

// getTotalIncludeFieldCount gets the count for all included Fields. May be used
// as a wait group counter.
func (s *Scope) getTotalIncludeFieldCount() int {
	return s.totalIncludeCount
}

/**

INCLUDED FIELD DEFINITION

*/

// IncludeField is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*models.StructField

	// Scope is the scope that contains the values and filters for given
	// include field
	Scope *Scope

	// RelatedScope is a pointer to the scope where the IncludedField is stored.
	RelatedScope *Scope

	NotInFieldset bool
}

func (i *IncludeField) copy(relatedScope *Scope, root *Scope) *IncludeField {
	included := &IncludeField{StructField: i.StructField, NotInFieldset: i.NotInFieldset}
	included.Scope = i.Scope.copy(false, root)
	included.RelatedScope = relatedScope
	return included
}

// GetMissingPrimaries gets the id values from the RelatedScope, checks which id values were
// already stored within the colleciton root scope and return new ones.
func (i *IncludeField) GetMissingPrimaries() ([]interface{}, error) {
	return i.getMissingPrimaries()
}

func newIncludeField(field *models.StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field

	// Set NewScope for given field

	includeField.Scope = scope.createModelsScope(field.Relationship().Struct())

	// Set the root collection scope for given scope
	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(field.Relationship().Struct())
	if _, ok := includeField.Scope.collectionScope.fieldset[includeField.ApiName()]; !ok {
		includeField.NotInFieldset = true
		scope.hasFieldNotInFieldset = true
	}

	// Set relatedScope for given incldudedField
	includeField.RelatedScope = scope

	includeField.Scope.rootScope.totalIncludeCount++

	return includeField
}

func (i *IncludeField) getMissingPrimaries() ([]interface{}, error) {
	// uniqueMissing makes it possible to get unique ids that are not already used
	uniqueMissing := map[interface{}]struct{}{}

	// Lock the SafeHashMap for given collection
	i.Scope.collectionScope.includedValues.Lock()
	defer i.Scope.collectionScope.includedValues.Unlock()

	// Get the value from the RelatedScope
	v := reflect.ValueOf(i.RelatedScope.Value)

	// RelatedScope Value must be a pointer type
	if v.Kind() != reflect.Ptr {
		return nil, internal.ErrUnexpectedType
	}

	// Check if is nil
	if !v.IsNil() {
		v = v.Elem()
		switch v.Kind() {
		case reflect.Slice:
			log.Debugf("Getting from slice")
			for j := 0; j < v.Len(); j++ {
				elem := v.Index(j)
				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}
				log.Debugf("i'th: %d element", j)
				if err := i.getMissingFromSingle(elem, uniqueMissing); err != nil {
					return nil, err
				}

			}
		case reflect.Struct:
			log.Debugf("Getting from single")
			if err := i.getMissingFromSingle(v, uniqueMissing); err != nil {
				return nil, err
			}

		default:
			err := internal.ErrUnexpectedType
			log.Errorf("Unexpected Included Scope Value type: %s", v.Type())
			return nil, err
		}
	}

	// Copy the notUsed map into array
	missingIDs := make([]interface{}, len(uniqueMissing))

	j := 0
	for uniqueID := range uniqueMissing {
		missingIDs[j] = uniqueID
		j++
	}

	return missingIDs, nil
}

func (i *IncludeField) getMissingFromSingle(
	value reflect.Value,
	uniqueMissing map[interface{}]struct{},
) error {

	var (
		fieldValue = value.FieldByIndex(i.FieldIndex())

		// get related model's primary index
		primIndex = models.FieldsRelatedModelStruct(i.StructField).PrimaryField().FieldIndex()

		// setCollectionValues sets the relationship field primary index into the uniqueMissing map
		setCollectionValues = func(model reflect.Value) {
			primValue := model.FieldByIndex(primIndex)

			if primValue.IsValid() {
				primary := primValue.Interface()

				if _, ok := i.Scope.collectionScope.includedValues.Values()[primary]; !ok {
					// add to collection IDs
					i.Scope.collectionScope.includedValues.UnsafeSet(primary, nil)
					if _, ok = uniqueMissing[primary]; !ok {
						uniqueMissing[primary] = struct{}{}
					} else {
						log.Debugf("Primary: '%v' already exists", primary)
					}
				}
			} else {
				log.Debugf("Primary value is not valid")
			}
		}
	)

	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil
		}
		fieldValue = fieldValue.Elem()
	}

	// Get the type of the value
	switch fieldValue.Kind() {
	case reflect.Slice:
		log.Debugf("Field is Slice")
		for j := 0; j < fieldValue.Len(); j++ {
			// set primary field within scope for given model struct

			// elem is the model at j'th index in the slice
			elem := fieldValue.Index(j)

			// it may be a pointer to struct
			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}
			setCollectionValues(elem)
		}
	case reflect.Struct:
		log.Debugf("Field is Struct")
		setCollectionValues(fieldValue)
	default:
		log.Debugf("Unexpect type: '%s' in the relationship field's value.", fieldValue.Type())
		err := internal.ErrUnexpectedType
		return err
	}

	return nil
}

func (i *IncludeField) setRelationshipValue(relatedValue reflect.Value) {
	var includedScopeValue reflect.Value
	fieldValue := relatedValue.FieldByIndex(i.FieldIndex())

	switch i.FieldType().Kind() {
	case reflect.Slice:
		if fieldValue.Len() == 0 {
			i.Scope.Value = reflect.New(i.FieldType()).Elem().Interface()
			return
		}
		includedScopeValue = reflect.New(i.FieldType()).Elem()
		includedScopeValue.Set(fieldValue)
	case reflect.Ptr:
		if fieldValue.IsNil() {
			i.Scope.Value = nil
			return
		}
		includedScopeValue = reflect.New(i.FieldType().Elem())
		includedScopeValue.Elem().Set(fieldValue.Elem())
	}

	i.Scope.Value = includedScopeValue.Interface()
	return
}

func (i *IncludeField) copyScopeBoundaries() {
	// copy primaries
	i.Scope.primaryFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.primaryFilters))
	copy(i.Scope.primaryFilters, i.Scope.collectionScope.primaryFilters)

	// copy attribute filters
	i.Scope.attributeFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.attributeFilters))
	copy(i.Scope.attributeFilters, i.Scope.collectionScope.attributeFilters)

	// copy filterKeyFilters
	i.Scope.keyFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.keyFilters))
	copy(i.Scope.keyFilters, i.Scope.collectionScope.keyFilters)

	// relationships
	i.Scope.relationshipFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.relationshipFilters))
	copy(i.Scope.relationshipFilters, i.Scope.collectionScope.relationshipFilters)

	//copy foreignKeyFilters
	i.Scope.foreignFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.foreignFilters))
	copy(i.Scope.foreignFilters, i.Scope.collectionScope.foreignFilters)

	// copy language filters

	if i.Scope.collectionScope.languageFilters != nil {
		i.Scope.languageFilters = i.Scope.collectionScope.languageFilters
	}

	// fieldset is taken by reference - copied if there is nested
	i.Scope.fieldset = i.Scope.collectionScope.fieldset

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.UseLinks); ok {
		i.Scope.fContainer.Set(flags.UseLinks, f)
	}

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.ReturnPatchContent); ok {
		i.Scope.fContainer.Set(flags.ReturnPatchContent, f)
	}

	i.Scope.store[internal.ControllerCtxKey] = i.Scope.collectionScope.store[internal.ControllerCtxKey]

	for _, nested := range i.Scope.includedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.fieldset) == len(i.Scope.collectionScope.fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.fieldset = make(map[string]*models.StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.fieldset {
					i.Scope.fieldset[key] = field
				}
			}

			//add nested
			i.Scope.fieldset[nested.ApiName()] = nested.StructField
		}

		nested.copyScopeBoundaries()
	}

}

func (i *IncludeField) copyPresetFullParameters() {
	// copy primaries
	i.Scope.primaryFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.primaryFilters))
	copy(i.Scope.primaryFilters, i.Scope.collectionScope.primaryFilters)

	// copy attribute filters
	i.Scope.attributeFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.attributeFilters))
	copy(i.Scope.attributeFilters, i.Scope.collectionScope.attributeFilters)

	// copy filterKeyFilters
	i.Scope.keyFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.keyFilters))
	copy(i.Scope.keyFilters, i.Scope.collectionScope.keyFilters)

	// relationships
	i.Scope.relationshipFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.relationshipFilters))
	copy(i.Scope.relationshipFilters, i.Scope.collectionScope.relationshipFilters)

	//copy foreignKeyFilters
	i.Scope.foreignFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.foreignFilters))
	copy(i.Scope.foreignFilters, i.Scope.collectionScope.foreignFilters)

	// copy language filters
	if i.Scope.collectionScope.languageFilters != nil {
		i.Scope.languageFilters = i.Scope.collectionScope.languageFilters
	}

	// fieldset is taken by reference - copied if there is nested
	// i.Scope.fieldset = i.Scope.collectionScope.fieldset

	i.Scope.sortFields = make([]*sorts.SortField, len(i.Scope.collectionScope.sortFields))
	copy(i.Scope.sortFields, i.Scope.collectionScope.sortFields)

	i.Scope.pagination = i.Scope.collectionScope.pagination
	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.UseLinks); ok {
		i.Scope.fContainer.Set(flags.UseLinks, f)
	}

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.ReturnPatchContent); ok {
		i.Scope.fContainer.Set(flags.ReturnPatchContent, f)
	}

	i.Scope.store[internal.ControllerCtxKey] = i.Scope.collectionScope.store[internal.ControllerCtxKey]

	for _, nested := range i.Scope.includedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.fieldset) == len(i.Scope.collectionScope.fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.fieldset = make(map[string]*models.StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.fieldset {
					i.Scope.fieldset[key] = field
				}
			}

			//add nested
			i.Scope.fieldset[nested.ApiName()] = nested.StructField
		}

		nested.copyPresetFullParameters()
	}
}
