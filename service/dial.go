package service

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/log"
)

// Dialer is an interface that starts the connection for the store.
type Dialer interface {
	Dial(ctx context.Context) error
}

// Dial establish connection for all dialers in the service.
func (s *Service) Dial(ctx context.Context) error {
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

	jobs := s.dialJobsCreator(ctx, wg)
	// Create error channel.
	errChan := make(chan error)
	// Dial to all repositories.
	for job := range jobs {
		s.dial(ctx, job, wg, errChan)
	}
	// Create wait group channel finish function.
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

func (s *Service) dialJobsCreator(ctx context.Context, wg *sync.WaitGroup) <-chan Dialer {
	out := make(chan Dialer)
	go func() {
		defer close(out)

		if dialer, ok := s.DB.(Dialer); ok {
			wg.Add(1)
			out <- dialer
		}
		// Iterate over stores and try to establish connection.
		for _, s := range s.Stores {
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

		// Iterate over file stores.
		for _, s := range s.FileStores {
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
	return out
}

func (s *Service) dial(ctx context.Context, dialer Dialer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := dialer.Dial(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}
