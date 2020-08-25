package core

import "C"
import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/filestore"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/store"
)

// Controller is the structure that contains, initialize and control the flow of the application.
// It contains repositories, model definitions.
type Controller struct {
	// Options are the settings for the controller.
	Options *Options

	// Repositories are the controller registered repositories.
	Repositories map[string]repository.Repository
	// DefaultRepository is the default repository for given controller.
	DefaultRepository repository.Repository
	// ModelRepositories is the mapping between the model and related repositories.
	ModelRepositories map[*mapping.ModelStruct]repository.Repository
	// ModelMap is a mapping for the model's structures.
	ModelMap *mapping.ModelMap
	// AccountModel is the account model used by Authenticator, Tokener and Verifier.
	AccountModel *mapping.ModelStruct

	// Stores is the mapping of the stores with their names.
	Stores map[string]store.Store
	// DefaultStore is the default key-value store used by this controller.
	DefaultStore store.Store

	// Authenticator is the service authenticator.
	Authenticator auth.Authenticator
	// Tokener is the service authentication token creator.
	Tokener auth.Tokener
	// Verifier is the authorization verifier.
	Verifier auth.Verifier

	// Initializers are the components that needs to initialized, that are not repository, store or one of auth interfaces.
	Initializers []Initializer

	// FileStores is the mapping of the stores with their names.
	FileStores map[string]filestore.Store
	// DefaultFileStore is the default file store used by this controller.
	DefaultFileStore filestore.Store
}

// New creates new controller for provided options.
func New(options *Options) *Controller {
	return newController(options)
}

// NewDefault creates new default controller based on the default config.
func NewDefault() *Controller {
	return newController(defaultOptions())
}

// Now gets and returns current timestamp. If the configs specify the function might return UTC timestamp.
func (c *Controller) Now() time.Time {
	ts := time.Now()
	if c.Options.UTCTimestamps {
		ts = ts.UTC()
	}
	return ts
}

type ctxKeyS struct{}

var ctxKey = &ctxKeyS{}

// CtxSetController stores the controller in the context.
func CtxSetController(parent context.Context, c *Controller) context.Context {
	return context.WithValue(parent, ctxKey, c)
}

// CtxGetController gets the controller from the context.
func CtxGetController(ctx context.Context) (*Controller, bool) {
	c, ok := ctx.Value(ctxKey).(*Controller)
	return c, ok
}

func newController(options *Options) *Controller {
	c := &Controller{
		Repositories:      map[string]repository.Repository{},
		ModelRepositories: map[*mapping.ModelStruct]repository.Repository{},
		Stores:            map[string]store.Store{},
	}
	if options == nil {
		options = &Options{}
		options.NamingConvention = mapping.SnakeCase
	}
	c.Options = options

	// Set the naming convention for the model map.
	mapOptions := []mapping.MapOption{mapping.WithNamingConvention(c.Options.NamingConvention)}
	// Set all not null models.
	for model := range c.Options.ModelNotNullFields {
		mapOptions = append(mapOptions, mapping.WithDefaultNotNullModel(model))
	}
	// Set optional all not null fields.
	if c.Options.DefaultNotNullFields {
		mapOptions = append(mapOptions, mapping.WithDefaultNotNull)
	}
	c.ModelMap = mapping.NewModelMap(mapOptions...)
	return c
}
