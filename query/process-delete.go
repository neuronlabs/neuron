package query

import (
	"context"
	"reflect"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal"
)

func deleteFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return err
	}

	_, hasDeletedAt := s.Struct().DeletedAt()
	if hasDeletedAt {
		patcher, ok := repo.(Patcher)
		if !ok {
			log.Warningf("Repository for model: '%s' doesn't implement Patcher interface", s.Struct().Type())
			return errors.NewDetf(class.RepositoryNotImplementsPatcher, "repository: '%T' doesn't implement Patcher interface", repo)
		}

		if log.Level().IsAllowed(log.LDEBUG3) {
			log.Debug3f("SCOPE[%s][%s] Deleting by patching the 'DeletedAt' field: %s", s.ID().String(), s.Struct().Collection(), s.String())
		}
		if err = patcher.Patch(ctx, s); err != nil {
			return err
		}
		return nil
	}

	dRepo, ok := repo.(Deleter)
	if !ok {
		log.Warningf("Repository for model: '%v' doesn't implement Deleter interface", s.Struct().Type().Name())
		return errors.NewDetf(class.RepositoryNotImplementsDeleter, "repository: %T doesn't implement Deleter interface", repo)
	}

	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s][%s] deleting: %s", s.ID().String(), s.Struct().Collection(), s.String())
	}
	// do the delete operation
	if err := dRepo.Delete(ctx, s); err != nil {
		return err
	}
	return nil
}

func beforeDeleteFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
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
	if s.Error != nil {
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
	if s.Error != nil {
		return nil
	}

	// get only relationship fields
	var relationships []*mapping.StructField
	for _, field := range s.Struct().RelationFields() {
		if field.Relationship().Kind() != mapping.RelBelongsTo {
			relationships = append(relationships, field)
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
			case mapping.RelHasOne:
				go deleteHasOneRelationshipsChan(ctx, s, field, results)
			case mapping.RelHasMany:
				go deleteHasManyRelationshipsChan(ctx, s, field, results)
			case mapping.RelMany2Many:
				go deleteMany2ManyRelationshipsChan(ctx, s, field, results)
			case mapping.RelBelongsTo:
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
	if s.Error != nil {
		return nil
	}

	// get only relationship fields
	var relationships []*mapping.StructField
	for _, field := range s.Struct().RelationFields() {
		if field.Relationship().Kind() != mapping.RelBelongsTo {
			relationships = append(relationships, field)
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
			case mapping.RelHasOne:
				err = deleteHasOneRelationships(ctx, s, field)
			case mapping.RelHasMany:
				err = deleteHasManyRelationships(ctx, s, field)
			case mapping.RelMany2Many:
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

func deleteHasOneRelationshipsChan(ctx context.Context, s *Scope, field *mapping.StructField, results chan<- interface{}) {
	if err := deleteHasOneRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteHasOneRelationships(ctx context.Context, s *Scope, field *mapping.StructField) error {
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
		clearScope = newScopeWithModel(s.Controller(), rel.Struct(), false)
	}

	// the selected field would be only the foreign key -> zero valued
	clearScope.Fieldset[rel.ForeignKey().NeuronName()] = rel.ForeignKey()
	for _, prim := range s.PrimaryFilters {
		clearScope.ForeignFilters = append(clearScope.ForeignFilters, &FilterField{StructField: rel.ForeignKey(), Values: prim.Values})
	}

	// patch the clearScope
	if err = clearScope.PatchContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			if e.Class() == class.QueryValueNoResult {
				err = nil
			}
		}
	}
	return err
}

func deleteHasManyRelationshipsChan(ctx context.Context, s *Scope, field *mapping.StructField, results chan<- interface{}) {
	if err := deleteHasManyRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteHasManyRelationships(ctx context.Context, s *Scope, field *mapping.StructField) error {
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
		clearScope = newScopeWithModel(s.Controller(), rel.Struct(), false)
	}

	// the selected field would be only the foreign key -> zero valued
	clearScope.Fieldset[rel.ForeignKey().NeuronName()] = rel.ForeignKey()

	for _, prim := range s.PrimaryFilters {
		clearScope.ForeignFilters = append(clearScope.ForeignFilters, &FilterField{StructField: rel.ForeignKey(), Values: prim.Values})
	}

	// patch the clearScope
	err = clearScope.PatchContext(ctx)
	if err != nil {
		e, ok := err.(errors.ClassError)
		if ok && e.Class() == class.QueryValueNoResult {
			err = nil
		}
	}
	return err
}

func deleteMany2ManyRelationshipsChan(ctx context.Context, s *Scope, field *mapping.StructField, results chan<- interface{}) {
	if err := deleteMany2ManyRelationships(ctx, s, field); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func deleteMany2ManyRelationships(ctx context.Context, s *Scope, field *mapping.StructField) error {
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
		clearScope = newScopeWithModel(s.Controller(), rel.JoinModel(), false)
	}

	// add the backreference filter
	foreignKeyFilter := &FilterField{StructField: rel.ForeignKey()}

	// with the values of the primary filter
	for _, prim := range s.PrimaryFilters {
		foreignKeyFilter.Values = append(foreignKeyFilter.Values, prim.Values...)
	}
	clearScope.ForeignFilters = append(clearScope.ForeignFilters, foreignKeyFilter)

	// delete the entries in the join model
	err = clearScope.DeleteContext(ctx)
	if err != nil {
		e, ok := err.(errors.ClassError)
		if ok && e.Class() == class.QueryValueNoResult {
			err = nil
		}
	}
	return err
}

// reducePrimaryFilters is the process func that changes the delete scope filters so that
// if the root model contains any nonBelongsTo relationship then the filters must be converted into primary field filter
func reducePrimaryFilters(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	reducedPrimariesInterface, alreadyReduced := s.StoreGet(internal.ReducedPrimariesStoreKey)
	if alreadyReduced {
		switch s.processMethod {
		case pmDelete:
			_, isBeforeDeleter := s.Value.(BeforeDeleter)
			if !isBeforeDeleter {
				return nil
			}
		case pmPatch:
			_, isBeforePatcher := s.Value.(BeforePatcher)
			if !isBeforePatcher {
				return nil
			}
		}
	}
	var reduce bool
	// get the primary field values from the scope's value
	if s.Value != nil {
		primaryValues, err := mapping.PrimaryValues(s.Struct(), reflect.ValueOf(s.Value))
		if err != nil {
			return err
		}

		if len(primaryValues) > 0 {
			// if there are any primary values in the scope values
			// create a filter field with these values
			primaryFilter := s.getOrCreatePrimaryFilter()
			_, ok := s.Fieldset[s.Struct().Primary().NeuronName()]
			if ok {
				// remove primary key from fieldset
				delete(s.Fieldset, s.Struct().Primary().NeuronName())
			}
			primaryFilter.Values = append(primaryFilter.Values, &OperatorValues{Operator: OpIn, Values: primaryValues})
		}
	}
	var primaries []interface{}
	// iterate over all primary filters
	// get the primary filter values
	// if the joinOperatorValues is true change the OpEqual into OpIn
	// if the operator is different than OpIn and OpEqual
	// then the reduction is necessary
	for _, pk := range s.PrimaryFilters {
		for _, fv := range pk.Values {
			switch fv.Operator {
			case OpIn:
				primaries = append(primaries, fv.Values...)
			case OpEqual:
				fv.Operator = OpIn
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
		if len(s.AttributeFilters) != 0 || len(s.RelationFilters) != 0 || s.LanguageFilters != nil || len(s.ForeignFilters) != 0 {
			reduce = true
		}
	}

	if len(primaries) == 0 && !reduce {
		err := errors.NewDet(class.QueryFilterMissingRequired, "no primary filter or values provided")
		err.SetDetails("No primary field nor primary filters provided while patching the model")
		return err
	}

	// if no reduction needed just add the primary filter and deselect the
	if !reduce {
		// just set the primary values in the store and finish
		s.StoreSet(internal.ReducedPrimariesStoreKey, primaries)
		return nil
	}

	primaryScope := NewModelC(s.Controller(), s.Struct(), true)

	// get only the primary field
	primaryScope.Fieldset = map[string]*mapping.StructField{"id": primaryScope.Struct().Primary()}
	// set the filters to the primary scope - they would get cleared after
	if err := s.setFiltersTo(primaryScope); err != nil {
		return err
	}

	// list the primary fields
	if err := primaryScope.ListContext(ctx); err != nil {
		return err
	}

	// overwrite the primaries
	primaries, err := primaryScope.getPrimaryFieldValues()
	if err != nil {
		log.Errorf("Getting primary field values failed: %v", err)
		return err
	}

	if len(primaries) == 0 {
		return errors.NewDet(class.QueryValueNoResult, "query value no results")
	}

	// clear all filters in the root scope
	s.clearFilters()

	// remove primary key from fieldset if exists
	delete(s.Fieldset, s.Struct().Primary().NeuronName())

	// reduce the filters as the primary filters in the root scope
	primaryFilter := s.getOrCreatePrimaryFilter()
	primaryFilter.Values = append(primaryFilter.Values, &OperatorValues{Operator: OpIn, Values: primaries})

	s.StoreSet(internal.ReducedPrimariesStoreKey, primaries)
	s.StoreSet(internal.PrimariesAlreadyChecked, struct{}{})

	return nil
}

func setDeletedAtField(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}
	deletedAt, hasDeletedAt := s.Struct().DeletedAt()
	if !hasDeletedAt {
		return nil
	}

	v := reflect.ValueOf(s.Value).Elem().FieldByIndex(deletedAt.ReflectField().Index)
	t := time.Now()
	v.Set(reflect.ValueOf(&t))

	s.Fieldset[deletedAt.NeuronName()] = deletedAt
	return nil
}
