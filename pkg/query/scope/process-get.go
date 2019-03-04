package scope

import (
	"github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
	"github.com/kucjac/jsonapi/pkg/log"
)

var (
	get Process = Process{
		Name: "whiz:get",
		Func: getFunc,
	}

	beforeGet Process = Process{
		Name: "whiz:hook_before_get",
		Func: beforeGetFunc,
	}

	afterGet Process = Process{
		Name: "whiz:hook_after_get",
		Func: afterGetFunc,
	}
)

// get returns the single value for the provided scope
func getFunc(s *Scope) error {
	var c *controller.Controller = (*controller.Controller)(s.Controller())

	repo, ok := c.RepositoryByModel((*scope.Scope)(s).Struct())
	if !ok {
		log.Errorf("No repository found for model: %v", (*scope.Scope)(s).Struct().Collection())
		return ErrNoRepositoryFound
	}

	getter, ok := repo.(getter)
	if !ok {
		log.Errorf("No Getter repository found for the model: %s", (*scope.Scope)(s).Struct().Collection())
		return ErrNoGetterRepoFound
	}

	// 	Get the value from the getter
	err := getter.Get(s)
	if err != nil {
		return err
	}
	return nil
}

// processHookBeforeGet is the function that makes the beforeGet hook
func beforeGetFunc(s *Scope) error {
	hookBeforeGetter, ok := s.Value.(wBeforeGetter)
	if !ok {
		return nil
	}

	log.Debugf("hookBeforeGetter: %T", hookBeforeGetter)

	if err := hookBeforeGetter.HBeforeGet(s); err != nil {
		return err
	}

	return nil
}

func afterGetFunc(s *Scope) error {
	hookAfterGetter, ok := s.Value.(wAfterGetter)
	if !ok {
		return nil
	}

	if err := hookAfterGetter.HAfterGet(s); err != nil {
		return err
	}

	return nil
}
