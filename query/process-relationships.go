package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/query/filters"
	"github.com/neuronlabs/neuron-core/internal/query/scope"
)

var (
	// ProcessGetIncluded is the process that gets the included scope values.
	ProcessGetIncluded = &Process{
		Name: "neuron:get_included",
		Func: getIncludedFunc,
	}

	// ProcessGetIncludedSafe is the gouroutine safe process that gets the included scope values.
	ProcessGetIncludedSafe = &Process{
		Name: "neuron:get_included_safe",
		Func: getIncludedSafeFunc,
	}

	// ProcessSetBelongsToRelationships is the Process that sets the BelongsToRelationships.
	ProcessSetBelongsToRelationships = &Process{
		Name: "neuron:set_belongs_to_relationships",
		Func: setBelongsToRelationshipsFunc,
	}

	// ProcessGetForeignRelationships is the Process that gets the foreign relationships.
	ProcessGetForeignRelationships = &Process{
		Name: "neuron:get_foreign_relationships",
		Func: getForeignRelationshipsFunc,
	}

	// ProcessGetForeignRelationshipsSafe is the goroutine safe Process that gets the foreign relationships.
	ProcessGetForeignRelationshipsSafe = &Process{
		Name: "neuron:get_foreign_relationships_safe",
		Func: getForeignRelationshipsSafeFunc,
	}

	// ProcessConvertRelationshipFilters converts the relationship filters into a primary or foreign key filters of the root scope.
	ProcessConvertRelationshipFilters = &Process{
		Name: "neuron:convert_relationship_filters",
		Func: convertRelationshipFiltersFunc,
	}

	// ProcessConvertRelationshipFiltersSafe goroutine safe process that converts the relationship filters into a primary or foreign key filters of the root scope.
	ProcessConvertRelationshipFiltersSafe = &Process{
		Name: "neuron:convert_relationship_filters_safe",
		Func: convertRelationshipFiltersSafeFunc,
	}
)

func convertRelationshipFiltersFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	relationshipFilters := s.internal().RelationshipFilters()
	if len(relationshipFilters) == 0 {
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

	for i, rf := range s.internal().RelationshipFilters() {
		// cast the filter to the internal structure definition
		filter := (*filters.FilterField)(rf)

		// check the relationship kind
		switch rf.StructField().Relationship().Kind() {
		case models.RelBelongsTo:
			go convertBelongsToRelationshipFilterChan(ctx, s, i, filter, results)
		case models.RelHasMany:
			go convertHasManyRelationshipFilterChan(ctx, s, i, filter, results)
		case models.RelHasOne:
			go convertHasOneRelationshipFilterChan(ctx, s, i, filter, results)
		case models.RelMany2Many:
			go convertMany2ManyRelationshipFilterChan(ctx, s, i, filter, results)
		default:
			err := errors.NewDetf(class.InternalQueryInvalidField, "invalid field's relationship kind. Model: %s, Field: %s", s.Struct().Type().Name(), rf.StructField().Name())
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

			if ctr == len(relationshipFilters) {
				done = true
			}
		}
		// finish the for loop when done
		if done {
			break
		}
	}

	// if all relationship filters are done we can clear them
	s.internal().ClearRelationshipFilters()

	return nil
}

func convertRelationshipFiltersSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	relationshipFilters := s.internal().RelationshipFilters()
	if len(relationshipFilters) == 0 {
		return nil
	}

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

	var err error

	for i, rf := range s.internal().RelationshipFilters() {
		// cast the filter to the internal structure definition
		filter := (*filters.FilterField)(rf)

		// check the relationship kind
		switch rf.StructField().Relationship().Kind() {
		case models.RelBelongsTo:
			_, err = convertBelongsToRelationshipFilter(ctx, s, i, filter)
		case models.RelHasMany:
			_, err = convertHasManyRelationshipFilter(ctx, s, i, filter)
		case models.RelHasOne:
			_, err = convertHasOneRelationshipFilter(ctx, s, i, filter)
		case models.RelMany2Many:
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
	s.internal().ClearRelationshipFilters()

	return nil
}

func convertBelongsToRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
	index, err := convertBelongsToRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

func convertBelongsToRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField) (int, error) {
	log.Debug2f("[SCOPE][%s] convertBelongsToRelationshipFilter field: '%s'", s.ID(), filter.StructField().Name())
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
		foreignKey := filter.StructField().Relationship().ForeignKey()
		filterField := s.internal().GetOrCreateForeignKeyFilter(foreignKey)

		for _, nested := range filter.NestedFields() {
			filterField.AddValues(nested.Values()...)
		}
		// send the index of the filter to remove from the scope
		return index, nil
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := relScope.internal().AddFilterField(nested); err != nil {
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

	primaries, err := relScope.internal().GetPrimaryFieldValues()
	if err != nil {
		return 0, err
	}

	if len(primaries) == 0 {
		log.Debugf("SCOPE[%s] getting BelongsTo relationship filters no results found.", s.ID().String())
		err := errors.NewDet(class.QueryValueNoResult, "no result found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField().NeuronName())
		return 0, err
	}

	// convert the filters into the foreign key filters of the root scope
	foreignKey := filter.StructField().Relationship().ForeignKey()
	filterField := s.internal().GetOrCreateForeignKeyFilter(foreignKey)
	filterField.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	// send the index of the filter to remove from the scope
	return index, nil
}

func convertHasManyRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
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
func convertHasManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField) (int, error) {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("[SCOPE][%s] convertHasManyRelationshipFilter field: '%s'", s.ID(), filter.StructField().Name())
	}
	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	onlyForeign := true
	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := s.internal().GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		return index, nil
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := relScope.internal().AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)
			return 0, err
		}
	}

	// the only required field in fieldset is a foreign key
	if err := relScope.internal().SetFields(foreignKey); err != nil {
		return 0, err
	}

	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	foreignValues, err := relScope.internal().GetUniqueForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		return 0, err
	}

	if len(foreignValues) == 0 {
		log.Debugf("No results found for the relationship filter for field: %s", filter.StructField().NeuronName())
		err := errors.NewDet(class.QueryValueNoResult, "no relationship filter values found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField().NeuronName())
		return 0, err
	}

	primaryFilter := s.internal().GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	return index, nil
}

func convertHasOneRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
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
func convertHasOneRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField) (int, error) {
	log.Debug3f("[SCOPE][%s] convertHasOneRelationshipFilter field: '%s'", s.ID(), filter.StructField().Name())
	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	onlyForeign := true
	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := s.internal().GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		return index, nil
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := relScope.internal().AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)
			// send the error to the results
			return 0, err
		}
	}

	// the only required field in fieldset is a foreign key
	if err := relScope.internal().SetFields(foreignKey); err != nil {
		return 0, err
	}

	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	foreignValues, err := relScope.internal().GetForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		return 0, err
	}

	if len(foreignValues) == 0 {
		err := errors.NewDet(class.QueryValueNoResult, "no result found")
		err.SetDetailsf("relationship: '%s' values not found", filter.StructField().NeuronName())
		return 0, err
	}

	if len(foreignValues) > 1 && !s.internal().IsMany() {
		log.Warningf("Multiple foreign key values found for a single valued - has one scope.")
	}

	primaryFilter := s.internal().GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	return index, nil
}

func convertMany2ManyRelationshipFilterChan(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
	index, err := convertMany2ManyRelationshipFilter(ctx, s, index, filter)
	if err != nil {
		results <- err
	} else {
		results <- index
	}
}

