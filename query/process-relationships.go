package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

func convertRelationshipFiltersFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}
	if len(s.RelationFilters) == 0 {
		// end fast if there is no relationship filters
		return nil
	}

	bufSize := 10
	if len(s.RelationFilters) < bufSize {
		bufSize = len(s.RelationFilters)
	}

	var results = make(chan interface{}, bufSize)
	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, rel := range s.RelationFilters {
		if rel.StructField.Relationship().Struct().Config() == nil {
			continue
		}

		if modelRepo := rel.StructField.Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	for i, filter := range s.RelationFilters {
		// cast the filter to the internal structure definition

		// check the relationship kind
		switch filter.StructField.Relationship().Kind() {
		case mapping.RelBelongsTo:
			go convertBelongsToRelationshipFilterChan(ctx, s, i, filter, results)
		case mapping.RelHasMany:
			go convertHasManyRelationshipFilterChan(ctx, s, i, filter, results)
		case mapping.RelHasOne:
			go convertHasOneRelationshipFilterChan(ctx, s, i, filter, results)
		case mapping.RelMany2Many:
			go convertMany2ManyRelationshipFilterChan(ctx, s, i, filter, results)
		default:
			err := errors.NewDetf(class.InternalQueryInvalidField, "invalid field's relationship kind. Model: %s, Field: %s", s.Struct().Type().Name(), filter.StructField.Name())
			results <- err
		}
	}

	var (
		ctr  int
		done bool
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-results:
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(s.RelationFilters) {
				done = true
			}
		}
		// finish the for loop when done
		if done {
			break
		}
	}
	// if all relationship filters are done we can clear them
	s.RelationFilters = Filters{}
	return nil
}

func convertRelationshipFiltersSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	if len(s.RelationFilters) == 0 {
		return nil
	}

	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, rel := range s.RelationFilters {
		if rel.StructField.Relationship().Struct().Config() == nil {
			continue
		}

		if modelRepo := rel.StructField.Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	var err error
	for i, filter := range s.RelationFilters {
		// check the relationship kind
		switch filter.StructField.Relationship().Kind() {
		case mapping.RelBelongsTo:
			_, err = convertBelongsToRelationshipFilter(ctx, s, i, filter)
		case mapping.RelHasMany:
			_, err = convertHasManyRelationshipFilter(ctx, s, i, filter)
		case mapping.RelHasOne:
			_, err = convertHasOneRelationshipFilter(ctx, s, i, filter)
		case mapping.RelMany2Many:
			_, err = convertMany2ManyRelationshipFilter(ctx, s, i, filter)
		}
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	// if all relationship filters are done we can clear them
	s.RelationFilters = Filters{}
	return nil
}

func convertBelongsToRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *FilterField, results chan<- interface{}) {
	index, err := convertBelongsToRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

func convertBelongsToRelationshipFilter(ctx context.Context, s *Scope, index int, filter *FilterField) (int, error) {
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("[SCOPE][%s] convertBelongsToRelationshipFilter field: '%s'", s.ID(), filter.StructField.Name())
	}
	// for the belongs to relationships if the relationship filters only over the relations primary
	// key then we can change the filter from relationship into the foreign key for the root scope
	// otherwise we need to get the
	var onlyPrimes = true

	for _, nested := range filter.Nested {
		if nested.StructField.FieldKind() != mapping.KindPrimary {
			onlyPrimes = false
			break
		}
	}

	if onlyPrimes {
		// convert the filters into the foreign key filters of the root scope
		foreignKey := filter.StructField.Relationship().ForeignKey()
		filterField := s.getOrCreateForeignKeyFilter(foreignKey)

		for _, nested := range filter.Nested {
			filterField.Values = append(filterField.Values, nested.Values...)
		}
		// send the index of the filter to remove from the scope
		return index, nil
	}
	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField.Relationship().Struct()), true)

	for _, nested := range filter.Nested {
		if err := relScope.addFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)
			return 0, err
		}
	}

	// we only need primary keys
	if err := relScope.SetFieldset(relScope.Struct().Primary()); err != nil {
		return 0, err
	}

	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	primaries, err := relScope.getPrimaryFieldValues()
	if err != nil {
		return 0, err
	}

	if len(primaries) == 0 {
		log.Debugf("SCOPE[%s] getting BelongsTo relationship filters no results found.", s.ID().String())
		err := errors.NewDet(class.QueryValueNoResult, "no result found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField.NeuronName())
		return 0, err
	}

	// convert the filters into the foreign key filters of the root scope
	foreignKey := filter.StructField.Relationship().ForeignKey()
	filterField := s.getOrCreateForeignKeyFilter(foreignKey)
	filterField.Values = append(filterField.Values, &OperatorValues{Operator: OpIn, Values: primaries})
	// send the index of the filter to remove from the scope
	return index, nil
}

func convertHasManyRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *FilterField, results chan<- interface{}) {
	index, err := convertHasManyRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func convertHasManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *FilterField) (int, error) {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("[SCOPE][%s] convertHasManyRelationshipFilter field: '%s'", s.ID(), filter.StructField.Name())
	}
	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	onlyForeign := true
	foreignKey := filter.StructField.Relationship().ForeignKey()

	for _, nested := range filter.Nested {
		if nested.StructField != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := s.getOrCreatePrimaryFilter()
		for _, nested := range filter.Nested {
			primaryFilter.Values = append(primaryFilter.Values, nested.Values...)
		}
		return index, nil
	}

	relScope := NewModelC(s.Controller(), filter.StructField.Relationship().Struct(), true)

	for _, nested := range filter.Nested {
		if err := relScope.addFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)
			return 0, err
		}
	}

	// the only required field in fieldset is a foreign key
	if err := relScope.setFields(foreignKey); err != nil {
		log.Errorf("Setting foreign field: %v failed: '%v'", foreignKey, err)
		return 0, err
	}
	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	foreignValues, err := relScope.getUniqueForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		return 0, err
	}

	if len(foreignValues) == 0 {
		log.Debugf("No results found for the relationship filter for field: %s", filter.StructField.NeuronName())
		err := errors.NewDet(class.QueryValueNoResult, "no relationship filter values found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField.NeuronName())
		return 0, err
	}
	primaryFilter := s.getOrCreatePrimaryFilter()
	primaryFilter.Values = append(primaryFilter.Values, &OperatorValues{foreignValues, OpIn})
	return index, nil
}

func convertHasOneRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *FilterField, results chan<- interface{}) {
	index, err := convertHasOneRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func convertHasOneRelationshipFilter(ctx context.Context, s *Scope, index int, filter *FilterField) (int, error) {
	log.Debug3f("[SCOPE][%s] convertHasOneRelationshipFilter field: '%s'", s.ID(), filter.StructField.Name())
	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	onlyForeign := true
	foreignKey := filter.StructField.Relationship().ForeignKey()

	for _, nested := range filter.Nested {
		if nested.StructField != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := s.getOrCreatePrimaryFilter()

		for _, nested := range filter.Nested {
			primaryFilter.Values = append(primaryFilter.Values, nested.Values...)
		}

		return index, nil
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField.Relationship().Struct()), true)

	for _, nested := range filter.Nested {
		if err := relScope.addFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)
			return 0, err
		}
	}

	// the only required field in fieldset is a foreign key
	if err := relScope.setFields(foreignKey); err != nil {
		return 0, err
	}

	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	foreignValues, err := relScope.getForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		return 0, err
	}

	if len(foreignValues) == 0 {
		err := errors.NewDet(class.QueryValueNoResult, "no result found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField.NeuronName())
		return 0, err
	}

	if len(foreignValues) > 1 && !s.isMany {
		log.Warningf("Multiple foreign key values found for a single valued - has one scope.")
	}

	primaryFilter := s.getOrCreatePrimaryFilter()
	primaryFilter.Values = append(primaryFilter.Values, &OperatorValues{foreignValues, OpIn})

	return index, nil
}

