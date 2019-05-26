package query

import (
	"context"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

var (
	// ProcessGet is the process that does the repository Get method
	ProcessGet = &Process{
		Name: "neuron:get",
		Func: getFunc,
	}

	// ProcessBeforeGet is the process that does the  hook HBeforeGet
	ProcessBeforeGet = &Process{
		Name: "neuron:hook_before_get",
		Func: beforeGetFunc,
	}

	// ProcessAfterGet is the process that does the hook HAfterGet
	ProcessAfterGet = &Process{
		Name: "neuron:hook_after_get",
		Func: afterGetFunc,
	}
)

// get returns the single value for the provided scope
func getFunc(ctx context.Context, s *Scope) error {

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", (*scope.Scope)(s).Struct().Collection())
		return ErrNoRepositoryFound
	}

	getter, ok := repo.(Getter)
	if !ok {
		log.Errorf("No Getter repository found for the model: %s", (*scope.Scope)(s).Struct().Collection())
		return ErrNoGetterRepoFound
	}

	// 	Get the value from the getter
	if err = getter.Get(ctx, s); err != nil {
		return err
	}

	return nil
}

// processHookBeforeGet is the function that makes the beforeGet hook
func beforeGetFunc(ctx context.Context, s *Scope) error {
	hookBeforeGetter, ok := s.Value.(BeforeGetter)
	if !ok {
		return nil
	}

	log.Debugf("hookBeforeGetter: %T", hookBeforeGetter)

	if err := hookBeforeGetter.HBeforeGet(ctx, s); err != nil {
		return err
	}

	return nil
}

func afterGetFunc(ctx context.Context, s *Scope) error {
	hookAfterGetter, ok := s.Value.(AfterGetter)
	if !ok {
		return nil
	}

	if err := hookAfterGetter.HAfterGet(ctx, s); err != nil {
		return err
	}

	return nil
}
