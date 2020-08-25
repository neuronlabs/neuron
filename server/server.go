package server

import (
	"context"
)

// Server is the interface used
type Server interface {
	EndpointsGetter
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
