package processes

import (
	"github.com/neuronlabs/neuron/internal/query/scope"
	irepos "github.com/neuronlabs/neuron/internal/repositories"
	"github.com/neuronlabs/neuron/log"
	scp "github.com/neuronlabs/neuron/query/scope"
	"github.com/neuronlabs/neuron/repositories"

	"github.com/pkg/errors"
)

// Get returns the single value for the provided scope
func Get(ctrl irepos.RepositoryGetter, s *scope.Scope) error {
	repo, ok := ctrl.RepositoryByModel(s.Struct())
	if !ok {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return ErrNoRepositoryFound
	}

	getter, ok := repo.(repositories.Getter)
	if !ok {
		log.Errorf("No Getter repository found for the model: %s", s.Struct().Collection())
		return ErrNoGetterRepoFound
	}

	// 	Get the value from the getter
	err := getter.Get((*scp.Scope)(s))
	if err != nil {
		return err
	}
	return nil
}
