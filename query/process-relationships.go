package query

import (
	"context"
	"fmt"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/uni-db"

	"github.com/neuronlabs/neuron/internal/query/filters"
	scope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"reflect"
)

var (
	// ProcessGetIncluded is the process that gets the included scope values
	ProcessGetIncluded = &Process{
		Name: "neuron:get_included",
		Func: getIncludedFunc,
	}

	// ProcessSetBelongsToRelationships is the Process that sets the BelongsToRelationships
	ProcessSetBelongsToRelationships = &Process{
		Name: "neuron:set_belongs_to_relationships",
		Func: setBelongsToRelationshipsFunc,
	}

	// ProcessGetForeignRelationships is the Process that gets the foreign relationships
	ProcessGetForeignRelationships = &Process{
		Name: "neuron:get_foreign_relationships",
		Func: getForeignRelationshipsFunc,
	}

	// ProcessConvertRelationshipFilters converts the relationship filters into a primary or foreign key filters of the root scope
	ProcessConvertRelationshipFilters = &Process{
		Name: "neuron:convert_relationship_filters",
		Func: convertRelationshipFilters,
	}
)

func convertRelationshipFilters(ctx context.Context, s *Scope) error {
	relationshipFilters := s.internal().RelationshipFilters()

	if len(relationshipFilters) == 0 {
		// end fast if no relationship filters
		return nil
	}

	bufSize := 10

	if len(relationshipFilters) < bufSize {
		bufSize = len(relationshipFilters)
	}

	var results = make(chan interface{}, bufSize)

	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, rel := range relationshipFilters {
		if rel.StructField().Relationship().Struct().Config() == nil {
			continue
		}
		if modelRepo := rel.StructField().Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	for i, rf := range (*scope.Scope)(s).RelationshipFilters() {

		// cast the filter to the internal structure definition
		filter := (*filters.FilterField)(rf)

		// check the relatinoship kind
		switch rf.StructField().Relationship().Kind() {
		case models.RelBelongsTo:
			go handleBelongsToRelationshipFilter(ctx, s, i, filter, results)

		case models.RelHasMany:
			// these shoud contain the foreign key in their definitions
			go handleHasManyRelationshipFilter(ctx, s, i, filter, results)

		case models.RelHasOne:
			// these shoud contain the foreign key in their definitions
			go handleHasOneRelationshipFilter(ctx, s, i, filter, results)

		case models.RelMany2Many:
			go handleMany2ManyRelationshipFilter(ctx, s, i, filter, results)
		}
	}

	var ctr int
fl:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-results:
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(relationshipFilters) {
				break fl
			}
		}
	}

	// if all relationship filters are done we can clear them
	(*scope.Scope)(s).ClearRelationshipFilters()

	return nil
}

func handleBelongsToRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
	// for the belongs to relationships if the relationship filters only over the relations primary
	// key then we can change the filter from relationship into the foreign key for the root scope
	// otherwise we need to get the

	var onlyPrimes = true
	for _, nested := range filter.NestedFields() {
		if nested.StructField().FieldKind() != models.KindPrimary {
			onlyPrimes = false
			break
		}
	}

	if onlyPrimes {
		// convert the filters into the foreign key filters of the root scope
		var foreignKey = filter.StructField().Relationship().ForeignKey()

		filterField := (*scope.Scope)(s).GetOrCreateForeignKeyFilter(foreignKey)

		for _, nested := range filter.NestedFields() {
			filterField.AddValues(nested.Values()...)
		}

		// send the index of the filter to remove from the scope
		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// we only need primary keys
	relScope.SetFieldset(relScope.Struct().Primary())

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	primaries, err := (*scope.Scope)(relScope).GetPrimaryFieldValues()
	if err != nil {
		results <- err
		return
	}

	if len(primaries) == 0 {
		log.Debugf("SCOPE[%s] getting BelongsTo relationship filters no results found.", s.ID().String())
		results <- unidb.ErrNoResult.New()
		return
	}

	// convert the filters into the foreign key filters of the root scope
	var foreignKey = filter.StructField().Relationship().ForeignKey()

	filterField := (*scope.Scope)(s).GetOrCreateForeignKeyFilter(foreignKey)

	filterField.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	// send the index of the filter to remove from the scope
	results <- index

	return
}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func handleHasManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {

	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	var onlyForeign = true

	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// the only required field in fieldset is a foreign key
	(*scope.Scope)(relScope).SetFields(foreignKey)

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	foreignValues, err := (*scope.Scope)(relScope).GetForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		results <- err
		return
	}

	if len(foreignValues) == 0 {
		results <- unidb.ErrNoResult.New()
		log.Debugf("No results found for the relationship filter for field: %s", filter.StructField().ApiName())
		return
	}

	primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	results <- index

}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func handleHasOneRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {

	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	var onlyForeign = true

	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// the only required field in fieldset is a foreign key
	(*scope.Scope)(relScope).SetFields(foreignKey)

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	foreignValues, err := (*scope.Scope)(relScope).GetForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		results <- err
		return
	}

	if len(foreignValues) == 0 {
		results <- unidb.ErrNoResult.New()
		return
	}

	if len(foreignValues) > 1 && !(*scope.Scope)(s).IsMany() {
		log.Warningf("Multiple foreign key values found for a single valued - has one scope.")
	}

	primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	results <- index
}

