package ncore

import (
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/query"
)

// Controller creates new controller for provided 'cfg' config.
func Controller(cfg *config.Controller) (*controller.Controller, error) {
	return controller.New(cfg)
}

// Query creates new query scope for the provided 'model' using the default controller.
func Query(model interface{}) (*query.Scope, error) {
	return query.New(model)
}

// QueryC creates new query scope for the provided 'model' and controller 'c'.
func QueryC(c *controller.Controller, model interface{}) (*query.Scope, error) {
	return query.NewC(c, model)
}
