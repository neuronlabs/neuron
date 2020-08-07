package database

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

func findIncludedRelations(ctx context.Context, db DB, s *query.Scope) (err error) {
	if len(s.IncludedRelations) == 0 {
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3(logFormat(s, "no included fields found"))
		}
		return nil
	}
	ts := time.Now()
	defer func() {
		log.Debugf(logFormat(s, "Finding included relations taken: '%s'"), time.Since(ts))
	}()

	if db.Controller().Options.SynchronousConnections {
		return findIncludedRelationsSynchronous(ctx, db, s)
	}
	return findIncludedRelationsAsynchronous(ctx, db, s)
}

func findIncludedRelationsSynchronous(ctx context.Context, db DB, s *query.Scope) (err error) {
	for _, included := range s.IncludedRelations {
		if err = findIncludedRelation(ctx, db, s, included); err != nil {
			return err
		}
	}
	return nil
}

func findIncludedRelation(ctx context.Context, db DB, s *query.Scope, included *query.IncludedRelation) (err error) {
	switch included.StructField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		err = findBelongsToRelation(ctx, db, s, included)
	case mapping.RelHasOne, mapping.RelHasMany:
		err = findHasRelation(ctx, db, s, included)
	case mapping.RelMany2Many:
		err = findManyToManyRelation(ctx, db, s, included)
	default:
		return errors.Newf(query.ClassInternal, "invalid relationship: '%s' kind: '%s'", included.StructField, included.StructField.Relationship().Kind())
	}
	return err
}

func findIncludedRelationsAsynchronous(ctx context.Context, db DB, s *query.Scope) (err error) {
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
	// Generate included relation jobs.
	jobs := includedJobCreator(ctx, s, wg)

	errChan := make(chan error, 1)
	// Execute jobs with respect to given wait group and error channel.
	for job := range jobs {
		go includedJobFind(ctx, db, s, job, wg, errChan)
	}

	// Wait channel waits until all jobs marks wait group done.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Debug2(logFormat(s, "Getting includes - context done"))
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Find include error: %v", e)
		return e
	case <-waitChan:
		if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
			log.Debug2f(logFormat(s, "Finished finding asynchronous included relations"))
		}
	}
	return nil
}

func includedJobCreator(ctx context.Context, s *query.Scope, wg *sync.WaitGroup) (jobs <-chan *query.IncludedRelation) {
	out := make(chan *query.IncludedRelation)
	go func() {
		defer close(out)

		for _, included := range s.IncludedRelations {
			wg.Add(1)
			select {
			case out <- included:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func includedJobFind(ctx context.Context, db DB, s *query.Scope, include *query.IncludedRelation, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	if err := findIncludedRelation(ctx, db, s, include); err != nil {
		errChan <- err
	}
}
