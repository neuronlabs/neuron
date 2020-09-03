package database

import (
	"context"
	"sync"
	"time"

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
	jobs := healthCheckJobsCreator(ctx, db.(repositoryMapper), wg)
	if jobs == nil {
		return &repository.HealthResponse{Status: repository.StatusPass}, nil
	}

	hc := make(chan *repository.HealthResponse, len(jobs))
	errChan := make(chan error)
	for job := range jobs {
		healthCheckRepo(ctx, job, wg, errChan, hc)
	}

	go func() {
		wg.Wait()
		close(hc)
	}()

	var responses []*repository.HealthResponse
lp:
	for {
		select {
		case <-ctx.Done():
			log.Errorf("HealthCheck - context deadline exceeded: %v", ctx.Err())
			return nil, ctx.Err()
		case e := <-errChan:
			log.Debugf("HealthCheck error: %v", e)
			return nil, e
		case resp, ok := <-hc:
			if !ok {
				break lp
			}
			responses = append(responses, resp)
		}
	}
	commonHealthCheck := &repository.HealthResponse{
		Status: repository.StatusPass,
	}
	for _, healthCheck := range responses {
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
	}
	return commonHealthCheck, nil
}

func healthCheckJobsCreator(ctx context.Context, mapper repositoryMapper, wg *sync.WaitGroup) <-chan repository.HealthChecker {
	if len(mapper.mapper().Repositories) == 0 {
		return nil
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
	return out
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
