package neuron

import (
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/server"
)

// NewVersioned creates a new service that handles multiple versions as a single service.
// VersionedService uses single server that handles different services with different controllers.
func NewVersioned(options ...core.VersionedOption) *core.VersionedService {
	return core.NewVersioned(options...)
}

// ServiceVersion is an option used for VersionedService that adds the 'service' mapped with it's 'version' to the VersionedService.
func ServiceVersion(version string, service *core.Service) core.VersionedOption {
	return func(o *core.VersionedOptions) {
		o.Versions = append(o.Versions, core.ServiceVersion{Version: version, Service: service})
	}
}

// VersionedServer is an option used for VersionedService that sets the VersionedServer.
func VersionedServer(s server.VersionedServer) core.VersionedOption {
	return func(o *core.VersionedOptions) {
		o.Server = s
	}
}

// VersionedHandleSignal is the option that determines if the os signals should be handled by the versioned service.
func VersionedHandleSignal(handle bool) core.VersionedOption {
	return func(o *core.VersionedOptions) {
		o.HandleSignals = handle
	}
}
