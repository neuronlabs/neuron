package core

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/store"
)

// DefaultStoreName is the name of the default store.
const DefaultStoreName = "nrn_default_store"

// RegisterStore registers store 's' at 'name'.
func (c *Controller) RegisterStore(name string, s store.Store) error {
	_, ok := c.Stores[name]
	if !ok {
		return ErrStoreAlreadySet
	}
	c.Stores[name] = s
	if name == DefaultStoreName {
		if c.DefaultStore != nil && c.DefaultStore != s {
			return errors.Wrap(ErrStoreAlreadySet, "default store already set")
		}
		c.DefaultStore = s
	}
	return nil
}

// SetDefaultStore sets the default store in the controller.
func (c *Controller) SetDefaultStore(s store.Store) error {
	if c.DefaultStore != nil {
		return errors.Wrap(ErrStoreAlreadySet, "default store already exists")
	}
	c.DefaultStore = s
	c.Stores[DefaultStoreName] = s
	return nil
}
