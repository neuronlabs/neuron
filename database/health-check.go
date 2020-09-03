package database

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

// HealthCheck checks all repositories health.
func HealthCheck(ctx context.Context, db DB) (*repository.HealthResponse, error) {
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
	jobs, err := healthCheckJobsCreator(ctx, db.(repositoryMapper), wg)
	if err != nil {
		return nil, err
	}

	hc := make(chan *repository.HealthResponse)
	errChan := make(chan error)
	for job := range jobs {
		healthCheckRepo(ctx, job, wg, errChan, hc)
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
		return healthCheckAll(ctx, hc), nil
	}
}

func healthCheckJobsCreator(ctx context.Context, mapper repositoryMapper, wg *sync.WaitGroup) (<-chan repository.HealthChecker, error) {
	if len(mapper.mapper().Repositories) == 0 {
		return nil, errors.WrapDetf(ErrRepositoryNotFound, "no repositories found for the model")
	}
	out := make(chan repository.HealthChecker)
	go func() {
		defer close(out)

		for _, repo := range mapper.mapper().Repositories {
			healthChecker, isHealthChecker := repo.(repository.HealthChecker)
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

func healthCheckAll(ctx context.Context, hc <-chan *repository.HealthResponse) *repository.HealthResponse {
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

func healthCheckRepo(ctx context.Context, repo repository.HealthChecker, wg *sync.WaitGroup, errChan chan<- error, hc chan<- *repository.HealthResponse) {
	go func() {
		defer wg.Done()
		resp, err := repo.HealthCheck(ctx)
		if err != nil {
			errChan <- err
			return
		}
		hc <- resp
	}()
}
