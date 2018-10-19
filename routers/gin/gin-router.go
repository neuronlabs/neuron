package ginjsonapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gwatts/gin-adapter"
	"github.com/kucjac/jsonapi"
	"net/http"
)

func RouteHandler(router *gin.Engine, handler *jsonapi.Handler) error {
	for _, model := range handler.ModelHandlers {
		// mStruct := handler.Controller.Models.Get(model.ModelType)
		mStruct := handler.Controller.Models.Get(model.ModelType)

		if mStruct == nil {
			return fmt.Errorf("Model:'%s' not precomputed.", model.ModelType.Name())
		}

		base := handler.Controller.APIURLBase + "/" + mStruct.GetCollectionType()

		getMiddlewares := func(middlewares ...jsonapi.MiddlewareFunc) gin.HandlersChain {
			ginMiddlewares := []gin.HandlerFunc{}
			for _, middleware := range middlewares {
				ginMiddlewares = append(ginMiddlewares, adapter.Wrap(middleware))
			}
			return ginMiddlewares
		}

		var (
			handlers       gin.HandlersChain
			handlerFunc    http.HandlerFunc
			ginHandlerFunc gin.HandlerFunc
		)

		// CREATE
		if model.Create != nil {
			if model.Create.CustomHandlerFunc != nil {
				handlerFunc = model.Create.CustomHandlerFunc
			} else {
				handlerFunc = handler.Create(model, model.Create)
			}
			ginHandlerFunc = gin.WrapF(handlerFunc)
			handlers = getMiddlewares(model.Create.Middlewares...)
			handlers = append(handlers, ginHandlerFunc)

			router.POST(base, handlers...)

		} else {
			router.POST(base, gin.WrapF(handler.EndpointForbidden(model, jsonapi.Create)))
		}

		// GET
		if model.Get != nil {
			if model.Get.CustomHandlerFunc != nil {
				handlerFunc = model.Get.CustomHandlerFunc
			} else {
				handlerFunc = handler.Get(model, model.Get)
			}
			ginHandlerFunc = gin.WrapF(handlerFunc)
			handlers = getMiddlewares(model.Get.Middlewares...)
			handlers = append(handlers, ginHandlerFunc)
			router.GET(base+"/:id", handlers...)
		} else {
			router.GET(base+"/:id", gin.WrapF(handler.EndpointForbidden(model, jsonapi.Get)))

		}

		// LIST
		if model.List != nil {
			if model.List.CustomHandlerFunc != nil {
				handlerFunc = model.List.CustomHandlerFunc
			} else {
				handlerFunc = handler.List(model, model.List)
			}
			ginHandlerFunc = gin.WrapF(handlerFunc)
			handlers = getMiddlewares(model.List.Middlewares...)
			handlers = append(handlers, ginHandlerFunc)
			router.GET(base, handlers...)
		} else {
			router.GET(base, gin.WrapF(handler.EndpointForbidden(model, jsonapi.List)))
		}

		// PATCH
		if model.Patch != nil {
			if model.Patch.CustomHandlerFunc != nil {
				handlerFunc = model.Patch.CustomHandlerFunc
			} else {
				handlerFunc = handler.Patch(model, model.Patch)
			}

			ginHandlerFunc = gin.WrapF(handlerFunc)
			handlers = getMiddlewares(model.Patch.Middlewares...)
			handlers = append(handlers, ginHandlerFunc)
			router.PATCH(base+"/:id", handlers...)
		} else {
			router.PATCH(base+"/:id", gin.WrapF(handler.EndpointForbidden(model, jsonapi.Patch)))
		}

		// DELETE
		if model.Delete != nil {
			if model.Delete.CustomHandlerFunc != nil {
				handlerFunc = model.Delete.CustomHandlerFunc
			} else {
				handlerFunc = handler.Delete(model, model.Delete)
			}
			ginHandlerFunc = gin.WrapF(handlerFunc)
			handlers = getMiddlewares(model.Delete.Middlewares...)
			handlers = append(handlers, ginHandlerFunc)
			router.DELETE(base+"/:id", handlers...)
		} else {
			router.DELETE(base+"/:id", gin.WrapF(handler.EndpointForbidden(model, jsonapi.Delete)))
		}

		for _, rel := range mStruct.ListRelationshipNames() {
			if model.GetRelated != nil {
				if model.GetRelated.CustomHandlerFunc != nil {
					handlerFunc = model.GetRelated.CustomHandlerFunc
				} else {
					handlerFunc = handler.GetRelated(model, model.GetRelated)
				}
				ginHandlerFunc = gin.WrapF(handlerFunc)
				handlers = getMiddlewares(model.GetRelated.Middlewares...)
				handlers = append(handlers, ginHandlerFunc)
				router.GET(base+"/:id/"+rel, handlers...)
			} else {
				router.GET(base+"/:id/"+rel, gin.WrapF(handler.EndpointForbidden(model, jsonapi.GetRelated)))
			}

			if model.GetRelationship != nil {
				if model.GetRelationship.CustomHandlerFunc != nil {
					handlerFunc = model.GetRelationship.CustomHandlerFunc
				} else {
					handlerFunc = handler.GetRelationship(model, model.GetRelationship)
				}
				ginHandlerFunc = gin.WrapF(handlerFunc)
				handlers = getMiddlewares(model.GetRelationship.Middlewares...)
				handlers = append(handlers, ginHandlerFunc)
				router.GET(base+"/:id/relationships/"+rel, handlers...)
			} else {
				router.GET(base+"/:id/relationships"+rel,
					gin.WrapF(handler.EndpointForbidden(model, jsonapi.GetRelationship)))
			}

			if model.PatchRelated != nil {
				if model.PatchRelated.CustomHandlerFunc != nil {
					handlerFunc = model.PatchRelated.CustomHandlerFunc
				} else {
					handlerFunc = handler.PatchRelated(model, model.PatchRelated)
				}
				ginHandlerFunc = gin.WrapF(handlerFunc)
				handlers = getMiddlewares(model.PatchRelated.Middlewares...)
				handlers = append(handlers, ginHandlerFunc)
				router.PATCH(base+"/:id/"+rel, handlers...)
			} else {
				router.PATCH(base+"/:id/"+rel, gin.WrapF(handler.EndpointForbidden(model, jsonapi.PatchRelated)))
			}

			if model.PatchRelationship != nil {
				if model.PatchRelationship.CustomHandlerFunc != nil {
					handlerFunc = model.PatchRelationship.CustomHandlerFunc
				} else {
					handlerFunc = handler.PatchRelationship(model, model.PatchRelationship)
				}
				ginHandlerFunc = gin.WrapF(handlerFunc)
				handlers = getMiddlewares(model.PatchRelationship.Middlewares...)
				handlers = append(handlers, ginHandlerFunc)
				router.PATCH(base+"/id/relationships"+rel, handlers...)
			} else {
				router.PATCH(base+"/:id/relationships/"+rel, gin.WrapF(handler.EndpointForbidden(model, jsonapi.PatchRelationship)))
			}

		}
	}
	return nil
}
