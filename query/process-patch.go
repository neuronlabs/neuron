package query

import (
	"context"
	"github.com/kucjac/uni-db"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
	"reflect"
)

var (
	// ProcessPatch is the Process that does the repository Patch method
	ProcessPatch = &Process{
		Name: "neuron:patch",
		Func: patchFunc,
	}

	// ProcessBeforePatch is the Process that does the hook HBeforePatch
	ProcessBeforePatch = &Process{
		Name: "neuron:hook_before_patch",
		Func: beforePatchFunc,
	}

	// ProcessAfterPatch is the Process that does the hook HAfterPatch
	ProcessAfterPatch = &Process{
		Name: "neuron:hook_after_patch",
		Func: afterPatchFunc,
	}

	// ProcessPatchBelongsToRelationships is the process that patches the belongs to relationships
	ProcessPatchBelongsToRelationships = &Process{
		Name: "neuron:patch_belongs_to_relationships",
		Func: patchBelongsToRelationshipsFunc,
	}

	// ProcessPatchForeignRelationships is the Process that patches the foreign relationships
	ProcessPatchForeignRelationships = &Process{
		Name: "neuron:patch_foreign_relationships",
		Func: patchForeignRelationshipsFunc,
	}
)

func patchFunc(ctx context.Context, s *Scope) error {

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return ErrNoRepositoryFound
	}

	patchRepo, ok := repo.(Patcher)
	if !ok {
		log.Errorf("Repository for current model: '%s' doesn't support Patch method", s.Struct().Type().Name())
		return ErrNoPatcherFound
	}

	if err := patchRepo.Patch(ctx, s); err != nil {
		return err
	}

	return nil
}

