package scope

import (
	"context"
	"errors"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query/tx"
)

// ErrRepositoryNotACommiter is an error returned when the repository doesn't implement Commiter interface
var (
	ErrRepositoryNotACommiter   = errors.New("Repository doesn't implement Commiter interface")
	ErrRepositoryNotARollbacker = errors.New("Repository doesn't implement Rollbacker interface")
)

func (s *Scope) commitSingle(ctx context.Context, results chan<- interface{}) {
	var err error
	defer func() {
		if err != nil {
			results <- err
		} else {
			results <- struct{}{}
		}

		s.tx().State = tx.Commit
	}()

	log.Debugf("Scope: %s, tx.State: %s", s.ID().String(), s.tx().State)
	if txn := s.tx(); txn.State != tx.Begin {
		return
	}
	log.Debugf("Scope: %s for model: %s commiting tx: %s", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())

	var c *controller.Controller = (*controller.Controller)(s.Controller())
	repo, ok := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))
	if !ok {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		err = ErrNoRepositoryFound
		return
	}

	cm, ok := repo.(Committer)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Commiter interface", repo.RepositoryName())
		err = ErrRepositoryNotACommiter
		return
	}

	if err = cm.Commit(ctx, s); err != nil {
		return
	}
}

func (s *Scope) rollbackSingle(ctx context.Context, results chan<- interface{}) {
	var err error
	defer func() {
		s.tx().State = tx.Rollback
		if err != nil {
			results <- err

		} else {
			results <- struct{}{}
		}

	}()

	log.Debugf("Scope: %s, tx state: %s", s.ID().String(), s.tx().State)

	if txn := s.tx(); txn.State == tx.Rollback {
		return
	}

	log.Debugf("Scope: %s for model: %s rolling back tx: %s", s.ID().String(), s.Struct().Collection(), s.tx().ID.String())
	var c *controller.Controller = (*controller.Controller)(s.Controller())
	repo, ok := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))
	if !ok {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		err = ErrNoRepositoryFound
		return
	}

	rb, ok := repo.(Rollbacker)
	if !ok {
		log.Debugf("Repository for model: '%s' doesn't implement Rollbacker interface", repo.RepositoryName())
		err = ErrRepositoryNotARollbacker
		return
	}

	if err = rb.Rollback(ctx, s); err != nil {
		return
	}
}
