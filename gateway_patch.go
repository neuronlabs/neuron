package jsonapi

import (
	"context"
	"github.com/kucjac/uni-db"
	"net/http"
	"reflect"
)

// Patch the patch endpoint is used to patch given entity.
// It accepts only the models that matches the provided model Handler.
// If the incoming model
// PRESETTING:
//	- Preset values using PresetScope
//	- Precheck values using PrecheckScope
func (h *Handler) Patch(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		// UnmarshalScope from the request body.
		SetContentType(rw)

		/**

		  PATCH: UNMARSHAL SCOPE

		*/

		scope := h.UnmarshalScope(model.ModelType, rw, req)
		if scope == nil {
			return
		}

		h.log.Debugf("Patching the scope.Value: %#v", scope.Value)

		h.setScopeFlags(scope, endpoint, model)

		/**

		  PATCH: GET and SET ID

		  Set the ID for given model's scope
		*/

		err := h.Controller.GetAndSetID(req, scope)
		if err != nil {
			if err == IErrInvalidType {
				h.log.Errorf("ControllerGetAndSetID failed. %v", err)
				h.MarshalInternalError(rw)
				return
			}
			var (
				errObj *ErrorObject
				ok     bool
			)
			if errObj, ok = err.(*ErrorObject); !ok {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = "Provided invalid id parameter."
			}

			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  PATCH: PRESET PAIRS

		*/
		for _, presetPair := range endpoint.PresetPairs {
			presetScope, presetField := presetPair.GetPair()
			if presetPair.Key != nil {
				if !h.getPresetFilter(presetPair.Key, presetScope, req, model) {
					continue
				}
			}
			values, err := h.GetPresetValues(presetScope, rw)
			if err != nil {
				if hErr := err.(*HandlerError); hErr != nil {
					if hErr.Code == HErrNoValues {
						errObj := ErrInsufficientAccPerm.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					if !h.handleHandlerError(hErr, rw) {
						return
					}
				} else {
					h.log.Error(err)
					h.MarshalInternalError(rw)
					return
				}
				continue
			}

			if err := h.PresetScopeValue(scope, presetField, values...); err != nil {
				h.log.Errorf("Cannot preset value while creating model: '%s'.'%s'", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  PATCH: PRESET FILTERS

		*/

		if !h.SetPresetFilters(scope, model, req, rw, endpoint.PresetFilters...) {
			return
		}

		/**

		  PATCH: PRECHECK PAIRS

		*/

		/**

		CREATE: PRECHECK PAIRS

		*/

		// for _, pair := range endpoint.PrecheckPairs {
		// 	presetScope, presetField := pair.GetPair()
		// 	if pair.Key != nil {
		// 		if !h.getPrecheckFilter(pair.Key, presetScope, req, model) {
		// 			continue
		// 		}
		// 	}

		// 	values, err := h.GetPresetValues(presetScope, rw)
		// 	if err != nil {
		// 		if hErr := err.(*HandlerError); hErr != nil {
		// 			if hErr.Code == HErrNoValues {
		// 				errObj := ErrInsufficientAccPerm.Copy()
		// 				h.MarshalErrors(rw, errObj)
		// 				return
		// 			}
		// 			if !h.handleHandlerError(hErr, rw) {
		// 				return
		// 			}
		// 		} else {
		// 			h.log.Error(err)
		// 			h.MarshalInternalError(rw)
		// 			return
		// 		}
		// 		continue
		// 	}

		// 	if err := h.SetPresetFilterValues(presetField, values...); err != nil {
		// 		h.log.Error("Cannot preset values to the filter value. %s", err)
		// 		h.MarshalInternalError(rw)
		// 		return
		// 	}

		// 	if err := h.CheckPrecheckValues(scope, presetField); err != nil {
		// 		h.log.Debugf("Precheck value error: '%s'", err)
		// 		if err == IErrValueNotValid {
		// 			errObj := ErrInvalidJSONFieldValue.Copy()
		// 			errObj.Detail = "One of the field values are not valid."
		// 			h.MarshalErrors(rw, errObj)
		// 		} else {
		// 			h.MarshalInternalError(rw)
		// 		}
		// 		return
		// 	}
		// }
		// if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
		// 	return
		// }

		/**

		  PATCH: PRECHECK FILTERS

		*/

		// if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
		// 	return
		// }

		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		err = scope.setBelongsToForeignKeyWithFields(scope.SelectedFields...)
		if err != nil {
			h.log.Errorf("scope.setBelongsToForeignKey failed. Scope: %#v, %v", scope, err)
			h.MarshalInternalError(rw)
			return
		}

		// Get the Repository for given model
		repo := h.GetRepositoryByType(model.ModelType)

		/**

		  PATCH: HOOK BEFORE PATCH

		*/
		if beforePatcher, ok := scope.Value.(HookBeforePatcher); ok {
			if err = beforePatcher.JSONAPIBeforePatch(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Error in HookBeforePatch for model: %v. Error: %v", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  PATCH: REPOSITORY PATCH

		*/
		// Use Patch Method on given model's Repository for given scope.
		err = repo.Patch(scope)
		if err != nil {
			if dbErr, ok := err.(*unidb.Error); ok {
				if dbErr.Compare(unidb.ErrNoResult) && endpoint.HasPrechecks() {
					errObj := ErrInsufficientAccPerm.Copy()
					errObj.Detail = "Given object is not available for this account or it does not exists."
					h.MarshalErrors(rw, errObj)
					return
				}
				h.manageDBError(rw, dbErr)
				return
			} else if errObj, ok := err.(*ErrorObject); ok {
				h.MarshalErrors(rw, errObj)
				return
			}
			h.log.Errorf("Repository.Patch unknown error: %v. Scope: %#v", err, scope)
			h.MarshalInternalError(rw)
			return
		}

		/**

		  PATCH: HOOK AFTER PATCH

		*/
		if afterPatcher, ok := scope.Value.(HookAfterPatcher); ok {
			if err = afterPatcher.JSONAPIAfterPatch(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Error in HookAfterPatcher for model: %v. Error: %v", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		CREATE FOREIGN RELATIONSHIPS

		*/

		primVal, err := scope.getFieldValue(scope.Struct.primary)
		if err != nil {
			h.log.Errorf("CreateHandler - Primary for Foreign Relationships - scope.getFieldValue() failed. Unexpected error while getting field value. Err: %v. Scope: %v", err, scope)

			// TO DO:
			// Remove the value with id

			h.MarshalInternalError(rw)
		}
		ctx, cancelFunc := context.WithCancel(scope.ctx)
		defer cancelFunc()

		// Iterate over the relationships and get the foreign relationships that
		// are not nil
	ForeignRelations:
		for _, relField := range scope.Struct.relationships {
			// Patch only the relationship that are flaged as sync
			if rel := relField.relationship; rel != nil {

				relValue, err := scope.getFieldValue(relField)
				if err != nil {
					h.log.Errorf("CreatHandler- getFieldValue of relField. Unexpected error occurred. Err: %v. Scope %#v, relField: %#v", err, scope, relField)
					// TO DO:
					// UNDO changes
					h.MarshalInternalError(rw)
					return
				}

				h.log.Debugf("Relation: %s", relField.jsonAPIName)

				// Check if the relation is empty
				if rel.isToMany() {
					// check if not empty
					if reflect.DeepEqual(relValue.Interface(), reflect.Zero(relField.refStruct.Type).Interface()) {
						h.log.Debugf("Relation: %s empty", relField.fieldName)
						continue
					}
					h.log.Debugf("Relation: %s not empty. Value: %v", relField.fieldName, relValue)
				} else {
					// check if not nil
					if relValue.IsNil() {
						continue
					}
				}

				if rel.Sync != nil && *rel.Sync {

					// get scope for single patch
					relPrim, err := relField.getRelationshipPrimariyValues(relValue)
					if err != nil {
						h.log.Errorf("Unexpected error occurred. relField.getRelationshipPrimaryValues failed. Err: %v. Scope: %#v. RelfieldValue: %#v", err, scope, relValue)
						return
					}

					// Prepare the patch scope
					// if relationship is too many
					switch rel.Kind {
					case RelBelongsTo:

						// if sync then it should take backreference and update it
						// if the backreferenced relation is non synced (HasOne and HasMany)
						continue
					case RelHasOne, RelHasMany:
						err := h.patchRelationScope(ctx, scope, primVal, relPrim, relField)
						if err != nil {
							switch e := err.(type) {
							case *unidb.Error:
								h.manageDBError(rw, err)
							case *ErrorObject:
								h.MarshalErrors(rw, e)
							default:
								h.MarshalInternalError(rw)
							}
							return
						}
						// patch the relation value

					case RelMany2Many:
						if rel.BackReferenceField == nil {
							continue
						}
						// Get the backreferenced id values
						relScope := h.getRelationshipScope(ctx, scope, relField, relPrim)
						relScope.newValueMany()
						relScope.Fieldset = map[string]*StructField{
							rel.BackReferenceField.jsonAPIName: rel.BackReferenceField,
						}

						relRepo := h.GetRepositoryByType(relField.relatedStruct.modelType)

						err := relRepo.List(relScope)
						if err != nil {
							switch e := err.(type) {
							case *unidb.Error:
								h.manageDBError(rw, e)
							case *ErrorObject:
								h.MarshalErrors(rw, e)
							default:
								h.log.Errorf("Unexpected error relRepo.List relation. Scope: %#v, Relation: %v, RelScope: %#v, Err: %v", scope, relScope, err)
								h.MarshalInternalError(rw)
							}
							return
						}

						// relPatchScope := h.getRelationshipScope(ctx, scope, relField, relPrim)
						// relPatchScope.newValueSingle()

						// m2mVal := relPatchScope.getFieldValue(relField)
						// // getBackreferencePrimaryValues

						relVal := reflect.ValueOf(relScope.Value)

						// iterate over relScope instances
						for i := 0; i < relVal.Len(); i++ {
							elem := relVal.Index(i)
							if elem.Kind() != reflect.Ptr {
								h.log.Warningf("Elem: %v. Index: %d, is not of ptr type. Model: %s, Relation: %v, elemType: %v", elem.Interface(), i, model.ModelType.Name(), relField.fieldName, elem.Type().Name())
								continue ForeignRelations
							}
							elem = elem.Elem()

							// create new relPatchScope on the base of it's value
							relPatchScope := newRootScope(relField.mStruct)
							relPatchScope.ctx = ctx
							relPatchScope.logger = scope.logger

							// get the value of the Backreference relationship
							backRefElemValue := reflect.New(model.ModelType)
							brrv := backRefElemValue.Elem()
							prim := brrv.FieldByIndex(scope.Struct.primary.refStruct.Index)
							prim.Set(primVal)

							// append the rootscope value to the backreference relationship
							backRef := elem.FieldByIndex(rel.BackReferenceField.refStruct.Index)
							backRef.Set(reflect.Append(backRef, backRefElemValue))

							relPatchScope.Value = relVal.Index(i).Interface()
							relPatchScope.SelectedFields = []*StructField{rel.BackReferenceField}

							// patch the relPatchScope
							err := relRepo.Patch(relPatchScope)
							if err != nil {
								switch e := err.(type) {
								case *unidb.Error:
									h.manageDBError(rw, e)
								case *ErrorObject:
									h.MarshalErrors(rw, e)
								default:
									h.log.Errorf("Unexpected error relRepo.List relation. Scope: %#v, Relation: %v, RelScope: %#v, Err: %v", scope, relScope, err)
									h.MarshalInternalError(rw)
								}
								return
							}
						}

					default:
						continue
					}
				}
			}
		}

		/**

		  PATCH: MARSHAL RESULT

		*/
		if scope.FlagReturnPatchContent != nil {
			if *scope.FlagReturnPatchContent {

				h.MarshalScope(scope, rw, req)
			} else {
				rw.WriteHeader(http.StatusNoContent)
			}
		} else {
			rw.WriteHeader(http.StatusNoContent)
		}
		return
	}
}
