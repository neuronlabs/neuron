package scope

import (
	"fmt"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/pkg/errors"
	"reflect"
)

/**

FILTERS

*/

// AddFilterField adds the filter to the given scope
func (s *Scope) AddFilterField(filter *filters.FilterField) error {
	return s.addFilterField(filter)
}

// AttributeFilters returns scopes attribute filters
func (s *Scope) AttributeFilters() []*filters.FilterField {
	return s.attributeFilters
}

// ClearRelationshipFilters clears the relationship filters for the scope
func (s *Scope) ClearRelationshipFilters() {
	s.relationshipFilters = []*filters.FilterField{}
}

// FilterKeyFilters return key filters for the scope
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	return s.keyFilters
}

// GetOrCreateAttributeFilter creates or gets existing attribute filter for given sField
func (s *Scope) GetOrCreateAttributeFilter(
	sField *models.StructField,
) (filter *filters.FilterField) {
	return s.getOrCreateAttributeFilter(sField)
}

//GetOrCreateFilterKeyFilter creates or get an existing filter field
func (s *Scope) GetOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateFilterKeyFilter(sField)
}

//GetOrCreateForeignKeyFilter creates or get an existing filter field
func (s *Scope) GetOrCreateForeignKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateForeignKeyFilter(sField)
}

// GetOrCreateIDFilter gets or creates new filterField
func (s *Scope) GetOrCreateIDFilter() *filters.FilterField {
	return s.getOrCreateIDFilter()
}

// GetOrCreateLanguageFilter used to get or if yet not found create the language filter field
func (s *Scope) GetOrCreateLanguageFilter() (filter *filters.FilterField) {
	return s.getOrCreateLangaugeFilter()
}

// GetOrCreateRelationshipFilter creates or gets existing fitler field for given struct field.
func (s *Scope) GetOrCreateRelationshipFilter(sField *models.StructField) (filter *filters.FilterField) {
	return s.getOrCreateRelationshipFilter(sField)
}

// LanguageFilter return language filters for given scope
func (s *Scope) LanguageFilter() *filters.FilterField {
	return s.languageFilters
}

// PrimaryFilters returns scopes primary filter values
func (s *Scope) PrimaryFilters() []*filters.FilterField {
	return s.primaryFilters
}

// RelationshipFilters returns scopes relationship filters
func (s *Scope) RelationshipFilters() []*filters.FilterField {
	return s.relationshipFilters
}

// RemoveRelationshipFilter at index
func (s *Scope) RemoveRelationshipFilter(at int) error {
	if at > len(s.relationshipFilters)-1 {
		return errors.New("Removing relationship filter out of possible range")
	}

	s.relationshipFilters = append(s.relationshipFilters[:at], s.relationshipFilters[at+1:]...)
	return nil
}

// SetRelationshipFilters sets the relationship filters
func (s *Scope) SetRelationshipFilters(fs []*filters.FilterField) {
	s.relationshipFilters = fs
}

// SetIDFilters sets the ID Filter for given values.
func (s *Scope) SetIDFilters(idValues ...interface{}) {
	s.setIDFilterValues(idValues...)
}

// SetPrimaryFilters sets the primary filter for given values.
func (s *Scope) SetPrimaryFilters(values ...interface{}) {
	s.setIDFilterValues(values...)
}

// SetLanguageFilter the LanguageFilter for given scope.
// If the scope's model does not support i18n it does not create language filter, and ends fast.
func (s *Scope) SetLanguageFilter(languages ...interface{}) {
	s.setLanguageFilterValues(languages...)
}

// SetBelongsToForeignKeyFields sets the fields of type foreign key to the belongs of relaitonships
func (s *Scope) SetBelongsToForeignKeyFields() error {
	if s.Value == nil {
		return internal.ErrNilValue
	}

	setField := func(v reflect.Value) ([]*models.StructField, error) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Type() != s.Struct().Type() {
			return nil, internal.ErrInvalidType
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

/**

PRIVATE METHODS

*/

func (s *Scope) addFilterField(filter *filters.FilterField) error {
	if models.FieldsStruct(filter.StructField()).ID() != s.mStruct.ID() {
		err := fmt.Errorf("Filter Struct does not match with the scope. Model: %v, filterField: %v", s.mStruct.Type().Name(), filter.StructField().Name())
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
		err := fmt.Errorf("Provided filter field of invalid kind. Model: %v. FilterField: %v", s.mStruct.Type().Name(), filter.StructField().Name())
		return err
	}
	return nil
}

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
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

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

func (s *Scope) getOrCreateIDFilter() (filter *filters.FilterField) {
	return s.getOrCreatePrimaryFilter(models.StructPrimary(s.mStruct))
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

	filter = filters.NewFilter(models.StructLanguage(s.mStruct))
	s.languageFilters = filter
	return

}

func (s *Scope) getOrCreateAttributeFilter(
	sField *models.StructField,
) (filter *filters.FilterField) {

	s.filterLock.Lock()
	defer s.filterLock.Unlock()

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

func (s *Scope) getOrCreateFilterKeyFilter(sField *models.StructField) (filter *filters.FilterField) {
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

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
	s.filterLock.Lock()
	defer s.filterLock.Unlock()

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

			return
		}
	}

	filter = filters.NewFilter(sField)
	s.relationshipFilters = append(s.relationshipFilters, filter)
	return filter
}
