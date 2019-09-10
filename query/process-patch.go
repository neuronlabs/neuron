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

func patchFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return err
	}

	patchRepo, ok := repo.(Patcher)
	if !ok {
		log.Errorf("Repository for current model: '%s' doesn't support Patch method", s.Struct().Type().Name())
		return errors.NewDetf(class.RepositoryNotImplementsPatcher, "repository: '%T' doesn't implement Patcher interface", repo)
	}

	updatedAt, hasUpdatedAt := s.Struct().UpdatedAt()
	// if there are any selected fields that are not a foreign relationships
	// (attributes, foreign keys etc, relationship-belongs-to...) do the patch process
	if !hasUpdatedAt && (len(s.SelectedFields) == 0 || s.SelectedFields == nil) {
		return errors.NewDet(class.QuerySelectedFieldsNotSelected, "no fields selected for patch process")
	}

	onlyForeignRelationships := true
	var updatedAtSelected bool
	for _, selected := range s.SelectedFields {
		if selected.IsPrimary() {
			if len(s.SelectedFields) == 1 {
				if !hasUpdatedAt {
					return nil
				}
				onlyForeignRelationships = false
			}
			continue
		}

		if !selected.IsRelationship() {
			onlyForeignRelationships = false
		} else if selected.Relationship().Kind() == mapping.RelBelongsTo {
			onlyForeignRelationships = false
		}

		if hasUpdatedAt && selected == updatedAt {
			updatedAtSelected = true
		}
	}

	if !onlyForeignRelationships || hasUpdatedAt {
		if hasUpdatedAt {
			v := reflect.ValueOf(s.Value).Elem().FieldByIndex(updatedAt.ReflectField().Index)

			var setUpdatedAt bool
			if s.autoSelectedFields {
				setUpdatedAt = reflect.DeepEqual(v.Interface(), reflect.Zero(updatedAt.ReflectField().Type).Interface())
			} else {
				setUpdatedAt = !updatedAtSelected
			}
			if setUpdatedAt {
				t := time.Now()
				switch {
				case updatedAt.IsTimePointer():
					v.Set(reflect.ValueOf(&t))
				case updatedAt.IsTime():
					v.Set(reflect.ValueOf(t))
				}
				if !updatedAtSelected {
					s.SelectedFields = append(s.SelectedFields, updatedAt)
				}
			}
		}
		if log.Level().IsAllowed(log.LDEBUG3) {
			log.Debug3f("SCOPE[%s][%s] patching: %s", s.ID().String(), s.Struct().Collection(), s.String())
		}
		if err := patchRepo.Patch(ctx, s); err != nil {
			return err
		}
		return nil
	}

	// check if the primaries were already checked
	_, ok = s.StoreGet(internal.PrimariesAlreadyChecked)
	if ok {
		return nil
	}

	var listScope *Scope
	if tx := s.Tx(); tx != nil {
		listScope, err = tx.NewModelC(s.Controller(), s.Struct(), true)
		if err != nil {
			return err
		}
	} else {
		listScope = newScopeWithModel(s.Controller(), s.Struct(), true)
	}

	if err = s.setFiltersTo(listScope); err != nil {
		return err
	}
	listScope.Fieldset = map[string]*mapping.StructField{"id": listScope.Struct().Primary()}

	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s][%s] check patch with the list scope: '%s'", s.ID().String(), s.Struct().Collection(), listScope.String())
	}
	if err = listScope.ListContext(ctx); err != nil {
		return err
	}

	return nil
}

func beforePatchFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

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
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

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
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}
	err := s.setBelongsToForeignKeyFields()
	if err != nil {
		log.Debugf("[Patch] SetBelongsToForeignKey failed: %v", err)
		return err
	}
	return nil
}

func patchForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	var relationships []*mapping.StructField
	for _, relField := range s.SelectedFields {
		if !relField.IsRelationship() {
			continue
		}
		if relField.Relationship().Kind() == mapping.RelBelongsTo {
			continue
		}
		relationships = append(relationships, relField)
	}

	if len(relationships) == 0 {
		return nil
	}
	primaryValues, ok := s.StoreGet(internal.ReducedPrimariesStoreKey)
	if !ok {
		err := errors.NewDet(class.InternalQueryNoStoredValue, "no primaries context key set in the store")
		log.Errorf("Scope[%s] %s", s.ID(), err.Error())
		return err
	}

	primaries, ok := primaryValues.([]interface{})
	if !ok {
		err := errors.NewDetf(class.InternalQueryNoStoredValue, "primaries not of a type []interface{}")
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
		case mapping.RelHasOne:
			go patchHasOneRelationshipChan(ctx, s, relField, primaries, results)
		case mapping.RelHasMany:
			go patchHasManyRelationshipChan(ctx, s, relField, primaries, results)
		case mapping.RelMany2Many:
			go patchMany2ManyRelationshipChan(ctx, s, relField, primaries, results)
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

func patchForeignRelationshipsSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	var relationships []*mapping.StructField
	for _, relField := range s.SelectedFields {
		if !relField.IsRelationship() {
			continue
		}

		if relField.Relationship().Kind() == mapping.RelBelongsTo {
			continue
		}
		relationships = append(relationships, relField)
	}

	if len(relationships) > 0 {
		primaryValues, ok := s.StoreGet(internal.ReducedPrimariesStoreKey)
		if !ok {
			err := errors.NewDet(class.InternalQueryNoStoredValue, "no primaries context key set in the store")
			log.Errorf("Scope[%s] %s", s.ID(), err.Error())
			return err
		}

		primaries, ok := primaryValues.([]interface{})
		if !ok {
			err := errors.NewDetf(class.InternalQueryNoStoredValue, "primaries not of a type []interface{}")
			log.Errorf("Scope[%s]  %s", s.ID(), err.Error())
			return err
		}

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

		var err error
		for _, relField := range relationships {
			switch relField.Relationship().Kind() {
			case mapping.RelHasOne:
				err = patchHasOneRelationship(ctx, s, relField, primaries)
			case mapping.RelHasMany:
				err = patchHasManyRelationship(ctx, s, relField, primaries)
			case mapping.RelMany2Many:
				err = patchMany2ManyRelationship(ctx, s, relField, primaries)
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

func patchHasOneRelationshipChan(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	if err := patchHasOneRelationship(ctx, s, relField, primaries); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func patchHasOneRelationship(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{},
) error {
	var err error
	// if the len primaries are greater than one
	// the foreign key must be in a join model
	if len(primaries) > 1 {
		log.Debugf("SCOPE[%s] Patching multiple primary with HasOne relationship is unsupported.", s.ID())
		err := errors.NewDet(class.QueryValuePrimary, "patching has one relationship for many primaries")
		err.SetDetailsf("Patching multiple primary values on the struct's: '%s' relationship: '%s' is not possible.", s.Struct().Collection(), relField.NeuronName())
		return err
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

	var relScope *Scope
	if tx := s.Tx(); tx != nil {
		relScope, err = tx.NewContextC(ctx, s.Controller(), relScopeValue)
		if err != nil {
			return err
		}
	} else {
		relScope, err = NewC(s.Controller(), relScopeValue)
		// create the relation scope that will patch the relation
		if err != nil {
			log.Errorf("Cannot create related scope for HasOne patch relationship: %v of type: %T", err, relScopeValue)
			err = errors.NewDet(class.InternalModelRelationNotMapped, err.Error())
			return err
		}
	}

	// the scope should have only primary field and foreign key selected
	relScope.SelectedFields = append(relScope.SelectedFields, relScope.Struct().Primary(), relField.Relationship().ForeignKey())

	log.Debug2f("SCOPE[%s] Patching foreign HasOne relationship: '%s'", s.ID(), relField.NeuronName())
	// the primary field filter would be added by the process makePrimaryFilters
	if err = relScope.PatchContext(ctx); err != nil {
		log.Debugf("SCOPE[%s] Patching HasOne relationship failed: %v", s.ID(), err)

		if e, ok := err.(errors.DetailedError); ok {
			// change the class of the error
			if e.Class() == class.QueryValueNoResult {
				e.WrapDetailsf("Patching relationship: '%s' failed. Related resource not found.", relField.NeuronName())
				err = errors.NewDet(class.QueryValueNoResult, e.Error())
			} else {
				err = errors.NewDet(class.QueryRelation, e.Error())
			}
		}
		return err
	}
	// send the struct to the counter
	return nil
}

func patchHasManyRelationshipChan(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	if err := patchHasManyRelationship(ctx, s, relField, primaries); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}

}

func patchHasManyRelationship(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{},
) error {
	var err error
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

			relatedPrimaryValue := single.FieldByIndex(relField.Relationship().Struct().Primary().ReflectField().Index)
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
		// 2) for the empty related field
		// clear the related scope
		log.Debug2f("SCOPE[%s] Patch HasMany relationship: '%s' clear relationships.", s.ID(), relField.NeuronName())
		if err = patchClearRelationshipWithForeignKey(ctx, s, relField, primaries); err != nil {
			return err
		}
		return nil
	}

	// 1) for related field with values
	// check if there are multiple primaries
	if len(primaries) > 1 {
		err := errors.NewDet(class.QueryValuePrimary, "multiple query primaries while patching has many relationship")
		err.SetDetails("Can't patch multiple instances with the relation of type hasMany")
		return err
	}

	var clearScope *Scope
	// clear the related scope - patch all related models with the foreign key filter equal to the given primaries
	// create clearScope for the relation field's model
	if tx := s.Tx(); tx != nil {
		clearScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().Struct(), false)
		if err != nil {
			return err
		}
	} else {
		clearScope = newScopeWithModel(s.Controller(), relField.Relationship().Struct(), false)
		if _, err = clearScope.BeginTx(ctx, nil); err != nil {
			return err
		}
	}

	// add selected field into the clear scope
	clearScope.SelectedFields = append(clearScope.SelectedFields, relField.Relationship().ForeignKey())
	// add the foreign key filter
	clearScope.ForeignFilters = append(clearScope.ForeignFilters, NewFilter(relField.Relationship().ForeignKey(), OpIn, primaries...))

	log.Debug2f("SCOPE[%s] Patch relationship: '%s' - clear current relationship values", s.ID(), relField.NeuronName())
	// clear the related scope.
	if err = clearScope.PatchContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			// if the error is no value result clear the error
			if e.Class() == class.QueryValueNoResult {
				err = nil
			}
		}
		if err != nil {
			if tx := s.Tx(); tx == nil {
				err = clearScope.RollbackContext(ctx)
			}
			return err
		}
	}

	// create the related value for the scope
	relatedValue := mapping.NewReflectValueSingle(relField.Relationship().Struct())
	// set the foreign key field
	relatedValue.Elem().FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index).Set(reflect.ValueOf(primaries[0]))
	// create the related scope that patches the relatedPrimaries
	var relatedScope *Scope

	if tx := s.Tx(); tx != nil {
		log.Debug3f("SCOPE[%s][%s] Creating patch relationship: '%s' scope with tx:'%s'", s.ID(), s.Struct().Collection(), relField.NeuronName(), tx.ID)
		relatedScope, err = tx.NewContextC(ctx, s.Controller(), relatedValue.Interface())
		if err != nil {
			err = errors.NewDet(class.InternalModelRelationNotMapped, err.Error())
			return err
		}
	} else {
		tx := clearScope.Tx()
		relatedScope, err = tx.NewContextC(ctx, s.Controller(), relatedValue.Interface())
		if err != nil {
			if err := clearScope.RollbackContext(ctx); err != nil {
				return err
			}
			err = errors.NewDet(class.InternalModelRelationNotMapped, err.Error())
			return err
		}
	}
	relatedScope.PrimaryFilters = append(relatedScope.PrimaryFilters, NewFilter(relField.Relationship().Struct().Primary(), OpIn, relatedPrimaries...))

	log.Debug2f("SCOPE[%s][%s] Patch HasMany relationship: '%s'", s.ID(), s.Struct().Collection(), relField.NeuronName())
	// patch the related Scope
	if err := relatedScope.PatchContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			if e.Class() == class.QueryValueNoResult {
				e, ok := err.(errors.DetailedError)
				if ok {
					e.WrapDetailsf("Patching related field: '%s' failed - no related resources found", relField.NeuronName())
				}
				return e
			}
		}
		if tx := s.Tx(); tx == nil {
			if err := clearScope.RollbackContext(ctx); err != nil {
				log.Errorf("Rollback failed: %v", err.Error())
			}
		}
		return err
	}

	if tx := s.Tx(); tx == nil {
		if err = clearScope.CommitContext(ctx); err != nil {
			return err
		}
	}
	return nil
}

