package mockrepo

import (
	"context"

	"github.com/neuronlabs/neuron/query"
)

// Options are the settings for the OnXXX functions.
type Options struct {
	Permanent bool
	Count     int
}

// Option is function that changes options.
type Option func(o *Options)

// Permanent is a permanent option
func Permanent() Option {
	return func(o *Options) {
		o.Permanent = true
	}
}

// Count sets the number of executions of given function.
func Count(count int) Option {
	return func(o *Options) {
		o.Count = count
	}
}

// TransExecuter is an executor of the transaction functions.
type TransExecuter struct {
	Options     *Options
	ExecuteFunc TransFunc
}

// TransFunc is trans execution function.
type TransFunc func(context.Context, *query.Transaction) error

// SavepointExecuter is an executor of the transaction functions.
type SavepointExecuter struct {
	Options     *Options
	ExecuteFunc SavepointFunc
}

// SavepointFunc is trans execution function.
type SavepointFunc func(context.Context, *query.Transaction, string) error

// ResultExecuter is an executor of the result function.
type ResultExecuter struct {
	Options     *Options
	ExecuteFunc ResultFunc
}

// ResultFunc is a repository function that returns int64 and error.
type ResultFunc func(ctx context.Context, s *query.Scope) (int64, error)

// CommonExecuter is an executor of the common function.
type CommonExecuter struct {
	Options     *Options
	ExecuteFunc CommonFunc
}

// CommonFunc is a common repository function.
type CommonFunc func(ctx context.Context, s *query.Scope) error
