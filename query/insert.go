package query

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// Insert stores the values within the given scope's value repository.
func (s *Scope) Insert(ctx context.Context) (err error) {
	startTS := s.DB().Controller().Now()
	if log.Level().IsAllowed(log.LevelDebug2) {
		log.Debug2f(s.logFormat("Insert %s with %d models begins."), s.mStruct.Collection(), len(s.Models))
	}
	if len(s.Models) == 0 {
		log.Debug(s.logFormat("provided empty models slice to insert"))
		return errors.New(ClassInvalidModels, "nothing to insert")
	}

	// Check if models repository implements Inserter interface.
	inserter, isInserter := s.repository().(Inserter)
	if !isInserter {
		log.Error(s.logFormat("model's: '%s' repository doesn't implement Inserter interface"), s.mStruct)
		return errors.Newf(repository.ClassNotImplements, "a repository for model: '%s' doesn't implement Inserter interface", s.mStruct)
	}

	// Execute BeforeInsert hook if model implements BeforeInserter interface.
	for i, model := range s.Models {
		beforeInserter, ok := model.(BeforeInserter)
		if !ok {
			if log.Level().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement BeforeInserter interface.", s.mStruct)
			}
			break
		}
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f(s.logFormat("Executing model[%d] BeforeInsert hook"), i)
		}
		if err = beforeInserter.BeforeInsert(ctx, s.db); err != nil {
			log.Debugf(s.logFormat("model[%d] before insert hook failed: %v"), i, err)
			return err
		}
	}
	// Check if given model has created at and updated at fields.
	createdAt, hasCreatedAt := s.mStruct.CreatedAt()
	updatedAt, hasUpdatedAt := s.mStruct.UpdatedAt()

	// If the fieldset was not set for the query, iterate over all models, find and select all non-zero fields.
	// In case when the model has 'created at' or 'updated at' fields which has zero values - set their values to
	// current timestamp.
	//
	// If the query global fieldset was defined, this function wouldn't check or set the timestamps.
	// In this case a user is responsible for setting these timestamp fields.
	if len(s.Fieldset) == 0 {
		s.ModelsFieldsets = make([]mapping.FieldSet, len(s.Models))
		for i, model := range s.Models {
			fielder, ok := model.(mapping.Fielder)
			if !ok {
				// If the model is not a fielder let's check if a primary is not zero.
				if !model.IsPrimaryKeyZero() {
					if log.Level().IsAllowed(log.LevelDebug3) {
						log.Debug3f(s.logFormat("model[%d] adding primary field to fieldset"), i)
					}
					s.ModelsFieldsets[i] = append(s.ModelsFieldsets[i], s.mStruct.Primary())
				}
				break
			}

			for _, field := range s.mStruct.Fields() {
				isZero, err := fielder.IsFieldZero(field)
				if err != nil {
					return err
				}

				if isZero {
					// If the field is a 'created at' or 'updated at' timestamps set their zero value to current timestamp.
					switch {
					case hasCreatedAt && field == createdAt:
						if log.Level().IsAllowed(log.LevelDebug3) {
							log.Debug3f(s.logFormat("model[%d], setting created at field to: '%s'"), i, startTS)
						}
						if err = fielder.SetFieldValue(createdAt, startTS); err != nil {
							return err
						}
					case hasUpdatedAt && field == updatedAt:
						if log.Level().IsAllowed(log.LevelDebug3) {
							log.Debug3f(s.logFormat("model[%d], setting updated at field to: '%s'"), i, startTS)
						}
						if err = fielder.SetFieldValue(updatedAt, startTS); err != nil {
							return err
						}
					default:
						if log.Level().IsAllowed(log.LevelDebug3) {
							log.Debug3f(s.logFormat("model[%d], field: '%s' has zero value"), i, field)
						}
						continue
					}
				}
				if log.Level().IsAllowed(log.LevelDebug3) {
					log.Debug3f(s.logFormat("model[%d] adding field: '%s' to fieldset"), i, field)
				}
				s.ModelsFieldsets[i] = append(s.ModelsFieldsets[i], field)
			}
		}
	}

	// Execute repository Insert method.
	err = inserter.Insert(ctx, s)
	if err != nil {
		log.Debugf(s.logFormat("inserting failed: '%s'"), err)
		return err
	}

	// Execute 'AfterInsert' hook if models implements AfterInserter interface.
	for i, model := range s.Models {
		afterInserter, ok := model.(AfterInserter)
		if !ok {
			if log.Level().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement After inserter interface", s.mStruct)
			}
			return nil
		}
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f(s.logFormat("Executing model[%d] AfterInsert hook"), i)
		}
		if err = afterInserter.AfterInsert(ctx, s.db); err != nil {
			return err
		}
	}
	if log.Level().IsAllowed(log.LevelDebug2) {
		log.Debug2f(s.logFormat("Insert of %s with %d models finished in '%s'."), s.mStruct.Collection(), len(s.Models), time.Since(startTS))
	}
	return nil
}
