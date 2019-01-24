package gateway

import (
	"net/http"
)

// Get returns a http.HandlerFunc that gets single entity from the "model's"
// repository.
func (h *Handler) Get(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		h.log.Debug("GW-GET BEGIN")
		defer func() { h.log.Debug("GW-GET FINISHED") }()
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}

		SetContentType(rw)

		/**

		GET: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeSingle(req, endpoint, model)
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}

		if errs != nil {
			h.MarshalErrors(rw, errs...)
			return
		}
		scope.NewValueSingle()

		/**

		GET: Language

		*/

		if scope.Struct.language != nil {
			if scope.LanguageFilters == nil {
				if len(h.SupportedLanguages) > 0 {
					tag := h.SupportedLanguages[0]

					b, _ := tag.Base()
					h.log.Debugf("Default Language Filter set to: %v", b.String())
					scope.LanguageFilters = &FilterField{
						StructField: scope.Struct.language,
						Values: []*FilterValues{
							{
								Operator: OpEqual,
								Values: []interface{}{
									b.String(),
								},
							},
						},
					}
				} else {
					h.log.Debugf("No supported languages.")
				}
			} else {
				h.log.Debugf("Language filter not nil: %#v", scope.LanguageFilters)
			}
		} else {
			h.log.Debugf("Language not supported within model: %v", model.ModelType.Name())
		}

		/**

		GET: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		GET: HOOK BEFORE

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET: RELATIONSHIP FILTERS

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

		repo := h.GetRepositoryByType(model.ModelType)
		// Set NewSingleValue for the scope

		/**

		GET: REPOSITORY GET

		*/
		dbErr := repo.Get(scope)
		if dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		GET: HOOK AFTER

		*/
		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET: GET FOREIGN RELATIONS
		The gateway should get the foreign relations not used within includes

		*/
		if err := h.getForeginRelationships(scope.Context(), scope); err != nil {
			h.manageDBError(rw, err)
			return
		}

		/**

		GET: GET INCLUDED FIELDS

		*/
		if correct := h.GetIncluded(scope, rw, req); !correct {
			return
		}

		h.MarshalScope(scope, rw, req)
		return
	}
}

