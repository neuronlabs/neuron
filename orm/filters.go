package orm

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// reduceRelationshipFilters reduces all foreign relationship filters into a model's primary field key filters.
func reduceRelationshipFilters(ctx context.Context, db DB, s *query.Scope) error {
	var filters query.Filters
	for _, filter := range s.Filters {
		if !filter.StructField.IsRelationship() {
			continue
		}
		filters = append(filters, filter)
	}
	if len(filters) == 0 {
		return nil
	}

	if db.Controller().Config.SynchronousConnections {
		return reduceRelationshipFiltersSynchronous(ctx, db, s, filters)
	}
	return reduceRelationshipFiltersAsynchronous(ctx, db, s, filters)
}

func reduceRelationshipFiltersAsynchronous(ctx context.Context, db DB, s *query.Scope, filters query.Filters) error {
	var cancelFunc context.CancelFunc
	if _, deadlineSet := ctx.Deadline(); !deadlineSet {
		// if no default timeout is already set - try with 30 second timeout.
		ctx, cancelFunc = context.WithTimeout(ctx, time.Second*30)
	} else {
		// otherwise create a cancel function.
		ctx, cancelFunc = context.WithCancel(ctx)
	}
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	jobs := reduceFilterJobCreator(ctx, wg, filters)
	errChan := make(chan error, 1)
	for job := range jobs {
		go reduceFilterJob(ctx, db, s, job, wg, errChan)
	}

	// Wait channel waits until all jobs marks wait group done.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Debug2(logFormat(s, "Reduce relationship filters - context done"))
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Reduce relationship filters error: %v", e)
		return e
	case <-waitChan:
		if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
			log.Debug2f(logFormat(s, "Finished finding asynchronous reduce relationship filterslic"))
		}
	}
	return nil
}

func reduceRelationshipFiltersSynchronous(ctx context.Context, db DB, s *query.Scope, filters query.Filters) (err error) {
	for _, filter := range filters {
		if err = reduceRelationshipFilter(ctx, db, s, filter); err != nil {
			return err
		}
	}
	return nil
}

func reduceRelationshipFilter(ctx context.Context, db DB, s *query.Scope, filter *query.FilterField) error {
	switch filter.StructField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		return reduceBelongsToRelationshipFilter(ctx, db, s, filter)
	case mapping.RelHasMany, mapping.RelHasOne:
		return reduceHasManyRelationshipFilter(ctx, db, s, filter)
	case mapping.RelMany2Many:
		return reduceMany2ManyRelationshipFilter(ctx, db, s, filter)
	default:
		return errors.Newf(query.ClassFilterField, "filter's field: '%s' is not a relationship", filter.StructField)
	}
}

func reduceBelongsToRelationshipFilter(ctx context.Context, db DB, s *query.Scope, filter *query.FilterField) error {
	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f(logFormat(s, "reduceBelongsToRelationshipFilter field: '%s'"), filter.StructField)
	}
	// For the belongs to relationships if the filters are only over the relations primary key then we can change
	// the filter from relationship into the foreign key for the scope 's'
	var onlyPrimes = true
	for _, nested := range filter.Nested {
		if nested.StructField.Kind() != mapping.KindPrimary {
			onlyPrimes = false
			break
		}
	}

	if onlyPrimes {
		// Convert the filters into the foreign key filters of the root scope
		foreignKey := filter.StructField.Relationship().ForeignKey()
		filterField := s.GetOrCreateFieldFilter(foreignKey)

		for _, nested := range filter.Nested {
			filterField.Values = append(filterField.Values, nested.Values...)
		}
		return nil
	}

	// Otherwise find all primary key values from the relationship's model that matches the filters.
	relatedQuery := db.QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nestedFilter := range filter.Nested {
		relatedQuery.Filter(nestedFilter)
	}

	models, err := relatedQuery.Select(filter.StructField.Relationship().Struct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(models) == 0 {
		log.Debug2f(logFormat(s, "no belongs to relationship: '%s' results found"), filter.StructField)
		return errors.Newf(query.ClassNoResult, "no relationship: '%s' filter results found", filter.StructField)
	}

	// Get primary key values and set as the scope's foreign key field filters.
	primaries := make([]interface{}, len(models))
	for i, model := range models {
		primaries[i] = model.GetPrimaryKeyHashableValue()
	}
	filterField := s.GetOrCreateFieldFilter(filter.StructField.Relationship().ForeignKey())
	filterField.Values = append(filterField.Values, query.OperatorValues{Operator: query.OpIn, Values: primaries})
	return nil
}

