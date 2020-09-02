package service

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/filestore"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/store"
)

// Service is the neuron service struct definition.
type Service struct {
	Options *Options

	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// DB contains the databases connection.
	DB database.DB

	// Stores is the mapping of the stores with their names.
	Stores map[string]store.Store
	// DefaultStore is the default key-value store used by this controller.
	DefaultStore store.Store

	// Authenticator is the service authenticator.
	Authenticator auth.Authenticator
	// Tokener is the service authentication token creator.
	Tokener auth.Tokener
	// Verifier is the authorization verifier.
	Verifier auth.Verifier

	// FileStores is the mapping of the stores with their names.
	FileStores map[string]filestore.Store
	// DefaultFileStore is the default file store used by this controller.
	DefaultFileStore filestore.Store
	// Server serves defined models.
	Server server.Server
}

// New creates new service for provided controller config.
func New(options ...Option) (*Service, error) {
	o := defaultOptions()
	for _, opt := range options {
		opt(o)
	}
	svc := &Service{
		Options:          o,
		Server:           o.Server,
		Stores:           o.Stores,
		DefaultStore:     o.DefaultStore,
		FileStores:       o.FileStores,
		DefaultFileStore: o.DefaultFileStore,
		Authenticator:    o.Authenticator,
		Tokener:          o.Tokener,
		Verifier:         o.Verifier,
	}

	modelMappingOptions := []mapping.MapOption{mapping.WithNamingConvention(o.NamingConvention)}
	if o.DefaultNotNull {
		modelMappingOptions = append(modelMappingOptions, mapping.WithDefaultNotNull)
	}
	svc.ModelMap = mapping.New(modelMappingOptions...)

	models := append(o.Models, o.MigrateModels...)
	for _, repoModels := range o.RepositoryModels {
		models = append(models, repoModels...)
	}
	if len(models) > 0 {
		if err := svc.ModelMap.RegisterModels(models...); err != nil {
			return nil, err
		}
	}

	if svc.Options.DefaultRepository != nil || len(svc.Options.RepositoryModels) != 0 {
		// Prepare database options.
		databaseOptions := []database.Option{
			database.WithModelMap(svc.ModelMap),
			database.WithMigrateModels(o.MigrateModels...),
			database.WithDefaultRepository(o.DefaultRepository),
		}
		for repo, repoModels := range o.RepositoryModels {
			databaseOptions = append(databaseOptions, database.WithRepositoryModels(repo, repoModels...))
		}
		var err error
		svc.DB, err = database.New(databaseOptions...)
		if err != nil {
			return nil, err
		}
	}
	return svc, nil
}

// Run starts the service and it's server.
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
		// The error from the server listen and serve
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
