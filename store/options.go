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
	// ConnectionURL is the optional url for store connection.
	ConnectionURL string
	// FileName is the an optional setting when the store's temporary value are being stored.
	FileName string
	// TimeFunc sets the time func for given options.
	TimeFunc func() time.Time
}

// DefaultOptions creates the default store options.
func DefaultOptions() *Options {
	return &Options{
		DefaultExpiration: -1,
		CleanupInterval:   -1,
		TimeFunc:          time.Now,
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

// WithConnectionURL sets the connectionURL for the store.
func WithConnectionURL(connectionURL string) Option {
	return func(o *Options) {
		o.ConnectionURL = connectionURL
	}
}

// WithFileName sets the filename setting for the store.
func WithFileName(fileName string) Option {
	return func(o *Options) {
		o.FileName = fileName
	}
}

// WithTimeFunc sets the time function used by the store.
func WithTimeFunc(tf func() time.Time) Option {
	return func(o *Options) {
		o.TimeFunc = tf
	}
}

// SetOptions are the options used for setting the record.
type SetOptions struct {
	TTL time.Duration
}

// SetOption is an option that sets the set options.
type SetOption func(o *SetOptions)

// SetWithTTL sets the TTL while setting the record.
func SetWithTTL(ttl time.Duration) SetOption {
	return func(o *SetOptions) {
		o.TTL = ttl
	}
}
