package service

import (
	"context"
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/gateway/routers"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"

	"net/http"

	"testing"
	"time"
)

var defaultRouterFunc = func(
	c *controller.Controller,
	cfg *config.RouterConfig,
) (http.Handler, error) {
	return http.NewServeMux(), nil
}

func TestDefaultGateway(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(unilogger.DEBUG)
	}
	routers.RegisterRouter("default", defaultRouterFunc)

	t.Run("WithValidRouter", func(t *testing.T) {
		var g *GatewayService
		if assert.NotPanics(t, func() { g = Default("default") }) {

			t.Run("CancelWithContext", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
				defer cancel()

				g.Run(ctx)
			})

		}

	})

}
