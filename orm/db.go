package orm

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// New creates new DB for given controller.
func New(c *controller.Controller) *db {
	return &db{c: c}
}

// Compile time check for the DB interface implementations.
var _ DB = &db{}

// Composer is the default query composer that implements DB interface.
type db struct {
	c *controller.Controller
}

// Begin starts new transaction with respect to the transaction context and transaction options with controller 'c'.
func (d *db) Begin(ctx context.Context, options *query.TxOptions) *Tx {
	return begin(ctx, d.c, options)
}

// Controller implements DB interface.
func (d *db) Controller() *controller.Controller {
	return d.c
}

// Query creates new query builder for given 'model' and it's optional instances 'models'.
func (d *db) Query(model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return d.query(context.Background(), model, models...)
}

// QueryCtx creates new query builder for given 'model' and it's optional instances 'models'.
func (d *db) QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return d.query(ctx, model, models...)
}

// Insert implements DB interface.
func (d *db) Insert(ctx context.Context, models ...mapping.Model) error {
	if len(models) == 0 {
		return errors.New(query.ClassNoModels, "nothing to insert")
	}
	mStruct, err := d.c.ModelStruct(models[0])
	if err != nil {
		return err
	}
	s := query.NewScope(mStruct, models...)
	return Insert(ctx, d, s)
}

func (d *db) Update(ctx context.Context, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.New(query.ClassNoModels, "nothing to update")
	}
	mStruct, err := d.c.ModelStruct(models[0])
	if err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	return Update(ctx, d, s)
}

func (d *db) Delete(ctx context.Context, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.New(query.ClassNoModels, "nothing to delete")
	}
	mStruct, err := d.c.ModelStruct(models[0])
	if err != nil {
		return 0, err
	}
	s := query.NewScope(mStruct, models...)
	return Delete(ctx, d, s)
}

func (d *db) query(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) *dbQuery {
	b := &dbQuery{ctx: ctx, db: d}
	b.scope = query.NewScope(model, models...)
	return b
}
