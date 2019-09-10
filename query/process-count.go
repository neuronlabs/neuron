package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

func countProcessFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Errorf("No repository found for model: %v", s.Struct().Collection())
		return err
	}

	counter, ok := repo.(Counter)
	if !ok {
		log.Errorf("No Counter repository found for the model: %s", s.Struct().Collection())
		return errors.NewDet(class.RepositoryNotImplementsCounter, "repository doesn't implement Getter interface")
	}

	value, err := counter.Count(ctx, s)
	if err != nil {
		return err
	}
	s.Value = value
	return nil
}

func beforeCountProcessFunc(ctx context.Context, s *Scope) error {
	var ok bool
	if s.Error != nil {
		return nil
	}

	if ok = s.Struct().IsBeforeCounter(); !ok {
		return nil
	}

	var hookBeforeCounter BeforeCounter
	if hookBeforeCounter, ok = s.Value.(BeforeCounter); !ok {
		hookBeforeCounter, ok = reflect.New(s.Struct().Type()).Interface().(BeforeCounter)
		if !ok {
			return errors.NewDetf(class.InternalModelNotCast, "Model: %s should cast into BeforeCounter interface.", s.Struct().Type().Name())
		}
	}
	if err := hookBeforeCounter.BeforeCount(ctx, s); err != nil {
		return err
	}
	return nil
}

func afterCountProcessFunc(ctx context.Context, s *Scope) error {
	var ok bool
	if s.Error != nil {
		return nil
	}
	if ok = s.Struct().IsAfterCounter(); !ok {
		return nil
	}

	hookAfterCounter, ok := reflect.New(s.Struct().Type()).Interface().(AfterCounter)
	if !ok {
		return errors.NewDetf(class.InternalModelNotCast, "Model: %s should cast into AfterCounter interface.", s.Struct().Type().Name())
	}
	if err := hookAfterCounter.AfterCount(ctx, s); err != nil {
		return err
	}
	return nil
}
