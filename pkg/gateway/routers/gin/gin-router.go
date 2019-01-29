package ginrouter

import (
	"github.com/gin-gonic/gin"
	"github.com/gwatts/gin-adapter"
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/gateway/handler"
	mids "github.com/kucjac/jsonapi/pkg/gateway/middlewares"
	"github.com/kucjac/jsonapi/pkg/gateway/routers"
	ictrl "github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"path"

	"net/http"
)

func init() {
	routers.RegisterRouter("gin-default", defaultGinRouterSetter())
	routers.RegisterRouter("gin", newGinRouterSetter())
}

func defaultGinRouterSetter() routers.RouterSetterFunc {
	g := gin.Default()
	return ginRouterSetter(g)
}

func newGinRouterSetter() routers.RouterSetterFunc {
	g := gin.New()
	return ginRouterSetter(g)
}

// ginRouterSetter is the routers.RouterSetterFunc for the gin Engine
func ginRouterSetter(g *gin.Engine) routers.RouterSetterFunc {
	return routers.RouterSetterFunc(func(
		c *controller.Controller,
		rCfg *config.RouterConfig,
	) (http.Handler, error) {

		// get the handler for the routes
		h := handler.New(c)

		for _, schema := range (*ictrl.Controller)(c).ModelSchemas().Schemas() {
			for _, model := range schema.Models() {

				// Get the url base path = Prefix + schema + model collection
				urlBasePath := path.Join(rCfg.Prefix, schema.Name, model.Collection())

				// get gin middlewares as gin.HandlerChain
				var (
					cfg        *config.ModelConfig = model.Config()
					customMids []string            = rCfg.DefaultMiddlewares
					hc         gin.HandlersChain
					err        error
					forbidden  bool
				)

				// Create Endpoint
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.Create.CustomMiddlewares...)
					forbidden = cfg.Endpoints.Create.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleCreate((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'Create' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}

					g.POST(urlBasePath, hc...)
				}

				// List endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.List.CustomMiddlewares...)
					forbidden = cfg.Endpoints.List.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleList((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'Create' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}

					g.GET(urlBasePath, hc...)
				}

				// Get endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.Get.CustomMiddlewares...)
					forbidden = cfg.Endpoints.Get.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleGet((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'Create' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}

					g.GET(path.Join(urlBasePath, ":id"), hc...)
				}

				// Get Related endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.GetRelated.CustomMiddlewares...)
					forbidden = cfg.Endpoints.GetRelated.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleGetRelated((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'GetRelated' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}
					// iterate over all relationship fields and create the 'get related' endpoints
					for _, rel := range model.RelatinoshipFields() {
						g.GET(path.Join(urlBasePath, ":id", rel.ApiName()), hc...)
					}
				}

				// Get Relationship endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.GetRelationship.CustomMiddlewares...)
					forbidden = cfg.Endpoints.GetRelationship.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleGetRelationship((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'GetRelationship' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}
					// iterate over all relationship fields and create the 'get related' endpoints
					for _, rel := range model.RelatinoshipFields() {
						g.GET(path.Join(urlBasePath, ":id", "relationship", rel.ApiName()), hc...)
					}
				}

				// Patch endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.Patch.CustomMiddlewares...)
					forbidden = cfg.Endpoints.Patch.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandlePatch((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'Patch' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}

					g.PATCH(path.Join(urlBasePath, ":id"), hc...)
				}

				// Delete endpoint
				customMids = rCfg.DefaultMiddlewares
				if cfg != nil {
					customMids = append(customMids, cfg.Endpoints.Delete.CustomMiddlewares...)
					forbidden = cfg.Endpoints.Delete.Forbidden
				}

				if !forbidden {
					hc, err = getHandlersChain(
						gin.WrapF(h.HandleDelete((*mapping.ModelStruct)(model))),
						customMids...,
					)
					if err != nil {
						log.Errorf("Routing Endpoint: 'Delete' for model: '%s' failed. %v", model.Collection(), err)
						return nil, err
					}

					g.DELETE(path.Join(urlBasePath, ":id"), hc...)
				}
			}
		}
		return g, nil
	})
}

// getHandlersChain gets the middlewares from the middleware container and then
// appends if exists custom middlewares
func getHandlersChain(
	f gin.HandlerFunc,
	middlewares ...string,
) (gin.HandlersChain, error) {
	var hc gin.HandlersChain

	for _, middleware := range middlewares {
		mid, err := mids.Get(middleware)
		if err != nil {
			return nil, err
		}

		hc = append(hc, adapter.Wrap(mid))
	}
	hc = append(hc, f)

	return hc, nil
}
