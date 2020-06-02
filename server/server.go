package server

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/orm"
)

// Server is the interface used
type Server interface {
	// Initialize apply controller and initialize given server.
	Initialize(options *Options) error
	// ListenAndServe starts listen and serve the requests.
	Serve() error
	// Shutdown defines gentle shutdown of the server and all it's connections.
	// The server shouldn't handle any more requests but let the remaining finish within given context.
	Shutdown(ctx context.Context) error
}

// Options are the server initialization options.
type Options struct {
	Auth       auth.Authorizator
	Controller *controller.Controller
	DB         orm.DB
}
