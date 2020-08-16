package memory

import (
	"context"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/store"
)

// Compile time check if memory implements store interface.
var _ store.Store = &Memory{}

// Memory is a in-memory store implementation for neuron store.
type Memory struct {
	c       *controller.Controller
	cache   *cache.Cache
	Options *store.Options
}

// New creates new in-memory store.
func New(options ...store.Option) *Memory {
	m := &Memory{
		Options: store.DefaultOptions(),
	}
	for _, option := range options {
		option(m.Options)
	}
	m.cache = cache.New(m.Options.DefaultExpiration, m.Options.CleanupInterval)
	return m
}

// Initialize implements core.Initializer interface.
func (m *Memory) Initialize(c *controller.Controller) error {
	m.c = c
	return nil
}

// Set implements store.Store interface.
func (m *Memory) Set(ctx context.Context, record *store.Record) error {
	key := m.key(record.Key)
	if record.ExpiresAt.IsZero() {
		record.ExpiresAt = m.c.Now().Add(m.Options.DefaultExpiration)
	}
	m.cache.Set(key, record.Value, m.Options.DefaultExpiration)
	return nil
}

// SetWithTTL implements store.Store interface.
func (m *Memory) SetWithTTL(ctx context.Context, record *store.Record, ttl time.Duration) error {
	key := m.key(record.Key)
	cp := &store.Record{Key: record.Key, Value: make([]byte, len(record.Value)), ExpiresAt: record.ExpiresAt}
	if cp.ExpiresAt.IsZero() {
		cp.ExpiresAt = m.c.Now().Add(ttl)
	}
	copy(cp.Value, record.Value)

	m.cache.Set(key, record.Value, ttl)
	return nil
}

// Get implements store.Store interface.
func (m *Memory) Get(ctx context.Context, key string) (*store.Record, error) {
	r, found := m.cache.Get(m.key(key))
	if !found {
		return nil, store.ErrValueNotFound
	}
	rec, ok := r.(*store.Record)
	if !ok {
		return nil, errors.Wrap(store.ErrInternal, "provided record is not a store.Record")
	}
	return rec.Copy(), nil
}

// Delete implements store.Store interface.
func (m *Memory) Delete(ctx context.Context, key string) error {
	m.cache.Delete(m.key(key))
	return nil
}

// Find implements store.Store.
func (m *Memory) Find(ctx context.Context, options ...store.FindOption) ([]*store.Record, error) {
	findOptions := &store.FindPattern{}
	for _, option := range options {
		option(findOptions)
	}
	items := m.cache.Items()
	var records []*store.Record
	prefix := m.Options.Prefix + findOptions.Prefix
	suffix := m.Options.Suffix + findOptions.Suffix
	for k, v := range items {
		if prefix != "" && !strings.HasPrefix(k, prefix) {
			continue
		}
		if suffix != "" && !strings.HasSuffix(k, suffix) {
			continue
		}
		if findOptions.Offset != 0 {
			findOptions.Offset--
			continue
		}
		rec, ok := v.Object.(*store.Record)
		if !ok {
			return nil, errors.Wrap(store.ErrInternal, "a record is not store.Record")
		}
		records = append(records, rec.Copy())
		findOptions.Limit--
		if findOptions.Limit == 0 {
			break
		}
	}
	return records, nil
}

// Close implements io.Closer interface.
func (m Memory) Close(context.Context) error {
	m.cache.Flush()
	return nil
}

func (m *Memory) key(key string) string {
	return m.Options.Prefix + key + m.Options.Suffix
}
