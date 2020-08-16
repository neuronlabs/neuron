package store

import (
	"time"
)

// Options are the initialization options for the store.
type Options struct {
	// DefaultExpiration is the default expiration time that the records use.
	DefaultExpiration time.Duration
	// CleanupInterval sets the cleanup interval when the expired keys are being deleted from store.
	CleanupInterval time.Duration
	// Prefix, Suffix are the default prefix, suffix for the record key.
	Prefix, Suffix string
}

// DefaultOptions creates the default store options.
func DefaultOptions() *Options {
	return &Options{
		DefaultExpiration: -1,
		CleanupInterval:   -1,
		Prefix:            "nrn_",
	}
}

// Option is an option function that changes Options.
type Option func(o *Options)

// WithDefaultExpiration sets the default expiration option.
func WithDefaultExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.DefaultExpiration = expiration
	}
}

// WithPrefix sets the default prefix for the keys using this store.
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithSuffix sets the default suffix for the keys using this store.
func WithSuffix(suffix string) Option {
	return func(o *Options) {
		o.Suffix = suffix
	}
}
