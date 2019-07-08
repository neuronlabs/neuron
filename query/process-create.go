package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/common"
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
	"github.com/neuronlabs/neuron-core/internal/query/scope"
)

var (
	// ProcessCreate is the process that does the Repository Create method.
	ProcessCreate = &Process{
		Name: "neuron:create",
		Func: createFunc,
	}

	// ProcessBeforeCreate is the process that does the hook BeforeCreate.
	ProcessBeforeCreate = &Process{
		Name: "neuron:hook_before_create",
		Func: beforeCreateFunc,
	}

	// ProcessAfterCreate is the Process that does the hook AfterCreate.
	ProcessAfterCreate = &Process{
		Name: "neuron:hook_after_create",
		Func: afterCreateFunc,
	}

	// ProcessStoreScopePrimaries gets the primary field values and sets into scope's store
	// under key: internal.ReducedPrimariesKeyCtx.
	ProcessStoreScopePrimaries = &Process{
		Name: "neuron:store_scope_primaries",
		Func: storeScopePrimaries,
	}
)

func createFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		return err
	}

	creater, ok := repo.(Creator)
	if !ok {
		log.Errorf("The repository deosn't implement Creator interface for model: %s", (*scope.Scope)(s).Struct().Collection())
		return errors.New(class.RepositoryNotImplementsCreator, "repository doesn't implement Creator interface")
	}

	if err := creater.Create(ctx, s); err != nil {
		return err
	}

	s.StoreSet(internal.JustCreated, struct{}{})

	return nil
}

// beforeCreate is the function that is used before the create process
func beforeCreateFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(common.ProcessError); ok {
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
	if _, ok := s.StoreGet(common.ProcessError); ok {
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
	if _, ok := s.StoreGet(common.ProcessError); ok {
		return nil
	}

	primaryValues, err := s.internal().GetPrimaryFieldValues()
	if err != nil {
		return err
	}

	s.StoreSet(internal.ReducedPrimariesStoreKey, primaryValues)

	return nil
}
