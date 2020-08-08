package database

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// reduceRelationshipFilters reduces all foreign relationship filters into a model's primary field key filters.
func reduceRelationshipFilters(ctx context.Context, db DB, s *query.Scope) error {
	var (
		relationFilters []filter.Relation
		filters         []filter.Filter
	)
	for _, f := range s.Filters {
		switch ft := f.(type) {
		case filter.Relation:
			relationFilters = append(relationFilters, ft)
		case filter.OrGroup:
			filters = append(filters, ft)
		case filter.Simple:
			filters = append(filters, ft)
		}
	}
	s.Filters = filters
	if len(relationFilters) == 0 {
		return nil
	}

	if db.Controller().Options.SynchronousConnections {
		return reduceRelationshipFiltersSynchronous(ctx, db, s, relationFilters...)
	}
	return reduceRelationshipFiltersAsynchronous(ctx, db, s, relationFilters...)
}

func reduceRelationshipFiltersAsynchronous(ctx context.Context, db DB, s *query.Scope, filters ...filter.Relation) error {
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
	jobs := reduceFilterJobCreator(ctx, wg, filters...)
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
			log.Debug2f(logFormat(s, "Finished finding asynchronous reduce relationship filters"))
		}
	}
	return nil
}

func reduceRelationshipFiltersSynchronous(ctx context.Context, db DB, s *query.Scope, filters ...filter.Relation) (err error) {
	for _, f := range filters {
		if err = reduceRelationshipFilter(ctx, db, s, f); err != nil {
			return err
		}
	}
	return nil
}

func reduceRelationshipFilter(ctx context.Context, db DB, s *query.Scope, relationFilter filter.Relation) error {
	switch relationFilter.StructField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		return reduceBelongsToRelationshipFilter(ctx, db, s, relationFilter)
	case mapping.RelHasMany, mapping.RelHasOne:
		return reduceHasManyRelationshipFilter(ctx, db, s, relationFilter)
	case mapping.RelMany2Many:
		return reduceMany2ManyRelationshipFilter(ctx, db, s, relationFilter)
	default:
		return errors.Wrapf(filter.ErrFilterField, "filter's field: '%s' is not a relationship", relationFilter.StructField)
	}
}

func reduceBelongsToRelationshipFilter(ctx context.Context, db DB, s *query.Scope, f filter.Relation) error {
	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f(logFormat(s, "reduceBelongsToRelationshipFilter field: '%s'"), f.StructField)
	}
	// For the belongs to relationships if the filters are only over the relations primary key then we can change
	// the filter from relationship into the foreign key for the scope 's'
	var onlyPrimes = true
	for _, nested := range f.Nested {
		if simple, ok := nested.(*filter.Simple); ok {
			if simple.StructField.Kind() != mapping.KindPrimary {
				onlyPrimes = false
				break
			}
		}
	}

	if onlyPrimes {
		// Convert the filters into the foreign key filters of the root scope
		for _, nested := range f.Nested {
			simple := nested.(filter.Simple)
			simple.StructField = f.StructField.Relationship().ForeignKey()
			s.Filters = append(s.Filters, simple)
		}
		return nil
	}

	// Otherwise find all primary key values from the relationship's model that matches the filters.
	relatedQuery := db.QueryCtx(ctx, f.StructField.Relationship().RelatedModelStruct())
	for _, nestedFilter := range f.Nested {
		relatedQuery.Filter(nestedFilter)
	}

	models, err := relatedQuery.Select(f.StructField.Relationship().RelatedModelStruct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(models) == 0 {
		log.Debug2f(logFormat(s, "no belongs to relationship: '%s' results found"), f.StructField)
		return errors.Wrapf(query.ErrQueryNoResult, "no relationship: '%s' filter results found", f.StructField)
	}

	// Get primary key values and set as the scope's foreign key field filters.
	primaries := make([]interface{}, len(models))
	for i, model := range models {
		primaries[i] = model.GetPrimaryKeyHashableValue()
	}
	s.Filters = append(s.Filters, filter.New(s.ModelStruct.Primary(), filter.OpIn, primaries...))
	return nil
}

// Has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func reduceHasManyRelationshipFilter(ctx context.Context, db DB, s *query.Scope, relationFilter filter.Relation) error {
	if log.CurrentLevel() == log.LevelDebug3 {
		log.Debug3f("[SCOPE][%s] reduceHasManyRelationshipFilter field: '%s'", s.ID, relationFilter.StructField.Name())
	}
	// Check if all the nested filter fields are the foreign key of this relationship. These filters might be
	// primary field filters for the main scope.
	onlyForeign := true
	foreignKey := relationFilter.StructField.Relationship().ForeignKey()
	for _, nested := range relationFilter.Nested {
		if simple, ok := nested.(filter.Simple); ok {
			if simple.StructField != foreignKey {
				onlyForeign = false
				break
			}
		}
	}

	if onlyForeign {
		// Convert foreign key filter fields into main scope primary key field filters.
		for _, nested := range relationFilter.Nested {
			simple := nested.(filter.Simple)
			simple.StructField = s.ModelStruct.Primary()
			s.Filters = append(s.Filters, simple)
		}
		return nil
	}

	relationQuery := db.QueryCtx(ctx, relationFilter.StructField.Relationship().RelatedModelStruct())
	for _, nested := range relationFilter.Nested {
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
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder", relationFilter.StructField.Relationship().RelatedModelStruct())
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
		log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), relationFilter.StructField.NeuronName())
		return errors.Wrapf(query.ErrQueryNoResult, "no relationship: '%s' filter filterValues found", relationFilter.StructField)
	}
	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := filter.Simple{StructField: s.ModelStruct.Primary(), Operator: filter.OpIn}
	for foreignKeyValue := range uniqueForeignKeyValues {
		primaryFilter.Values = append(primaryFilter.Values, foreignKeyValue)
	}
	return nil
}

