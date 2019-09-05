package query

import (
	"context"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
)

func createFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		return err
	}

	creator, ok := repo.(Creator)
	if !ok {
		log.Errorf("The repository doesn't implement Creator interface for model: %s", s.Struct().Collection())
		return errors.NewDet(class.RepositoryNotImplementsCreator, "repository doesn't implement Creator interface")
	}

	if err := creator.Create(ctx, s); err != nil {
		return err
	}

	s.StoreSet(internal.JustCreated, struct{}{})

	return nil
}

// beforeCreate is the function that is used before the create process
func beforeCreateFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	beforeCreator, ok := s.Value.(BeforeCreator)
	if !ok {
		return nil
	}

	// Use the hook function before create
	err := beforeCreator.BeforeCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

// afterCreate is the function that is used after the create process
// It uses AfterCreateR hook if the model implements it.
func afterCreateFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	afterCreator, ok := s.Value.(AfterCreator)
	if !ok {
		return nil
	}

	err := afterCreator.AfterCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

func storeScopePrimaries(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	primaryValues, err := s.getPrimaryFieldValues()
	if err != nil {
		return err
	}

	s.StoreSet(internal.ReducedPrimariesStoreKey, primaryValues)

	return nil
}
