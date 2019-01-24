package processes

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
	irepos "github.com/kucjac/jsonapi/pkg/internal/repositories"
	"github.com/kucjac/jsonapi/pkg/log"
	scp "github.com/kucjac/jsonapi/pkg/query/scope"
	"github.com/kucjac/jsonapi/pkg/repositories"

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
