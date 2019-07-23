package scope

import (
	"reflect"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/query/filters"
)

// AddFilterField adds the 'filter' to the given scope.
func (s *Scope) AddFilterField(filter *filters.FilterField) error {
	return s.addFilterField(filter)
}

// AttributeFilters returns scope's attribute filters.
func (s *Scope) AttributeFilters() []*filters.FilterField {
	return s.attributeFilters
}

// ClearAllFilters clears all filters within the scope.
func (s *Scope) ClearAllFilters() {
	s.clearAttributeFilters()
	s.clearForeignKeyFilters()
	s.clearFilterKeyFilters()
	s.clearLanguageFilters()
	s.clearPrimaryFilters()
	s.clearRelationshipFilters()
}

// ClearAttributeFilters clears all the attribute filters.
func (s *Scope) ClearAttributeFilters() {
	s.clearAttributeFilters()
}

// ClearForeignKeyFilters clears the foreign key filters.
func (s *Scope) ClearForeignKeyFilters() {
	s.clearForeignKeyFilters()
}

// ClearFilterKeyFilters clears the filter key filters.
func (s *Scope) ClearFilterKeyFilters() {
	s.clearFilterKeyFilters()
}

// ClearLanguageFilters clears the language filters.
func (s *Scope) ClearLanguageFilters() {
	s.clearLanguageFilters()
}

// ClearPrimaryFilters clears the primary field filters.
func (s *Scope) ClearPrimaryFilters() {
	s.clearPrimaryFilters()
}

// ClearRelationshipFilters clears the relationship filters.
func (s *Scope) ClearRelationshipFilters() {
	s.clearRelationshipFilters()
}

// SetFiltersTo set the filters to the scope with the same model struct.
func (s *Scope) SetFiltersTo(to *Scope) error {
	if s.mStruct != to.mStruct {
		log.Errorf("SetFiltersTo mismatch scope's structs. Is: '%s' should be: '%s'", to.mStruct.Collection(), s.mStruct.Collection())
		return errors.New(class.InternalQueryModelMismatch, "scope's model mismatch").SetOperation("SetFiltersTo")
	}

	to.primaryFilters = s.primaryFilters
	to.attributeFilters = s.attributeFilters
	to.languageFilters = s.languageFilters
	to.relationshipFilters = s.relationshipFilters
	to.foreignFilters = s.foreignFilters
	to.keyFilters = s.keyFilters

	return nil
}

// FilterKeyFilters returns all key filters.
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	return s.keyFilters
}

// ForeignKeyFilters returns all the foreign key filters.
func (s *Scope) ForeignKeyFilters() []*filters.FilterField {
	return s.foreignFilters
}

// GetOrCreateAttributeFilter creates or gets existing attribute filter for given 'sField' *model.StructField.
func (s *Scope) GetOrCreateAttributeFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateAttributeFilter(sField)
}

//GetOrCreateFilterKeyFilter creates or get an existing filter key filter for given 'sField' *models.StructField.
func (s *Scope) GetOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateFilterKeyFilter(sField)
}

//GetOrCreateForeignKeyFilter creates or get an existing foreign key filter for given 'sField' *models.StructField.
func (s *Scope) GetOrCreateForeignKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateForeignKeyFilter(sField)
}

// GetOrCreateIDFilter creates or gets an exististing primary field filter.
func (s *Scope) GetOrCreateIDFilter() *filters.FilterField {
	return s.getOrCreateIDFilter()
}

// GetOrCreateLanguageFilter creates or gets an existing language language field filter.
func (s *Scope) GetOrCreateLanguageFilter() (filter *filters.FilterField) {
	return s.getOrCreateLangaugeFilter()
}

// GetOrCreateRelationshipFilter creates or gets existing relationship field fitler for the 'sField' *models.StructField.
func (s *Scope) GetOrCreateRelationshipFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateRelationshipFilter(sField)
}

// LanguageFilter returns language field filter.
func (s *Scope) LanguageFilter() *filters.FilterField {
	return s.languageFilters
}

