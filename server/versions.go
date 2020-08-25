package server

import (
	"context"

	"github.com/neuronlabs/neuron/core"
)

// VersionInitializer is a server that allows to initializer separate versions.
type VersionInitializer interface {
	InitializeVersion(version string, c *core.Controller) error
}

// VersionedServer is a server that handles multiple versions at once.
type VersionedServer interface {
	VersionInitializer
	EndpointsGetter
	Serve() error
	Shutdown(ctx context.Context) error
}
