package core

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/filestore"
)

// DefaultFileStoreName is the name of the default store.
const DefaultFileStoreName = "nrn_default_file_store"

// RegisterFileStore registers file store 's' at 'name'.
func (c *Controller) RegisterFileStore(name string, s filestore.Store) error {
	_, ok := c.FileStores[name]
	if !ok {
		return ErrStoreAlreadySet
	}
	c.FileStores[name] = s
	if name == DefaultFileStoreName {
		if c.DefaultFileStore != nil && c.DefaultFileStore != s {
			return errors.Wrap(ErrStoreAlreadySet, "default store already set")
		}
		c.DefaultFileStore = s
	}
	return nil
}

// SetDefaultFileStore sets the default file store in the controller.
func (c *Controller) SetDefaultFileStore(s filestore.Store) error {
	if c.DefaultFileStore != nil {
		return errors.Wrap(ErrStoreAlreadySet, "default store already exists")
	}
	c.DefaultFileStore = s
	c.FileStores[DefaultFileStoreName] = s
	return nil
}
