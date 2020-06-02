package orm

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Update updates the models with selected fields or a single model with provided filters.
func Update(ctx context.Context, db DB, s *query.Scope) (modelsAffected int64, err error) {
	startTS := db.Controller().Now()
	if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
		log.Debug2f(logFormat(s, "Update %s begins."), s.ModelStruct.Collection())
	}

	// If any filter is applied use it as update query.
	if len(s.Filters) != 0 {
		modelsAffected, err = updateFiltered(ctx, db, s)
	} else {
		// Otherwise update all models from the scope values.
		modelsAffected, err = updateModels(ctx, db, s)
	}

	if err != nil {
		if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
			log.Debug2f(logFormat(s, "Update %s finished with error: '%v' in: '%s' affecting: '%d' models."), s.ModelStruct.Collection(), time.Since(startTS), err, modelsAffected)
		}
		return modelsAffected, err
	}
	if log.CurrentLevel().IsAllowed(log.LevelDebug2) {
		log.Debug2f(logFormat(s, "Update %s finished in: '%s' affecting: '%d' models."), s.ModelStruct.Collection(), time.Since(startTS), modelsAffected)
	}
	return modelsAffected, nil
}

// UpdateAll updates all the models in the repository, for the single model in the scope.
func UpdateAll(ctx context.Context, db DB, s *query.Scope) (modelsAffected int64, err error) {
	return updateFiltered(ctx, db, s)
}
func updateModels(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	if len(s.Models) == 0 {
		log.Debug(logFormat(s, "provided empty models slice to update"))
		return 0, errors.New(query.ClassNoModels, "no values provided to update")
	}

	// Get models Updater repository.
	updater := getRepository(db.Controller(), s)
	// Execute before update hook if model implements BeforeUpdater.
	for i, model := range s.Models {
		beforeUpdater, ok := model.(BeforeUpdater)
		if !ok {
			// If one model is not a before updater - break the loop faster.
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement BeforeUpdater interface.", s.ModelStruct)
			}
			break
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "Executing model[%d] BeforeUpdate hook"), i)
		}
		if err := beforeUpdater.BeforeUpdate(ctx, db); err != nil {
			return 0, err
		}
	}

	// If the fieldset is already is provided by the user don't create batch field sets.
	if len(s.FieldSet) == 0 {
		// If given model has zero value 'UpdatedAt' field set it to current timestamp.
		updatedAt, hasUpdatedAt := s.ModelStruct.UpdatedAt()
		tsNow := db.Controller().Now()
		s.BulkFieldSets = &query.BulkFieldSet{Indices: map[string][]int{}}
		for i, model := range s.Models {
			fieldSet := mapping.FieldSet{}
			fielder, ok := model.(mapping.Fielder)
			if !ok {
				return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct.String())
			}

			// Add all non zero fields to the batch fieldset.
			for _, field := range s.ModelStruct.Fields() {
				isZero, err := fielder.IsFieldZero(field)
				if err != nil {
					return 0, err
				}
				if isZero {
					switch {
					case field.IsPrimary():
						return 0, errors.Newf(query.ClassInvalidModels, "cannot update model at: '%d' index. The primary key field have zero value.", i)
					case hasUpdatedAt && field == updatedAt:
						if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
							log.Debug3f(logFormat(s, "model[%d], setting updated at field to: '%s'"), i, tsNow)
						}
						if err = fielder.SetFieldValue(field, tsNow); err != nil {
							return 0, err
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
			s.BulkFieldSets.AddIndex(fieldSet, i)
		}
	}

	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f(logFormat(s, "Update: '%d' models"), len(s.Models))
	}

	modelsAffected, err := updater.Update(ctx, s)
	if err != nil {
		log.Debugf(logFormat(s, "Update failed: '%s'"), err)
		return 0, err
	}
	// Affected models number must match the number of models updated.
	if modelsAffected != int64(len(s.Models)) {
		switch len(s.Models) {
		case 1:
			return 0, errors.Newf(query.ClassNoResult, "model with id: '%v' doesn't exists", s.Models[0].GetPrimaryKeyValue())
		default:
			return 0, errors.New(query.ClassNoResult, "one or more of provided models doesn't exists")
		}
	}

	// Execute after update hook if model implements AfterUpdater.
	for i, model := range s.Models {
		afterUpdater, ok := model.(AfterUpdater)
		if !ok {
			// If one model is not a AfterUpdater - break the loop faster.
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement AfterUpdater interface", s.ModelStruct)
			}
			break
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "Executing model[%d] AfterUpdate hook"), i)
		}
		if err = afterUpdater.AfterUpdate(ctx, db); err != nil {
			return 0, err
		}
	}
	return modelsAffected, nil
}

