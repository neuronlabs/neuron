package service

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/log"
)

// Closer is an interface that closes all connection for given instance.
type Closer interface {
	Close(ctx context.Context) error
}

// Close closes all connection within provided context.
func (s *Service) Close(ctx context.Context) error {
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
	jobs := s.closeJobsCreator(ctx, wg)

	errChan := make(chan error)
	for job := range jobs {
		log.Debugf("Closing: %T", job)
		s.closeCloser(ctx, job, wg, errChan)
	}

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

func (s *Service) closeJobsCreator(ctx context.Context, wg *sync.WaitGroup) <-chan Closer {
	out := make(chan Closer)
	go func() {
		defer close(out)

		// Added DB closer.
		if closer, isCloser := s.DB.(Closer); isCloser {
			wg.Add(1)
			out <- closer
		}

		// Close all stores.
		for _, s := range s.Stores {
			closer, isCloser := s.(Closer)
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

		// Close all file stores.
		for _, s := range s.FileStores {
			closer, isCloser := s.(Closer)
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
	return out
}

func (s *Service) closeCloser(ctx context.Context, closer Closer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := closer.Close(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}
