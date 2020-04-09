package repository

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

var ctr = newContainer()

// RegisterFactory registers provided Factory within the container.
func RegisterFactory(f Factory) error {
	log.Infof("Registering factory: '%s'", f.DriverName())
	return ctr.registerFactory(f)
}

// GetFactory gets the factory with given driver 'name'.
func GetFactory(name string) Factory {
	f, ok := ctr.factories[name]
	if ok {
		return f
	}
	return nil
}

// container is the container for the model repositories.
// It contains mapping between repository name as well as the repository mapped to
// the given ModelStruct
type container struct {
	factories map[string]Factory
}

// newContainer creates the repository container
func newContainer() *container {
	r := &container{
		factories: map[string]Factory{},
	}

	return r
}

// registerFactory registers the given factory within the container
func (c *container) registerFactory(f Factory) error {
	repoName := f.DriverName()

	_, ok := c.factories[repoName]
	if ok {
		log.Debugf("Repository already registered: %s", repoName)
		return errors.NewDetf(class.RepositoryFactoryAlreadyRegistered, "factory: '%s' already registered", repoName)
	}

	c.factories[repoName] = f

	log.Debugf("Repository Factory: '%s' registered successfully.", repoName)
	return nil
}
