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
	if s.Err != nil {
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

	_, hasUpdatedAt := s.Struct().UpdatedAt()
	onlyForeignRelationships := true
	for _, selected := range s.Fieldset {
		if selected.IsPrimary() {
			if len(s.Fieldset) == 1 {
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
	}

	if !onlyForeignRelationships || hasUpdatedAt {
		if log.Level().IsAllowed(log.LevelDebug3) {
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

	// create query for then many models
	q := s.query(ctx, s.c, mapping.NewValueMany(s.Struct()))
	// add filters from 's' scope.
	if err = s.setFiltersTo(q.Scope()); err != nil {
		return err
	}
	// set fields to primary only
	q.SetFields(q.Scope().Struct().Primary())
	if log.Level().IsAllowed(log.LevelDebug3) && q.Err() != nil {
		log.Debug3f("SCOPE[%s][%s] check patch with the list scope: '%s'", s.ID().String(), s.Struct().Collection(), q.Scope())
	}
	return q.List()
}

func beforePatchFunc(ctx context.Context, s *Scope) error {
	if s.Err != nil {
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
	if s.Err != nil {
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
	if s.Err != nil {
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
	if s.Err != nil {
		return nil
	}

	var relationships []*mapping.StructField
	for _, relField := range s.Fieldset {
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
	if s.Err != nil {
		return nil
	}

	var relationships []*mapping.StructField
	for _, relField := range s.Fieldset {
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
		errDetailed := errors.NewDet(class.QueryValuePrimary, "patching has one relationship for many primaries")
		errDetailed.SetDetailsf("Patching multiple primary values on the struct's: '%s' relationship: '%s' is not possible.", s.Struct().Collection(), relField.NeuronName())
		return errDetailed
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

	log.Debug2f("SCOPE[%s] Patching foreign HasOne relationship: '%s'", s.ID(), relField.NeuronName())
	// the primary field filter would be added by the process makePrimaryFilters

	b := s.query(ctx, s.c, relScopeValue)
	err = b.SetFields(b.Scope().Struct().Primary(), relField.Relationship().ForeignKey()).
		Patch()
	if err != nil {
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

func patchHasManyRelationship(ctx context.Context, s *Scope, relField *mapping.StructField, primaries []interface{}) error {
	var err error
	// 1) for related field with values
	// i.e. model with field hasMany = []*relatedModel{{ID: 4},{ID: 5}}
	//
	// for a single primary key:
	// - clear all the relatedModels with the foreign keys equal to the primary (set them to null if possible)
	// - set the relatedModels foreignkeys to primary key where ID in (4, 5)
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
		errDetailed := errors.NewDet(class.QueryValuePrimary, "multiple query primaries while patching has many relationship")
		errDetailed.SetDetails("Can't patch multiple instances with the relation of type hasMany")
		return err
	}
	// create the related value for the scope
	relatedValue := mapping.NewReflectValueSingle(relField.Relationship().Struct())
	// set the foreign key field
	relatedValue.Elem().FieldByIndex(relField.Relationship().ForeignKey().ReflectField().Index).Set(reflect.ValueOf(primaries[0]))
	// create the related scope that patches the relatedPrimaries

	log.Debug2f("SCOPE[%s][%s] Patch HasMany relationship: '%s'", s.ID(), s.Struct().Collection(), relField.NeuronName())
	// patch the related Scope
	err = s.query(ctx, s.c, relatedValue.Interface()).
		AddFilterField(NewFilterField(relField.Relationship().Struct().Primary(), OpIn, relatedPrimaries...)).
		Patch()
	if e, ok := err.(errors.ClassError); ok {
		if e.Class() == class.QueryValueNoResult {
			ed, ok := err.(errors.DetailedError)
			if ok {
				ed.WrapDetailsf("Patching related field: '%s' failed - no related resources found", relField.NeuronName())
			}
		}
	}
	return err
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
		// create the scope for clearing the relationship
		err = s.query(ctx, s.c, mapping.NewValueSingle(relField.Relationship().JoinModel())).
			AddFilterField(NewFilterField(relField.Relationship().ForeignKey(), OpIn, primaries...)).
			Delete()
		// delete the rows in the many2many relationship containing provided values
		if err != nil {
			e, ok := err.(errors.ClassError)
			if ok && e.Class() == class.QueryValueNoResult {
				err = nil
			}
		}
		return err
	} else if isEmpty && justCreated {
		return nil
	}
	// 1) clear the join model values
	// create the clearScope based on the join model
	if !justCreated {
		// it also should get all 'old' entries matched with root.primary
		// value as the backreference primary field
		// copy the root scope primary filters into backreference foreign key
		err = s.query(ctx, s.c, mapping.NewValueSingle(relField.Relationship().JoinModel())).
			AddFilterField(NewFilterField(relField.Relationship().ForeignKey(), OpIn, primaries...)).
			Delete()
		// delete the entries in the join model
		switch e := err.(type) {
		case errors.ClassError:
			// if the err is no result
			if e.Class() == class.QueryValueNoResult {
				err = nil
			}
		case nil:
		}
		if err != nil {
			return err
		}
	}

	// 2) check if the related fields exists in the relation model
	// then if correct insert the new entries into the the join model

	// get the primary field values from the relationship field
	// we need only primary fields
	log.Debug3f("SCOPE[%s][%s] checking many2many field: '%s' related values.", s.ID(), s.Struct().Collection(), relField.NeuronName())
	checkScope := s.query(ctx, s.c, mapping.NewValueMany(relField.Relationship().Struct())).
		SetFields("id").
		AddFilterField(NewFilterField(relField.Relationship().Struct().Primary(), OpIn, relatedPrimaries...)).
		Scope()
	err = checkScope.ListContext(ctx)
	switch e := err.(type) {
	case errors.ClassError:
		if e.Class() == class.QueryValueNoResult {
			if ed, ok := err.(errors.DetailedError); ok {
				ed.WrapDetails("No many2many relationship values found with the provided ids")
			}
		}
	case nil:
	}
	if err != nil {
		return err
	}
	var primaryMap = map[interface{}]bool{}
	// set the primaryMap with the relation primaries
	for _, primary := range relatedPrimaries {
		primaryMap[primary] = false
	}

	checkPrimaries, err := checkScope.getPrimaryFieldValues()
	if err != nil {
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
		// violation integrity constraint error
		errDetailed := errors.NewDet(class.QueryViolationIntegrityConstraint, "relationship values doesn't exists")
		errDetailed.SetDetailsf("Patching relationship field: '%s' failed. The relationships: %v doesn't exists", relField.NeuronName(), nonExistsPrimaries)
		return errDetailed
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

		err = s.query(ctx, s.Controller(), single.Interface()).
			SetFields(relField.Relationship().ManyToManyForeignKey(), relField.Relationship().ForeignKey()).
			Create()
		if err != nil {
			return err
		}
	}
	return nil
}

func patchClearRelationshipWithForeignKey(ctx context.Context, s *Scope, relField *mapping.StructField, primaries []interface{}) error {
	// create clearScope for the relation field's model
	err := s.query(ctx, s.c, relField.Relationship().Struct()).
		SetFields(relField.Relationship().ForeignKey()).
		AddFilterField(NewFilterField(relField.Relationship().ForeignKey(), OpIn, primaries...)).
		Patch()
	if e, ok := err.(errors.ClassError); ok {
		// if the error is no value result clear the error
		if e.Class() == class.QueryValueNoResult {
			err = nil
		}
	}
	return err
}

func setUpdatedAtField(ctx context.Context, s *Scope) error {
	if s.Err != nil {
		return nil
	}

	updatedAt, hasUpdatedAt := s.Struct().UpdatedAt()
	// if there are any selected fields that are not a foreign relationships
	// (attributes, foreign keys etc, relationship-belongs-to...) do the patch process
	if !hasUpdatedAt && len(s.Fieldset) == 0 {
		return errors.NewDet(class.QuerySelectedFieldsNotSelected, "no fields selected for patch process")
	}
	if !hasUpdatedAt {
		return nil
	}
	v := reflect.ValueOf(s.Value).Elem().FieldByIndex(updatedAt.ReflectField().Index)

	if !reflect.DeepEqual(v.Interface(), reflect.Zero(updatedAt.ReflectField().Type).Interface()) {
		return nil
	}

	t := time.Now()
	switch {
	case updatedAt.IsTimePointer():
		v.Set(reflect.ValueOf(&t))
	case updatedAt.IsTime():
		v.Set(reflect.ValueOf(t))
	}
	s.Fieldset[updatedAt.NeuronName()] = updatedAt
	return nil
}
