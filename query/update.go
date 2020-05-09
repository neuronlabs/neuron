package query

import (
	"context"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// Update patches all selected values for given scope value.
func (s *Scope) Update(ctx context.Context) (modelsAffected int64, err error) {
	startTS := s.DB().Controller().Now()
	if log.Level().IsAllowed(log.LevelDebug2) {
		log.Debug2f(s.logFormat("Update %s begins."), s.mStruct.Collection())
	}
	if len(s.mStruct.Fields()) == 0 {
		// A model may don't have any fields - it might be only a primary key or/with relations.
		log.Debugf(s.logFormat("provided model:'%s' have no fields to update"), s.mStruct)
		return 0, errors.Newf(class.QueryNothingToDo, "model: '%s' have no fields to update", s.mStruct.String())
	}

	// If any filter is applied use it as update query.
	if len(s.Filters) != 0 {
		modelsAffected, err = s.updateFiltered(ctx)
	} else {
		// Otherwise update all models from the scope values.
		modelsAffected, err = s.updateModels(ctx)
	}
	if err != nil {
		if log.Level().IsAllowed(log.LevelDebug2) {
			log.Debug2f(s.logFormat("Update %s finished with error: '%v' in: '%s' affecting: '%d' models."), s.mStruct.Collection(), time.Since(startTS), err, modelsAffected)
		}
		return modelsAffected, err
	}
	if log.Level().IsAllowed(log.LevelDebug2) {
		log.Debug2f(s.logFormat("Update %s finished in: '%s' affecting: '%d' models."), s.mStruct.Collection(), time.Since(startTS), modelsAffected)
	}
	return modelsAffected, nil
}

func (s *Scope) updateModels(ctx context.Context) (int64, error) {
	if len(s.Models) == 0 {
		log.Debug(s.logFormat("provided empty models slice to update"))
		return 0, errors.New(class.QueryNoModelValues, "no values provided to update")
	}

	// Get models Updater repository.
	updater, isUpdater := s.repository().(Updater)
	if !isUpdater {
		log.Errorf("Repository for current model: '%s' doesn't support Update method", s.Struct().Type().Name())
		return 0, errors.NewDetf(class.RepositoryNotImplements, "models: '%s' repository doesn't implement Updater interface", s.mStruct)
	}

	// Execute before update hook if model implements BeforeUpdater.
	for i, model := range s.Models {
		beforeUpdater, ok := model.(BeforeUpdater)
		if !ok {
			// If one model is not a before updater - break the loop faster.
			if log.Level().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement BeforeUpdater interface.", s.mStruct)
			}
			break
		}
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f(s.logFormat("Executing model[%d] BeforeUpdate hook"), i)
		}
		if err := beforeUpdater.BeforeUpdate(ctx, s.db); err != nil {
			return 0, err
		}
	}

	// If the fieldset is already is provided by the user don't create batch field sets.
	if len(s.Fieldset) == 0 {
		// If given model has zero value 'UpdatedAt' field set it to current timestamp.
		updatedAt, hasUpdatedAt := s.mStruct.UpdatedAt()
		tsNow := s.DB().Controller().Now()
		s.ModelsFieldsets = make([]mapping.FieldSet, len(s.Models))
		for i, model := range s.Models {
			fielder, ok := model.(mapping.Fielder)
			if !ok {
				return 0, errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.mStruct.String())
			}

			// Add all non zero fields to the batch fieldset.
			for _, field := range s.mStruct.Fields() {
				isZero, err := fielder.IsFieldZero(field)
				if err != nil {
					return 0, err
				}
				if isZero {
					switch {
					case field.IsPrimary():
						return 0, errors.Newf(class.QueryInvalidModels, "cannot update model at: '%d' index. The primary key field have zero value.", i)
					case hasUpdatedAt && field == updatedAt:
						if log.Level().IsAllowed(log.LevelDebug3) {
							log.Debug3f(s.logFormat("model[%d], setting updated at field to: '%s'"), i, tsNow)
						}
						if err = fielder.SetFieldValue(field, tsNow); err != nil {
							return 0, err
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

	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f(s.logFormat("Update: '%d' models"), len(s.Models))
	}
	modelsAffected, err := updater.Update(ctx, s)
	if err != nil {
		log.Debugf(s.logFormat("Update failed: '%s'"), err)
		return 0, err
	}

	// Execute after update hook if model implements AfterUpdater.
	for i, model := range s.Models {
		afterUpdater, ok := model.(AfterUpdater)
		if !ok {
			// If one model is not a AfterUpdater - break the loop faster.
			if log.Level().IsAllowed(log.LevelDebug3) {
				log.Debug3f("Model: '%s' doesn't implement AfterUpdater interface", s.mStruct)
			}
			break
		}
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f(s.logFormat("Executing model[%d] AfterUpdate hook"), i)
		}
		if err = afterUpdater.AfterUpdate(ctx, s.db); err != nil {
			return 0, err
		}
	}
	return modelsAffected, nil
}

func (s *Scope) updateFiltered(ctx context.Context) (int64, error) {
	switch len(s.Models) {
	case 0:
		return 0, errors.New(class.QueryNoModelValues, "no values provided to update")
	case 1:
		// Only a single model value is allowed to update with the filters.
	default:
		return 0, errors.New(class.QueryNoModelValues, "too many values for the query ")
	}
	if len(s.mStruct.Fields()) == 1 {
		return 0, errors.New(class.QueryNotValid, "cannot update a model without any fields")
	}
	// The first model would be used to change the values.
	model := s.Models[0]
	// Check if the model implements any hooks - if it is required to find all model values and batch update them.
	_, requireFind := model.(BeforeUpdater)
	if !requireFind {
		_, requireFind = model.(AfterUpdater)
		if !requireFind {
			_, requireFind = s.mStruct.UpdatedAt()
		}
	}
	if !model.IsPrimaryKeyZero() || s.Fieldset.Contains(s.mStruct.Primary()) {
		return 0, errors.New(class.QueryInvalidField, "cannot update filtered models with the primary model in the fieldset")
	}
	// If the model implements before or after update hooks we need to find all given models and then update their values.
	if requireFind {
		return s.updateFilteredWithFind(ctx, model)
	}
	// If the query have no selected fields to update - find all non zero model fields.
	if len(s.Fieldset) == 0 {
		// Check if the model implements Fielder - otherwise the query should be based on the
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return 0, errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.mStruct)
		}
		// Select all non zero, not primary fields from the 'model'.
		for _, field := range s.mStruct.Fields() {
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
			s.Fieldset = append(s.Fieldset, field)
		}
	}
	updatedAt, hasUpdatedAt := s.mStruct.UpdatedAt()
	if hasUpdatedAt {
		// Check if the model have already selected
		if !s.Fieldset.Contains(updatedAt) {
			s.Fieldset = append(s.Fieldset, updatedAt)
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return 0, errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.mStruct.String())
		}

		if err := fielder.SetFieldValue(updatedAt, s.DB().Controller().Now()); err != nil {
			return 0, err
		}
	}

	if len(s.Fieldset) == 0 {
		return 0, errors.New(class.QueryNothingToDo, "nothing to update - only primary key field in the fieldset")
	}
	// Reduce relationship filters.
	if err := s.reduceRelationshipFilters(ctx); err != nil {
		return 0, err
	}
	updater, ok := s.repository().(Updater)
	if !ok {
		return 0, errors.Newf(class.RepositoryNotImplements, "models: '%s' repository doesn't implement Updater interface", s.mStruct)
	}
	return updater.Update(ctx, s)
}

