package query

import (
	"context"
	"reflect"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal/query/scope"
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
	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Debug("RepositoryByModel failed: %v", s.Struct().Type().Name())
		return err
	}

	// Cast repository as Lister interface
	listerRepo, ok := repo.(Lister)
	if !ok {
		log.Debug("Repository: %T is not a Lister repository.", repo)
		return errors.New(class.RepositoryNotImplementsLister, "repository doesn't implement Lister interface")
	}

	// List the
	if err := listerRepo.List(ctx, s); err != nil {
		log.Debugf("List failed for model %v. Err: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func afterListFunc(ctx context.Context, s *Scope) error {
	if !(*scope.Scope)(s).Struct().IsAfterLister() {
		return nil
	}

	afterLister, ok := reflect.New(s.Struct().Type()).Interface().(AfterLister)
	if !ok {
		return errors.Newf(class.InternalModelNotCast, "Model: %s should cast into AfterLister interface.", s.Struct().Type().Name())
	}

	if err := afterLister.AfterList(ctx, s); err != nil {
		log.Debug("AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func beforeListFunc(ctx context.Context, s *Scope) error {
	if !(*scope.Scope)(s).Struct().IsBeforeLister() {
		return nil
	}

	beforeLister, ok := reflect.New(s.Struct().Type()).Interface().(BeforeLister)
	if !ok {
		return errors.Newf(class.InternalModelNotCast, "Model: %s should cast into BeforeLister interface.", s.Struct().Type().Name())
	}

	if err := beforeLister.BeforeList(ctx, s); err != nil {
		log.Debug("AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}