func patchMany2ManyRelationshipChan(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{}, results chan<- interface{},
) {
	if err := patchMany2ManyRelationship(ctx, s, relField, primaries); err != nil {
		results <- err
	} else {
		results <- struct{}{}
	}
}

func patchMany2ManyRelationship(
	ctx context.Context, s *Scope, relField *mapping.StructField,
	primaries []interface{},
) error {
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
	var err error
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

			relatedPrimaryValue := single.FieldByIndex(relField.Relationship().Struct().Primary().ReflectField().Index)
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

	// justCreated is a flag that this is a part of Create Process
	_, justCreated := s.StoreGet(internal.JustCreated)

	if isEmpty && !justCreated {
		var (
			clearScope *Scope
			rootTx     bool
		)
		// create the scope for clearing the relationship
		if tx := s.tx(); tx != nil {
			rootTx = true
			clearScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().JoinModel(), false)
			if err != nil {
				return err
			}
		} else {
			clearScope = newScopeWithModel(s.Controller(), relField.Relationship().JoinModel(), false)
			if _, err = clearScope.begin(ctx, nil, false); err != nil {
				return err
			}
		}
		clearScope.ForeignFilters = append(clearScope.ForeignFilters, NewFilter(relField.Relationship().ForeignKey(), OpIn, primaries...))

		// delete the rows in the many2many relationship containing provided values
		if err = clearScope.DeleteContext(ctx); err != nil {
			switch t := err.(type) {
			case errors.ClassError:
				if t.Class() == class.QueryValueNoResult {
					err = nil
				}
			}

			if err != nil {
				if !rootTx {
					if err := clearScope.RollbackContext(ctx); err != nil {
						log.Debugf("[SCOPE][%s]Rollback failed for clearScope: %s", clearScope.ID(), err)
					}
				}
				return err
			}
		}
		return nil
	} else if isEmpty && justCreated {
		return nil
	}

	// 1) clear the join model values
	// create the clearScope based on the join model
	var clearScope *Scope
	var rootTx bool

	if !justCreated {
		if tx := s.Tx(); tx != nil {
			rootTx = true
			clearScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().JoinModel(), false)
			if err != nil {
				return err
			}
		} else {
			clearScope = newScopeWithModel(s.Controller(), relField.Relationship().JoinModel(), false)
			if _, err = clearScope.begin(ctx, nil, false); err != nil {
				return err
			}
		}

		// it also should get all 'old' entries matched with root.primary
		// value as the backreference primary field
		// copy the root scope primary filters into backreference foreign key
		clearScope.ForeignFilters = append(clearScope.ForeignFilters, NewFilter(relField.Relationship().ForeignKey(), OpIn, primaries...))

		// delete the entries in the join model
		if err := clearScope.DeleteContext(ctx); err != nil {
			if e, ok := err.(errors.ClassError); ok {
				// if the err is no result
				if e.Class() == class.QueryValueNoResult {
					err = nil
				}
			}
			if err != nil {
				if rootTx {
					if err := clearScope.RollbackContext(ctx); err != nil {
						log.Errorf("[SCOPE][%s] clearScope failed to rollback: %v", clearScope.ID(), err)
					}
				}
				return err
			}
		}
	}

	// 2) check if the related fields exists in the relation model
	// then if correct insert the new entries into the the join model

	// get the primary field values from the relationship field
	var checkScope *Scope

	if tx := s.Tx(); tx != nil {
		checkScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().Struct(), true)
		if err != nil {
			return err
		}
	} else if !justCreated {
		if tx := clearScope.Tx(); tx != nil {
			checkScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().Struct(), true)
			if err != nil {
				return err
			}
		}
	} else {
		checkScope = newScopeWithModel(s.Controller(), relField.Relationship().Struct(), true)
	}

	// we need only primary fields
	checkScope.Fieldset = map[string]*mapping.StructField{"id": checkScope.Struct().Primary()}
	checkScope.PrimaryFilters = append(checkScope.PrimaryFilters, NewFilter(relField.Relationship().Struct().Primary(), OpIn, relatedPrimaries...))

	log.Debug3f("SCOPE[%s][%s] checking many2many field: '%s' related values in scope: '%s'", s.ID(), s.Struct().Collection(), relField.NeuronName(), checkScope.ID())
	// get the values from the checkScope
	err = checkScope.ListContext(ctx)
	if err != nil {
		if e, ok := err.(errors.ClassError); ok {
			if e.Class() == class.QueryValueNoResult {
				if e, ok := err.(errors.DetailedError); ok {
					e.WrapDetails("No many2many relationship values found with the provided ids")
				}
			}
			err = e
		}
		if !justCreated && !rootTx {
			err := clearScope.RollbackContext(ctx)
			if err != nil {
				return err
			}
		}
		return err
	}

	var primaryMap = map[interface{}]bool{}
	// set the primaryMap with the relation primaries
	for _, primary := range relatedPrimaries {
		primaryMap[primary] = false
	}

	checkPrimaries, err := checkScope.getPrimaryFieldValues()
	if err != nil {
		if !justCreated && !rootTx {
			err := clearScope.RollbackContext(ctx)
			if err != nil {
				return err
			}
		}
		return err
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
		if !justCreated && !rootTx {
			err := clearScope.RollbackContext(ctx)
			if err != nil {
				return err
			}
		}
		err := errors.NewDet(class.QueryViolationIntegrityConstraint, "relationship values doesn't exists")
		err.SetDetailsf("Patching relationship field: '%s' failed. The relationships: %v doesn't exists", relField.NeuronName(), nonExistsPrimaries)
		return err
	}

	// 3) when all the primaries exists in the relation model
	// insert the relation values into the join model
	joinModel := relField.Relationship().JoinModel()

	// create multiple instances of join models
	// TODO: change to CreateMany if implemented.
	for _, primary := range primaries {
		single := mapping.NewReflectValueSingle(joinModel)
		for _, relPrimary := range relatedPrimaries {
			// get the join model foreign key
			mtmFK := single.Elem().FieldByIndex(relField.Relationship().ManyToManyForeignKey().ReflectField().Index)
			// set it's value to primary value
			mtmFK.Set(reflect.ValueOf(relPrimary))

			// get the join model backreference key and set it with primary value
			fk := single.Elem().FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index)
			fk.Set(reflect.ValueOf(primary))
		}

		var insertScope *Scope
		if tx := s.Tx(); tx != nil {
			insertScope, err = tx.NewC(s.Controller(), single.Interface())

		} else if !justCreated {
			insertScope, err = clearScope.Tx().NewC(s.Controller(), single.Interface())
		} else {
			insertScope, err = NewC(s.Controller(), single.Interface())
		}

		if err != nil {
			log.Debugf("SCOPE[%s][%s] Many2Many Relationship InsertScope failed: %v", s.ID(), s.Struct().Collection(), err)
			if !justCreated && !rootTx {
				err := clearScope.RollbackContext(ctx)
				if err != nil {
					return err
				}
			}
			return err
		}
		insertScope.SelectedFields = append(insertScope.SelectedFields, relField.Relationship().ManyToManyForeignKey(), relField.Relationship().ForeignKey())

		// create the newly created relationships
		if err = insertScope.CreateContext(ctx); err != nil {
			if !justCreated && !rootTx {
				err := clearScope.RollbackContext(ctx)
				if err != nil {
					return err
				}
			}
			return err
		}
	}

	if !justCreated && !rootTx {
		if err = clearScope.CommitContext(ctx); err != nil {
			return err
		}
	}
	return nil
}

func patchClearRelationshipWithForeignKey(ctx context.Context, s *Scope, relField *mapping.StructField, primaries []interface{}) error {
	// create clearScope for the relation field's model
	var (
		clearScope *Scope
		err        error
	)
	if tx := s.Tx(); tx != nil {
		clearScope, err = tx.newModelC(ctx, s.Controller(), relField.Relationship().Struct(), false)
		if err != nil {
			return err
		}
	} else {
		clearScope = newScopeWithModel(s.Controller(), relField.Relationship().Struct(), false)
	}

	// add selected field into the clear scope
	clearScope.SelectedFields = append(clearScope.SelectedFields, relField.Relationship().ForeignKey())
	clearScope.ForeignFilters = append(clearScope.ForeignFilters, NewFilter(relField.Relationship().ForeignKey(), OpIn, primaries...))

	// clear the related scope
	if err = clearScope.PatchContext(ctx); err != nil {
		if e, ok := err.(errors.ClassError); ok {
			// if the error is no value result clear the error
			if e.Class() == class.QueryValueNoResult {
				err = nil
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}
