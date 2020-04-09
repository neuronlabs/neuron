package query

import (
	"context"
)

// TxQuery is the query builder for the transaction.
type TxQuery struct {
	tx    *Tx
	scope *Scope
}

// compile time check for the transaction query builder.
var _ Builder = &TxQuery{}

/**
 *
 * Basic methods
 *
 */

// Ctx returns query context.
func (b *TxQuery) Ctx() context.Context {
	return b.tx.ctx
}

// Scope returns query scope.
func (b *TxQuery) Scope() *Scope {
	return b.scope
}

// Err returns query error.
func (b *TxQuery) Err() error {
	return b.tx.err
}

/**
 *
 * Executable Methods
 *
 */

// Count returns the number of the values for the provided query scope.
func (b *TxQuery) Count() (int64, error) {
	if b.tx.err != nil {
		return 0, b.tx.err
	}
	cnt, err := b.scope.CountContext(b.tx.ctx)
	b.tx.err = err
	return cnt, err
}

// Create stores the values within the given scope's value repository, by starting
// the create process.
func (b *TxQuery) Create() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	b.tx.err = b.scope.CreateContext(b.tx.ctx)
	return b.tx.err
}

// Patch updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (b *TxQuery) Patch() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	b.tx.err = b.scope.PatchContext(b.tx.ctx)
	return b.tx.err
}

// List gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (b *TxQuery) List() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	b.tx.err = b.scope.ListContext(b.tx.ctx)
	return b.tx.err
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (b *TxQuery) Get() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	return b.scope.GetContext(b.tx.ctx)
}

// Delete deletes the values provided in the query's scope.
func (b *TxQuery) Delete() error {
	if b.tx.err != nil {
		return b.tx.err
	}
	b.tx.err = b.scope.DeleteContext(b.tx.ctx)
	return b.tx.err
}

/**
 *
 * CallBack methods
 *
 */

// AddFilterField inserts 'filterField' into the query scope.
func (b *TxQuery) AddFilterField(filterField *FilterField) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.FilterField(filterField)
	return b
}

// Filter parses the filter into the filters.FilterField and adds it to the given scope.
func (b *TxQuery) Filter(filter string, values ...interface{}) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Filter(filter, values...)
	return b
}

// Limit sets the maximum number of objects returned by the List process,
func (b *TxQuery) Limit(limit int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Limit(limit)
	return b
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
func (b *TxQuery) Offset(offset int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Offset(offset)
	return b
}

// PageSize defines pagination page size - maximum amount of returned objects.
func (b *TxQuery) PageSize(pageSize int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.PageSize(pageSize)
	return b
}

// PageNumber defines the pagination page number.
func (b *TxQuery) PageNumber(pageNumber int64) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.PageNumber(pageNumber)
	return b
}

// SetFields adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (b *TxQuery) SetFields(fields ...interface{}) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.SetFields(fields...)
	return b
}

// Sort creates the sort order of the result.
func (b *TxQuery) Sort(fields ...string) Builder {
	if b.tx.err != nil {
		return b
	}
	b.tx.err = b.scope.Sort(fields...)
	return b
}
