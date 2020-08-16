package store

import (
	"context"
	"time"
)

// Store is an interface for key - value stores.
type Store interface {
	// Set sets the record within the store.
	Set(ctx context.Context, record *Record) error
	// SetWithTTL sets the record with specified ttl value.
	SetWithTTL(ctx context.Context, record *Record, ttl time.Duration) error
	// Get gets the record stored under 'key'.
	Get(ctx context.Context, key string) (*Record, error)
	// Delete deletes the record stored using a 'key'.
	Delete(ctx context.Context, key string) error
	// Find finds the records stored using some specific pattern.
	Find(ctx context.Context, options ...FindOption) ([]*Record, error)
	// Close closes the store connection.
	Close(ctx context.Context) error
}

// Record is a single entry stored within a store.
type Record struct {
	// Key is the key at which the record would be stored
	Key string
	// Value is the value of the record.
	Value []byte
	// ExpiresAt defines the time when the record would be expired.
	ExpiresAt time.Time
}

// Copy creates a record copy.
func (r *Record) Copy() *Record {
	cp := &Record{Key: r.Key, Value: make([]byte, len(r.Value)), ExpiresAt: r.ExpiresAt}
	copy(cp.Value, r.Value)
	return cp
}
