package service

import (
	"context"
	"fmt"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/gateway/routers"
	"github.com/neuronlabs/neuron/i18n"
	"github.com/neuronlabs/neuron/log"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// GatewayService is the API gateway service structure
type GatewayService struct {
	c *controller.Controller

	// Config is the configuration for teh gateway
	Config *config.Gateway

	// Server is the default gateway service server
	Server *http.Server

	I18n *i18n.Support
}

// New creates new gateway config
func New(cfg *config.Config) (*GatewayService, error) {
	c, err := controller.New(cfg.Controller)
	if err != nil {
		log.Debugf("Creating controller failed: %v", err)
		return nil, err
	}

	g, err := newGatewayService(c, cfg.Gateway)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// NewWithC creates new GatewayService with the provided Controller and the GatewayConfig
func NewWithC(c *controller.Controller, cfg *config.Gateway) (*GatewayService, error) {
	return newGatewayService(c, cfg)
}

// Default creates the default gateway service for the provided router name
func Default(routerName string) *GatewayService {
	c := controller.Default()
	cfg := &(*config.ReadDefaultGatewayConfig())
	cfg.Router.Name = routerName
	cfg.Hostname = "localhost"
	g, err := newGatewayService(c, cfg)
	if err != nil {
		log.Debugf("NewGatewayService failed: %v", err)
		panic(err)
	}

	return g
}

// Controller returns the gateway controller
func (g *GatewayService) Controller() *controller.Controller {
	return g.c
}

// RegisterModels registers the model within the provided gateway's controller
func (g *GatewayService) RegisterModels(models ...interface{}) error {
	return g.Controller().RegisterModels(models...)
}

// SetHandlerAndRoutes sets the routes for the gateway service
// This step should be done as the last step before running the service
// It is based on the registered models and repositories
func (g *GatewayService) SetHandlerAndRoutes() (err error) {
	// set the routes and the router
	g.Server.Handler, err = routers.GetSettedRouter(g.c, g.Config, g.I18n)
	if err != nil {
		log.Debugf("Getting Setted Router failed for the router named: '%s'. %v", g.Config.Router.Name, err)
		return err
	}
	log.Debugf("Registered router: %s with following default middlewares: '%v'", g.Config.Router.Name, g.Config.Router.DefaultMiddlewares)
	return nil
}

// Run starts the GatewayService service and starts listening and serving
// If the Interrupt signal comes from the operating system
func (g *GatewayService) Run(ctx context.Context) {
	// Run the Server
	go func() {
		// Start TLS Server
		if g.Config.TLSCertPath != "" {
			log.Infof("Start listening with TLS at: %s...", g.Server.Addr)
			err := g.Server.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServeTLS failed: %v", err)
			}

		} else {
			log.Infof("Start listening at: %s...", g.Server.Addr)
			err := g.Server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("Listen failed: %v", err)
			}
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Infof("Shutting Down the server with context: %v", ctx.Err())
	case s := <-quit:
		log.Infof("Received Signal: '%s'. Shutdown Server begins...", s)
	}

	ctx, cancel := context.WithTimeout(ctx, g.Config.ShutdownTimeout)
	defer cancel()
	if err := g.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v", err)
	}
	log.Info("Server Shutdown succesfully.")
}

// setConfig sets the gateway config
func (g *GatewayService) setConfig() error {
	v := validator.New()

	if err := v.Struct(g.Config); err != nil {
		return err
	}

	// Get the server setting
	g.Server.IdleTimeout = g.Config.ShutdownTimeout
	g.Server.ReadTimeout = g.Config.ReadTimeout
	g.Server.WriteTimeout = g.Config.WriteTimeout
	g.Server.ReadHeaderTimeout = g.Config.ReadHeaderTimeout

	i, err := i18n.New(g.Config.I18n)
	if err != nil {
		log.Errorf("Creating I18n support failed. %v", err)
		return err
	}
	g.I18n = i

	g.Server.Addr = fmt.Sprintf("%s:%d", g.Config.Hostname, g.Config.Port)
	return nil
}

func newGatewayService(c *controller.Controller, cfg *config.Gateway) (*GatewayService, error) {
	g := &GatewayService{
		c:      c,
		Config: cfg,
		Server: &http.Server{},
	}

	// set the config
	if err := g.setConfig(); err != nil {
		log.Errorf("Setting Config failed: %v", err)
		return nil, err
	}

	return g, nil
}
