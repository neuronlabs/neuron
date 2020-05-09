package query

import (
	"context"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/mapping"
)

// txQuery is the query builder for the transaction.
type txQuery struct {
	tx    *Tx
	scope *Scope
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
	return b.tx.ctx
}

// Scope returns query scope.
func (b *txQuery) Scope() *Scope {
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
	cnt, err := b.scope.Count(b.tx.ctx)
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
	exists, err := b.scope.Exists(b.tx.ctx)
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
	err := b.scope.Insert(b.tx.ctx)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// Update updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (b *txQuery) Update() (modelsAffected int64, err error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}

	modelsAffected, err = b.scope.Update(b.tx.ctx)
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
	values, err := b.scope.Find(b.tx.ctx)
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
	result, err := b.scope.Get(b.tx.ctx)
	if err != nil {
		b.tx.err = err
		if classError, ok := err.(errors.ClassError); ok {
			// TODO: this might be invalid if the error is of class query Value NoResult.
			if classError.Class() == class.QueryValueNoResult {
				b.tx.err = err
			}
		}
		return nil, err
	}
	return result, nil
}

// Delete deletes the values provided in the query's scope.
func (b *txQuery) Delete() (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	var modelsAffected int64
	modelsAffected, b.tx.err = b.scope.Delete(b.tx.ctx)
	return modelsAffected, b.tx.err
}

/**
 *
 * CallBack methods
 *
 */

// Where inserts 'filterField' into the query scope.
func (b *txQuery) Filter(filterField *FilterField) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Filter(filterField)
	return b
}

// Where parses the filter into the filters.Filter and adds it to the given scope.
func (b *txQuery) Where(filter string, values ...interface{}) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Where(filter, values...)
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
func (b *txQuery) OrderBy(fields ...*SortField) Builder {
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
	err := b.scope.addRelations(b.tx.ctx, relationField, relations...)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// SetRelations implements Builder interface.
func (b *txQuery) SetRelations(relationField *mapping.StructField, relations ...mapping.Model) error {
	if b.tx.err != nil {
		return b.tx.err
	}
	err := b.scope.setRelations(b.tx.ctx, relationField, relations...)
	if err != nil {
		b.tx.err = err
	}
	return err
}

// RemoveRelations implements Builder interface.
func (b *txQuery) RemoveRelations(relationField *mapping.StructField) (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	modelsAffected, err := b.scope.removeRelations(b.tx.ctx, relationField)
	if err != nil {
		b.tx.err = err
	}
	return modelsAffected, err
}
