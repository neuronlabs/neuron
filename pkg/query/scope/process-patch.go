package scope

import (
	"context"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/uni-db"
	"reflect"
)

var (
	patch Process = Process{
		Name: "whiz:patch",
		Func: patchFunc,
	}

	beforePatch Process = Process{
		Name: "whiz:hook_before_patch",
		Func: beforePatchFunc,
	}

	afterPatch Process = Process{
		Name: "whiz:hook_after_patch",
		Func: afterPatchFunc,
	}

	patchBelongsToRelationships Process = Process{
		Name: "whiz:patch_belongs_to_relationships",
		Func: patchBelongsToRelationshipsFunc,
	}

	patchForeignRelationships Process = Process{
		Name: "whiz:patch_foreign_relationships",
		Func: patchForeignRelationshipsFunc,
	}
)

func patchFunc(s *Scope) error {
	var c *controller.Controller = (*controller.Controller)(s.Controller())
	repo, ok := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))
	if !ok {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return ErrNoRepositoryFound
	}

	patchRepo, ok := repo.(patcher)
	if !ok {
		log.Errorf("Repository for current model: '%s' doesn't support Patch method", s.Struct().Type().Name())
		return ErrNoPatcherFound
	}

	if err := patchRepo.Patch(s); err != nil {
		return err
	}

	return nil
}

func beforePatchFunc(s *Scope) error {
	beforePatcher, ok := s.Value.(wBeforePatcher)
	if !ok {
		return nil
	}
	if err := beforePatcher.HBeforePatch(s); err != nil {
		log.Debugf("AfterPatcher failed for scope value: %v", s.Value)
		return err
	}

	return nil
}

func afterPatchFunc(s *Scope) error {
	afterPatcher, ok := s.Value.(wAfterPatcher)
	if !ok {
		return nil
	}

	if err := afterPatcher.HAfterPatch(s); err != nil {
		log.Debugf("AfterPatcher failed for scope value: %v", s.Value)
		return err
	}

	return nil
}

func patchBelongsToRelationshipsFunc(s *Scope) error {
	err := (*scope.Scope)(s).SetBelongsToForeignKeyFields()
	if err != nil {
		log.Debugf("[Patch] SetBelongsToForeignKey failed: %v", err)
		return err
	}
	return nil
}