func beforePatchFunc(ctx context.Context, s *Scope) error {
	beforePatcher, ok := s.Value.(BeforePatcher)
	if !ok {
		return nil
	}
	if err := beforePatcher.HBeforePatch(ctx, s); err != nil {
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

	if err := afterPatcher.HAfterPatch(ctx, s); err != nil {
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
	iScope := (*scope.Scope)(s)
	primVal, err := iScope.GetPrimaryFieldValue()
	if err != nil {
		log.Debugf("Can't get scope's primary field value")
		return err
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var relationships []*models.StructField

	for _, relField := range iScope.SelectedFields() {
		if !relField.IsRelationship() {
			continue
		}
		relationships = append(relationships, relField)
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

		ctx, cancelFunc := context.WithTimeout(ctx, maxTimeout)
		defer cancelFunc()

		for _, relField := range relationships {
			go patchForeignRelationship(ctx, iScope, relField, v, primVal, results)
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

func patchForeignRelationship(ctx context.Context, iScope *scope.Scope, relField *models.StructField, v, primVal reflect.Value, results chan<- interface{}) {
	// Patch only the relationship that are flaged as sync
	var err error

	defer func() {
		if err != nil {
			results <- err
		} else {
			results <- struct{}{}
		}
	}()
	if rel := relField.Relationship(); rel != nil {

		if rel.Sync() != nil && *rel.Sync() {
			var relValue reflect.Value
			relValue, err = iScope.GetFieldValue(relField)
			if err != nil {
				log.Errorf("patch Foreign Relationships failed. GetFieldValue of relField. Unexpected error occurred. Err: %v.  relField: %#v", err, relField.Name())
				return
			}
			switch rel.Kind() {

			case models.RelBelongsTo:

				// if sync then it should take backreference and update it
				// if the backreferenced relation is non synced (HasOne and HasMany)
				return

			case models.RelHasOne:

				// divide the values for the zero and non-zero
				relValueInterface := relValue.Interface()
				if reflect.DeepEqual(
					relValueInterface,
					reflect.Zero(relValue.Type()).Interface(),
				) {

					// do the zero value update
					relScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

					// Add the filterField for the given scope
					err = relScope.AddFilterField(
						filters.NewFilter(

							// Filter's structField is the relationship ForeignKey
							rel.ForeignKey(),

							// Value would be root scope's PrimaryField value
							filters.NewOpValuePair(
								filters.OpEqual,
								primVal.Interface(),
							),
						),
					)
					if err != nil {
						log.Debugf("Create RelationshipScope FilterField failed: %v", err)
						return
					}

					// setting foreign key within relScope value is not necessary
					// because newValue contains ZeroValue type
					// It just need to be selected field
					relScope.AddSelectedField(rel.ForeignKey())

					if err = (*Scope)(relScope).PatchContext(ctx); err != nil {
						log.Debugf("Patching relationship Scope with HasOne relation failed: %v", err)
						return
					}

				} else {

					relPrim, er := rel.Struct().PrimaryValues(relValue)
					if er != nil {
						err = er
						log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), err)
						return
					}

					relScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

					prim := relPrim.Interface()
					var sl []interface{}
					sl, err = internal.ConvertToSliceInterface(&prim)
					if err != nil {
						relScope.SetIDFilters(prim)
					} else {
						relScope.SetIDFilters(sl...)
					}

					// add to the chain if the root is on the  transaction
					if tx := (*Scope)(iScope).tx(); tx != nil {
						iScope.AddChainSubscope(relScope)

						if err = (*Scope)(relScope).begin(ctx, &tx.Options, false); err != nil {
							log.Debugf("BeginTx for the related scope failed: %s", err)
							return
						}
					}

					// get and set foreign key value
					fkValue, er := relScope.GetFieldValue(rel.ForeignKey())
					if er != nil {
						log.Debugf("Getting ForeignKey field value failed: %v", er)
						err = er
						return
					}
					fkValue.Set(primVal)

					relScope.AddSelectedField(rel.ForeignKey())

					if err = (*Scope)(relScope).PatchContext(ctx); err != nil {
						log.Debugf("Patching Relationship Scope HasMany failed: %v", err)
						return
					}

				}

			case models.RelHasMany:
				// HasMany relationship

				relScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

				if tx := (*Scope)(iScope).tx(); tx != nil {
					iScope.AddChainSubscope(relScope)
					if err = (*Scope)(relScope).begin(ctx, &tx.Options, false); err != nil {
						log.Debug("BeginTx for the related scope failed: %s", err)
						return
					}
				}

				relValueInterface := relValue.Interface()

				// Check if the relationship value is zero'th value
				if reflect.DeepEqual(
					relValueInterface,
					reflect.Zero(relValue.Type()).Interface(),
				) {
					// If the values is zero'like
					relScope.AddSelectedField(rel.ForeignKey())
					err = relScope.AddFilterField(
						filters.NewFilter(
							rel.ForeignKey(),
							filters.NewOpValuePair(
								filters.OpEqual,
								primVal.Interface(),
							),
						),
					)
					if err != nil {
						log.Debugf("RelationshipScope('%s')->AddFilterField('%s') failed: %v", relField.ApiName(), rel.ForeignKey().Name(), err)
						return
					}

					if err = (*Scope)(relScope).PatchContext(ctx); err != nil {
						return
					}
				} else {
					// Add filter fields
					err = relScope.AddFilterField(
						filters.NewFilter(
							rel.ForeignKey(),
							filters.NewOpValuePair(
								filters.OpEqual,
								primVal.Interface(),
							),
						),
					)
					if err != nil {
						log.Debugf("RelationshipScope('%s')->AddFilterField('%s') failed: %v", relField.ApiName(), rel.ForeignKey().Name(), err)
						return
					}

					if err = (*Scope)(relScope).PatchContext(ctx); err != nil {
						log.Debugf("RelationshipScope('%s')->Patch failed. %v", relField.ApiName(), err)
						switch t := err.(type) {
						case *unidb.Error:
							// Return error if it isn't the 'ErrNoResult'
							if !t.Compare(unidb.ErrNoResult) {
								return
							}
							err = nil
						}

						return
					}

					relPrim, er := rel.Struct().PrimaryValues(relValue)
					if er != nil {
						log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), err)
						err = er
						return
					}

					patchRelScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

					prim := relPrim.Interface()
					sl, err := internal.ConvertToSliceInterface(&prim)
					if err != nil {
						patchRelScope.SetIDFilters(prim)
					} else {
						patchRelScope.SetIDFilters(sl...)
					}

					// get and set foreign key value
					fkValue, er := patchRelScope.GetFieldValue(rel.ForeignKey())
					if er != nil {
						log.Debugf("Getting ForeignKey field value failed: %v", err)
						err = er
						return
					}
					fkValue.Set(primVal)

					patchRelScope.AddSelectedField(rel.ForeignKey())

					if tx := (*Scope)(iScope).tx(); tx != nil {
						iScope.AddChainSubscope(patchRelScope)
						if err = (*Scope)(patchRelScope).begin(ctx, &tx.Options, false); err != nil {
							log.Debug("BeginTx for the related scope failed: %s", err)
							return
						}
					}

					if err = (*Scope)(patchRelScope).PatchContext(ctx); err != nil {
						log.Debugf("Patching Relationship Scope HasMany failed: %v", err)
						return
					}

				}
			case models.RelMany2Many:

				// Check if the Backreference field is set
				if rel.BackreferenceField() == nil {
					return
				}

				relFieldVal := v.FieldByIndex(relField.ReflectField().Index)

				// check if the values are clear
				if reflect.DeepEqual(
					relFieldVal.Interface(),
					reflect.Zero(relField.ReflectField().Type).Interface(),
				) {

					// create the scope for clearing the relationship
					clearScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

					f := filters.NewFilter(rel.BackreferenceField())

					nested := filters.NewFilter(
						rel.BackreferenceField().Relationship().Struct().PrimaryField(),
						filters.NewOpValuePair(filters.OpEqual, primVal.Interface()))
					f.AddNestedField(nested)

					err = clearScope.AddFilterField(f)
					if err != nil {
						return
					}

					clearScope.AddSelectedField(rel.BackreferenceField())
					if tx := (*Scope)(iScope).tx(); tx != nil {
						iScope.AddChainSubscope(clearScope)
						if err = (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
							log.Debug("BeginTx for the related scope failed: %s", err)
							return
						}
					}

					// Patch the clear scope
					if err = (*Scope)(clearScope).PatchContext(ctx); err != nil {
						switch t := err.(type) {
						case *unidb.Error:
							if !t.Compare(unidb.ErrNoResult) {
								return
							}
							err = nil
						}
						return
					}
				} else {
					relPrim, er := rel.Struct().PrimaryValues(relValue)
					if er != nil {
						log.Debug("Getting relationship primary values failed: %v", er)
						err = er
						return
					}

					clearScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

					var relationPrimaries []interface{}
					for i := 0; i < relPrim.Len(); i++ {
						relationPrimaries = append(relationPrimaries, relPrim.Index(i).Interface())
					}

					// patch relations which are not equal to those within
					// root.relation primaries

					clearScope.AddFilterField(
						filters.NewFilter(
							rel.Struct().PrimaryField(),
							filters.NewOpValuePair(
								filters.OpNotIn,
								relationPrimaries...,
							),
						),
					)

					// it also should get all 'old' entries matched with root.primary
					// value as the backreference primary field
					f := filters.NewFilter(
						rel.BackreferenceField(),
					)
					f.AddNestedField(
						filters.NewFilter(
							rel.BackreferenceField().Struct().PrimaryField(),
							filters.NewOpValuePair(
								filters.OpEqual,
								primVal.Interface(),
							),
						),
					)
					clearScope.AddFilterField(f)

					// ClearScope should clear only the provided relationship
					clearScope.AddSelectedField(rel.BackreferenceField())

					if tx := (*Scope)(iScope).tx(); tx != nil {
						iScope.AddChainSubscope(clearScope)
						if err = (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
							log.Debug("BeginTx for the related scope failed: %s", err)
							return
						}
					}

					if err = (*Scope)(clearScope).PatchContext(ctx); err != nil {
						switch t := err.(type) {
						case *unidb.Error:
							if !t.Compare(unidb.ErrNoResult) {
								return
							}
							err = nil
						}
						return
					}

					relPrim2, er := rel.Struct().PrimaryValues(relValue)
					if er != nil {
						log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), er)
						err = er
						return
					}

					relScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), true)

					prim := relPrim2.Interface()
					sl, er := internal.ConvertToSliceInterface(&prim)
					if er != nil {
						relScope.SetIDFilters(prim)
					} else {
						relScope.SetIDFilters(sl...)
					}

					relScope.SetFieldsetNoCheck(rel.BackreferenceField())

					if err = (*Scope)(relScope).ListContext(ctx); err != nil {
						switch e := err.(type) {
						case *unidb.Error:
							if !e.Compare(unidb.ErrNoResult) {
								log.Debugf("List Relationship Scope Many2Many failed: %v", err)
								return
							}
							err = nil
						default:
							log.Debugf("List Relationship Scope Many2Many failed: %v", err)
						}
						return
					}

					relVal := reflect.ValueOf(relScope.Value)

					// iterate over relScope instances
					for i := 0; i < relVal.Len(); i++ {

						elem := relVal.Index(i)
						if elem.Kind() != reflect.Ptr {
							return
						}

						if elem.IsNil() {
							log.Debugf("Patching relation Many2Many. List of related values contain nil value. Relation: %s", relField.Name())
							continue
						}
						elem = elem.Elem()

						// Get field by related struct primary index
						elemPrim := elem.FieldByIndex(
							rel.Struct().PrimaryField().ReflectField().Index,
						)

						elemPrimVal := elemPrim.Interface()
						if reflect.DeepEqual(
							elemPrimVal,
							reflect.Zero(elemPrim.Type()).Interface(),
						) {
							continue
						}

						// create new relPatchScope on the base of it's value

						relPatchScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)

						err = relPatchScope.AddFilterField(
							filters.NewFilter(
								rel.Struct().PrimaryField(),
								filters.NewOpValuePair(
									filters.OpEqual,
									elemPrimVal,
								),
							),
						)
						if err != nil {
							log.Debugf("relPatchScope Many2Many AddFilterField failed: %v", err)
							return
						}

						// get the value of the Backreference relationship
						rootBackref := reflect.New(rel.BackreferenceField().Relationship().Struct().Type())
						el := rootBackref.Elem()

						rootBackRefPrim := el.FieldByIndex(rel.BackreferenceField().Relationship().Struct().PrimaryField().ReflectField().Index)
						rootBackRefPrim.Set(primVal)

						// append the rootscope value to the backreference relationship
						backRef := elem.FieldByIndex(rel.BackreferenceField().ReflectField().Index)
						log.Debugf("Type of backref: %v", backRef.Type())
						backRef.Set(reflect.Append(backRef, rootBackref))

						relPatchScope.Value = relVal.Index(i).Interface()
						relPatchScope.AddSelectedField(rel.BackreferenceField())

						if tx := (*Scope)(iScope).tx(); tx != nil {
							iScope.AddChainSubscope(relPatchScope)
							if err = (*Scope)(relPatchScope).begin(ctx, &tx.Options, false); err != nil {
								log.Debug("BeginTx for the related scope failed: %s", err)
								return
							}
						}

						if err = (*Scope)(relPatchScope).PatchContext(ctx); err != nil {
							log.Debugf("Patching relationship scope in many2many failed: %v", err)
							return
						}

					}

				}
			}
		}
	}
}
