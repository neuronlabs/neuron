package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

var (
	// ProcessPatch is the Process that does the repository Patch method.
	ProcessPatch = &Process{
		Name: "neuron:patch",
		Func: patchFunc,
	}

	// ProcessBeforePatch is the Process that does the hook BeforePatch.
	ProcessBeforePatch = &Process{
		Name: "neuron:hook_before_patch",
		Func: beforePatchFunc,
	}

	// ProcessAfterPatch is the Process that does the hook AfterPatch.
	ProcessAfterPatch = &Process{
		Name: "neuron:hook_after_patch",
		Func: afterPatchFunc,
	}

	// ProcessPatchBelongsToRelationships is the process that patches the belongs to relationships.
	ProcessPatchBelongsToRelationships = &Process{
		Name: "neuron:patch_belongs_to_relationships",
		Func: patchBelongsToRelationshipsFunc,
	}

	// ProcessPatchForeignRelationships is the Process that patches the foreign relationships.
	ProcessPatchForeignRelationships = &Process{
		Name: "neuron:patch_foreign_relationships",
		Func: patchForeignRelationshipsFunc,
	}
)

func patchFunc(ctx context.Context, s *Scope) error {
	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return err
	}

	patchRepo, ok := repo.(Patcher)
	if !ok {
		log.Errorf("Repository for current model: '%s' doesn't support Patch method", s.Struct().Type().Name())
		return errors.Newf(class.RepositoryNotImplementsPatcher, "repository: '%T' doesn't implement Patcher interface", repo)
	}

	var onlyForeignRelationships = true

	// if there are any selected fields that are not a foreign relationships
	// (attributes, foreign keys etc, relationship-belongs-to...) do the patch process
	for _, selected := range s.internal().SelectedFields() {
		if !selected.IsRelationship() {
			onlyForeignRelationships = false
			break
		}

		if selected.Relationship().Kind() == models.RelBelongsTo {
			onlyForeignRelationships = false
			break
		}
	}

	if !onlyForeignRelationships {
		log.Debugf("SCOPE[%s] Patch process", s.ID().String())
		if err := patchRepo.Patch(ctx, s); err != nil {
			return err
		}
	}

	return nil
}

func beforePatchFunc(ctx context.Context, s *Scope) error {
	beforePatcher, ok := s.Value.(BeforePatcher)
	if !ok {
		return nil
	}
	if err := beforePatcher.BeforePatch(ctx, s); err != nil {
		log.Debugf("AfterPatcher failed for scope value: %v", s.Value)
		return err
	}

	return nil
}

func afterPatchFunc(ctx context.Context, s *Scope) error {
	afterPatcher, ok := s.Value.(AfterPatcher)
	if !ok {
		return nil
	}

	if err := afterPatcher.AfterPatch(ctx, s); err != nil {
		log.Debugf("AfterPatcher failed for scope value: %v", s.Value)
		return err
	}

	return nil
}

func patchBelongsToRelationshipsFunc(ctx context.Context, s *Scope) error {
	err := (*scope.Scope)(s).SetBelongsToForeignKeyFields()
	if err != nil {
		log.Debugf("[Patch] SetBelongsToForeignKey failed: %v", err)
		return err
	}
	return nil
}

func patchForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	var relationships []*models.StructField

	for _, relField := range s.internal().SelectedFields() {
		if !relField.IsRelationship() {
			continue
		}

		if relField.Relationship().Kind() == models.RelBelongsTo {
			continue
		}

		relationships = append(relationships, relField)
	}

	if len(relationships) > 0 {
		primaryValues, ok := s.internal().StoreGet(internal.ReducedPrimariesStoreKey)
		if !ok {
			err := errors.New(class.InternalQueryNoStoredValue, "no primaries context key set in the store")
			log.Errorf("Scope[%s] %s", s.ID(), err.Error())
			return err
		}

		primaries, ok := primaryValues.([]interface{})
		if !ok {
			err := errors.Newf(class.InternalQueryNoStoredValue, "primaries not of a type []interface{}")
			log.Errorf("Scope[%s]  %s", s.ID(), err.Error())
			return err
		}

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

		ctx, cancelFunc := context.WithTimeout(ctx, maxTimeout)
		defer cancelFunc()

		for _, relField := range relationships {
			switch relField.Relationship().Kind() {
			case models.RelHasOne:
				go patchHasOneRelationship(ctx, s, relField, primaries, results)
			case models.RelHasMany:
				go patchHasManyRelationship(ctx, s, relField, primaries, results)
			case models.RelMany2Many:
				go patchMany2ManyRelationship(ctx, s, relField, primaries, results)
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

func patchHasOneRelationship(
	ctx context.Context, s *Scope, relField *models.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	// if the len primaries are greater than one
	// the foreign key must be in a join model
	if len(primaries) > 1 {
		log.Debugf("SCOPE[%s] Patching multiple primary with HasOne relationship is unsupported.", s.ID())
		err := errors.New(class.QueryValuePrimary, "patching has one relationship for many primaries")
		err.SetDetailf("Patching multiple primary values on the struct's: '%s' relationship: '%s' is not possible.", s.Struct().Collection(), relField.NeuronName())
		results <- err
		return
	}

	// get the reflection value of the scope's value
	v := reflect.ValueOf(s.Value).Elem()

	// get the related field value
	relFieldValue := v.FieldByIndex(relField.ReflectField().Index)

	// set the foreign key
	if relFieldValue.Kind() == reflect.Ptr {
		relFieldValue = relFieldValue.Elem()
	}

	// find the foreign key and set it's value as primary of the root scope
	fkField := relFieldValue.FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index)
	fkField.Set(reflect.ValueOf(primaries[0]))

	relScopeValue := relFieldValue.Addr().Interface()

	// create the relation scope that will patch the relation
	relScope, err := NewC(s.Controller(), relScopeValue)
	if err != nil {
		log.Errorf("Cannot create related scope for HasOne patch relationship: %v of type: %T", err, relScopeValue)
		err = errors.New(class.InternalModelRelationNotMapped, err.Error())
		results <- err
		return
	}

	// the scope should have only primary field and foreign key selected
	relScope.internal().AddSelectedField(relScope.internal().Struct().PrimaryField())
	relScope.internal().AddSelectedField(relField.Relationship().ForeignKey())

	if tx := s.tx(); tx != nil {
		s.internal().AddChainSubscope(relScope.internal())
		if err = relScope.begin(ctx, &tx.Options, false); err != nil {
			results <- err
			return
		}
	}

	// the primary field filter would be added by the process makePrimaryFilters
	if err = relScope.PatchContext(ctx); err != nil {
		log.Debugf("SCOPE[%s] Patching HasOne relationship failed: %v", s.ID(), err)

		if e, ok := err.(*errors.Error); ok {
			if e.Class == class.QueryValueNoResult {
				e.WrapDetailf("Patching relationship: '%s' failed. Related resource not found.", relField.NeuronName())
			}
		}
		results <- err
		return
	}

	// send the struct to the counter
	results <- struct{}{}
	return
}

func patchHasManyRelationship(
	ctx context.Context, s *Scope, relField *models.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	// 1) for related field with values
	// i.e. model with field hasMany = []*relatedModel{{ID: 4},{ID: 5}}
	//
	// for a single primary key:
	// - clear all the relatedModels with the foreign keys equal to the primary (set them to null if possible)
	// - set the realtedModels foreignkeys to primary key where ID in (4, 5)
	//
	//
	// for multiple primaries the relatedModels would not have specified foreign key
	// - throw an error
	//
	// 2) for the empty related field
	// i.e. model with field hasMany = []*relatedModel{}
	//
	// for any primaries length
	// - clear all the relatedModels with the foreign keys equal to any of the primaries values

	rootValue := reflect.ValueOf(s.Value).Elem()
	relatedFieldValue := rootValue.FieldByIndex(relField.ReflectField().Index)

	var isEmpty bool
	if relatedFieldValue.Kind() == reflect.Ptr {
		if relatedFieldValue.IsNil() {
			isEmpty = true
		} else {
			relatedFieldValue = relatedFieldValue.Elem()
		}
	}

	var relatedPrimaries []interface{}
	if !isEmpty {
		for i := 0; i < relatedFieldValue.Len(); i++ {
			single := relatedFieldValue.Index(i)
			if single.Kind() == reflect.Ptr {
				if single.IsNil() {
					continue
				}
				single = single.Elem()
			}

			relatedPrimaryValue := single.FieldByIndex(relField.Relationship().Struct().PrimaryField().ReflectField().Index)
			relatedPrimary := relatedPrimaryValue.Interface()

			if reflect.DeepEqual(relatedPrimary, reflect.Zero(relatedPrimaryValue.Type()).Interface()) {
				continue
			}
			relatedPrimaries = append(relatedPrimaries, relatedPrimary)
		}
	}
	if len(relatedPrimaries) == 0 {
		isEmpty = true
	}

	if !isEmpty {
		// 1) for related field with values

		// check if there are multiple primaries
		if len(primaries) > 1 {
			err := errors.New(class.QueryValuePrimary, "multiple query primaries while patching has many relationship")
			err.SetDetail("Can't patch multiple instances with the relation of type hasMany")
			results <- err
			return
		}

		// clear the related scope - patch all related models with the foreign key filter equal to the given primaries
		if err := patchClearRelationshipWithForeignKey(ctx, s, relField, primaries); err != nil {
			results <- err
			return
		}

		// create the related value for the scope
		relatedValue := relField.Relationship().Struct().NewReflectValueSingle()

		// set the foreign key field
		relatedValue.Elem().FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index).Set(reflect.ValueOf(primaries[0]))

		// create the related scope that patches the relatedPrimaries
		relatedScope, err := NewC(s.Controller(), relatedValue.Interface())
		if err != nil {
			results <- errors.New(class.InternalModelRelationNotMapped, err.Error())
			return
		}

		err = relatedScope.internal().AddFilterField(filters.NewFilter(
			relField.Relationship().Struct().PrimaryField(),
			filters.NewOpValuePair(filters.OpIn, relatedPrimaries...),
		))
		if err != nil {
			results <- err
			return
		}

		if tx := s.tx(); tx != nil {
			s.internal().AddChainSubscope(relatedScope.internal())

			if err := relatedScope.begin(ctx, &tx.Options, false); err != nil {
				results <- err
				return
			}
		}

		// patch the related Scope
		if err := relatedScope.PatchContext(ctx); err != nil {
			if e, ok := err.(*errors.Error); ok {
				if e.Class == class.QueryValueNoResult {
					e.WrapDetailf("Patching related field: '%s' failed - no related resources found", relField.NeuronName())
					results <- e
					return
				}
			}
			results <- err
			return
		}
	} else {
		// 2) for the empty related field
		// clear the related scope
		if err := patchClearRelationshipWithForeignKey(ctx, s, relField, primaries); err != nil {
			results <- err
			return
		}
	}

	results <- struct{}{}
}

func patchMany2ManyRelationship(
	ctx context.Context, s *Scope, relField *models.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	// by patching the many2many relationship the join model entries are required to be patched
	//
	// if the patching is used to clear the relationships (the relationship field value is zero)
	// remove all the entries in the join model where the primary fields are in the backreference foreign key field
	//
	// if the patched relationship field contains any values
	// then all the current join model entries should be clear
	// and new entries from the model's relationship value should be inserted
	// (REPLACE relationship process)

	// get relatedFieldValue
	rootValue := reflect.ValueOf(s.Value).Elem()
	relatedFieldValue := rootValue.FieldByIndex(relField.ReflectField().Index)

	var isEmpty bool
	if relatedFieldValue.Kind() == reflect.Ptr {
		if relatedFieldValue.IsNil() {
			isEmpty = true
		} else {
			relatedFieldValue = relatedFieldValue.Elem()
		}
	}

	var relatedPrimaries []interface{}
	if !isEmpty {
		for i := 0; i < relatedFieldValue.Len(); i++ {
			single := relatedFieldValue.Index(i)
			if single.Kind() == reflect.Ptr {
				if single.IsNil() {
					continue
				}
				single = single.Elem()
			}

			relatedPrimaryValue := single.FieldByIndex(relField.Relationship().Struct().PrimaryField().ReflectField().Index)
			relatedPrimary := relatedPrimaryValue.Interface()

			if reflect.DeepEqual(relatedPrimary, reflect.Zero(relatedPrimaryValue.Type()).Interface()) {
				continue
			}
			relatedPrimaries = append(relatedPrimaries, relatedPrimary)
		}
	}
	if len(relatedPrimaries) == 0 {
		isEmpty = true
	}

	if isEmpty {
		// create the scope for clearing the relationship
		clearScope := newScopeWithModel(s.Controller(), relField.Relationship().JoinModel(), false)

		f := filters.NewFilter(relField.Relationship().ForeignKey(), filters.NewOpValuePair(filters.OpIn, primaries...))
		err := clearScope.AddFilterField(f)
		if err != nil {
			results <- err
			return
		}

		if tx := s.tx(); tx != nil {
			s.internal().AddChainSubscope(clearScope)
			if err = queryS(clearScope).begin(ctx, &tx.Options, false); err != nil {
				log.Debug("BeginTx for the related scope failed: %s", err)
				results <- err
				return
			}
		}

		// delete the rows in the many2many relationship containing provided values
		if err = queryS(clearScope).DeleteContext(ctx); err != nil {
			switch t := err.(type) {
			case *errors.Error:
				if t.Class == class.QueryValueNoResult {
					err = nil
				}
			}

			if err != nil {
				results <- err
				return
			}
		}

		results <- struct{}{}
		return
	}

	// 1) clear the join model values
	//
	// create the clearScope based on the join model
	clearScope := newScopeWithModel(s.Controller(), relField.Relationship().JoinModel(), false)

	// it also should get all 'old' entries matched with root.primary
	// value as the backreference primary field
	// copy the root scope primary filters into backreference foreign key
	f := filters.NewFilter(relField.Relationship().ForeignKey(), filters.NewOpValuePair(filters.OpIn, primaries...))
	if err := clearScope.AddFilterField(f); err != nil {
		results <- err
		return
	}

	if tx := s.tx(); tx != nil {
		s.internal().AddChainSubscope(clearScope)
		if err := queryS(clearScope).begin(ctx, &tx.Options, false); err != nil {
			log.Debug("BeginTx for the related scope failed: %s", err)
			results <- err
			return
		}
	}

	// delete the entries in the join model
	if err := queryS(clearScope).DeleteContext(ctx); err != nil {
		if e, ok := err.(*errors.Error); ok {
			// if the err is no result
			if e.Class == class.QueryValueNoResult {
				err = nil
			}
		}

		if err != nil {
			results <- err
			return
		}
	}

	// 2) check if the related fields exists in the relation model
	// then if correct insert the new entries into the the join model

	// get the primary field values from the relationship field
	checkScope := newScopeWithModel(s.Controller(), relField.Relationship().Struct(), true)

	// we would need only primary fields
	checkScope.SetEmptyFieldset()
	checkScope.SetFieldsetNoCheck(relField.Relationship().Struct().PrimaryField())

	if tx := s.tx(); tx != nil {
		s.internal().AddChainSubscope(checkScope)
		if err := queryS(checkScope).begin(ctx, &tx.Options, false); err != nil {
			log.Debug("BeginTx for the related scope failed: %s", err)
			results <- err
			return
		}
	}

	// get the values from the checkScope
	if err := queryS(checkScope).ListContext(ctx); err != nil {
		if e, ok := err.(*errors.Error); ok {
			if e.Class == class.QueryValueNoResult {
				e.WrapDetail("No many2many relationship values found with the provided ids")
			}
		}
		results <- err
		return
	}

	var primaryMap = map[interface{}]bool{}

	// set the primaryMap with the relation primaries
	for _, primary := range relatedPrimaries {
		primaryMap[primary] = false
	}

	checkPrimaries, err := checkScope.GetPrimaryFieldValues()
	if err != nil {
		results <- err
		return
	}

	for _, primary := range checkPrimaries {
		primaryMap[primary] = true
	}

	var nonExistsPrimaries []interface{}

	// check if all the primary values were in the check scope
	for primary, exists := range primaryMap {
		if !exists {
			nonExistsPrimaries = append(nonExistsPrimaries, primary)
		}
	}

	if len(nonExistsPrimaries) > 0 {
		// violation integrity constraint erorr
		err = errors.New(class.QueryViolationIntegrityConstraint, "relationship values doesn't exists").SetDetailf("Patching relationship field: '%s' failed. The relationships: %v doesn't exists", relField.NeuronName(), nonExistsPrimaries)
		results <- err
		return
	}

	// clear the memory from slices and maps
	checkPrimaries = nil
	primaryMap = nil

	// 3) when all the primaries exists in the relation model
	// insert the relation values into the join model
	instances := relField.Relationship().JoinModel().NewReflectValueMany()
	dereferencedInstances := instances.Elem()

	// create multiple instances of join models
	for _, primary := range primaries {
		for _, relPrimary := range relatedPrimaries {
			joinModelInstance := relField.Relationship().JoinModel().NewReflectValueSingle()

			// get the join model foreign key
			mtmFK := joinModelInstance.Elem().FieldByIndex(relField.Relationship().ManyToManyForeignKey().ReflectField().Index)
			// set it's value to primary value
			mtmFK.Set(reflect.ValueOf(relPrimary))

			// get the join model backreference key and set it with primary value
			fk := joinModelInstance.Elem().FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index)
			fk.Set(reflect.ValueOf(primary))
			dereferencedInstances.Set(reflect.Append(dereferencedInstances, joinModelInstance))
		}
	}

	insertScope, err := NewC(s.Controller(), instances.Interface())
	if err != nil {
		results <- err
		return
	}

	insertScope.internal().AddSelectedField(relField.Relationship().ManyToManyForeignKey())
	insertScope.internal().AddSelectedField(relField.Relationship().ForeignKey())

	// create the newly created relationships
	if err = insertScope.CreateContext(ctx); err != nil {
		results <- err
		return
	}

	results <- struct{}{}
}

func patchClearRelationshipWithForeignKey(ctx context.Context, s *Scope, relField *models.StructField, primaries []interface{}) error {
	// create clearScope for the relation field's model
	clearScope := newScopeWithModel(s.Controller(), relField.Relationship().Struct(), false)

	// add selected field into the clear scope
	clearScope.AddSelectedField(relField.Relationship().ForeignKey())

	err := clearScope.AddFilterField(
		filters.NewFilter(
			relField.Relationship().ForeignKey(),
			filters.NewOpValuePair(filters.OpIn, primaries...),
		),
	)
	if err != nil {
		return err
	}

	// if the root scope has a transaction 'on' begin the transaction on the clearScope
	if tx := s.tx(); tx != nil {
		s.internal().AddChainSubscope(clearScope)
		if err := (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
			return err
		}
	}

	// clear the related scope
	if err = queryS(clearScope).PatchContext(ctx); err != nil {
		if e, ok := err.(*errors.Error); ok {
			// if the error is no value result clear the error
			if e.Class == class.QueryValueNoResult {
				err = nil
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}
