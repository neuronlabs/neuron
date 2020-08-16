package controller

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// Dialer is an interface that starts the connection for the store.
type Dialer interface {
	Dial(ctx context.Context) error
}

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
	errChan := make(chan error)
	// dial to all repositories
	for job := range jobs {
		c.dial(ctx, job, wg, errChan)
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
	case e := <-errChan:
		log.Errorf("Dial error: %v", e)
		return e
	case <-waitChan:
		log.Debug("Successful dial to all repositories")
	}
	return nil
}

func (c *Controller) dialJobsCreator(ctx context.Context, wg *sync.WaitGroup) (<-chan Dialer, error) {
	if len(c.Repositories) == 0 {
		return nil, errors.WrapDetf(ErrRepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan Dialer)
	go func() {
		defer close(out)

		for _, repo := range c.Repositories {
			dialer, isDialer := repo.(Dialer)
			if !isDialer {
				continue
			}
			wg.Add(1)
			select {
			case out <- dialer:
			case <-ctx.Done():
				return
			}
		}
		// Iterate over stores and try to establish connection.
		for _, s := range c.Stores {
			dialer, isDialer := s.(Dialer)
			if !isDialer {
				continue
			}
			wg.Add(1)
			select {
			case out <- dialer:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (c *Controller) dial(ctx context.Context, dialer Dialer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := dialer.Dial(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}
