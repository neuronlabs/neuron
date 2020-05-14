package server

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
)

// Server is the interface used
type Server interface {
	// Initialize apply controller and initialize given server.
	Initialize(c *controller.Controller)
	// ListenAndServe starts listen and serve the requests.
	Serve() error
	// Shutdown defines gentle shutdown of the server and all it's connections.
	// The server shouldn't handle any more requests but let the remaining finish within given context.
	Shutdown(ctx context.Context) error
}
