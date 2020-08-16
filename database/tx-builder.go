package database

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// txQuery is the query builder for the transaction.
type txQuery struct {
	tx    *Tx
	scope *query.Scope
}

// compile time check for the transaction query builder.
var _ Builder = &txQuery{}

/**
 *
 * Basic methods
 *
 */

// Ctx returns query context.
func (b *txQuery) Ctx() context.Context {
	return b.tx.Transaction.Ctx
}

// Scope returns query scope.
func (b *txQuery) Scope() *query.Scope {
	return b.scope
}

// Err returns query error.
func (b *txQuery) Err() error {
	return b.tx.err
}

/**
 *
 * Executable Methods
 *
 */

// Count returns the number of the values for the provided query scope.
func (b *txQuery) Count() (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	cnt, err := Count(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
	}
	return cnt, err
}

// Exists returns true or false depending if there are any rows matching the query.
func (b *txQuery) Exists() (bool, error) {
	if b.tx.err != nil {
		return false, b.tx.err
	}
	exists, err := Exists(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
	}
	return exists, err
}

// Insert stores the values within the given scope's value repository, by starting
// the create process.
func (b *txQuery) Insert() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	err := queryInsert(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// update updates the scope's attribute and relationship values based on the scope models or filters.
func (b *txQuery) Update() (modelsAffected int64, err error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}

	modelsAffected, err = queryUpdate(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
	}
	return modelsAffected, err
}

// Find gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (b *txQuery) Find() ([]mapping.Model, error) {
	if b.tx.err != nil {
		return nil, b.tx.err
	}
	values, err := queryFind(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
		return nil, err
	}
	return values, nil
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (b *txQuery) Get() (mapping.Model, error) {
	if b.tx.err != nil {
		return nil, b.tx.err
	}
	result, err := queryGet(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		if !errors.Is(err, query.ErrNoResult) {
			b.tx.err = err
		}
		return nil, err
	}
	return result, nil
}

// deleteQuery deletes the values provided in the query's scope.
func (b *txQuery) Delete() (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	var modelsAffected int64
	modelsAffected, b.tx.err = deleteQuery(b.tx.Transaction.Ctx, b.tx, b.scope)
	return modelsAffected, b.tx.err
}

// refreshQuery implements Builder interface.
func (b *txQuery) Refresh() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	err := refreshQuery(b.tx.Transaction.Ctx, b.tx, b.scope)
	if err != nil {
		b.tx.err = err
		return err
	}
	return nil
}

/**
 *
 * CallBack methods
 *
 */

// Where inserts 'filterField' into the query scope.
func (b *txQuery) Filter(f filter.Filter) Builder {
	if b.tx.err != nil {
		return b
	}
	b.scope.Filter(f)
	return b
}

// Where parses the filter into the filters.Filter and adds it to the given scope.
func (b *txQuery) Where(where string, values ...interface{}) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Where(where, values...)
	return b
}

// Where parses the filter into the filters.Filter and adds it to the given scope.
func (b *txQuery) WhereOr(filters ...filter.Simple) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.WhereOr(filters...)
	return b
}

// Include includes provided 'relation' field in the query result with respect to provided 'relationFieldset'.
func (b *txQuery) Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Include(relation, relationFieldset...)
	return b
}

// Limit sets the maximum number of objects returned by the Find process,
func (b *txQuery) Limit(limit int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.scope.Limit(limit)
	return b
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
func (b *txQuery) Offset(offset int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.scope.Offset(offset)
	return b
}

// Select adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (b *txQuery) Select(fields ...*mapping.StructField) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Select(fields...)
	return b
}

// OrderBy creates the sort order of the result.
func (b *txQuery) OrderBy(fields ...query.Sort) Builder {
	if b.tx.err != nil {
		return b
	}
	b.scope.SortingOrder = fields
	return b
}

//
// Relations
//

// AddRelations implements Builder interface.
func (b *txQuery) AddRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.tx.err != nil {
		return b.tx.err
	}
	err := queryAddRelations(b.tx.Transaction.Ctx, b.tx, b.scope, relationField, relations...)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// querySetRelations implements Builder interface.
func (b *txQuery) SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.tx.err != nil {
		return b.tx.err
	}
	err := querySetRelations(b.tx.Transaction.Ctx, b.tx, b.scope, relationField, relations...)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// ClearRelations implements Builder interface.
func (b *txQuery) RemoveRelations(relationField *mapping.StructField) (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	modelsAffected, err := queryClearRelations(b.tx.Transaction.Ctx, b.tx, b.scope, relationField)
	if err != nil {
		b.tx.err = err
	}
	return modelsAffected, err
}

// GetRelations implements Builder interface.
func (b *txQuery) GetRelations(relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	if b.tx.err != nil {
		return nil, b.tx.err
	}
	models, err := queryGetRelations(b.tx.Transaction.Ctx, b.tx, b.scope.ModelStruct, b.scope.Models, relationField, relationFieldset...)
	if err != nil {
		b.tx.err = err
	}
	return models, err
}