func updateFiltered(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	switch len(s.Models) {
	case 0:
		return 0, errors.New(query.ClassNoModels, "no values provided to update")
	case 1:
		// Only a single model value is allowed to update with the filters.
	default:
		return 0, errors.New(query.ClassInvalidModels, "too many values for the query ")
	}
	if len(s.ModelStruct.Fields()) == 1 {
		return 0, errors.New(query.ClassInvalidInput, "cannot update a model without any fields")
	}
	// The first model would be used to change the values.
	model := s.Models[0]
	// Check if the model implements any hooks - if it is required to find all model values and batch update them.
	_, requireFind := model.(BeforeUpdater)
	if !requireFind {
		_, requireFind = model.(AfterUpdater)
		if !requireFind {
			_, requireFind = s.ModelStruct.UpdatedAt()
		}
	}
	if !model.IsPrimaryKeyZero() || s.FieldSet.Contains(s.ModelStruct.Primary()) {
		return 0, errors.New(query.ClassInvalidField, "cannot update filtered models with the primary model in the fieldset")
	}
	// If the model implements before or after update hooks we need to find all given models and then update their values.
	if requireFind {
		return updateFilteredWithFind(ctx, db, s, model)
	}
	// If the query have no selected fields to update - find all non zero model fields.
	if len(s.FieldSet) == 0 {
		// Check if the model implements Fielder - otherwise the query should be based on the
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
		}
		// Select all non zero, not primary fields from the 'model'.
		for _, field := range s.ModelStruct.Fields() {
			if field.IsPrimary() {
				continue
			}
			isZero, err := fielder.IsFieldZero(field)
			if err != nil {
				return 0, err
			}
			if isZero {
				continue
			}
			s.FieldSet = append(s.FieldSet, field)
		}
	}
	updatedAt, hasUpdatedAt := s.ModelStruct.UpdatedAt()
	if hasUpdatedAt {
		// Check if the model have already selected
		if !s.FieldSet.Contains(updatedAt) {
			s.FieldSet = append(s.FieldSet, updatedAt)
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct.String())
		}

		if err := fielder.SetFieldValue(updatedAt, db.Controller().Now()); err != nil {
			return 0, err
		}
	}

	if len(s.FieldSet) == 0 {
		return 0, errors.New(query.ClassNoModels, "nothing to update - only primary key field in the fieldset")
	}
	// Reduce relationship filters.
	if err := reduceRelationshipFilters(ctx, db, s); err != nil {
		return 0, err
	}
	return getRepository(db.Controller(), s).Update(ctx, s)
}

func updateFilteredWithFind(ctx context.Context, db DB, s *query.Scope, model mapping.Model) (int64, error) {
	// Find all models that matches given query filters.
	findFieldset := mapping.FieldSet{s.ModelStruct.Primary()}
	fielder, ok := model.(mapping.Fielder)
	if !ok {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
	}

	if len(s.FieldSet) == 0 {
		for _, field := range s.ModelStruct.Fields() {
			if field.IsPrimary() {
				// Primary key is already included in the fieldset.
				continue
			}
			isZero, err := fielder.IsFieldZero(field)
			if err != nil {
				return 0, err
			}
			if isZero {
				continue
			}
			findFieldset = append(findFieldset, field)
		}
	} else {
		findFieldset = append(findFieldset, s.FieldSet...)
	}
	// Check if there are any field other than primary key in the fieldset.
	// This would mean which fields were marked to update.
	if len(findFieldset) == 1 {
		return 0, errors.Newf(query.ClassNoModels, "nothing to update - only primary key field in the fieldset")
	}
	// Find all models for given query.
	findQuery := db.QueryCtx(ctx, s.ModelStruct).Select(findFieldset...)
	for _, filter := range s.Filters {
		findQuery.Filter(filter)
	}
	if s.Pagination != nil {
		findQuery.Scope().Pagination = s.Pagination
	}
	models, err := findQuery.Find()
	if err != nil {
		return 0, err
	}
	// If no models were found return without error.
	if len(models) == 0 {
		log.Debug2(logFormat(s, "No models were found to update in the query"))
		return 0, nil
	}
	// Get update fieldset that without first primary key value.
	updateFields := findFieldset[1:]
	// addModel all fieldset fields from the 'model' to each result model.
	for _, resultModel := range models {
		findModelFielder := resultModel.(mapping.Fielder)
		for _, field := range updateFields {
			fieldValue, err := fielder.GetHashableFieldValue(field)
			if err != nil {
				return 0, err
			}
			if err = findModelFielder.SetFieldValue(field, fieldValue); err != nil {
				return 0, err
			}
		}
	}
	updatedAt, hasUpdatedAt := s.ModelStruct.UpdatedAt()
	// If given model has a 'UpdatedAt' field and it is not in the fieldset, set it's value for all models.
	if hasUpdatedAt && !findFieldset.Contains(updatedAt) {
		tsNow := db.Controller().Now()
		for i, singleModel := range models {
			fielder, ok = singleModel.(mapping.Fielder)
			if !ok {
				return 0, errors.Newf(mapping.ClassModelNotImplements, "singleModel: '%s' doesn't implement Fielder interface", s.ModelStruct)
			}
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f(logFormat(s, "model[%d], setting updated at field to: '%s'"), i, tsNow)
			}
			if err = fielder.SetFieldValue(updatedAt, tsNow); err != nil {
				return 0, err
			}
		}
		if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
			log.Debug3f(logFormat(s, "adding field: '%s' to the fieldset"), updatedAt)
		}
		findFieldset = append(findFieldset, updatedAt)
	}
	// Update all result models with the fieldset from the find query.
	s.Models = models
	s.FieldSet = findFieldset
	s.ClearFilters()
	return updateModels(ctx, db, s)
}
