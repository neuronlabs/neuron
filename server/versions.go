package server

import (
	"context"
)

// VersionInitializer is a server that allows to initializer separate versions.
type VersionInitializer interface {
	InitializeVersion(version string, options Options) error
}

// VersionedServer is a server that handles multiple versions at once.
type VersionedServer interface {
	VersionInitializer
	EndpointsGetter
	Serve() error
	Shutdown(ctx context.Context) error
}