func convertMany2ManyRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *FilterField, results chan<- interface{}) {
	index, err := convertMany2ManyRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

func convertMany2ManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *FilterField) (int, error) {
	log.Debug3f("[SCOPE][%s] convertMany2ManyRelationshipFilter field: '%s'", s.ID(), filter.StructField.Name())
	// if filter is only on the 'id' field then

	// if all the nested filters are the primary key of the related model
	// the only query should be used on the join model foreign key
	onlyPrimaries := true
	relatedPrimaryKey := filter.StructField.Relationship().Struct().Primary()
	joinModelForeignKey := filter.StructField.Relationship().ForeignKey()

	for _, nested := range filter.Nested {
		if nested.StructField != relatedPrimaryKey && nested.StructField != joinModelForeignKey {
			onlyPrimaries = false
			break
		}
	}

	// if only the primaries of the related model were used list only the values from the join table
	if onlyPrimaries {
		// convert the foreign into root scope primary key filter
		joinScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField.Relationship().JoinModel()), true)

		// create the joinScope foreign key filter with the values of the original filter
		fkFilter := joinScope.getOrCreateForeignKeyFilter(filter.StructField.Relationship().ManyToManyForeignKey())
		for _, nested := range filter.Nested {
			fkFilter.Values = append(fkFilter.Values, nested.Values...)
		}

		// the fieldset should contain only a backreference field
		if err := joinScope.SetFieldset(filter.StructField.Relationship().ForeignKey()); err != nil {
			return 0, err
		}

		// Do the ListProcess with the context
		if err := joinScope.ListContext(ctx); err != nil {
			return 0, err
		}

		// get the foreign key values for the given joinScope
		fkValues, err := joinScope.getForeignKeyValues(filter.StructField.Relationship().ForeignKey())
		if err != nil {
			return 0, err
		}

		// Add (or get) the ID filter for the primary key in the root scope
		rootIDFilter := s.getOrCreatePrimaryFilter()
		// add the backreference field values into primary key filter
		rootIDFilter.Values = append(rootIDFilter.Values, &OperatorValues{fkValues, OpIn})
		return index, nil
	}

	// create the scope for the related model
	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField.Relationship().Struct()), true)

	// add then nested fields from the filter field into the
	for _, nested := range filter.Nested {
		var nestedFilter *FilterField
		switch nested.StructField.FieldKind() {
		case mapping.KindPrimary:
			nestedFilter = relScope.getOrCreatePrimaryFilter()
		case mapping.KindAttribute:
			if nested.StructField.IsLanguage() {
				nestedFilter = relScope.getOrCreateLangaugeFilter()
			} else {
				nestedFilter = relScope.getOrCreateAttributeFilter(nested.StructField)
			}
		case mapping.KindForeignKey:
			nestedFilter = relScope.getOrCreateForeignKeyFilter(nested.StructField)
		case mapping.KindFilterKey:
			nestedFilter = relScope.getOrCreateFilterKeyFilter(nested.StructField)
		case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
			nestedFilter = relScope.getOrCreateRelationshipFilter(nested.StructField)
		default:
			log.Debugf("Nested Filter for field of unknown type: %v", nested.StructField.NeuronName())
			continue
		}

		nestedFilter.Values = append(nestedFilter.Values, nested.Values...)
	}

	// only the primary field should be in a fieldset
	relScope.Fieldset = map[string]*mapping.StructField{"id": relScope.Struct().Primary()}
	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	primaries, err := relScope.getPrimaryFieldValues()
	if err != nil {
		return 0, err
	}

	if len(primaries) == 0 {
		err = errors.NewDet(class.QueryValueNoResult, "related filter values doesn't exist")
		return 0, err
	}

	joinScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField.Relationship().JoinModel()), true)

	mtmFKFilter := joinScope.getOrCreateForeignKeyFilter(filter.StructField.Relationship().ManyToManyForeignKey())
	mtmFKFilter.Values = append(mtmFKFilter.Values, &OperatorValues{primaries, OpIn})

	fk := filter.StructField.Relationship().ForeignKey()
	if err := joinScope.setFields(fk); err != nil {
		log.Errorf("Setting Foreign key: '%s' field into joinScope failed: %v", fk, err)
		return 0, err
	}

	if err := joinScope.ListContext(ctx); err != nil {
		return 0, err
	}

	fkValues, err := joinScope.getUniqueForeignKeyValues(filter.StructField.Relationship().ForeignKey())
	if err != nil {
		return 0, err
	}

	rootIDFilter := s.getOrCreatePrimaryFilter()
	rootIDFilter.Values = append(rootIDFilter.Values, &OperatorValues{fkValues, OpIn})

	return index, nil
}

func getForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}
	relations := map[string]*mapping.StructField{}

	// get all relationship field from the fieldset
	for _, field := range s.Fieldset {
		if field.Relationship() != nil {
			if _, ok := relations[field.NeuronName()]; !ok {
				relations[field.NeuronName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range s.includedFields {
		if _, ok := relations[included.NeuronName()]; !ok {
			relations[included.NeuronName()] = included.StructField
		}
	}

	if len(relations) == 0 {
		// finish fast if there is no relations in fieldset
		return nil
	}

	v := reflect.ValueOf(s.Value)
	if v.IsNil() {
		log.Infof("[SCOPE][%s] Nil scope's value while taking foreign relationships", s.ID())
		return errors.NewDet(class.QueryNoValue, "scope has nil value")
	}
	v = v.Elem()

	results := make(chan interface{}, len(relations))

	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, rel := range relations {
		if rel.Relationship().Struct().Config() == nil {
			continue
		}

		if modelRepo := rel.Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	for _, rel := range relations {
		switch rel.Relationship().Kind() {
		case mapping.RelHasOne:
			go getForeignRelationshipHasOneChan(ctx, s, v, rel, results)
		case mapping.RelHasMany:
			go getForeignRelationshipHasManyChan(ctx, s, v, rel, results)
		case mapping.RelMany2Many:
			go getForeignRelationshipManyToManyChan(ctx, s, v, rel, results)
		case mapping.RelBelongsTo:
			go getForeignRelationshipBelongsToChan(ctx, s, v, rel, results)
		default:
			log.Errorf("Invalid relationship type: %v", rel.Relationship())
			results <- struct{}{}
		}
	}

	var ctr int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-results:
			if !ok {
				return nil
			}

			if err, ok := v.(error); ok {
				return err
			}

			ctr++
			if ctr == len(relations) {
				return nil
			}
		}
	}
}

func getForeignRelationshipsSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}
	relations := map[string]*mapping.StructField{}

	// get all relationship field from the fieldset
	for _, field := range s.Fieldset {
		if field.Relationship() != nil {
			if _, ok := relations[field.NeuronName()]; !ok {
				relations[field.NeuronName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range s.includedFields {
		if _, ok := relations[included.NeuronName()]; !ok {
			relations[included.NeuronName()] = included.StructField
		}
	}

	if len(relations) == 0 {
		// finish fast if there is no relation in fieldset
		return nil
	}

	v := reflect.ValueOf(s.Value)
	if v.IsNil() {
		log.Infof("[SCOPE][%s] Nil scope's value while taking foreign relationships", s.ID())
		return errors.NewDet(class.QueryNoValue, "scope has nil value")
	}
	v = v.Elem()

	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, rel := range relations {
		if rel.Relationship().Struct().Config() == nil {
			continue
		}

		if modelRepo := rel.Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	var err error
	for _, rel := range relations {
		switch rel.Relationship().Kind() {
		case mapping.RelHasOne:
			err = getForeignRelationshipHasOne(ctx, s, v, rel)
		case mapping.RelHasMany:
			err = getForeignRelationshipHasMany(ctx, s, v, rel)
		case mapping.RelMany2Many:
			err = getForeignRelationshipManyToMany(ctx, s, v, rel)
		case mapping.RelBelongsTo:
			err = getForeignRelationshipBelongsTo(ctx, s, v, rel)
		}

		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func getForeignRelationshipHasOneChan(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField, results chan<- interface{}) {
	if err := getForeignRelationshipHasOne(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipHasOne(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField) error {
	// the scope should contain only foreign key in the fieldset
	var (
		err          error
		op           *Operator
		filterValues []interface{}
	)

	if !s.isMany {
		// if scope value is single use Get
		// Value already checked if nil
		op = OpEqual

		primVal := v.FieldByIndex(s.Struct().Primary().ReflectField().Index)
		prim := primVal.Interface()

		// check if primary field is zero
		if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
			return nil
		}
		filterValues = []interface{}{prim}
	} else {
		// if scope value is many use List
		op = OpIn
		// Iterate over a slice of values
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.IsNil() {
				continue
			}

			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			primVal := elem.FieldByIndex(s.Struct().Primary().ReflectField().Index)
			prim := primVal.Interface()

			// Check if primary field is zero valued
			if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
				continue
			}
			// if not add it to the filter values
			filterValues = append(filterValues, prim)
		}

		if len(filterValues) == 0 {
			return nil
		}
	}
	rel := relField.Relationship()
	fk := rel.ForeignKey()

	relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), s.isMany)
	// set the fieldset
	relatedScope.Fieldset = map[string]*mapping.StructField{}
	relatedScope.setFieldsetNoCheck(rel.Struct().Primary(), fk)
	relatedScope.collectionScope = relatedScope

	// set filterfield
	filter := NewFilter(fk, op, filterValues...)
	if err := relatedScope.addFilterField(filter); err != nil {
		return err
	}

	log.Debug2f("SCOPE[%s][%s] Getting hasOne relationship: '%s'", s.ID(), s.Struct().Collection(), relField.NeuronName())
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("hasOne Scope: %s", relatedScope.String())
	}
	// distinguish if the root scope has a single or many values
	if !s.isMany {
		err = relatedScope.GetContext(ctx)
		if err != nil {
			if e, ok := err.(errors.ClassError); ok {
				if e.Class() == class.QueryValueNoResult {
					return nil
				}
			}
			return err
		}
		// After getting the values set the scope relation value into the value of the relatedscope
		rf := v.FieldByIndex(relField.ReflectField().Index)
		if rf.Kind() == reflect.Ptr {
			rf.Set(reflect.ValueOf(relatedScope.Value))
		} else {
			rf.Set(reflect.ValueOf(relatedScope.Value).Elem())
		}
		return nil
	}

	// list scope type
	err = relatedScope.ListContext(ctx)
	if err != nil {
		e, ok := err.(errors.ClassError)
		if ok {
			if e.Class() == class.QueryValueNoResult {
				return nil
			}
		}
		return err
	}

	// iterate over relatedScope values and match the fk's with scope's primaries.
	relVal := reflect.ValueOf(relatedScope.Value)
	if relVal.IsNil() {
		log.Errorf("Relationship field's scope has nil value after List. Rel: %s", relField.NeuronName())
		return errors.NewDet(class.InternalQueryNilValue, "related scope has nil value")
	}

	// iterate over related values and match their foreign key's with the root primaries.
	relVal = relVal.Elem()
	for i := 0; i < relVal.Len(); i++ {
		elem := relVal.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		// ForeignKey Field Value
		fkVal := elem.FieldByIndex(fk.ReflectField().Index)
		// Primary Field from the Related Scope Value
		pkVal := elem.FieldByIndex(relatedScope.Struct().Primary().ReflectField().Index)
		pk := pkVal.Interface()

		if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
			log.Debugf("Relationship HasMany. Elem value with Zero Value for foreign key. Elem: %#v, FKName: %s", elem.Interface(), fk.NeuronName())
			continue
		}

		for j := 0; j < v.Len(); j++ {
			// iterate over root values and set their relationship values with the
			// values from the relVal.
			rootElem := v.Index(j)
			if rootElem.Kind() == reflect.Ptr {
				rootElem = rootElem.Elem()
			}

			rootPrim := rootElem.FieldByIndex(s.Struct().Primary().ReflectField().Index)
			// check if the root primary value matches the foreign key value
			if rootPrim.Interface() != fkVal.Interface() {
				continue
			}

			// get the relationship value
			rootElem.FieldByIndex(relField.ReflectField().Index).Set(relVal.Index(i))
			// add the new instance of the related model
			break
		}
	}
	return nil
}

func getForeignRelationshipHasManyChan(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField, results chan<- interface{}) {
	if err := getForeignRelationshipHasMany(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipHasMany(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField) error {
	// if relationship is synced get it from the related repository
	// Use a relatedRepo.List method
	// the scope should contain only foreign key as fieldset
	var (
		err          error
		op           *Operator
		filterValues []interface{}
		primMap      map[interface{}]int
	)

	rel := relField.Relationship()
	primIndex := s.Struct().Primary().ReflectField().Index

	// check if the scope is a single value
	if !s.isMany {
		op = OpEqual
		pkVal := v.FieldByIndex(primIndex)
		pk := pkVal.Interface()

		if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
			// if the primary value is not set the function should not enter here
			err = errors.NewDet(class.InternalQueryInvalidField, "primary field value should not be 'zero'")
			log.Errorf("Getting related HasMany scope failed. Primary field should not be zero. pk:%v, Scope: %s", pk, s)
			return err
		}

		rf := v.FieldByIndex(relField.ReflectField().Index)
		rf.Set(reflect.Zero(relField.ReflectField().Type))

		filterValues = append(filterValues, pk)
	} else {
		primMap = make(map[interface{}]int)
		op = OpIn
		// iterate over model instances and get the primary field mapping.
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)

			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}

			pkVal := elem.FieldByIndex(primIndex)
			pk := pkVal.Interface()

			// if the value is Zero like continue
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				continue
			}

			primMap[pk] = i

			rf := elem.FieldByIndex(relField.ReflectField().Index)
			rf.Set(reflect.Zero(relField.ReflectField().Type))
			filterValues = append(filterValues, pk)
		}
	}

	if len(filterValues) == 0 {
		return nil
	}

	relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), true)
	relatedScope.collectionScope = relatedScope

	fk := rel.ForeignKey()
	filterField := NewFilter(fk, op, filterValues...)

	if err = relatedScope.addFilterField(filterField); err != nil {
		log.Errorf("AddingFilterField failed. %v", err)
		return err
	}
	// clear the fieldset
	relatedScope.Fieldset = map[string]*mapping.StructField{}
	// set the fields to primary field and foreign key
	relatedScope.setFieldsetNoCheck(relatedScope.Struct().Primary(), fk)

	log.Debug2f("SCOPE[%s][%s] Getting hasMany relationship: '%s'", s.ID(), s.Struct().Collection(), relField.NeuronName())
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("hasMany Scope: %s", relatedScope.String())
	}
	if err = relatedScope.ListContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			if e.Class() == class.QueryValueNoResult {
				log.Debug2f("NoResult: %v", err)
				return nil
			}
		}
		log.Debugf("Error while getting foreign field: '%s'. Err: %v", rel.ForeignKey(), err)
		return err
	}

	relVal := reflect.ValueOf(relatedScope.Value).Elem()
	for i := 0; i < relVal.Len(); i++ {
		elem := relVal.Index(i)

		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		// get the foreignkey value
		fkVal := elem.FieldByIndex(fk.ReflectField().Index)
		// pkVal := elem.FieldByIndex(relatedScope.Struct.primary.refStruct.Index)

		if !s.isMany {
			elemRelVal := v.FieldByIndex(relField.ReflectField().Index)
			elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
			continue
		}
		// foreign key in relation would be a primary key within root scope
		j, ok := primMap[fkVal.Interface()]
		if !ok {
			log.Debugf("Relationship HasMany Foreign Key not found in the primaries map.")
			continue
		}
		// rootPK should be at index 'j'
		scopeElem := v.Index(j)
		if scopeElem.Kind() == reflect.Ptr {
			scopeElem = scopeElem.Elem()
		}

		elemRelVal := scopeElem.FieldByIndex(relField.ReflectField().Index)
		elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
	}
	return nil
}

