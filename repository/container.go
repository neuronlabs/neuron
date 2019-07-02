package repository

import (
	"context"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
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
	factories      map[string]Factory
	defaultFactory Factory
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
		return errors.Newf(class.RepositoryFactoryAlreadyRegistered, "factory: '%s' already registered", repoName)
	}

	c.factories[repoName] = f

	log.Debugf("Repository Factory: '%s' registered succesfully.", repoName)
	return nil
}

// CloseAll closes all repositories
func CloseAll(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan interface{}, len(ctr.factories))
	for _, factory := range ctr.factories {
		go factory.Close(ctx, done)
	}
	var ct int
fl:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-done:
			if err, ok := v.(error); ok {
				return err
			}
			ct++
			if ct == len(ctr.factories) {
				break fl
			}
		}
	}
	return nil
}
