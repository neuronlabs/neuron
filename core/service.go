package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/server"
)

// Service is the neuron service struct definition.
type Service struct {
	Options *Options

	// Controller controls service model definitions, repositories and configuration.
	Controller *controller.Controller
	// Server serves defined models.
	Server        server.Server
	DB            database.DB
	Verifier      auth.Verifier
	Authenticator auth.Authenticator
	Tokener       auth.Tokener
}

// New creates new service for provided controller config.
func New(options ...Option) *Service {
	svc := &Service{Options: defaultOptions()}
	for _, opt := range options {
		opt(svc.Options)
	}
	svc.Controller = controller.New(svc.Options.controllerOptions())
	svc.DB = database.New(svc.Controller)
	svc.Server = svc.Options.Server
	svc.Authenticator = svc.Options.Authenticator
	svc.Verifier = svc.Options.Verifier
	svc.Tokener = svc.Options.Tokener

	if len(svc.Options.Models) == 0 {
		log.Fatal("no models defined for the service")
	}
	if svc.Options.DefaultRepository == nil && len(svc.Options.RepositoryModels) == 0 {
		log.Fatal("no repositories defined for the service")
	}
	return svc
}

// Run starts the service.
func (s *Service) Run(ctx context.Context) error {
	if s.Server == nil {
		return errors.Wrap(server.ErrServer, "no server defined for the service")
	}
	if !s.Options.HandleSignals {
		return s.Server.Serve()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM)

	errorChan := make(chan error, 1)
	go func() {
		var err error
		if err = s.Server.Serve(); err != nil && err != http.ErrServerClosed {
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
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown failed: %v", err)
		return err
	}
	log.Info("Server had shutdown successfully.")
	return nil
}

// CloseAll closes all connections with the repositories, proxies and services.
func (s *Service) CloseAll(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.Controller.CloseAll(ctx)
}

// Initialize registers all repositories and models, establish the connection for each repository.
func (s *Service) Initialize(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var cancelFunc context.CancelFunc
	if _, deadlineSet := ctx.Deadline(); !deadlineSet {
		// if no default timeout is already set - try with 30 second timeout.
		ctx, cancelFunc = context.WithTimeout(ctx, time.Second*30)
	} else {
		// otherwise create a cancel function.
		ctx, cancelFunc = context.WithCancel(ctx)
	}
	defer cancelFunc()

	// Register all models.
	if err = s.Controller.RegisterModels(s.Options.Models...); err != nil {
		return err
	}

	// Set the default repository.
	if s.Options.DefaultRepository != nil {
		if err = s.Controller.SetDefaultRepository(s.Options.DefaultRepository); err != nil {
			return err
		}
	}

	// For all non default repositories map related models.
	for repo, models := range s.Options.RepositoryModels {
		if err = s.Controller.MapRepositoryModels(repo, models...); err != nil {
			return err
		}
	}

	// Verify if all models have mapped repository or if there is a default one.
	if err = s.Controller.SetUnmappedModelRepositories(); err != nil {
		return err
	}

	// Register models in their repositories.
	if err = s.Controller.RegisterRepositoryModels(); err != nil {
		return err
	}

	// Initialize all repositories that implements Initializer interface.
	for _, repo := range s.Controller.Repositories {
		initializer, ok := repo.(controller.Initializer)
		if ok {
			if err = initializer.Initialize(s.Controller); err != nil {
				return err
			}
		}
	}

	// Set default store if exists.
	if s.Options.DefaultStore != nil {
		if err = s.Controller.SetDefaultStore(s.Options.DefaultStore); err != nil {
			return err
		}
	}

	// Register named stores.
	for name, store := range s.Options.Stores {
		if err = s.Controller.RegisterStore(name, store); err != nil {
			return err
		}
	}

	// Establish connection with all repositories.
	if err = s.Controller.DialAll(ctx); err != nil {
		return err
	}

	// Migrate defined models.
	if len(s.Options.MigrateModels) > 0 {
		if err = s.Controller.MigrateModels(ctx, s.Options.MigrateModels...); err != nil {
			return err
		}
	}

	// Initialize all collections.
	for _, collection := range s.Options.Collections {
		if err = collection.InitializeCollection(s.Controller); err != nil {
			return err
		}
	}

	if s.Server != nil {
		o := s.serverOptions()
		if err = s.Server.Initialize(o); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) serverOptions() server.Options {
	o := server.Options{
		Tokener:       s.Tokener,
		Verifier:      s.Verifier,
		Authenticator: s.Authenticator,
		Controller:    s.Controller,
		DB:            s.DB,
	}
	return o
}
