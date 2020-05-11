package neuron

import (
	"context"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/repository"
)

// Neuron is the main structure that stores the controller, create queries, registers models and repositories.
type Neuron struct {
	c *controller.Controller
}

// New creates new Neuron controller.
func New(cfg *config.Controller) (*Neuron, error) {
	c, err := controller.NewController(cfg)
	if err != nil {
		return nil, err
	}
	return &Neuron{c: c}, nil
}

// NewC creates new Neuron service with the provided 'c' controller.
func NewC(c *controller.Controller) *Neuron {
	return &Neuron{c: c}
}

// Controller gets neuron controller.
func (n *Neuron) Controller() *controller.Controller {
	return n.c
}

/**
 *
 * Models
 *
 */

// MigrateModels updates or creates provided models representation in their related repositories.
// A representation of model might be a database table, collection etc.
// Model's repository must implement repository.Migrator.
func (n *Neuron) MigrateModels(ctx context.Context, models ...interface{}) error {
	return n.c.MigrateModels(ctx, models...)
}

// RegisterModels registers all models into default controller.
// This function requires repositories to be registered before.
// Returns error if the model was already registered.
func (n *Neuron) RegisterModels(models ...interface{}) error {
	return n.c.RegisterModels(models...)
}

/**
 *
 * Repositories
 *
 */

// CloseAll gently closes repositories connections.
func (n *Neuron) CloseAll(ctx context.Context) error {
	return n.c.CloseAll(ctx)
}

// DialAll establish connections for all neuron repositories.
func (n *Neuron) DialAll(ctx context.Context) error {
	return n.c.DialAll(ctx)
}

// HealthCheck checks all repositories health for the default controller.
func (n *Neuron) HealthCheck(ctx context.Context) (*repository.HealthResponse, error) {
	return n.c.HealthCheck(ctx)
}

// RegisterRepository registers repository into default controller.
// Returns error if the repository with given name was already registered.
// By default the first registered repository is set to the default repository
// for all models that doesn't define their repositories, unless default
// controller's config DisallowDefaultRepository is set to false.
func (n *Neuron) RegisterRepository(name string, repo *config.Repository) error {
	return n.c.RegisterRepository(name, repo)
}
