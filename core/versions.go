package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/server"
)

// ServiceVersion is a structure that maps the service with it's version.
type ServiceVersion struct {
	Version string
	Service *Service
}

// VersionedService is a service that supports multiple versions for the models server by single server.
type VersionedService struct {
	Options *VersionedOptions
}

// VersionedOptions are the multi version service options.
type VersionedOptions struct {
	HandleSignals bool
	Versions      []ServiceVersion
	Server        server.VersionedServer
}

type VersionedOption func(o *VersionedOptions)

// NewVersioned creates new versioned service.
func NewVersioned(options ...VersionedOption) *VersionedService {
	v := &VersionedService{
		Options: &VersionedOptions{},
	}
	for _, option := range options {
		option(v.Options)
	}
	return v
}

// Initialize sets up all version services, register their models and repositories, and establish their connections.
func (v *VersionedService) Initialize(ctx context.Context) error {
	if len(v.Options.Versions) == 0 {
		log.Fatalf("no version services defined for the VersionedService")
	}
	if v.Options.Server == nil {
		log.Fatalf("versioned server not defined")
	}
	for _, versioned := range v.Options.Versions {
		if err := versioned.Service.Initialize(ctx); err != nil {
			return err
		}
		if err := v.Options.Server.InitializeVersion(versioned.Version, versioned.Service.serverOptions()); err != nil {
			return err
		}
	}
	return nil
}

func (v *VersionedService) Run(ctx context.Context) error {
	if !v.Options.HandleSignals {
		return v.Options.Server.Serve()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM)

	errorChan := make(chan error, 1)
	go func() {
		var err error
		if err = v.Options.Server.Serve(); err != nil && err != http.ErrServerClosed {
			log.Errorf("ListenAndServe failed: %v", err)
			errorChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Infof("Service context had finished.")
	case sig := <-quit:
		log.Infof("Received Signal: '%s'. Shutdown Server begins...", sig.String())
	case err := <-errorChan:
		// the error from the server listen and serve
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err := v.Options.Server.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown failed: %v", err)
		return err
	}
	log.Info("Server had shutdown successfully.")
	return nil
}

// CloseAll closes all connections with the repositories, proxies and services.
func (v *VersionedService) CloseAll(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for _, versioned := range v.Options.Versions {
		if err := versioned.Service.Controller.CloseAll(ctx); err != nil {
			return err
		}
	}
	return nil
}
