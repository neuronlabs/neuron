package scope

import (
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"reflect"
)

var (
	list Process = Process{
		Name: "neuron:list",
		Func: listFunc,
	}

	beforeList Process = Process{
		Name: "neuron:hook_before_list",
		Func: beforeListFunc,
	}

	afterList Process = Process{
		Name: "neuron:hook_after_list",
		Func: afterListFunc,
	}
)

func listFunc(s *Scope) error {
	log.Debugf("ListFunc")
	var c *controller.Controller = (*controller.Controller)(s.Controller())

	repo, ok := c.RepositoryByModel(((*scope.Scope)(s).Struct()))
	if !ok {
		log.Debug("processList RepositoryByModel failed: %v", s.Struct().Type().Name())
		return ErrNoRepositoryFound
	}

	// Cast repository as Lister interface
	listerRepo, ok := repo.(lister)
	if !ok {
		log.Debug("processList repository is not a Lister repository.")
		return ErrNoListerRepoFound
	}

	// List the
	if err := listerRepo.List(s); err != nil {
		log.Debugf("processList listerRepo.List failed for model %v. Err: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func afterListFunc(s *Scope) error {
	log.Debugf("AfterListFunc")
	if !(*scope.Scope)(s).Struct().IsAfterLister() {
		return nil
	}

	afterLister := reflect.New(s.Struct().Type()).Interface().(wAfterLister)

	if err := afterLister.HAfterList(s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}

func beforeListFunc(s *Scope) error {
	log.Debugf("BeforeListFunc")
	if !(*scope.Scope)(s).Struct().IsAfterLister() {
		return nil
	}

	beforeLister := reflect.New(s.Struct().Type()).Interface().(wBeforeLister)

	if err := beforeLister.HBeforeList(s); err != nil {
		log.Debug("processHookAfterList AfterListR for model: %v. Failed: %v", s.Struct().Collection(), err)
		return err
	}

	return nil
}