func patchForeignRelationshipsFunc(s *Scope) error {
	iScope := (*scope.Scope)(s)
	primVal, err := iScope.GetPrimaryFieldValue()
	if err != nil {
		log.Debugf("Can't get scope's primary field value")
		return err
	}

	ctx, cancelFunc := context.WithCancel(s.Context())
	defer cancelFunc()

	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

ForeignRelations:
	for _, relField := range iScope.SelectedFields() {
		if !relField.IsRelationship() {
			continue
		}
		/**

		TO DO:

		- make patching relationships concurrent

		*/

		// Patch only the relationship that are flaged as sync
		if rel := relField.Relationship(); rel != nil {

			if rel.Sync() != nil && *rel.Sync() {

				relValue, err := iScope.GetFieldValue(relField)
				if err != nil {
					log.Errorf("patch Foreign Relationships failed. GetFieldValue of relField. Unexpected error occurred. Err: %v.  relField: %#v", err, relField.Name())
					// TO DO:
					// UNDO changes
					return err
				}
				switch rel.Kind() {

				case models.RelBelongsTo:

					// if sync then it should take backreference and update it
					// if the backreferenced relation is non synced (HasOne and HasMany)
					continue

				case models.RelHasOne:

					// divide the values for the zero and non-zero
					relValueInterface := relValue.Interface()
					if reflect.DeepEqual(
						relValueInterface,
						reflect.Zero(relValue.Type()).Interface(),
					) {

						// do the zero value update

						relScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

						relScope.NewValueSingle()

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
							return err
						}

						// setting foreign key within relScope value is not necessary
						// because newValue contains ZeroValue type
						// It just need to be selected field
						relScope.AddSelectedField(rel.ForeignKey())

						if err := (*Scope)(relScope).Patch(); err != nil {
							log.Debugf("Patching relationship Scope with HasOne relation failed: %v", err)

							/**

							TO DO:

							- Start REDO transaction
							*/
							return err
						}

					} else {

						relPrim, err := rel.Struct().PrimaryValues(relValue)
						if err != nil {
							log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), err)
							return err
						}

						relScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

						prim := relPrim.Interface()
						sl, err := internal.ConvertToSliceInterface(&prim)
						if err != nil {
							relScope.SetIDFilters(prim)
						} else {
							relScope.SetIDFilters(sl...)
						}

						relScope.NewValueSingle()

						// get and set foreign key value
						fkValue, err := relScope.GetFieldValue(rel.ForeignKey())
						if err != nil {
							log.Debugf("Getting ForeignKey field value failed: %v", err)
							return err
						}
						fkValue.Set(primVal)

						relScope.AddSelectedField(rel.ForeignKey())

						if err := (*Scope)(relScope).Patch(); err != nil {
							log.Debugf("Patching Relationship Scope HasMany failed: %v", err)
							return err
						}

					}

				case models.RelHasMany:
					// HasMany relationship

					relScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

					relScope.NewValueSingle()

					// Check if the relationship value is zero'th value
					relValueInterface := relValue.Interface()
					if reflect.DeepEqual(
						relValueInterface,
						reflect.Zero(relValue.Type()).Interface(),
					) {
						// If the values is zero'like
						relScope.AddSelectedField(rel.ForeignKey())
						err := relScope.AddFilterField(
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
							return err
						}

						if err := (*Scope)(relScope).Patch(); err != nil {
							return err
						}
					} else {
						// Add filter fields
						err := relScope.AddFilterField(
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
							return err
						}

						if err := (*Scope)(relScope).Patch(); err != nil {
							log.Debugf("RelationshipScope('%s')->Patch failed. %v", relField.ApiName(), err)
							switch t := err.(type) {
							case *unidb.Error:
								// Return error if it isn't the 'ErrNoResult'
								if !t.Compare(unidb.ErrNoResult) {
									return err
								}
							default:
								return err
							}

							return err
						}

						relPrim, err := rel.Struct().PrimaryValues(relValue)
						if err != nil {
							log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), err)
							return err
						}

						patchRelScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

						prim := relPrim.Interface()
						sl, err := internal.ConvertToSliceInterface(&prim)
						if err != nil {
							patchRelScope.SetIDFilters(prim)
						} else {
							patchRelScope.SetIDFilters(sl...)
						}

						patchRelScope.NewValueSingle()

						// get and set foreign key value
						fkValue, err := patchRelScope.GetFieldValue(rel.ForeignKey())
						if err != nil {
							log.Debugf("Getting ForeignKey field value failed: %v", err)
							return err
						}
						fkValue.Set(primVal)

						patchRelScope.AddSelectedField(rel.ForeignKey())

						if err := (*Scope)(patchRelScope).Patch(); err != nil {
							log.Debugf("Patching Relationship Scope HasMany failed: %v", err)
							return err
						}

					}
				case models.RelMany2Many:

					// Check if the Backreference field is set
					if rel.BackreferenceField() == nil {
						continue
					}

					relFieldVal := v.FieldByIndex(relField.ReflectField().Index)

					// check if the values are clear
					if reflect.DeepEqual(
						relFieldVal.Interface(),
						reflect.Zero(relField.ReflectField().Type).Interface(),
					) {

						clearScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

						f := filters.NewFilter(rel.BackreferenceField())

						nested := filters.NewFilter(
							rel.BackreferenceField().Relationship().Struct().PrimaryField(),
							filters.NewOpValuePair(filters.OpEqual, primVal.Interface()))
						f.AddNestedField(nested)

						err := clearScope.AddFilterField(f)
						if err != nil {
							return err
						}

						clearScope.NewValueSingle()
						clearScope.AddSelectedField(rel.BackreferenceField())

						// Patch the clear scope
						if err := (*Scope)(clearScope).Patch(); err != nil {
							switch t := err.(type) {
							case *unidb.Error:
								if !t.Compare(unidb.ErrNoResult) {
									return err
								}
							default:
								return err
							}
						}
					} else {
						relPrim, err := rel.Struct().PrimaryValues(relValue)
						if err != nil {
							log.Debug("Getting relationship primary values failed: %v", err)
							return err
						}

						clearScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

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
						clearScope.NewValueSingle()

						if err := (*Scope)(clearScope).Patch(); err != nil {
							switch t := err.(type) {
							case *unidb.Error:
								if !t.Compare(unidb.ErrNoResult) {
									return err
								}
							default:
								return err
							}
						}

						relPrim2, err := rel.Struct().PrimaryValues(relValue)
						if err != nil {
							log.Debugf("Getting Relationship: '%s' PrimaryValues failed: %v", relField.ApiName(), err)
							return err
						}

						relScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

						prim := relPrim2.Interface()
						sl, err := internal.ConvertToSliceInterface(&prim)
						if err != nil {
							relScope.SetIDFilters(prim)
						} else {
							relScope.SetIDFilters(sl...)
						}

						relScope.NewValueMany()
						relScope.SetFieldsetNoCheck(rel.BackreferenceField())

						if err := (*Scope)(relScope).List(); err != nil {
							switch e := err.(type) {
							case *unidb.Error:
								if !e.Compare(unidb.ErrNoResult) {
									log.Debugf("List Relationship Scope Many2Many failed: %v", err)
									return err
								}
							default:
								log.Debugf("List Relationship Scope Many2Many failed: %v", err)
								return err
							}
						}

						relVal := reflect.ValueOf(relScope.Value)

						// iterate over relScope instances
						for i := 0; i < relVal.Len(); i++ {

							elem := relVal.Index(i)
							if elem.Kind() != reflect.Ptr {
								continue ForeignRelations
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
							relPatchScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

							err := relPatchScope.AddFilterField(
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
								return err
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

							if err := (*Scope)(relPatchScope).Patch(); err != nil {
								log.Debugf("Patching relationship scope in many2many failed: %v", err)
								return err
							}

						}

					}
				default:
					continue
				}
			}
		}

	}

	return nil
}
