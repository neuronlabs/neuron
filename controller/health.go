package controller

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/repository"
)

// HealthCheck checks all repositories health.
func (c *Controller) HealthCheck(ctx context.Context) (*repository.HealthResponse, error) {
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
	jobs, err := c.healthCheckJobsCreator(ctx, wg)
	if err != nil {
		return nil, err
	}

	hc := make(chan *repository.HealthResponse)
	errChan := make(chan error)
	for job := range jobs {
		c.healthCheckRepo(ctx, job, wg, errChan, hc)
	}

	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		waitChan <- struct{}{}
		close(hc)
	}()

	select {
	case <-ctx.Done():
		log.Errorf("HealthCheck - context deadline exceeded: %v", ctx.Err())
		return nil, ctx.Err()
	case e := <-errChan:
		log.Debugf("HealthCheck error: %v", e)
		return nil, e
	case <-waitChan:
		log.Debug("HealthCheck done for all repositories")
		return c.healthCheckAll(ctx, hc), nil
	}
}

func (c *Controller) healthCheckJobsCreator(ctx context.Context, wg *sync.WaitGroup) (<-chan repository.Repository, error) {
	if len(c.Repositories) == 0 {
		return nil, errors.NewDetf(class.RepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan repository.Repository)
	go func() {
		defer close(out)

		for _, repo := range c.Repositories {
			wg.Add(1)
			select {
			case out <- repo:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (c *Controller) healthCheckAll(ctx context.Context, hc <-chan *repository.HealthResponse) *repository.HealthResponse {
	commonHealthCheck := &repository.HealthResponse{
		Status: repository.StatusPass,
	}
	for healthCheck := range hc {
		if healthCheck.Status > commonHealthCheck.Status {
			commonHealthCheck.Status = healthCheck.Status
		}
		if healthCheck.Output != "" {
			if commonHealthCheck.Output != "" {
				commonHealthCheck.Output += ". "
			}
			commonHealthCheck.Output += healthCheck.Output
		}
		if len(healthCheck.Notes) != 0 {
			commonHealthCheck.Notes = append(commonHealthCheck.Notes, healthCheck.Notes...)
		}
		select {
		case <-ctx.Done():
			return commonHealthCheck
		default:
		}
	}
	return commonHealthCheck
}

func (c *Controller) healthCheckRepo(ctx context.Context, repo repository.Repository, wg *sync.WaitGroup, errc chan<- error, hc chan<- *repository.HealthResponse) {
	go func() {
		defer wg.Done()
		resp, err := repo.HealthCheck(ctx)
		if err != nil {
			errc <- err
			return
		}
		hc <- resp
	}()
}