func handleMany2ManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
	// if filter is only on the 'id' field then

	// if all the nested filters are the primary key of the related model
	// the only query should be used on the join model foreign key
	var onlyPrimaries = true

	relatedPrimaryKey := filter.StructField().Relationship().Struct().PrimaryField()
	joinModelForeignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != relatedPrimaryKey && nested.StructField() != joinModelForeignKey {
			onlyPrimaries = false
			break
		}
	}

	// if only the primaries of the related model were used list only the values from the join table
	if onlyPrimaries {

		// convert the foreign into root scope primary key filter
		joinScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().JoinModel()), true)

		// create the joinScope foreign key filter with the values of the original filter
		fkFilter := (*scope.Scope)(joinScope).GetOrCreateForeignKeyFilter(filter.StructField().Relationship().ForeignKey())
		for _, nested := range filter.NestedFields() {
			fkFilter.AddValues(nested.Values()...)
		}

		// the fieldset should contain only a backreference field
		joinScope.SetFieldset(filter.StructField().Relationship().BackreferenceForeignKey())

		// Do the ListProcess with the context
		if err := joinScope.ListContext(ctx); err != nil {
			results <- err
			return
		}

		// get the foreign key values for the given joinScope
		fkValues, err := (*scope.Scope)(joinScope).GetForeignKeyValues(filter.StructField().Relationship().BackreferenceForeignKey())
		if err != nil {
			results <- err
			return
		}

		// Add (or get) the ID filter for the primary key in the root scope
		rootIDFilter := (*scope.Scope)(s).GetOrCreateIDFilter()

		// add the backreference field values into primary key filter
		rootIDFilter.AddValues(filters.NewOpValuePair(filters.OpIn, fkValues...))

		results <- index
		return
	}

	// create the scope for the related model
	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	// add then nested fields from the filter field into the
	for _, nested := range filter.NestedFields() {
		var nestedFilter *filters.FilterField
		switch nested.StructField().FieldKind() {
		case models.KindPrimary:
			nestedFilter = (*scope.Scope)(relScope).GetOrCreateIDFilter()
		case models.KindAttribute:
			if nested.StructField().IsLanguage() {
				nestedFilter = (*scope.Scope)(relScope).GetOrCreateLanguageFilter()
			} else {
				nestedFilter = (*scope.Scope)(relScope).GetOrCreateAttributeFilter(nested.StructField())
			}
		case models.KindForeignKey:
			nestedFilter = (*scope.Scope)(relScope).GetOrCreateForeignKeyFilter(nested.StructField())
		case models.KindFilterKey:
			nestedFilter = (*scope.Scope)(relScope).GetOrCreateFilterKeyFilter(nested.StructField())
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:
			nestedFilter = (*scope.Scope)(relScope).GetOrCreateRelationshipFilter(nested.StructField())
		default:
			log.Debugf("Nested Filter for field of unknown type: %v", nested.StructField().ApiName())
			continue
		}

		nestedFilter.AddValues(nested.Values()...)
	}

	// only the primary field should be in a fieldset
	relScope.internal().SetEmptyFieldset()
	(*scope.Scope)(relScope).SetFieldsetNoCheck((*models.StructField)(relScope.Struct().Primary()))

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	primaries, err := relScope.internal().GetPrimaryFieldValues()
	if err != nil {
		results <- err
		return
	}

	joinScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().JoinModel()), true)

	idFilter := joinScope.internal().GetOrCreateIDFilter()
	idFilter.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	joinScope.internal().SetNilFieldset()
	joinScope.internal().SetFieldsetNoCheck(filter.StructField().Relationship().BackreferenceForeignKey())
	if err := joinScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	fkValues, err := joinScope.internal().GetForeignKeyValues(filter.StructField().Relationship().BackreferenceForeignKey())
	if err != nil {
		results <- err
		return
	}

	rootIDFilter := s.internal().GetOrCreateIDFilter()
	rootIDFilter.AddValues(filters.NewOpValuePair(filters.OpIn, fkValues...))

	results <- index

}