func (s *Scope) updateFilteredWithFind(ctx context.Context, model mapping.Model) (int64, error) {
	// Find all models that matches given query filters.
	findFieldset := mapping.FieldSet{s.mStruct.Primary()}
	fielder, ok := model.(mapping.Fielder)
	if !ok {
		return 0, errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.mStruct)
	}

	if len(s.Fieldset) == 0 {
		for _, field := range s.mStruct.Fields() {
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
		findFieldset = append(findFieldset, s.Fieldset...)
	}
	// Check if there are any field other than primary key in the fieldset.
	// This would mean which fields were marked to update.
	if len(findFieldset) == 1 {
		return 0, errors.Newf(class.QueryNothingToDo, "nothing to update - only primary key field in the fieldset")
	}
	// Find all models for given query.
	findQuery := s.DB().QueryCtx(ctx, s.mStruct).Select(findFieldset...)
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
		log.Debug2(s.logFormat("No models were found to update in the query"))
		return 0, nil
	}
	// Get update fieldset that without first primary key value.
	updateFields := findFieldset[1:]
	// AddModel all fieldset fields from the 'model' to each result model.
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
	updatedAt, hasUpdatedAt := s.mStruct.UpdatedAt()
	// If given model has a 'UpdatedAt' field and it is not in the fieldset, set it's value for all models.
	if hasUpdatedAt && !findFieldset.Contains(updatedAt) {
		tsNow := s.DB().Controller().Now()
		for i, singleModel := range models {
			fielder, ok = singleModel.(mapping.Fielder)
			if !ok {
				return 0, errors.Newf(class.ModelNotImplements, "singleModel: '%s' doesn't implement Fielder interface", s.mStruct)
			}
			if log.Level().IsAllowed(log.LevelDebug3) {
				log.Debug3f(s.logFormat("model[%d], setting updated at field to: '%s'"), i, tsNow)
			}
			if err = fielder.SetFieldValue(updatedAt, tsNow); err != nil {
				return 0, err
			}
		}
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3f(s.logFormat("adding field: '%s' to the fieldset"), updatedAt)
		}
		findFieldset = append(findFieldset, updatedAt)
	}
	// Update all result models with the fieldset from the find query.
	s.Models = models
	s.Fieldset = findFieldset
	s.clearFilters()
	return s.updateModels(ctx)
}