func reduceMany2ManyRelationshipFilter(ctx context.Context, db DB, s *query.Scope, relationFilter filter.Relation) error {
	log.Debug3f(logFormat(s, "reduceMany2ManyRelationshipFilter field: '%s'"), relationFilter.StructField)

	// If all the nested filters are the primary key of the related model
	// the only query should be used on the join model's foreign key.
	onlyPrimaries := true
	relatedPrimaryKey := relationFilter.StructField.Relationship().RelatedModelStruct().Primary()
	foreignKey := relationFilter.StructField.Relationship().ForeignKey()

	for _, nested := range relationFilter.Nested {
		if simple, ok := nested.(filter.Simple); ok {
			if simple.StructField != relatedPrimaryKey && simple.StructField != foreignKey {
				onlyPrimaries = false
				break
			}
		}
	}

	// If only the primaries of the related model were used find only the values from the join table
	joinModel := relationFilter.StructField.Relationship().JoinModel()
	relatedModelForeignKey := relationFilter.StructField.Relationship().ManyToManyForeignKey()
	if onlyPrimaries {
		// Convert the foreign into root scope primary key filter
		joinModelQuery := db.QueryCtx(ctx, joinModel)

		joinModelScope := joinModelQuery.Scope()
		for _, nested := range relationFilter.Nested {
			simple := nested.(filter.Simple)
			simple.StructField = relatedModelForeignKey
			joinModelScope.Filters = append(joinModelScope.Filters, simple)
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
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder", relationFilter.StructField.Relationship().RelatedModelStruct())
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
			log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), relationFilter.StructField.NeuronName())
			return errors.Wrapf(query.ErrQueryNoResult, "no relationship: '%s' filter filterValues found", relationFilter.StructField)
		}

		// Create primary filter that matches all relationship foreign key values.
		primaryFilter := filter.Simple{StructField: s.ModelStruct.Primary(), Operator: filter.OpIn, Values: make([]interface{}, len(uniqueForeignKeyValues))}
		var i int
		for foreignKeyValue := range uniqueForeignKeyValues {
			primaryFilter.Values[i] = foreignKeyValue
			i++
		}
		s.Filters = append(s.Filters, primaryFilter)
		return nil
	}

	relationQuery := db.QueryCtx(ctx, relationFilter.StructField.Relationship().RelatedModelStruct())
	for _, nestedFilter := range relationFilter.Nested {
		relationQuery.Filter(nestedFilter)
	}
	relatedModels, err := relationQuery.Select(relationFilter.StructField.Relationship().RelatedModelStruct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(relatedModels) == 0 {
		log.Debug2f(logFormat(s, "No relationship: '%s' filter results found"), relationFilter.StructField)
		return errors.Wrapf(query.ErrQueryNoResult, "no relationship: '%s' filter results", relationFilter.StructField)
	}
	var primaries []interface{}
	for _, model := range relatedModels {
		primaries = append(primaries, model.GetPrimaryKeyHashableValue())
	}

	joinModels, err := db.QueryCtx(ctx, joinModel).
		Filter(filter.New(relatedModelForeignKey, filter.OpIn, primaries...)).
		Select(foreignKey).Find()
	if err != nil {
		return err
	}

	uniqueForeignKeyValues := make(map[interface{}]struct{})
	for _, joinModel := range joinModels {
		relationFielder, ok := joinModel.(mapping.Fielder)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder", relationFilter.StructField.Relationship().RelatedModelStruct())
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
		log.Debug2f(logFormat(s, "No results found for the relationship filter for field: %s"), relationFilter.StructField.NeuronName())
		return errors.Wrapf(query.ErrQueryNoResult, "no relationship: '%s' filter filterValues found", relationFilter.StructField)
	}

	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := filter.Simple{StructField: s.ModelStruct.Primary(), Operator: filter.OpIn, Values: make([]interface{}, len(uniqueForeignKeyValues))}
	var i int
	for foreignKeyValue := range uniqueForeignKeyValues {
		primaryFilter.Values[i] = foreignKeyValue
	}
	s.Filters = append(s.Filters, primaryFilter)
	return nil
}

func reduceFilterJobCreator(ctx context.Context, wg *sync.WaitGroup, filters ...filter.Relation) (jobs <-chan filter.Relation) {
	out := make(chan filter.Relation)
	go func() {
		defer close(out)
		for _, relationFilter := range filters {
			wg.Add(1)
			select {
			case out <- relationFilter:
			case <-ctx.Done():
				return

			}
		}
	}()
	return out
}

func reduceFilterJob(ctx context.Context, db DB, s *query.Scope, filter filter.Relation, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	if err := reduceRelationshipFilter(ctx, db, s, filter); err != nil {
		errChan <- err
	}
}