// processGetIncluded gets the included fields for the
func getIncludedFunc(ctx context.Context, s *Scope) error {
	iScope := (*scope.Scope)(s)

	if iScope.IsRoot() && len(iScope.IncludedScopes()) == 0 {
		return nil
	}

	if err := iScope.SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, incScope := range iScope.IncludedScopes() {

		if incScope.Struct().Config() == nil {
			continue
		}
		if modelRepo := incScope.Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	includedFields := iScope.IncludedFields()

	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) {

		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			results <- err
			return
		}
		log.Debugf("Missing primaries: %v", missing)

		if len(missing) > 0 {
			includedScope := includedField.Scope

			includedScope.SetIDFilters(missing...)
			includedScope.NewValueMany()

			if err = (*Scope)(includedScope).ListContext(ctx); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				results <- err
				return
			}

		}
		results <- struct{}{}
	}

	// send the jobs
	for _, includedField := range includedFields {
		go getInclude(includedField, results)
	}

	// collect the results
	var ctr int
	for {
		select {
		case <-ctx.Done():
		case v, ok := <-results:
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(includedFields) {
				break
			}
		}
	}
}

func getForeignRelationshipsFunc(ctx context.Context, s *Scope) error {

	relations := map[string]*models.StructField{}

	// get all relationship field from the fieldset
	for _, field := range (*scope.Scope)(s).Fieldset() {
		if field.Relationship() != nil {
			if _, ok := relations[field.ApiName()]; !ok {
				relations[field.ApiName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range (*scope.Scope)(s).IncludedFields() {
		if _, ok := relations[included.ApiName()]; !ok {
			relations[included.ApiName()] = included.StructField
		}
	}

	if len(relations) > 0 {

		v := reflect.ValueOf(s.Value)
		if v.IsNil() {
			log.Debugf("Scope has a nil value.")
			return internal.ErrNilValue
		}
		v = v.Elem()

		// log.Debugf("Internal check for the ctrl: %p", ctx.Value(internal.ControllerCtxKey))

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
			go getForeignRelationshipValue(ctx, s, v, rel, results)
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
	return nil
}

func getForeignRelationshipValue(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	res chan<- interface{},
) {

	var err error

	defer func() {
		if err != nil {
			res <- err
		} else {
			res <- struct{}{}
		}
	}()
	// Check the relationships
	rel := relField.Relationship()

	switch rel.Kind() {
	case models.RelHasOne:
		// if relationship is synced get the relations primaries from the related repository
		// if scope value is many use List
		//
		// if scope value is single use Get
		//
		// the scope should contain only foreign key as fieldset
		// Create new relation scope
		var (
			op           *filters.Operator
			filterValues []interface{}
		)

		if !(*scope.Scope)(s).IsMany() {

			// Value already checked if nil
			op = filters.OpEqual

			primVal := v.FieldByIndex(s.Struct().Primary().ReflectField().Index)
			prim := primVal.Interface()

			// check if primary field is zero
			if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
				return
			}

			filterValues = []interface{}{prim}
		} else {
			op = filters.OpIn
			// Iterate over an slice of values

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
				return
			}
		}

		// Create the
		fk := rel.ForeignKey()

		relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), (*scope.Scope)(s).IsMany())

		// set the fieldset
		relatedScope.SetEmptyFieldset()
		relatedScope.SetFieldsetNoCheck(rel.Struct().PrimaryField(), fk)
		relatedScope.SetFlagsFrom((*scope.Scope)(s).Flags())
		relatedScope.SetCollectionScope(relatedScope)

		// set filterfield
		filter := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		relatedScope.AddFilterField(filter)

		if (*scope.Scope)(s).IsMany() {
			err = (*Scope)(relatedScope).ListContext(ctx)
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok {
					if dberr.Compare(unidb.ErrNoResult) {
						err = nil
						return
					}
				}

				return
			}

			// iterate over relatedScope values and match the fk's with scope's primaries.
			relVal := reflect.ValueOf(relatedScope.Value)
			if relVal.IsNil() {
				log.Errorf("Relationship field's scope has nil value after List. Rel: %s", relField.ApiName())
				err = internal.ErrNilValue
				return
			}

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
				pkVal := elem.FieldByIndex(relatedScope.Struct().PrimaryField().ReflectField().Index)

				pk := pkVal.Interface()

				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					log.Debugf("Relationship HasMany. Elem value with Zero Value for foreign key. Elem: %#v, FKName: %s", elem.Interface(), fk.ApiName())
					continue
				}

			rootLoop:
				for j := 0; j < v.Len(); j++ {
					rootElem := v.Index(j)
					if rootElem.Kind() == reflect.Ptr {
						rootElem = rootElem.Elem()
					}

					rootPrim := rootElem.FieldByIndex(s.Struct().Primary().ReflectField().Index)
					if rootPrim.Interface() == fkVal.Interface() {
						rootElem.FieldByIndex(relField.ReflectField().Index).Set(relVal.Index(i))
						break rootLoop
					}
				}

			}
		} else {

			err = (*Scope)(relatedScope).GetContext(ctx)
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok && dberr.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
				return
			}
			// After getting the values set the scope relation value into the value of the relatedscope
			rf := v.FieldByIndex(relField.ReflectField().Index)
			if rf.Kind() == reflect.Ptr {
				rf.Set(reflect.ValueOf(relatedScope.Value))
			} else {
				rf.Set(reflect.ValueOf(relatedScope.Value).Elem())
			}
		}

	case models.RelHasMany:

		// if relationship is synced get it from the related repository
		// Use a relatedRepo.List method
		// the scope should contain only foreign key as fieldset
		var (
			op           *filters.Operator
			filterValues []interface{}

			primMap   map[interface{}]int
			primIndex = s.Struct().Primary().ReflectField().Index
		)

		// If is single scope's value
		if !(*scope.Scope)(s).IsMany() {

			op = filters.OpEqual
			pkVal := v.FieldByIndex(primIndex)
			pk := pkVal.Interface()
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				// if the primary value is not set the function should not enter here
				err = fmt.Errorf("Empty Scope Primary Value")
				log.Debugf("Err: %v. pk:%v, Scope: %#v", err, pk, s)
				return
			}

			rf := v.FieldByIndex(relField.ReflectField().Index)
			rf.Set(reflect.Zero(relField.ReflectField().Type))

			filterValues = append(filterValues, pk)
		} else {
			primMap = map[interface{}]int{}
			op = filters.OpIn
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
		log.Debugf("Filter Values: %v", filterValues)
		if len(filterValues) == 0 {
			return
		}
		relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), true)
		relatedScope.SetCollectionScope(relatedScope)

		fk := rel.ForeignKey()

		filterField := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		if err = relatedScope.AddFilterField(filterField); err != nil {
			log.Errorf("AddingFilterField failed. %v", err)
			return
		}
		// clear the fieldset
		relatedScope.SetEmptyFieldset()

		// set the fields to primary field and foreign key
		relatedScope.SetFieldsetNoCheck(relatedScope.Struct().PrimaryField(), fk)

		if err = (*Scope)(relatedScope).ListContext(ctx); err != nil {
			dberr, ok := err.(*unidb.Error)
			if ok {
				if dberr.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			log.Debugf("Error while getting foreign field: '%s'. Err: %v", rel.BackreferenceForeignKeyName(), err)
			return
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

			if (*scope.Scope)(s).IsMany() {

				// foreign key in relation would be a primary key within root scope
				j, ok := primMap[fkVal.Interface()]
				if !ok {

					// TODO: write some good debug log
					log.Debugf("Foreign Key not found in the primaries map.")
					continue
				}

				// rootPK should be at index 'j'
				var scopeElem reflect.Value
				scopeElem = v.Index(j)
				if scopeElem.Kind() == reflect.Ptr {
					scopeElem = scopeElem.Elem()
				}

				elemRelVal := scopeElem.FieldByIndex(relField.ReflectField().Index)
				elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
			} else {
				elemRelVal := v.FieldByIndex(relField.ReflectField().Index)
				elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
			}

		}

	case models.RelMany2Many:
		// if the relationship is synced get the values from the relationship

		// when sync get the relationships value from the backreference relationship
		var (
			filterValues = []interface{}{}
			primMap      map[interface{}]int
			primIndex    = s.Struct().Primary().ReflectField().Index
			op           *filters.Operator
		)

		if !(*scope.Scope)(s).IsMany() {
			// set isMany to false just to see the difference

			op = filters.OpEqual

			pkVal := v.FieldByIndex(primIndex)
			pk := pkVal.Interface()

			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				err = fmt.Errorf("Error no primary valoue set for the scope value.  Scope %#v. Value: %+v", s, s.Value)
				return
			}

			filterValues = append(filterValues, pk)

		} else {
			// if the scope is 'many' valued map the primary keys with their indexes
			// Operator is 'in'
			op = filters.OpIn

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
		backReference := rel.BackreferenceForeignKey()

		// Foreign Key in the join model
		fk := backReference.Relationship().ForeignKey()

		// prepare new scope
		joinScope := newScopeWithModel(s.Controller(), rel.JoinModel(), true)

		joinScope.SetEmptyFieldset()
		joinScope.SetFieldsetNoCheck(backReference, fk)

		joinScope.SetCollectionScope(joinScope)

		// Add filter on the backreference foreign key (root scope primary keys) to the join scope
		filterField := filters.NewFilter(backReference, filters.NewOpValuePair(op, filterValues...))
		joinScope.AddFilterField(filterField)

		// do the List process on the join scope
		if err = (*Scope)(joinScope).ListContext(ctx); err != nil {
			dberr, ok := err.(*unidb.Error)
			if ok {
				// if the error is ErrNoResult don't throw an error - no relationship values
				if dberr.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			return
		}

		// get the joinScope value reflection
		relVal := reflect.ValueOf(joinScope.Value)
		if relVal.IsNil() {
			log.Errorf("Related Field scope: %s is nil after getting m2m relationships.", relField.ApiName())
			err = unidb.ErrInternalError.NewWithMessage("get many2Many foreign relationship - nil value after list")
			return
		}

		relVal = relVal.Elem()

		// iterate over relatedScoep Value
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

			// get the backreference foreign key value - which in fact is our root scope's primary key value
			backReferenceValue := relElem.FieldByIndex(backReference.ReflectField().Index)
			backReferenceID := backReferenceValue.Interface()

			// if the value is Zero continue
			if reflect.DeepEqual(backReferenceID, reflect.Zero(backReferenceValue.Type()).Interface()) {
				continue
			}

			// get the join model foreign key value which is the relation's foreign key
			fk := relElem.FieldByIndex(fk.ReflectField().Index)

			var relationField reflect.Value

			// get scope's value at index
			if s.internal().IsMany() {

				index, ok := primMap[backReferenceID]
				if !ok {
					log.Errorf("Get Many2Many relationship. Can't find primary field within the mapping, for ID: %v", backReferenceID)
					continue
				}

				single := relVal.Index(index).Elem()
				relationField = single.FieldByIndex(relField.ReflectField().Index)
			} else {
				relationField = relVal.FieldByIndex(relField.ReflectField().Index)
			}

			// create new instance of the relationship's model that would be appended into the relationships value
			toAppend := relField.Relationship().Struct().NewReflectValueSingle()

			// set the primary field value for the newly created instance
			prim := toAppend.Elem().FieldByIndex(relField.Relationship().Struct().PrimaryField().ReflectField().Index)

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

	case models.RelBelongsTo:

		// if scope value is a slice
		// iterate over all entries and transfer all foreign keys into relationship primaries.

		// for single value scope do it once

		fkField := rel.ForeignKey()

		if !(*scope.Scope)(s).IsMany() {

			fkVal := v.FieldByIndex(fkField.ReflectField().Index)

			// check if foreign key is zero
			if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {

				return
			}

			relVal := v.FieldByIndex(relField.ReflectField().Index)

			if relVal.Kind() == reflect.Ptr {
				if relVal.IsNil() {
					relVal.Set(reflect.New(relVal.Type().Elem()))
				}
				relVal = relVal.Elem()
			} else if relVal.Kind() != reflect.Struct {
				err = fmt.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
				log.Warning(err)
				return
			}

			relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
			relPrim.Set(fkVal)

		} else {
			// is a list type scope
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
					err = fmt.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
					log.Warning(err)
					return
				}
				relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
				relPrim.Set(fkVal)
			}
		}
	}

}

// setBelongsToRelationshipsFunc sets the value from the belongs to relationship ID's to the foreign keys
func setBelongsToRelationshipsFunc(ctx context.Context, s *Scope) error {
	return (*scope.Scope)(s).SetBelongsToForeignKeyFields()
}
