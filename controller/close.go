package controller

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/service"
)

// CloseAll gently closes repository connections.
func (c *Controller) CloseAll(ctx context.Context) error {
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
	jobs, err := c.closeJobsCreator(ctx, wg)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	for job := range jobs {
		c.closeRepo(ctx, job, wg, errChan)
	}

	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Errorf("Close - context deadline exceeded: %v", ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Close error: %v", e)
		return e
	case <-waitChan:
		log.Debug("Closed all repositories with success")
	}
	return nil
}

func (c *Controller) closeJobsCreator(ctx context.Context, wg *sync.WaitGroup) (<-chan service.Closer, error) {
	if len(c.Services) == 0 {
		return nil, errors.NewDetf(ClassRepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan service.Closer)
	go func() {
		defer close(out)

		for _, repo := range c.Services {
			closer, isCloser := repo.(service.Closer)
			if !isCloser {
				continue
			}
			wg.Add(1)
			select {
			case out <- closer:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (c *Controller) closeRepo(ctx context.Context, repo service.Closer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := repo.Close(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}
