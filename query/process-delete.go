package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/neuron/common"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
)

var (
	// ProcessDelete is the process that does the Repository Delete method
	ProcessDelete = &Process{
		Name: "neuron:delete",
		Func: deleteFunc,
	}

	// ProcessReducePrimaryFilters is the process that reduces the primary filters for the given process
	ProcessReducePrimaryFilters = &Process{
		Name: "neuron:reduce_primary_filters",
		Func: reducePrimaryFilters,
	}

	// ProcessBeforeDelete is the Process that does the BeforeDelete hook
	ProcessBeforeDelete = &Process{
		Name: "neuron:hook_before_delete",
		Func: beforeDeleteFunc,
	}

	// ProcessAfterDelete is the Process that does the AfterDelete hook
	ProcessAfterDelete = &Process{
		Name: "neuron:hook_after_delete",
		Func: afterDeleteFunc,
	}

	// ProcessDeleteForeignRelationships is the Process that deletes the foreing relatioionships
	ProcessDeleteForeignRelationships = &Process{
		Name: "neuron:delete_foreign_relationships",
		Func: deleteForeignRelationshipsFunc,
	}

	// ProcessDeleteForeignRelationshipsSafe is the Process that deletes the foreing relatioionships
	ProcessDeleteForeignRelationshipsSafe = &Process{
		Name: "neuron:delete_foreign_relationships_safe",
		Func: deleteForeignRelationshipsSafeFunc,
	}
)

func deleteFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return err
	}

	dRepo, ok := repo.(Deleter)
	if !ok {
		log.Warningf("Repository for model: '%v' doesn't implement Deleter interface", s.Struct().Type().Name())
		return errors.Newf(class.RepositoryNotImplementsDeleter, "repository: %T doesn't implement Deleter interface", repo)
	}

	// do the delete operation
	if err := dRepo.Delete(ctx, s); err != nil {
		return err
	}

	return nil
}

func beforeDeleteFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	beforeDeleter, ok := s.Value.(BeforeDeleter)
	if !ok {
		return nil
	}

	if err := beforeDeleter.BeforeDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func afterDeleteFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	afterDeleter, ok := s.Value.(AfterDeleter)
	if !ok {
		return nil
	}

	if err := afterDeleter.AfterDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func deleteForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	// get only relationship fields
	var relationships []*models.StructField
	for _, field := range s.internal().Struct().Fields() {
		if field.IsRelationship() {
			if field.Relationship().Kind() != models.RelBelongsTo {
				relationships = append(relationships, field)
			}
		}
	}

	if len(relationships) > 0 {
		var results = make(chan interface{}, len(relationships))

		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Processor.DefaultTimeout
		for _, rel := range relationships {
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

		// delete foreign relationships
		for _, field := range relationships {
			switch field.Relationship().Kind() {
			case models.RelHasOne:
				go deleteHasOneRelationshipsChan(ctx, s, field, results)
			case models.RelHasMany:
				go deleteHasManyRelationshipsChan(ctx, s, field, results)
			case models.RelMany2Many:
				go deleteMany2ManyRelationshipsChan(ctx, s, field, results)
			case models.RelBelongsTo:
				continue
			}
		}

		var ctr int

		for {
			select {
			case <-ctx.Done():
			case v, ok := <-results:
				if !ok {
					return nil
				}
				if err, ok := v.(error); ok {
					return err
				}
				ctr++
				if ctr == len(relationships) {
					return nil
				}
			}
		}
	}
	return nil
}

func deleteForeignRelationshipsSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	// get only relationship fields
	var relationships []*models.StructField
	for _, field := range s.internal().Struct().Fields() {
		if field.IsRelationship() {
			if field.Relationship().Kind() != models.RelBelongsTo {
				relationships = append(relationships, field)
			}
		}
	}

	if len(relationships) > 0 {
		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Processor.DefaultTimeout
		for _, rel := range relationships {
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
		// delete foreign relationships
		for _, field := range relationships {
			switch field.Relationship().Kind() {
			case models.RelHasOne:
				err = deleteHasOneRelationships(ctx, s, field)
			case models.RelHasMany:
				err = deleteHasManyRelationships(ctx, s, field)
			case models.RelMany2Many:
				err = deleteMany2ManyRelationships(ctx, s, field)
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

func deleteHasOneRelationshipsChan(ctx context.Context, s *Scope, field *models.StructField, results chan<- interface{}) {
	if err := deleteHasOneRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteHasOneRelationships(ctx context.Context, s *Scope, field *models.StructField) error {
	// clearScope clears the foreign key values for the relationships
	var (
		clearScope *Scope
		err        error
	)

	rel := field.Relationship()

	if tx := s.Tx(); tx != nil {
		clearScope, err = tx.newModelC(ctx, s.Controller(), rel.Struct(), false)
		if err != nil {
			return err
		}
	} else {
		clearScope = queryS(newScopeWithModel(s.Controller(), rel.Struct(), false))
	}

	// the selected field would be only the foreign key -> zero valued
	clearScope.internal().AddSelectedField(rel.ForeignKey())

	for _, prim := range s.internal().PrimaryFilters() {
		err = clearScope.internal().AddFilterField(filters.NewFilter(rel.ForeignKey(), prim.Values()...))
		if err != nil {
			log.Debugf("Adding Relationship's foreign key failed: %v", err)
			return err
		}
	}

	// patch the clearScope
	if err = clearScope.PatchContext(ctx); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			if e.Class == class.QueryValueNoResult {
				err = nil
			}
		}
	}
	return err
}

func deleteHasManyRelationshipsChan(ctx context.Context, s *Scope, field *models.StructField, results chan<- interface{}) {
	if err := deleteHasManyRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteHasManyRelationships(ctx context.Context, s *Scope, field *models.StructField) error {
	// clearScope clears the foreign key values for the relationships
	var (
		clearScope *Scope
		err        error
	)

	rel := field.Relationship()

	if tx := s.Tx(); tx != nil {
		clearScope, err = tx.newModelC(ctx, s.Controller(), rel.Struct(), true)
		if err != nil {
			return err
		}
	} else {
		clearScope = queryS(newScopeWithModel(s.Controller(), rel.Struct(), true))
	}

	// the selected field would be only the foreign key -> zero valued
	clearScope.internal().AddSelectedField(rel.ForeignKey())

	for _, prim := range s.internal().PrimaryFilters() {
		err = clearScope.internal().AddFilterField(filters.NewFilter(rel.ForeignKey(), prim.Values()...))
		if err != nil {
			log.Debugf("Adding Relationship's foreign key failed: %v", err)
			return err
		}
	}

	// patch the clearScope
	err = clearScope.PatchContext(ctx)
	if err != nil {
		switch e := err.(type) {
		case *errors.Error:
			if e.Class == class.QueryValueNoResult {
				err = nil
			}
		}
	}
	return err
}

func deleteMany2ManyRelationshipsChan(ctx context.Context, s *Scope, field *models.StructField, results chan<- interface{}) {
	if err := deleteMany2ManyRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteMany2ManyRelationships(ctx context.Context, s *Scope, field *models.StructField) error {
	var (
		clearScope *Scope
		err        error
	)
	rel := field.Relationship()
	// there is assumption that only the primary filters exists on the delete scope
	// delete the many2many rows within the join model
	if tx := s.Tx(); tx != nil {
		clearScope, err = tx.newModelC(ctx, s.Controller(), rel.JoinModel(), false)
		if err != nil {
			return err
		}
	} else {
		clearScope = (*Scope)(newScopeWithModel(s.Controller(), rel.JoinModel(), false))
	}

	// add the backreference filter
	foreignKeyFilter := filters.NewFilter(rel.ForeignKey())

	// with the values of the primary filter
	for _, prim := range s.internal().PrimaryFilters() {
		foreignKeyFilter.AddValues(prim.Values()...)
	}

	if err = clearScope.internal().AddFilterField(foreignKeyFilter); err != nil {
		log.Debugf("Deleting relationship: '%s' AddFilterField failed: %v ", field.Name(), err)
		return err
	}

	// delete the entries in the join model
	err = clearScope.DeleteContext(ctx)
	if err != nil {
		switch e := err.(type) {
		case *errors.Error:
			if e.Class == class.QueryValueNoResult {
				err = nil
			}
		}
	}
	return err
}

// reducePrimaryFilters is the process func that changes the delete scope filters so that
// if the root model contains any nonBelongsTo relationship then the filters must be converted into primary field filter
func reducePrimaryFilters(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	reducedPrimariesInterface, alreadyReduced := s.internal().StoreGet(internal.ReducedPrimariesStoreKey)
	if alreadyReduced {
		previousProcess, ok := s.StoreGet(internal.PreviousProcessStoreKey)
		if ok {
			switch previousProcess.(*Process) {
			case ProcessBeforeDelete:
				_, isBeforeDeleter := s.Value.(BeforeDeleter)
				if !isBeforeDeleter {
					return nil
				}
			case ProcessBeforePatch:
				_, isBeforePatcher := s.Value.(BeforePatcher)
				if !isBeforePatcher {
					return nil
				}
			}
		}
	}

	var reduce bool

	// get the primary field values from the scope's value
	if s.Value != nil {
		primaryValues, err := s.internal().Struct().PrimaryValues(reflect.ValueOf(s.Value))
		if err != nil {
			return err
		}

		if len(primaryValues) > 0 {
			// if there are any primary values in the scope values
			// create a filter field with these values
			primaryFilter := s.internal().GetOrCreateIDFilter()
			if s.internal().IsPrimaryFieldSelected() {
				if err = s.internal().UnselectFields(s.internal().Struct().PrimaryField()); err != nil {
					return err
				}
			}
			primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, primaryValues...))
		}
	}

	var primaries []interface{}

	// iterate over all primary filters
	// get the primary filter values
	// if the joinOperatorValues is true change the OpEqual into OpIn
	// if the operator is different than filters.OpIn and filters.OpEqual
	// then the reduction is necessary
	for _, pk := range s.internal().PrimaryFilters() {
		for _, fv := range pk.Values() {
			switch fv.Operator() {
			case filters.OpIn:
				primaries = append(primaries, fv.Values...)
			case filters.OpEqual:
				fv.SetOperator(filters.OpIn)
				primaries = append(primaries, fv.Values...)
			default:
				reduce = true
			}
		}
	}

	if alreadyReduced {
		reducedPrimaries := reducedPrimariesInterface.([]interface{})
		if len(primaries) == len(reducedPrimaries) {
			var primaryMap = map[interface{}]struct{}{}

			for _, prim := range reducedPrimaries {
				primaryMap[prim] = struct{}{}
			}

			var allFound = true
			for _, prim := range primaries {
				_, ok := primaryMap[prim]
				if !ok {
					allFound = false
					break
				}
			}

			if allFound {
				return nil
			}
		}
	}

	if !reduce {
		// if the scope has any non primary field filters set the reduction true
		if len(s.internal().AttributeFilters()) != 0 || len(s.internal().RelationshipFilters()) != 0 || s.internal().LanguageFilter() != nil || len(s.internal().ForeignKeyFilters()) != 0 {
			reduce = true
		}
	}

	if len(primaries) == 0 && !reduce {
		return errors.New(class.QueryFilterMissingRequired, "no primary filter or values provided").SetDetail("No primary field nor primary filters provided while patching the model")
	}

	// if no reduction needed just add the primary filter and deselect the
	if !reduce {
		// just set the primary values in the store and finish
		s.StoreSet(internal.ReducedPrimariesStoreKey, primaries)
		return nil
	}

	primaryScope := NewModelC(s.Controller(), s.Struct(), true)

	// get only the primary field
	primaryScope.internal().SetEmptyFieldset()
	primaryScope.internal().SetFieldsetNoCheck((*models.StructField)(s.Struct().Primary()))

	// set the filters to the primary scope - they would get cleared after
	if err := s.internal().SetFiltersTo(primaryScope.internal()); err != nil {
		return err
	}

	// list the primary fields
	if err := primaryScope.ListContext(ctx); err != nil {
		return err
	}

	// overwrite the primaries
	primaries, err := primaryScope.internal().GetPrimaryFieldValues()
	if err != nil {
		log.Errorf("Getting primary field values failed: %v", err)
		return err
	}

	if len(primaries) == 0 {
		return errors.New(class.QueryValueNoResult, "query value no results")
	}

	// clear all filters in the root scope
	s.internal().ClearAllFilters()
	s.internal().UnselectFieldIfSelected(s.internal().Struct().PrimaryField())

	// reduce the filters as the primary filters in the root scope
	primaryFilter := s.internal().GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	s.StoreSet(internal.ReducedPrimariesStoreKey, primaries)
	s.StoreSet(internal.PrimariesAlreadyChecked, struct{}{})

	return nil
}
