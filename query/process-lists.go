package query

import (
	"context"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
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
	log.Debugf("ListFunc")

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
	log.Debugf("AfterListFunc")
	if !(*scope.Scope)(s).Struct().IsAfterLister() {
		return nil
	}

	afterLister := reflect.New(s.Struct().Type()).Interface().(AfterLister)

	if err := afterLister.HAfterList(ctx, s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func beforeListFunc(ctx context.Context, s *Scope) error {
	log.Debugf("BeforeListFunc")
	if !(*scope.Scope)(s).Struct().IsAfterLister() {
		return nil
	}

	beforeLister := reflect.New(s.Struct().Type()).Interface().(BeforeLister)

	if err := beforeLister.HBeforeList(ctx, s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}
