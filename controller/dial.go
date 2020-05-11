package controller

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

// DialAll establish connections for all repositories.
func (c *Controller) DialAll(ctx context.Context) error {
	var cancelFunc context.CancelFunc
	if _, deadlineSet := ctx.Deadline(); !deadlineSet {
		// if no default timeout is already set - try with 30 second timeout.
		ctx, cancelFunc = context.WithTimeout(ctx, time.Second*30)
	} else {
		// otherwise create a cancel function.
		ctx, cancelFunc = context.WithCancel(ctx)
	}
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	waitChan := make(chan struct{})

	jobs, err := c.dialJobsCreator(ctx, wg)
	if err != nil {
		return err
	}
	// create error channel
	errc := make(chan error)
	// dial to all repositories
	for job := range jobs {
		c.dial(ctx, job, wg, errc)
	}
	// create wait group channel finish function.
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Errorf("Dial - context deadline exceeded: %v", ctx.Err())
		return ctx.Err()
	case e := <-errc:
		log.Errorf("Dial error: %v", e)
		return e
	case <-waitChan:
		log.Debug("Successful dial to all repositories")
	}
	return nil
}

// dialJob is very simple structure that matches repository with its name.
type dialJob struct {
	name string
	repo repository.Repository
}

func (c *Controller) dialJobsCreator(ctx context.Context, wg *sync.WaitGroup) (<-chan dialJob, error) {
	if len(c.Repositories) == 0 {
		return nil, errors.NewDetf(ClassRepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan dialJob)
	go func() {
		defer close(out)

		for name, repo := range c.Repositories {
			job := dialJob{
				name: name,
				repo: repo,
			}
			wg.Add(1)
			select {
			case out <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (c *Controller) dial(ctx context.Context, job dialJob, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := job.repo.Dial(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}