func getForeignRelationshipManyToManyChan(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField, results chan<- interface{}) {
	if err := getForeignRelationshipManyToMany(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipManyToMany(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField) error {
	// TODO: reduce the calls to join model if the related field were already filtered
	// if the relationship is synced get the values from the relationship
	var (
		err     error
		primMap map[interface{}]int
		op      *Operator
	)

	filterValues := []interface{}{}
	rel := relField.Relationship()
	primIndex := s.Struct().Primary().ReflectField().Index

	if !s.isMany {
		// set isMany to false just to see the difference
		op = OpEqual

		pkVal := v.FieldByIndex(primIndex)
		pk := pkVal.Interface()

		if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
			err = errors.NewDet(class.InternalQueryFilter, "many2many relationship, no primary values are set for the scope")
			return err
		}

		filterValues = append(filterValues, pk)
	} else {
		// if the scope is 'many' valued map the primary keys with their indexes
		// Operator is 'in'
		op = OpIn

		// primMap contains the index of the proper scope value for
		//given primary key value
		primMap = map[interface{}]int{}
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)

			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}

			pkVal := elem.FieldByIndex(primIndex)
			pk := pkVal.Interface()

			// if the value is Zero like continue
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				continue
			}

			primMap[pk] = i

			filterValues = append(filterValues, pk)
		}
	}

	// backReference Foreign Key
	fk := rel.ForeignKey()

	// Foreign Key in the join model
	mtmFK := rel.ManyToManyForeignKey()

	// prepare new scope
	joinScope := newScopeWithModel(s.Controller(), rel.JoinModel(), true)

	joinScope.Fieldset = map[string]*mapping.StructField{}
	joinScope.setFieldsetNoCheck(fk, mtmFK)
	joinScope.collectionScope = joinScope

	// Add filter on the backreference foreign key (root scope primary keys) to the join scope
	filterField := NewFilter(fk, op, filterValues...)
	if err := joinScope.addFilterField(filterField); err != nil {
		return err
	}

	// do the List process on the join scope
	if err = joinScope.ListContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			// if the error is ErrNoResult don't throw an error - no relationship values
			if e.Class() == class.QueryValueNoResult {
				return nil
			}
		}
		return err
	}

	// get the joinScope value reflection
	relVal := reflect.ValueOf(joinScope.Value)
	if relVal.IsNil() {
		log.Errorf("Related Field scope: %s is nil after getting m2m relationships.", relField.NeuronName())
		err = errors.NewDet(class.InternalQueryNilValue, "nil value after listing many to many relationships")
		return err
	}

	relVal = relVal.Elem()

	// iterate over relatedScope Value
	for i := 0; i < relVal.Len(); i++ {
		// Get Each element from the relatedScope Value
		relElem := relVal.Index(i)

		// Dereference the ptr
		if relElem.Kind() == reflect.Ptr {
			if relElem.IsNil() {
				continue
			}
			relElem = relElem.Elem()
		}

		// get the foreign key value - which is our root scope's primary keys value
		fkValue := relElem.FieldByIndex(fk.ReflectField().Index)
		fkID := fkValue.Interface()

		// if the value is Zero continue
		if reflect.DeepEqual(fkID, reflect.Zero(fkValue.Type()).Interface()) {
			continue
		}

		// get the join model foreign key value which is the relation's foreign key
		fk := relElem.FieldByIndex(mtmFK.ReflectField().Index)

		var relationField reflect.Value

		// get scope's value at index
		if s.isMany {
			index, ok := primMap[fkID]
			if !ok {
				log.Errorf("Get Many2Many relationship. Can't find primary field within the mapping, for ID: %v", fkID)
				continue
			}

			single := v.Index(index).Elem()
			relationField = single.FieldByIndex(relField.ReflectField().Index)
		} else {
			relationField = relElem.FieldByIndex(relField.ReflectField().Index)
		}

		// create new instance of the relationship's model that would be appended into the relationships value
		toAppend := mapping.NewReflectValueSingle(relField.Relationship().Struct())

		// set the primary field value for the newly created instance
		prim := toAppend.Elem().FieldByIndex(relField.Relationship().Struct().Primary().ReflectField().Index)
		// the primary key would be the value of foreign key in join model ('fk')
		prim.Set(fk)

		// type check if the relation is a ptr to slice
		if relationField.Kind() == reflect.Ptr {
			// if the relation is ptr to slice - dereference it
			relationField = relationField.Elem()
		}
		// append the newly created instance to the relationship field
		relationField.Set(reflect.Append(relationField, toAppend))
	}
	return nil
}

