package db

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Insert stores the values within the given scope's value repository.
func Insert(ctx context.Context, db DB, s *query.Scope) (err error) {
	startTS := db.Controller().Now()
	if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
		log.Debug2f(logFormat(s, "Insert %s with %d models begins."), s.ModelStruct.Collection(), len(s.Models))
	}
	if len(s.Models) == 0 {
		log.Debug(logFormat(s, "provided empty models slice to insert"))
		return errors.New(query.ClassInvalidModels, "nothing to insert")
	}

	// Check if models repository implements Inserter interface.
	// Execute BeforeInsert hook if model implements BeforeInserter interface.
	for i, model := range s.Models {
		beforeInserter, ok := model.(BeforeInserter)
		if !ok {
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement BeforeInserter interface.", s.ModelStruct)
			}
			break
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "Executing model[%d] BeforeInsert hook"), i)
		}
		if err = beforeInserter.BeforeInsert(ctx, db); err != nil {
			log.Debugf(logFormat(s, "model[%d] before insert hook failed: %v"), i, err)
			return err
		}
	}

	// If the fieldset was not set for the query, iterate over all models, find and select all non-zero fields.
	// In case when the model has 'created at' or 'updated at' fields which has zero values - set their values to
	// current timestamp.
	//
	// If the query global fieldset was defined, this function wouldn't check or set the timestamps.
	// In this case a user is responsible for setting these timestamp fields.
	switch len(s.FieldSets) {
	case 0:
		s.FieldSets = make([]mapping.FieldSet, len(s.Models))
		for i := range s.Models {
			s.FieldSets[i], err = createSingleInsertFieldSet(i, startTS, s)
			if err != nil {
				return err
			}
		}
	case 1, len(s.Models):
		// Common fieldset for the insert or each fieldset per model. Do nothing.
	default:
		return errors.NewDetf(query.ClassInvalidFieldSet, "provided invalid field sets. Models len: %d, FieldSets len: %d", len(s.Models), len(s.FieldSets))
	}
	// Execute repository Insert method.
	err = getRepository(db.Controller(), s).Insert(ctx, s)
	if err != nil {
		log.Debugf(logFormat(s, "inserting failed: '%s'"), err)
		return err
	}

	// Execute 'AfterInsert' hook if models implements AfterInserter interface.
	for i, model := range s.Models {
		afterInserter, ok := model.(AfterInserter)
		if !ok {
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement After inserter interface", s.ModelStruct)
			}
			return nil
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "Executing model[%d] AfterInsert hook"), i)
		}
		if err = afterInserter.AfterInsert(ctx, db); err != nil {
			return err
		}
	}
	if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
		log.Debug2f(logFormat(s, "Insert of %s with %d models finished in '%s'."), s.ModelStruct.Collection(), len(s.Models), time.Since(startTS))
	}
	return nil
}

func createSingleInsertFieldSet(i int, startTS time.Time, s *query.Scope) (mapping.FieldSet, error) {
	model := s.Models[i]
	fieldSet := mapping.FieldSet{}
	fielder, ok := model.(mapping.Fielder)
	if !ok {
		// If the model is not a fielder let's check if a primary is not zero.
		if !model.IsPrimaryKeyZero() {
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f(logFormat(s, "model[%d] adding primary field to fieldset"), i)
			}
			fieldSet = append(fieldSet, s.ModelStruct.Primary())
		}
		return fieldSet, nil
	}

	// Check if given model has created at and updated at fields.
	createdAt, hasCreatedAt := s.ModelStruct.CreatedAt()
	updatedAt, hasUpdatedAt := s.ModelStruct.UpdatedAt()

	for _, field := range s.ModelStruct.Fields() {
		isZero, err := fielder.IsFieldZero(field)
		if err != nil {
			return nil, err
		}

		if isZero {
			// If the field is a 'created at' or 'updated at' timestamps set their zero value to current timestamp.
			switch {
			case hasCreatedAt && field == createdAt:
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f(logFormat(s, "model[%d], setting created at field to: '%s'"), i, startTS)
				}
				if err = fielder.SetFieldValue(createdAt, startTS); err != nil {
					return nil, err
				}
			case hasUpdatedAt && field == updatedAt:
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f(logFormat(s, "model[%d], setting updated at field to: '%s'"), i, startTS)
				}
				if err = fielder.SetFieldValue(updatedAt, startTS); err != nil {
					return nil, err
				}
			default:
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f(logFormat(s, "model[%d], field: '%s' has zero value"), i, field)
				}
				continue
			}
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "model[%d] adding field: '%s' to fieldset"), i, field)
		}
		fieldSet = append(fieldSet, field)
	}
	return fieldSet, nil
}
