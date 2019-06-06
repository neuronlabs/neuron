package query

import (
	"context"
	"fmt"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/uni-db"
	"reflect"
)

var (
	// ProcessList is the Process that do the Repository List method
	ProcessList = &Process{
		Name: "neuron:list",
		Func: listFunc,
	}

	// ProcessBeforeList is the Process that do the Hook Before List
	ProcessBeforeList = &Process{
		Name: "neuron:hook_before_list",
		Func: beforeListFunc,
	}

	// ProcessAfterList is the Process that do the Hook After List
	ProcessAfterList = &Process{
		Name: "neuron:hook_after_list",
		Func: afterListFunc,
	}
)

func listFunc(ctx context.Context, s *Scope) error {
	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Debug("processList RepositoryByModel failed: %v", s.Struct().Type().Name())
		return ErrNoRepositoryFound
	}

	// Cast repository as Lister interface
	listerRepo, ok := repo.(Lister)
	if !ok {
		log.Debug("processList repository is not a Lister repository.")
		return ErrNoListerRepoFound
	}

	// List the
	if err := listerRepo.List(ctx, s); err != nil {
		log.Debugf("processList listerRepo.List failed for model %v. Err: %v", s.Struct().Collection(), err)
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
		return unidb.ErrInternalError.NewWithError(fmt.Errorf("Model: '%s' doesn't cast to AfterLister interface", s.Struct().Type().Name()))
	}

	if err := afterLister.HAfterList(ctx, s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
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
		return unidb.ErrInternalError.NewWithError(fmt.Errorf("Model: '%s' doesn't cast to BeforeLister interface", s.Struct().Type().Name()))
	}

	if err := beforeLister.HBeforeList(ctx, s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}
