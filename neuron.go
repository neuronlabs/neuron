package neuron

import (
	"context"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/query"
	"github.com/neuronlabs/neuron-core/repository"
)

// Neuron is the main structure that stores the controller, create queries, registers models and repositories.
type Neuron struct {
	c *controller.Controller
}

// New creates new Neuron controller.
func New(cfg *config.Controller) (*Neuron, error) {
	c, err := controller.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Neuron{c: c}, nil
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

/*
 *
 * Queries
 *
 */

// Query creates new query for provided 'model'. A model must be registered within given neuron.
func (n *Neuron) Query(model interface{}) query.Builder {
	return query.NewQuery(context.Background(), n.c, model)
}

// QueryCtx creates new query for provided 'model'. A model must be registered within given neuron.
// Whole query would be affected by given context 'ctx'.
func (n *Neuron) QueryCtx(ctx context.Context, model interface{}) query.Builder {
	return query.NewQuery(ctx, n.c, model)
}

/*
 *
 * Transactions
 *
 */

// Begin creates new transactions.
func (n *Neuron) Begin() *query.Tx {
	return query.Begin(context.Background(), n.c, nil)
}

// BeginCtx creates new transaction for given context 'cts'.
// The 'opts' argument is optional, for nil it would take default values.
func (n *Neuron) BeginCtx(ctx context.Context, opts *query.TxOptions) *query.Tx {
	return query.Begin(ctx, n.c, opts)
}

// RunInTransaction runs transaction function 'txFunction' in a single transaction. Commits on exit or
// rolls back on error.
func (n *Neuron) RunInTransaction(ctx context.Context, txFunction TxFn) error {
	return runInTransaction(ctx, n.c, n, nil, txFunction)
}