// PrimaryFilters returns scope's primary filters.
func (s *Scope) PrimaryFilters() []*filters.FilterField {
	return s.primaryFilters
}

// RelationshipFilters returns scope's relationship filters.
func (s *Scope) RelationshipFilters() []*filters.FilterField {
	return s.relationshipFilters
}

// RemoveRelationshipFilter removes the relationship filter 'at' index in relationshipFilters array.
func (s *Scope) RemoveRelationshipFilter(at int) error {
	if at > len(s.relationshipFilters)-1 {
		return errors.New(class.InternalQueryFilter, "removing relationship filter out of possible range").SetOperation("RemoveRelationshipFilter")
	}

	s.relationshipFilters = append(s.relationshipFilters[:at], s.relationshipFilters[at+1:]...)
	return nil
}

// SetRelationshipFilters sets the relationship filters for 'fs' filter fields.
func (s *Scope) SetRelationshipFilters(fs []*filters.FilterField) {
	s.relationshipFilters = fs
}

// SetIDFilters sets the primary field filter with the operator OpIn for given 'idValues'.
func (s *Scope) SetIDFilters(idValues ...interface{}) {
	s.setIDFilterValues(idValues...)
}

// SetPrimaryFilters sets the primary field filter with the operator OpIn for given 'idValues'.
func (s *Scope) SetPrimaryFilters(values ...interface{}) {
	s.setIDFilterValues(values...)
}

// SetLanguageFilter the language filter with the OpIn filter operator and 'languages' values.
// If the scope's model does not support i18n it does not create language filter, and ends fast.
func (s *Scope) SetLanguageFilter(languages ...interface{}) {
	s.setLanguageFilterValues(languages...)
}

