package neuron

import (
	"github.com/neuronlabs/neuron/server"
	"github.com/neuronlabs/neuron/service"
)

// NewVersioned creates a new service that handles multiple versions as a single service.
// VersionedService uses single server that handles different services with different controllers.
func NewVersioned(options ...service.VersionedOption) *service.VersionedService {
	return service.NewVersioned(options...)
}

// ServiceVersion is an option used for VersionedService that adds the 'service' mapped with it's 'version' to the VersionedService.
func ServiceVersion(version string, s *service.Service) service.VersionedOption {
	return func(o *service.VersionedOptions) {
		o.Versions = append(o.Versions, service.Version{Version: version, Service: s})
	}
}

// VersionedServer is an option used for VersionedService that sets the VersionedServer.
func VersionedServer(s server.VersionedServer) service.VersionedOption {
	return func(o *service.VersionedOptions) {
		o.Server = s
	}
}

// VersionedHandleSignal is the option that determines if the os signals should be handled by the versioned service.
func VersionedHandleSignal(handle bool) service.VersionedOption {
	return func(o *service.VersionedOptions) {
		o.HandleSignals = handle
	}
}