// Has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func reduceHasManyRelationshipFilter(ctx context.Context, db DB, s *query.Scope, filter *query.FilterField) error {
	if log.CurrentLevel() == log.LevelDebug3 {
		log.Debug3f("[SCOPE][%s] reduceHasManyRelationshipFilter field: '%s'", s.ID, filter.StructField.Name())
	}
	// Check if all the nested filter fields are the foreign key of this relationship. These filters might be
	// primary field filters for the main scope.
	onlyForeign := true
	foreignKey := filter.StructField.Relationship().ForeignKey()
	for _, nested := range filter.Nested {
		if nested.StructField != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// Convert foreign key filter fields into main scope primary key field filters.
		primaryFilter := s.GetOrCreateFieldFilter(s.ModelStruct.Primary())
		for _, nested := range filter.Nested {
			primaryFilter.Values = append(primaryFilter.Values, nested.Values...)
		}
		return nil
	}

	relationQuery := db.QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nested := range filter.Nested {
		relationQuery.Filter(nested)
	}
	relationModels, err := relationQuery.Select(foreignKey).Find()
	if err != nil {
		return err
	}
	uniqueForeignKeyValues := make(map[interface{}]struct{})
	for _, relationModel := range relationModels {
		relationFielder, ok := relationModel.(mapping.Fielder)
		if !ok {
			return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
		}
		// Check if the foreign key field is not zero value.
		isZero, err := relationFielder.IsFieldZero(foreignKey)
		if err != nil {
			return err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
		if err != nil {
			return err
		}
		uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
	}
	// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
	if len(uniqueForeignKeyValues) == 0 {
		log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
		return errors.Newf(query.ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
	}
	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := s.GetOrCreateFieldFilter(s.ModelStruct.Primary())
	filterValues := query.OperatorValues{Operator: query.OpIn}
	for foreignKeyValue := range uniqueForeignKeyValues {
		filterValues.Values = append(filterValues.Values, foreignKeyValue)
	}
	primaryFilter.Values = append(primaryFilter.Values, filterValues)
	return nil
}

func reduceMany2ManyRelationshipFilter(ctx context.Context, db DB, s *query.Scope, filter *query.FilterField) error {
	log.Debug3f(logFormat(s, "reduceMany2ManyRelationshipFilter field: '%s'"), filter.StructField)

	// If all the nested filters are the primary key of the related model
	// the only query should be used on the join model's foreign key.
	onlyPrimaries := true
	relatedPrimaryKey := filter.StructField.Relationship().Struct().Primary()
	foreignKey := filter.StructField.Relationship().ForeignKey()

	for _, nested := range filter.Nested {
		if nested.StructField != relatedPrimaryKey && nested.StructField != foreignKey {
			onlyPrimaries = false
			break
		}
	}

	// If only the primaries of the related model were used find only the values from the join table
	joinModel := filter.StructField.Relationship().JoinModel()
	relatedModelForeignKey := filter.StructField.Relationship().ManyToManyForeignKey()
	if onlyPrimaries {
		// Convert the foreign into root scope primary key filter
		joinModelQuery := db.QueryCtx(ctx, joinModel)

		fkFilter := joinModelQuery.Scope().GetOrCreateFieldFilter(relatedModelForeignKey)
		for _, nested := range filter.Nested {
			fkFilter.Values = append(fkFilter.Values, nested.Values...)
		}
		// Select only foreign key field - the primary key value for the main scope.
		joinModels, err := joinModelQuery.Select(foreignKey).Find()
		if err != nil {
			return err
		}

		uniqueForeignKeyValues := make(map[interface{}]struct{})
		for _, joinModel := range joinModels {
			relationFielder, ok := joinModel.(mapping.Fielder)
			if !ok {
				return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
			}
			// Check if the foreign key field is not zero value.
			isZero, err := relationFielder.IsFieldZero(foreignKey)
			if err != nil {
				return err
			}
			if isZero {
				continue
			}
			foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
			if err != nil {
				return err
			}
			uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
		}
		// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
		if len(uniqueForeignKeyValues) == 0 {
			log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
			return errors.Newf(query.ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
		}

		// Create primary filter that matches all relationship foreign key values.
		primaryFilter := s.GetOrCreateFieldFilter(s.ModelStruct.Primary())
		filterValues := query.OperatorValues{Operator: query.OpIn}
		for foreignKeyValue := range uniqueForeignKeyValues {
			filterValues.Values = append(filterValues.Values, foreignKeyValue)
		}
		primaryFilter.Values = append(primaryFilter.Values, filterValues)
		return nil
	}

	relationQuery := db.QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nestedFilter := range filter.Nested {
		relationQuery.Filter(nestedFilter)
	}
	relatedModels, err := relationQuery.Select(filter.StructField.Relationship().Struct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(relatedModels) == 0 {
		log.Debug2f(logFormat(s, "No relationship: '%s' filter results found"), filter.StructField)
		return errors.Newf(query.ClassNoResult, "no relationship: '%s' filter results", filter.StructField)
	}
	var primaries []interface{}
	for _, model := range relatedModels {
		primaries = append(primaries, model.GetPrimaryKeyHashableValue())
	}

	joinModels, err := db.QueryCtx(ctx, joinModel).
		Filter(query.NewFilterField(relatedModelForeignKey, query.OpIn, primaries...)).
		Select(foreignKey).Find()
	if err != nil {
		return err
	}

	uniqueForeignKeyValues := make(map[interface{}]struct{})
	for _, joinModel := range joinModels {
		relationFielder, ok := joinModel.(mapping.Fielder)
		if !ok {
			return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
		}
		// Check if the foreign key field is not zero value.
		isZero, err := relationFielder.IsFieldZero(foreignKey)
		if err != nil {
			return err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
		if err != nil {
			return err
		}
		uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
	}
	// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
	if len(uniqueForeignKeyValues) == 0 {
		log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
		return errors.Newf(query.ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
	}

	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := s.GetOrCreateFieldFilter(s.ModelStruct.Primary())
	filterValues := query.OperatorValues{Operator: query.OpIn}
	for foreignKeyValue := range uniqueForeignKeyValues {
		filterValues.Values = append(filterValues.Values, foreignKeyValue)
	}
	primaryFilter.Values = append(primaryFilter.Values, filterValues)
	return nil
}

func reduceFilterJobCreator(ctx context.Context, wg *sync.WaitGroup, filters query.Filters) (jobs <-chan *query.FilterField) {
	out := make(chan *query.FilterField)
	go func() {
		defer close(out)
		for _, filter := range filters {
			wg.Add(1)
			select {
			case out <- filter:
			case <-ctx.Done():
				return

			}
		}
	}()
	return out
}

func reduceFilterJob(ctx context.Context, db DB, s *query.Scope, filter *query.FilterField, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	if err := reduceRelationshipFilter(ctx, db, s, filter); err != nil {
		errChan <- err
	}
}