func getForeignRelationshipBelongsToChan(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField, results chan<- interface{}) {
	if err := getForeignRelationshipBelongsTo(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipBelongsTo(ctx context.Context, s *Scope, v reflect.Value, relField *mapping.StructField) error {
	// if scope value is a slice iterate over all entries and transfer all foreign keys into relationship primaries.
	// for single value scope do it once
	var err error
	rel := relField.Relationship()
	fkField := rel.ForeignKey()

	if !s.isMany {
		// single scope value case
		fkVal := v.FieldByIndex(fkField.ReflectField().Index)
		// check if foreign key is zero
		if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {
			return nil
		}

		relVal := v.FieldByIndex(relField.ReflectField().Index)

		if relVal.Kind() == reflect.Ptr {
			if relVal.IsNil() {
				relVal.Set(reflect.New(relVal.Type().Elem()))
			}
			relVal = relVal.Elem()
		} else if relVal.Kind() != reflect.Struct {
			err = errors.NewDetf(class.InternalQueryInvalidField, "belongs to field with invalid type: '%s'", relVal.Type().Name())
			log.Warning(err)
			return err
		}

		relPrim := relVal.FieldByIndex(rel.Struct().Primary().ReflectField().Index)
		relPrim.Set(fkVal)
		return nil
	}
	// list type scope
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		// check if foreign key is not a zero
		fkVal := elem.FieldByIndex(fkField.ReflectField().Index)
		if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {
			continue
		}

		relVal := elem.FieldByIndex(relField.ReflectField().Index)
		if relVal.Kind() == reflect.Ptr {
			if relVal.IsNil() {
				relVal.Set(reflect.New(relVal.Type().Elem()))
			}
			relVal = relVal.Elem()
		} else if relVal.Kind() != reflect.Struct {
			err = errors.NewDetf(class.InternalQueryInvalidField, "relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
			log.Warning(err)
			return err
		}
		relPrim := relVal.FieldByIndex(rel.Struct().Primary().ReflectField().Index)
		relPrim.Set(fkVal)
	}
	return nil
}

// setBelongsToRelationshipsFunc sets the value from the belongs to relationship ID's to the foreign keys
func setBelongsToRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}
	return s.setBelongsToForeignKeyFields()
}
