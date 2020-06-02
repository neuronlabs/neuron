package core

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/server"
)

// Options is the structure that contains service options.
type Options struct {
	Name                      string
	NamingConvention          mapping.NamingConvention
	Repositories              []Repository
	Models                    []mapping.Model
	MigrateModels             []mapping.Model
	NonRepositoryModels       []mapping.Model
	Server                    server.Server
	Auth                      auth.Authorizator
	HandleSignals             bool
	Context                   context.Context
	DefaultRepositoryName     string
	SynchronousORM            bool
	DisallowDefaultRepository bool
	UTCTimestamps             bool
	ExternalController        bool
}

func (o *Options) config() (*config.Controller, error) {
	cfg := &config.Controller{
		Services:                  make(map[string]*config.Service),
		NamingConvention:          o.NamingConvention.String(),
		DefaultRepositoryName:     o.DefaultRepositoryName,
		DisallowDefaultRepository: o.DisallowDefaultRepository,
		SynchronousConnections:    o.SynchronousORM,
		UTCTimestamps:             o.UTCTimestamps,
	}
	for _, repo := range o.Repositories {
		_, ok := cfg.Services[repo.Name]
		if ok {
			return nil, errors.Newf(controller.ClassRepositoryAlreadyRegistered, "duplicated repository with name: '%s'", repo.Name)
		}
		cfg.Services[repo.Name] = repo.Config
	}
	return cfg, nil
}

func defaultOptions() *Options {
	return &Options{
		HandleSignals:    true,
		Context:          context.Background(),
		NamingConvention: mapping.SnakeCase,
	}
}

// Option is the function that sets the options for the service.
type Option func(o *Options)

// Repository is the pair of the name with related config used by the service.
type Repository struct {
	Name   string
	Config *config.Service
}
