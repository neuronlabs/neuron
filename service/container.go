package service

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

var ctr = newContainer()

var (
	// MjrRepository is the major error repository classification.
	MjrRepository errors.Major
	// ClassFactoryAlreadyRegistered is the error classification for the factories already registered.
	ClassFactoryAlreadyRegistered errors.Class
	// ClassNotImplements is the error classification for the repositories that doesn't implement some interface.
	ClassNotImplements errors.Class
	// ClassConnection is the error classification related with repository connection.
	ClassConnection errors.Class
	// ClassAuthorization is the error classification related with repository authorization.
	ClassAuthorization errors.Class
	// ClassReservedName is the error classification related with using reserved name.
	ClassReservedName errors.Class
	// ClassService is the error related with the repository service.
	ClassService errors.Class
)

func init() {
	// Initialize error classes.
	MjrRepository = errors.MustNewMajor()
	ClassFactoryAlreadyRegistered = errors.MustNewMajorClass(MjrRepository)
	ClassNotImplements = errors.MustNewMajorClass(MjrRepository)
	ClassAuthorization = errors.MustNewMajorClass(MjrRepository)
	ClassConnection = errors.MustNewMajorClass(MjrRepository)
	ClassReservedName = errors.MustNewMajorClass(MjrRepository)
	ClassService = errors.MustNewMajorClass(MjrRepository)
}

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
		return errors.NewDetf(ClassFactoryAlreadyRegistered, "factory: '%s' already registered", repoName)
	}

	c.factories[repoName] = f

	log.Debugf("Repository Factory: '%s' registered successfully.", repoName)
	return nil
}
