package service

import (
	"context"
	"github.com/kucjac/uni-logger"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/gateway/routers"
	"github.com/neuronlabs/neuron/i18n"
	"github.com/neuronlabs/neuron/log"
	_ "github.com/neuronlabs/neuron/query/mocks"
	"github.com/stretchr/testify/assert"

	"net/http"

	"testing"
	"time"
)

var defaultRouterFunc = func(
	c *controller.Controller,
	cfg *config.Gateway,
	i *i18n.Support,
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
			g = Default("default")

			t.Run("CancelWithContext", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
				defer cancel()

				g.Run(ctx)
			})

		}

	})

}
