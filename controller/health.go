package controller

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/service"
)

// HealthCheck checks all repositories health.
func (c *Controller) HealthCheck(ctx context.Context) (*service.HealthResponse, error) {
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

	hc := make(chan *service.HealthResponse)
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

func (c *Controller) healthCheckJobsCreator(ctx context.Context, wg *sync.WaitGroup) (<-chan service.HealthChecker, error) {
	if len(c.Services) == 0 {
		return nil, errors.NewDetf(ClassRepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan service.HealthChecker)
	go func() {
		defer close(out)

		for _, repo := range c.Services {
			healthChecker, isHealthChecker := repo.(service.HealthChecker)
			if !isHealthChecker {
				continue
			}
			wg.Add(1)
			select {
			case out <- healthChecker:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (c *Controller) healthCheckAll(ctx context.Context, hc <-chan *service.HealthResponse) *service.HealthResponse {
	commonHealthCheck := &service.HealthResponse{
		Status: service.StatusPass,
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

func (c *Controller) healthCheckRepo(ctx context.Context, repo service.HealthChecker, wg *sync.WaitGroup, errc chan<- error, hc chan<- *service.HealthResponse) {
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
