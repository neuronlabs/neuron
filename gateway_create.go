package jsonapi

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kucjac/uni-db"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"strings"
)

// Create returns http.HandlerFunc that creates new 'model' entity within it's repository.
//
// I18n
// It supports i18n of the model. The language in the request is being checked
// if the value provided is supported by the server. If the match is confident
// the language is converted.
//
// Correctly Response with status '201' Created.
func (h *Handler) Create(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}

		SetContentType(rw)
		scope := h.UnmarshalScope(model.ModelType, rw, req)
		if scope == nil {
			return
		}

		scope.setFlags(endpoint, model, h.Controller)

		h.log.Debugf("CREATE - model %v. Unmarshaled scope %#v", scope.Struct.modelType, scope.Value)
		/**

		CLIENT-ID

		*/

		primVal, err := scope.getFieldValue(scope.Struct.primary)
		if err != nil {
			// scope.getFieldValue fails if the model value is not a pointer to the type
			errObj := ErrInvalidInput.Copy()
			errObj.Detail = "The server does not allow creation of multiple objects at once."
			h.MarshalErrors(rw, errObj)
			return
		}

		primUsed := scope.IsPrimaryFieldSelected()
		if primUsed {
			h.log.Debugf("Client Posted ID value: %v", primVal.Interface())
		}

		if scope.Struct.AllowClientID() {
			h.log.Debug("Struct allows ClientID")
			// check if the value is uuid
			if primUsed {

				prim := scope.Struct.primary
				if prim.refStruct.Type.Kind() == reflect.String {
					_, err := uuid.Parse(primVal.String())
					if err != nil {
						errObj := ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = "Client-Generated ID must be of UUID type"
						h.MarshalErrors(rw, errObj)
						return
					}
				}
			}

		} else {
			// check if is non zero
			if primUsed {
				h.log.Debugf("Client Generated ID is not allowed for this model %v", model.ModelType.Name())
				errObj := ErrInvalidJSONFieldValue.Copy()
				errObj.Detail = "Client-Generated ID is not allowed for this model."
				h.MarshalErrors(rw, errObj)
				return
			}
		}

		h.log.Debugf("Create scope value: %+v", scope.Value)

		/**

		CREATE: PRESET PAIRS

		*/

		for _, pair := range endpoint.PresetPairs {
			presetScope, presetField := pair.GetPair()
			if pair.Key != nil {
				if !h.getPresetFilter(pair.Key, presetScope, req, model) {
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

		CREATE: PRESET FILTERS

		*/
		for _, filter := range endpoint.PresetFilters {
			value := req.Context().Value(filter.Key)
			if value != nil {
				if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
					h.log.Errorf("Cannot preset value for the Create.PresetFilter. Model %v. Field %v. Value: %v", model.ModelType.Name(), filter.StructField.GetFieldName(), value)
					h.MarshalInternalError(rw)
					return
				}
				if err := h.PresetScopeValue(scope, filter.FilterField, value); err != nil {
					h.log.Errorf("Cannot preset value for the model: '%s'. FilterField: %v. Error: %v", model.ModelType.Name(), filter.GetFieldName(), err)
					h.MarshalInternalError(rw)
					return
				}
			}
		}

		/**

		CREATE: VALIDATE MODEL

		*/

		if err := h.CreateValidator.Struct(scope.Value); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				errObj := ErrInvalidJSONFieldValue.Copy()
				h.MarshalErrors(rw, errObj)
				return
			}

			validateErrors, ok := err.(validator.ValidationErrors)
			if !ok || ok && len(validateErrors) == 0 {
				h.log.Debugf("Unknown error type while validating. %v", err)
				h.MarshalErrors(rw, ErrInvalidJSONFieldValue.Copy())
				return
			}

			var errs []*ErrorObject
			for _, verr := range validateErrors {
				tag := verr.Tag()

				var errObj *ErrorObject
				if tag == "required" {
					// if field is required but and the field tag is empty
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrMissingRequiredJSONField.Copy()
					errObj.Detail = fmt.Sprintf("The field: %s, is required.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if tag == "isdefault" {
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The field: '%s' must be empty.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if strings.HasPrefix(tag, "len") {
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value of the field: %s is of invalid length.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else {
					errObj = ErrInvalidJSONFieldValue.Copy()
					if verr.Field() != "" {
						errObj.Detail = fmt.Sprintf("Invalid value for the field: '%s'.", verr.Field())
					}
					errs = append(errs, errObj)
					continue
				}
			}
			h.MarshalErrors(rw, errs...)
			return
		}

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

		/**

		CREATE: PRECHECK FILTERS

		*/

		// for _, filter := range endpoint.PrecheckFilters {
		// 	value := req.Context().Value(filter.Key)
		// 	if value != nil {
		// 		if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
		// 			h.log.Errorf("Cannot preset value for the Create.PresetFilter. Model %v. Field %v. Value: %v", model.ModelType.Name(), filter.StructField.GetFieldName(), value)
		// 		}

		// 		if err := scope.AddFilterField(filter.FilterField); err != nil {
		// 			h.log.Error(err)
		// 			h.MarshalInternalError(rw)
		// 			return
		// 		}
		// 	}
		// }

		/**

		CREATE: RELATIONSHIP FILTERS

		*/
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

		err = scope.setBelongsToForeignKeyWithFields()
		if err != nil {
			h.log.Errorf("scope.setBelongsToForeignKey failed. Scope: %#v, %v", scope, err)
			h.MarshalInternalError(rw)
			return
		}

		/**

		CREATE: HOOK BEFORE

		*/

		beforeCreater, ok := scope.Value.(HookBeforeCreator)
		if ok {
			if err = beforeCreater.JSONAPIBeforeCreate(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.MarshalInternalError(rw)
				return
			}
		}

		repo := h.GetRepositoryByType(model.ModelType)

		/**

		CREATE: REPOSITORY CREATE

		*/
		if err = repo.Create(scope); err != nil {
			h.manageDBError(rw, err)
			return
		}

		/**

		CREATE: HOOK AFTER

		*/
		afterCreator, ok := scope.Value.(HookAfterCreator)
		if ok {
			if err = afterCreator.JSONAPIAfterCreate(scope); err != nil {
				// the value should not be created?
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Debugf("Error in HookAfterCreator: %v", err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		CREATE FOREIGN RELATIONSHIPS

		*/

		primVal, err = scope.getFieldValue(scope.Struct.primary)
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

					h.log.Debugf("Patching Foreign Relation: %s for model: %s", relField.jsonAPIName, scope.Struct.collectionType)
					// Prepare the patch scope
					// if relationship is too many
					switch rel.Kind {
					case RelBelongsTo:

						// if sync then it should take backreference and update it
						// if the backreferenced relation is non synced (HasOne and HasMany)
						continue
					case RelHasOne, RelHasMany:
						h.log.Debugf("Patching Relationship of type: %s", rel.Kind.String())
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

		rw.WriteHeader(http.StatusCreated)
		h.MarshalScope(scope, rw, req)
	}
}