func convertMany2ManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField) (int, error) {
	log.Debug3f("[SCOPE][%s] convertMany2ManyRelationshipFilter field: '%s'", s.ID(), filter.StructField().Name())
	// if filter is only on the 'id' field then

	// if all the nested filters are the primary key of the related model
	// the only query should be used on the join model foreign key
	onlyPrimaries := true
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
		fkFilter := joinScope.internal().GetOrCreateForeignKeyFilter(filter.StructField().Relationship().ManyToManyForeignKey())
		for _, nested := range filter.NestedFields() {
			fkFilter.AddValues(nested.Values()...)
		}

		// the fieldset should contain only a backreference field
		if err := joinScope.SetFieldset(filter.StructField().Relationship().ForeignKey()); err != nil {
			return 0, err
		}

		// Do the ListProcess with the context
		if err := joinScope.ListContext(ctx); err != nil {
			return 0, err
		}

		// get the foreign key values for the given joinScope
		fkValues, err := joinScope.internal().GetForeignKeyValues(filter.StructField().Relationship().ForeignKey())
		if err != nil {
			return 0, err
		}

		// Add (or get) the ID filter for the primary key in the root scope
		rootIDFilter := s.internal().GetOrCreateIDFilter()

		// add the backreference field values into primary key filter
		rootIDFilter.AddValues(filters.NewOpValuePair(filters.OpIn, fkValues...))

		return index, nil
	}

	// create the scope for the related model
	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	// add then nested fields from the filter field into the
	for _, nested := range filter.NestedFields() {
		var nestedFilter *filters.FilterField
		switch nested.StructField().FieldKind() {
		case models.KindPrimary:
			nestedFilter = relScope.internal().GetOrCreateIDFilter()
		case models.KindAttribute:
			if nested.StructField().IsLanguage() {
				nestedFilter = relScope.internal().GetOrCreateLanguageFilter()
			} else {
				nestedFilter = relScope.internal().GetOrCreateAttributeFilter(nested.StructField())
			}
		case models.KindForeignKey:
			nestedFilter = relScope.internal().GetOrCreateForeignKeyFilter(nested.StructField())
		case models.KindFilterKey:
			nestedFilter = relScope.internal().GetOrCreateFilterKeyFilter(nested.StructField())
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:
			nestedFilter = relScope.internal().GetOrCreateRelationshipFilter(nested.StructField())
		default:
			log.Debugf("Nested Filter for field of unknown type: %v", nested.StructField().NeuronName())
			continue
		}

		nestedFilter.AddValues(nested.Values()...)
	}

	// only the primary field should be in a fieldset
	relScope.internal().SetEmptyFieldset()
	relScope.internal().SetFieldsetNoCheck((*models.StructField)(relScope.Struct().Primary()))

	if err := relScope.ListContext(ctx); err != nil {
		return 0, err
	}

	primaries, err := relScope.internal().GetPrimaryFieldValues()
	if err != nil {
		return 0, err
	}

	if len(primaries) == 0 {
		err = errors.NewDet(class.QueryValueNoResult, "related filter values doesn't exist")
		return 0, err
	}

	joinScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().JoinModel()), true)

	mtmFKFilter := joinScope.internal().GetOrCreateForeignKeyFilter(filter.StructField().Relationship().ManyToManyForeignKey())
	mtmFKFilter.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	joinScope.internal().SetEmptyFieldset()
	joinScope.internal().SetFieldsetNoCheck(filter.StructField().Relationship().ForeignKey())
	if err := joinScope.ListContext(ctx); err != nil {
		return 0, err
	}

	fkValues, err := joinScope.internal().GetUniqueForeignKeyValues(filter.StructField().Relationship().ForeignKey())
	if err != nil {
		return 0, err
	}

	rootIDFilter := s.internal().GetOrCreateIDFilter()
	rootIDFilter.AddValues(filters.NewOpValuePair(filters.OpIn, fkValues...))

	return index, nil
}

// processGetIncluded gets the included fields for the
func getIncludedFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	if s.internal().IsRoot() && len(s.internal().IncludedScopes()) == 0 {
		return nil
	}

	if err := s.internal().SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, incScope := range s.internal().IncludedScopes() {
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

	includedFields := s.internal().IncludedFields()
	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) {
		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			results <- err
			return
		}

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

func getIncludedSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	if s.internal().IsRoot() && len(s.internal().IncludedScopes()) == 0 {
		return nil
	}

	if err := s.internal().SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, incScope := range s.internal().IncludedScopes() {
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

	includedFields := s.internal().IncludedFields()
	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) error {
		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			return err
		}

		if len(missing) > 0 {
			includedScope := includedField.Scope
			includedScope.SetIDFilters(missing...)
			includedScope.NewValueMany()

			if err = (*Scope)(includedScope).ListContext(ctx); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				return err
			}
		}
		return nil
	}

	// send the jobs
	for _, includedField := range includedFields {
		err := getInclude(includedField, results)
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

func getForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	relations := map[string]*models.StructField{}

	// get all relationship field from the fieldset
	for _, field := range s.internal().Fieldset() {
		if field.Relationship() != nil {
			if _, ok := relations[field.NeuronName()]; !ok {
				relations[field.NeuronName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range s.internal().IncludedFields() {
		if _, ok := relations[included.NeuronName()]; !ok {
			relations[included.NeuronName()] = included.StructField
		}
	}

	if len(relations) > 0 {
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
			case models.RelHasOne:
				go getForeignRelationshipHasOneChan(ctx, s, v, rel, results)
			case models.RelHasMany:
				go getForeignRelationshipHasManyChan(ctx, s, v, rel, results)
			case models.RelMany2Many:
				go getForeignRelationshipManyToManyChan(ctx, s, v, rel, results)
			case models.RelBelongsTo:
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
	return nil
}

func getForeignRelationshipsSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	relations := map[string]*models.StructField{}

	// get all relationship field from the fieldset
	for _, field := range s.internal().Fieldset() {
		if field.Relationship() != nil {
			if _, ok := relations[field.NeuronName()]; !ok {
				relations[field.NeuronName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range s.internal().IncludedFields() {
		if _, ok := relations[included.NeuronName()]; !ok {
			relations[included.NeuronName()] = included.StructField
		}
	}

	if len(relations) > 0 {
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
			case models.RelHasOne:
				err = getForeignRelationshipHasOne(ctx, s, v, rel)
			case models.RelHasMany:
				err = getForeignRelationshipHasMany(ctx, s, v, rel)
			case models.RelMany2Many:
				err = getForeignRelationshipManyToMany(ctx, s, v, rel)
			case models.RelBelongsTo:
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
	}
	return nil
}

func getForeignRelationshipHasOneChan(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	results chan<- interface{},
) {
	if err := getForeignRelationshipHasOne(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipHasOne(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
) error {
	// if relationship is synced get the relations primaries from the related repository
	//
	// the scope should contain only foreign key as fieldset
	// Create new relation scope
	var (
		err          error
		op           *filters.Operator
		filterValues []interface{}
	)

	if !s.internal().IsMany() {
		// if scope value is single use Get
		// Value already checked if nil
		op = filters.OpEqual

		primVal := v.FieldByIndex(s.Struct().Primary().ReflectField().Index)
		prim := primVal.Interface()

		// check if primary field is zero
		if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
			return nil
		}

		filterValues = []interface{}{prim}
	} else {
		// if scope value is many use List
		op = filters.OpIn

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

	// Create the
	fk := rel.ForeignKey()

	relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), s.internal().IsMany())

	// set the fieldset
	relatedScope.SetEmptyFieldset()
	relatedScope.SetFieldsetNoCheck(rel.Struct().PrimaryField(), fk)
	relatedScope.SetCollectionScope(relatedScope)

	// set filterfield
	filter := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

	if err := relatedScope.AddFilterField(filter); err != nil {
		return err
	}

	// distinguish if the root scope has a single or many values
	if s.internal().IsMany() {
		err = (*Scope)(relatedScope).ListContext(ctx)
		if err != nil {
			e, ok := err.(errors.DetailedError)
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
			pkVal := elem.FieldByIndex(relatedScope.Struct().PrimaryField().ReflectField().Index)
			pk := pkVal.Interface()

			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				log.Debugf("Relationship HasMany. Elem value with Zero Value for foreign key. Elem: %#v, FKName: %s", elem.Interface(), fk.NeuronName())
				continue
			}

		rootLoop:
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
				break rootLoop
			}
		}
	} else {
		err = (*Scope)(relatedScope).GetContext(ctx)
		if err != nil {
			if e, ok := err.(errors.DetailedError); ok {
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
	}
	return nil
}

func getForeignRelationshipHasManyChan(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	results chan<- interface{},
) {
	if err := getForeignRelationshipHasMany(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipHasMany(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
) error {
	// if relationship is synced get it from the related repository
	// Use a relatedRepo.List method
	// the scope should contain only foreign key as fieldset
	var (
		err          error
		op           *filters.Operator
		filterValues []interface{}
		primMap      map[interface{}]int
	)

	rel := relField.Relationship()
	primIndex := s.Struct().Primary().ReflectField().Index

	// check if the scope is a single value
	if !s.internal().IsMany() {
		op = filters.OpEqual
		pkVal := v.FieldByIndex(primIndex)
		pk := pkVal.Interface()

		if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
			// if the primary value is not set the function should not enter here
			err = errors.NewDet(class.InternalQueryInvalidField, "primary field value should not be 'zero'")
			log.Errorf("Getting related HasMany scope failed. Primary field should not be zero. pk:%v, Scope: %#v", pk, s)
			return err
		}

		rf := v.FieldByIndex(relField.ReflectField().Index)
		rf.Set(reflect.Zero(relField.ReflectField().Type))

		filterValues = append(filterValues, pk)
	} else {
		primMap = make(map[interface{}]int)
		op = filters.OpIn

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
	relatedScope.SetCollectionScope(relatedScope)

	fk := rel.ForeignKey()
	filterField := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

	if err = relatedScope.AddFilterField(filterField); err != nil {
		log.Errorf("AddingFilterField failed. %v", err)
		return err
	}
	// clear the fieldset
	relatedScope.SetEmptyFieldset()

	// set the fields to primary field and foreign key
	relatedScope.SetFieldsetNoCheck(relatedScope.Struct().PrimaryField(), fk)

	if err = (*Scope)(relatedScope).ListContext(ctx); err != nil {
		if e, ok := err.(errors.DetailedError); ok {
			if e.Class() == class.QueryValueNoResult {
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

		if s.internal().IsMany() {
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
		} else {
			elemRelVal := v.FieldByIndex(relField.ReflectField().Index)
			elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
		}
	}
	return nil
}

func getForeignRelationshipManyToManyChan(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	results chan<- interface{},
) {
	if err := getForeignRelationshipManyToMany(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipManyToMany(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
) error {
	// TODO: reduce the calls to join model if the related field were already filtered
	// if the relationship is synced get the values from the relationship
	var (
		err     error
		primMap map[interface{}]int
		op      *filters.Operator
	)

	filterValues := []interface{}{}
	rel := relField.Relationship()
	primIndex := s.Struct().Primary().ReflectField().Index

	if !s.internal().IsMany() {
		// set isMany to false just to see the difference
		op = filters.OpEqual

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
	fk := rel.ForeignKey()

	// Foreign Key in the join model
	mtmFK := rel.ManyToManyForeignKey()

	// prepare new scope
	joinScope := newScopeWithModel(s.Controller(), rel.JoinModel(), true)

	joinScope.SetEmptyFieldset()
	joinScope.SetFieldsetNoCheck(fk, mtmFK)
	joinScope.SetCollectionScope(joinScope)

	// Add filter on the backreference foreign key (root scope primary keys) to the join scope
	filterField := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))
	if err := joinScope.AddFilterField(filterField); err != nil {
		return err
	}

	// do the List process on the join scope
	if err = (*Scope)(joinScope).ListContext(ctx); err != nil {
		if e, ok := err.(errors.DetailedError); ok {
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
		if s.internal().IsMany() {
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
	return nil
}

func getForeignRelationshipBelongsToChan(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	results chan<- interface{},
) {
	if err := getForeignRelationshipBelongsTo(ctx, s, v, relField); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func getForeignRelationshipBelongsTo(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
) error {
	// if scope value is a slice
	// iterate over all entries and transfer all foreign keys into relationship primaries.
	// for single value scope do it once
	var err error

	rel := relField.Relationship()
	fkField := rel.ForeignKey()

	if !s.internal().IsMany() {
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
				err = errors.NewDetf(class.InternalQueryInvalidField, "relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
				log.Warning(err)
				return err
			}
			relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
			relPrim.Set(fkVal)
		}
	}
	return nil
}

// setBelongsToRelationshipsFunc sets the value from the belongs to relationship ID's to the foreign keys
func setBelongsToRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	return s.internal().SetBelongsToForeignKeyFields()
}
