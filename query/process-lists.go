package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

var (
	// ProcessList is the Process that do the Repository List method.
	ProcessList = &Process{
		Name: "neuron:list",
		Func: listFunc,
	}

	// ProcessBeforeList is the Process that do the Hook Before List.
	ProcessBeforeList = &Process{
		Name: "neuron:hook_before_list",
		Func: beforeListFunc,
	}

	// ProcessAfterList is the Process that do the Hook After List.
	ProcessAfterList = &Process{
		Name: "neuron:hook_after_list",
		Func: afterListFunc,
	}
)

func listFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Debug("RepositoryByModel failed: %v", s.Struct().Type().Name())
		return err
	}

	// Cast repository as Lister interface
	listerRepo, ok := repo.(Lister)
	if !ok {
		log.Debug("Repository: %T is not a Lister repository.", repo)
		return errors.NewDet(class.RepositoryNotImplementsLister, "repository doesn't implement Lister interface")
	}

	// List the
	if err := listerRepo.List(ctx, s); err != nil {
		log.Debugf("List failed for model %v. Err: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func afterListFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	if !s.internal().Struct().IsAfterLister() {
		return nil
	}

	afterLister, ok := reflect.New(s.Struct().Type()).Interface().(AfterLister)
	if !ok {
		return errors.NewDetf(class.InternalModelNotCast, "Model: %s should cast into AfterLister interface.", s.Struct().Type().Name())
	}

	if err := afterLister.AfterList(ctx, s); err != nil {
		log.Debug("AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func beforeListFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	if !s.internal().Struct().IsBeforeLister() {
		return nil
	}

	beforeLister, ok := reflect.New(s.Struct().Type()).Interface().(BeforeLister)
	if !ok {
		return errors.NewDetf(class.InternalModelNotCast, "Model: %s should cast into BeforeLister interface.", s.Struct().Type().Name())
	}

	if err := beforeLister.BeforeList(ctx, s); err != nil {
		log.Debug("AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}
