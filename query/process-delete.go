package query

import (
	"context"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/uni-db"
	"reflect"

	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
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

	// ProcessBeforeDelete is the Process that does the HBeforeDelete hook
	ProcessBeforeDelete = &Process{
		Name: "neuron:hook_before_delete",
		Func: beforeDeleteFunc,
	}

	// ProcessAfterDelete is the Process that does the HAfterDelete hook
	ProcessAfterDelete = &Process{
		Name: "neuron:hook_after_delete",
		Func: afterDeleteFunc,
	}

	// ProcessDeleteForeignRelationships is the Process that deletes the foreing relatioionships
	ProcessDeleteForeignRelationships = &Process{
		Name: "neuron:delete_foreign_relationships",
		Func: deleteForeignRelationshipsFunc,
	}
)

func deleteFunc(ctx context.Context, s *Scope) error {

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return ErrNoRepositoryFound
	}

	dRepo, ok := repo.(Deleter)
	if !ok {
		log.Warningf("Repository for model: '%v' doesn't implement Deleter interface", s.Struct().Type().Name())
		return ErrNoDeleterFound
	}

	// do the delete operation
	if err := dRepo.Delete(ctx, s); err != nil {
		return err
	}

	return nil
}

func beforeDeleteFunc(ctx context.Context, s *Scope) error {
	beforeDeleter, ok := s.Value.(BeforeDeleter)
	if !ok {
		return nil
	}

	if err := beforeDeleter.HBeforeDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func afterDeleteFunc(ctx context.Context, s *Scope) error {
	afterDeleter, ok := s.Value.(AfterDeleter)
	if !ok {
		return nil
	}

	if err := afterDeleter.HAfterDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func deleteForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	iScope := (*scope.Scope)(s)

	// get only relationship fields
	var relationships []*models.StructField
	for _, field := range iScope.Struct().Fields() {
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
			go deleteForeignRelationships(ctx, iScope, field, results)
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

func deleteForeignRelationships(ctx context.Context, iScope *scope.Scope, field *models.StructField, results chan<- interface{}) {
	var err error
	defer func() {
		if err != nil {
			results <- err
		} else {
			results <- struct{}{}
		}
	}()

	rel := field.Relationship()
	switch rel.Kind() {
	case models.RelBelongsTo:
		// belongs to relationship was deleted in the root scope
		return
	case models.RelHasOne, models.RelHasMany:

		// has relationships are deleted by ereasing the
		var isMany bool
		if rel.Kind() == models.RelHasMany {
			isMany = true
		}

		// clearScope clears the foreign key values for the relationships
		clearScope := newScopeWithModel(queryS(iScope).Controller(), rel.Struct(), isMany)

		// the selected field would be only the foreign key -> zero valued
		clearScope.AddSelectedField(rel.ForeignKey())

		for _, prim := range iScope.PrimaryFilters() {
			err = clearScope.AddFilterField(filters.NewFilter(rel.ForeignKey(), prim.Values()...))
			if err != nil {
				log.Debugf("Adding Relationship's foreign key failed: %v", err)
				return
			}
		}

		if tx := queryS(iScope).tx(); tx != nil {
			iScope.AddChainSubscope(clearScope)
			if err = queryS(clearScope).begin(ctx, &tx.Options, false); err != nil {
				log.Debug("BeginTx for the related scope failed: %s", err)
				return
			}
		}

		// patch the clearScope
		if err = queryS(clearScope).PatchContext(ctx); err != nil {
			switch e := err.(type) {
			case *unidb.Error:
				if e.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			return
		}
	case models.RelMany2Many:

		// there is assumption that only the primary filters exists on the delete scope

		// delete the many2many rows within the join model
		clearScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.JoinModel(), false)

		// add the backreference filter
		backReferenceFilter := filters.NewFilter(rel.BackreferenceForeignKey())

		// with the values of the primary filter
		for _, prim := range iScope.PrimaryFilters() {
			backReferenceFilter.AddValues(prim.Values()...)
		}

		if err = clearScope.AddFilterField(backReferenceFilter); err != nil {
			log.Debugf("Deleting relationship: '%s' AddFilterField failed: %v ", field.Name(), err)
			return
		}

		if tx := (*Scope)(iScope).tx(); tx != nil {
			iScope.AddChainSubscope(clearScope)
			if err = (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
				log.Debug("BeginTx for the related scope failed: %s", err)
				return
			}
		}

		// delete the entries in the join model
		if err = (*Scope)(clearScope).DeleteContext(ctx); err != nil {
			switch e := err.(type) {
			case *unidb.Error:
				if e.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			return
		}

	}
}

// reducePrimaryFilters is the process func that changes the delete scope filters so that
// if the root model contains any nonBelongsTo relationship then the filters must be converted into primary field filter
func reducePrimaryFilters(ctx context.Context, s *Scope) error {
	reducedPrimariesInterface, alreadyReduced := s.internal().StoreGet(internal.ReducedPrimariesCtxKey)
	if alreadyReduced {
		previousProcess, ok := s.StoreGet(internal.PreviousProcessCtxKey)
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
			return unidb.ErrInternalError.NewWithError(err)
		}

		if len(primaryValues) > 0 {
			// if there are any primary values in the scope values
			// create a filter field with these values
			primaryFilter := s.internal().GetOrCreateIDFilter()
			if s.internal().IsPrimaryFieldSelected() {
				if err = s.internal().UnselectFields(s.internal().Struct().PrimaryField()); err != nil {
					return unidb.ErrInternalError.NewWithError(err)
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
		return errors.ErrMissingRequiredQueryParam.Copy().WithDetail("No primary field nor primary filters provided while patching the model")
	}

	// if no reduction needed just add the primary filter and deselect the
	if !reduce {
		// just set the primary values in the store and finish
		s.StoreSet(internal.ReducedPrimariesCtxKey, primaries)
		return nil
	}

	primaryScope := NewModelC(s.Controller(), s.Struct(), true)

	// get only the primary field
	primaryScope.internal().SetEmptyFieldset()
	primaryScope.internal().SetFieldsetNoCheck((*models.StructField)(s.Struct().Primary()))

	// set the filters to the primary scope - they would get cleared after
	if err := s.internal().SetFiltersTo(primaryScope.internal()); err != nil {
		return unidb.ErrInternalError.NewWithError(err)
	}

	// list the primary fields
	if err := primaryScope.ListContext(ctx); err != nil {
		return err
	}

	// overwrite the primaries
	primaries, err := primaryScope.internal().GetPrimaryFieldValues()
	if err != nil {
		log.Errorf("Getting primary field values failed: %v", err)
		return unidb.ErrInternalError.NewWithError(err)
	}

	if len(primaries) == 0 {
		return unidb.ErrNoResult.New()
	}

	// clear all filters in the root scope
	s.internal().ClearAllFilters()

	// reduce the filters as the primary filters in the root scope
	primaryFilter := s.internal().GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	s.StoreSet(internal.ReducedPrimariesCtxKey, primaries)

	return nil
}