// GetRelated returns a http.HandlerFunc that returns the related field for the 'root' model
// It prepares the scope rooted with 'root' model with some id and gets the 'related' field from
// url. Related field must be a relationship, otherwise an error would be returned.
// The handler gets the root and the specific related field 'id' from the repository
// and then gets the related object from it's repository.
// If no error occurred an related object is being returned
func (h *Handler) GetRelated(root *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[root.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)

		/**

		GET RELATED: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeRelated(req, endpoint, root)
		if err != nil {
			h.log.Errorf("An internal error occurred while building related scope for model: '%v'. %v", root.ModelType, err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}

		/**

		GET RELATED: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, root, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET RELATED: PRECHECK FILTERS

		*/
		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		GET RELATED:  RELATIONSHIP FILTERS

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

		h.log.Debugf("Scope value: %#v", scope.Value)

		relatedField := scope.IncludedFields[0]
		h.log.Debugf("RelatedField: %#v, relatinoship: %#v", relatedField.StructField, relatedField.relationship)
		if relatedField.IsRelationship() && (relatedField.relationship.Kind == RelBelongsTo || (relatedField.relationship.isMany2Many() && (relatedField.relationship.Sync != nil && !*relatedField.relationship.Sync) || relatedField.relationship.Sync == nil)) {
			/**

			  GET RELATED: HOOK BEFORE READ

			*/

			h.log.Debugf("Get relatedField values: %v", relatedField.jsonAPIName)
			if errObj := h.HookBeforeReader(scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			// Get root repository
			rootRepository := h.GetRepositoryByType(root.ModelType)
			// Get the root for given id
			// Select the related field inside

			/**

			GET RELATED: REPOSITORY GET ROOT

			*/
			dbErr := rootRepository.Get(scope)
			if dbErr != nil {
				h.manageDBError(rw, dbErr)
				return
			}

			/**

			  GET RELATED: ROOT HOOK AFTER READ

			*/
			if errObj := h.HookAfterReader(scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}
		}

		/**

		GET: GET FOREIGN RELATIONS
		The gateway should get the foreign relations not used within includes

		*/
		if err := h.getForeginRelationships(scope.Context(), scope); err != nil {
			h.manageDBError(rw, err)
			return
		}

		/**

		GET RELATED: BUILD RELATED SCOPE

		*/

		relatedScope, err := scope.GetRelatedScope()
		if err != nil {
			h.log.Errorf("Error while getting Related Scope: %v", err)
			h.MarshalInternalError(rw)
			return
		}

		// if there is any primary filter
		if relatedScope.Value != nil && len(relatedScope.PrimaryFilters) != 0 {

			relatedRepository := h.GetRepositoryByType(relatedScope.Struct.GetType())

			/**

			  GET RELATED: HOOK BEFORE READER

			*/
			if errObj := h.HookBeforeReader(relatedScope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			// SELECT METHOD TO GET
			if relatedScope.IsMany {
				h.log.Debug("The related scope isMany.")
				err = relatedRepository.List(relatedScope)
			} else {
				h.log.Debug("The related scope isSingle.")
				h.log.Debugf("The value of related scope before get: %#v", relatedScope.Value)
				h.log.Debugf("Fieldset %+v", relatedScope.Fieldset)
				err = relatedRepository.Get(relatedScope)
			}
			if err != nil {
				h.manageDBError(rw, err)
				return
			}
			h.log.Debugf("The value of related scope after get: %#v", relatedScope.Value)

			/**

			HOOK AFTER READER

			*/
			if errObj := h.HookAfterReader(relatedScope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			/**

			GET: GET FOREIGN RELATIONS
			The gateway should get the foreign relations not used within includes

			*/
			if err := h.getForeginRelationships(relatedScope.Context(), relatedScope); err != nil {
				h.manageDBError(rw, err)
				return
			}
		}
		h.MarshalScope(relatedScope, rw, req)
		return

	}
}

// GetRelationship returns a http.HandlerFunc that returns in the response the relationship field
// for the root model
func (h *Handler) GetRelationship(root *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[root.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}

		/**

		GET RELATIONSHIP: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeRelationship(req, endpoint, root)
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}

		scope.setFlags(endpoint, root, h.Controller)

		/**

		  GET RELATIONSHIP: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, root, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET RELATIONSHIP: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		relatedField := scope.IncludedFields[0]

		h.log.Debugf("Related Field: %#v", relatedField)

		if relatedField.IsRelationship() && relatedField.relationship.Kind == RelBelongsTo {

			/**

			  GET RELATIONSHIP: ROOT HOOK BEFORE READ

			*/

			if errObj := h.HookBeforeReader(scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			/**

			GET RELATIONSHIP: GET RELATIONSHIP FILTERS

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

			/**

			  GET RELATIONSHIP: GET ROOT FROM REPOSITORY

			*/

			rootRepository := h.GetRepositoryByType(scope.Struct.GetType())
			dbErr := rootRepository.Get(scope)
			if dbErr != nil {
				h.manageDBError(rw, dbErr)
				return
			}

			/**

			  GET RELATIONSHIP: ROOT HOOK AFTER READ

			*/

			if errObj := h.HookAfterReader(scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}
		}

		/**

		GET RELATIONSHIP: ROOT GET FOREIGN RELATIONS
		The gateway should get the foreign relations not used within includes

		*/
		if err := h.getForeginRelationships(scope.Context(), scope); err != nil {
			h.manageDBError(rw, err)
			return
		}

		h.log.Debugf("Scope val: %#v", scope.Value)
		/**

		  GET RELATIONSHIP: GET RELATIONSHIP SCOPE

		*/
		relationshipScope, err := scope.GetRelationshipScope()
		if err != nil {
			h.log.Errorf("Error while getting RelationshipScope for model: %v. %v", scope.Struct.GetType(), err)
			h.MarshalInternalError(rw)
			return
		}

		/**

		  GET RELATIONSHIP: MARSHAL SCOPE

		*/
		h.MarshalScope(relationshipScope, rw, req)
	}
}
