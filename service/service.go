package service

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/server"
)

var (
	// MjrService is the major error classification for the neuron service.
	MjrService errors.Major
	// ClassInvalidOptions is the error classification for undefined server.
	ClassInvalidOptions errors.Class
)

func init() {
	MjrService = errors.MustNewMajor()
	ClassInvalidOptions = errors.MustNewMajorClass(MjrService)
}

// Service is the neuron service struct definition.
type Service struct {
	// Context is the
	Context context.Context
	Options *Options

	// Controller controls service model definitions, repositories and configuration.
	Controller *controller.Controller
	// Server serves defined models.
	Server server.Server
	DB     *query.Creator
}

// New creates new service for provided controller config.
func New(options ...Option) *Service {
	svc := &Service{Options: defaultOptions()}
	for _, opt := range options {
		opt(svc.Options)
	}
	cfg, err := svc.Options.config()
	if err != nil {
		panic(err)
	}
	if svc.Options.ExternalController {
		if svc.Controller, err = controller.NewController(cfg); err != nil {
			panic(err)
		}
	} else {
		if err = controller.New(cfg); err != nil {
			panic(err)
		}
		svc.Controller = controller.Default()
	}
	svc.Context = svc.Options.Context
	if svc.Server == nil {
		panic(errors.Newf(ClassInvalidOptions, "no server defined for the service"))
	}
	if len(svc.Options.Models) == 0 {
		panic(errors.Newf(ClassInvalidOptions, "no models defined for the service"))
	}
	if len(svc.Options.Repositories) == 0 {
		panic(errors.Newf(ClassInvalidOptions, "no repositories defined for the service"))
	}
	return svc
}

// Run starts the service.
func (s *Service) Run() error {
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
	case <-s.Context.Done():
		log.Infof("Service context had finished.")
	case <-quit:
		log.Infof("Received Signal: '%s'. Shutdown Server begins...", s)
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

	// Establish connection with all repositories.
	if err = s.Controller.DialAll(ctx); err != nil {
		return err
	}

	// Register all models.
	if err = s.Controller.RegisterModels(s.Options.Models...); err != nil {
		return err
	}

	// Migrate defined models.
	if len(s.Options.MigrateModels) > 0 {
		if err = s.Controller.MigrateModels(ctx, s.Options.MigrateModels...); err != nil {
			return err
		}
	}
	return nil
}