// SetBelongsToForeignKeyFields sets the foreign key fields for the 'belongs to' relationships.
func (s *Scope) SetBelongsToForeignKeyFields() error {
	if s.Value == nil {
		return errors.New(class.QueryNoValue, "nil query scope value provided")
	}

	setField := func(v reflect.Value) ([]*models.StructField, error) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Type() != s.Struct().Type() {
			return nil, errors.New(class.QueryValueType, "model's struct mismatch")
		}

		var fks []*models.StructField
		for _, field := range s.selectedFields {
			relField, ok := s.mStruct.RelationshipField(field.NeuronName())
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
			if elem.IsNil() {
				continue
			}
			fks, err := setField(elem)
			if err != nil {
				return err
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

/**

PRIVATE METHODS

*/

func (s *Scope) addFilterField(filter *filters.FilterField) error {
	if filter.StructField().Struct() != s.mStruct {
		log.Debugf("Filter's ModelStruct does not match scope's model. Scope's Model: %v, filterField: %v, filterModel: %v", s.mStruct.Type().Name(), filter.StructField().Name(), filter.StructField().Struct().Type().Name())
		err := errors.New(class.QueryFitlerNonMatched, "provied filter field's model structure doesn't match scope's model")
		return err
	}
	switch filter.StructField().FieldKind() {
	case models.KindPrimary:
		s.primaryFilters = append(s.primaryFilters, filter)
	case models.KindAttribute:
		if filter.StructField().IsLanguage() {
			if s.languageFilters == nil {
				s.languageFilters = filter
			} else {
				s.languageFilters.AddValues(filter.Values()...)
			}
		} else {
			s.attributeFilters = append(s.attributeFilters, filter)
		}
	case models.KindForeignKey:
		s.foreignFilters = append(s.foreignFilters, filter)
	case models.KindRelationshipMultiple, models.KindRelationshipSingle:
		s.relationshipFilters = append(s.relationshipFilters, filter)
	case models.KindFilterKey:
		s.keyFilters = append(s.keyFilters, filter)
	default:
		err := errors.Newf(class.QueryFilterFieldKind, "unknown field kind: %v", filter.StructField().FieldKind())
		return err
	}
	return nil
}

func (s *Scope) clearAttributeFilters() {
	s.attributeFilters = []*filters.FilterField{}
}
func (s *Scope) clearForeignKeyFilters() {
	s.foreignFilters = []*filters.FilterField{}
}
func (s *Scope) clearFilterKeyFilters() {
	s.keyFilters = []*filters.FilterField{}
}
func (s *Scope) clearLanguageFilters() {
	s.languageFilters = nil
}
func (s *Scope) clearPrimaryFilters() {
	s.primaryFilters = []*filters.FilterField{}
}
func (s *Scope) clearRelationshipFilters() {
	s.relationshipFilters = []*filters.FilterField{}
}

func (s *Scope) setIDFilterValues(values ...interface{}) {
	s.setPrimaryFilterValues(s.mStruct.PrimaryField(), values...)
}

func (s *Scope) setLanguageFilterValues(values ...interface{}) {
	filter := s.getOrCreateLangaugeFilter()
	if filter == nil {
		return
	}

	fv := &filters.OpValuePair{}
	fv.SetOperator(filters.OpIn)

	fv.Values = append(fv.Values, values...)
	filter.AppendValues(fv)
}

func (s *Scope) setPrimaryFilterValues(primField *models.StructField, values ...interface{}) {
	filter := s.getOrCreatePrimaryFilter(primField)
	filter.AppendValues(filters.NewOpValuePair(filters.OpIn, values...))
}

func (s *Scope) getOrCreatePrimaryFilter(primField *models.StructField) (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	if s.primaryFilters == nil {
		s.primaryFilters = []*filters.FilterField{}
	}

	for _, pf := range s.primaryFilters {
		if pf.StructField() == primField {
			return pf
		}
	}

	// if not found within primary filters
	filter = filters.NewFilter(primField)
	s.primaryFilters = append(s.primaryFilters, filter)
	return filter
}

func (s *Scope) getOrCreateIDFilter() (filter *filters.FilterField) {
	return s.getOrCreatePrimaryFilter(s.mStruct.PrimaryField())
}

func (s *Scope) getOrCreateLangaugeFilter() (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	if !s.mStruct.UseI18n() {
		return nil
	}

	if s.languageFilters != nil {
		return s.languageFilters
	}

	filter = filters.NewFilter(s.mStruct.LanguageField())
	s.languageFilters = filter
	return filter

}

func (s *Scope) getOrCreateAttributeFilter(sField *models.StructField) (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	if s.attributeFilters == nil {
		s.attributeFilters = []*filters.FilterField{}
	}

	for _, attrFilter := range s.attributeFilters {
		if attrFilter.StructField() == sField {
			filter = attrFilter
			return filter
		}
	}
	filter = filters.NewFilter(sField)
	s.attributeFilters = append(s.attributeFilters, filter)
	return filter
}

func (s *Scope) getOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	if s.keyFilters == nil {
		s.keyFilters = []*filters.FilterField{}
	}

	for _, fkFilter := range s.keyFilters {
		if fkFilter.StructField() == sField {
			filter = fkFilter
			return filter
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
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	if s.foreignFilters == nil {
		s.foreignFilters = []*filters.FilterField{}
	}

	for _, fkFilter := range s.foreignFilters {
		if fkFilter.StructField() == sField {
			filter = fkFilter
			return filter
		}
	}
	filter = filters.NewFilter(sField)
	s.foreignFilters = append(s.foreignFilters, filter)
	return filter
}

func (s *Scope) getOrCreateRelationshipFilter(sField *models.StructField) (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

	// Create if empty
	if s.relationshipFilters == nil {
		s.relationshipFilters = []*filters.FilterField{}
	}

	// Check if no relationship filter already exists
	for _, relFilter := range s.relationshipFilters {
		if relFilter.StructField() == sField {
			filter = relFilter
			return filter
		}
	}

	filter = filters.NewFilter(sField)
	s.relationshipFilters = append(s.relationshipFilters, filter)
	return filter
}
