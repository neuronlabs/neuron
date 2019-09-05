package query

import (
	"context"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// get returns the single value for the provided scope
func getFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return err
	}

	getter, ok := repo.(Getter)
	if !ok {
		log.Errorf("No Getter repository found for the model: %s", s.Struct().Collection())
		return errors.NewDet(class.RepositoryNotImplementsGetter, "repository doesn't implement Getter interface")
	}

	// 	Get the value from the getter
	if err = getter.Get(ctx, s); err != nil {
		return err
	}

	return nil
}

// processHookBeforeGet is the function that makes the beforeGet hook.
func beforeGetFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	hookBeforeGetter, ok := s.Value.(BeforeGetter)
	if !ok {
		return nil
	}

	if err := hookBeforeGetter.BeforeGet(ctx, s); err != nil {
		return err
	}

	return nil
}

func afterGetFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	hookAfterGetter, ok := s.Value.(AfterGetter)
	if !ok {
		return nil
	}

	if err := hookAfterGetter.AfterGet(ctx, s); err != nil {
		return err
	}

	return nil
}

// fillEmptyFieldset fills the fieldset for the given query if none fields are already set.
func fillEmptyFieldset(ctx context.Context, s *Scope) error {
	s.fillFieldsetIfNotSet()
	return nil
}
