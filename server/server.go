package server

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
)

// Server is the interface used
type Server interface {
	EndpointsGetter
	// Initialize apply controller and initialize given server.
	Initialize(options Options) error
	// ListenAndServe starts listen and serve the requests.
	Serve() error
	// Shutdown defines gentle shutdown of the server and all it's connections.
	// The server shouldn't handle any more requests but let the remaining finish within given context.
	Shutdown(ctx context.Context) error
}

// EndpointsGetter is an interface that allows to get endpoint information.
type EndpointsGetter interface {
	// GetEndpoints is a method that gets server endpoints after initialization process.
	GetEndpoints() []*Endpoint
}

// Options are the server initialization options.
type Options struct {
	Authorizer    auth.Authorizer
	Authenticator auth.Authenticator
	Controller    *controller.Controller
	DB            database.DB
}
