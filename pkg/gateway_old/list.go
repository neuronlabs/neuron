package gateway

import (
	"github.com/kucjac/uni-db"
	"net/http"
)

// List returns a http.HandlerFunc that response with the model's entities taken
// from it's repository.
// QueryParameters:
//	- filter - filter parameter must be followed by the collection name within brackets
// 		i.e. '[collection]' and the field scoped for the filter within brackets, i.e. '[id]'
//		i.e. url: http://myapiurl.com/api/blogs?filter[blogs][id]=4
func (h *Handler) List(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)

		/**

		  LIST: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeList(req, endpoint, model)
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}
		scope.NewValueMany()
		scope.setFlags(endpoint, model, h.Controller)

		// LIST Estimate Language Filter
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

		  LIST: PRECHECK PAIRS

		*/
		if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		  LIST: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		  LIST: HOOK BEFORE READER

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  LIST: GET RELATIONSHIP FILTERS

		*/
		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if hErr.Code == HErrNoValues {
					scope.NewValueMany()
					h.MarshalScope(scope, rw, req)
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
		}

		repo := h.GetRepositoryByType(model.ModelType)

		/**

		  LIST: INCLUDE COUNT

		  Include count into meta data
		*/
		// if endpoint.FlagMetaCountList {
		// scope.PageTotal = true
		// }

		/**

		  LIST: DEFAULT PAGINATION

		*/

		if endpoint.PresetPaginate != nil && scope.Pagination == nil {
			scope.Pagination = endpoint.PresetPaginate
		}

		/**

		  LIST: DEFAULT SORT

		*/
		if len(endpoint.PresetSort) != 0 {
			scope.Sorts = append(endpoint.PresetSort, scope.Sorts...)
		}

		/**

		  LIST: LIST FROM REPOSITORY

		*/
		err = repo.List(scope)
		if err != nil {
			if dbErr, ok := err.(*unidb.Error); ok {
				if !dbErr.Compare(unidb.ErrNoResult) {
					h.manageDBError(rw, err)
					return
				}
			} else {
				h.manageDBError(rw, err)
				return
			}

		}

		/**

		  LIST: HOOK AFTER READ

		*/
		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		LIST: GET FOREIGN RELATIONSHIPS

		*/
		if err := h.getForeginRelationships(scope.Context(), scope); err != nil {
			h.manageDBError(rw, err)
			return
		}

		/**

		  LIST: GET INCLUDED

		*/
		if correct := h.GetIncluded(scope, rw, req); !correct {
			return
		}

		/**

		  LIST: MARSHAL SCOPE

		*/
		h.MarshalScope(scope, rw, req)
		return
	}
}
