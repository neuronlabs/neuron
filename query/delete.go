package query

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// Delete deletes the values provided in the query's scope.
func (s *Scope) Delete(ctx context.Context) (int64, error) {
	if len(s.Filters) > 0 {
		return s.deleteFiltered(ctx)
	}
	return s.deleteModels(ctx)
}

func (s *Scope) deleteFiltered(ctx context.Context) (int64, error) {
	// If the model has 'DeletedAt' timestamp - do the soft delete - updates models with the 'DeletedAt' timestamp.
	model := mapping.NewModel(s.mStruct)
	// For models that implements delete hooks the function needs to find all models that matches given filters so
	// that the hook could be executed.
	_, hasHooks := model.(BeforeDeleter)
	if !hasHooks {
		_, hasHooks = model.(AfterDeleter)
	}
	if hasHooks {
		return s.deleteFilteredWithHooks(ctx)
	}

	// Reduce relationship filters into scope models filters.
	if err := s.reduceRelationshipFilters(ctx); err != nil {
		return 0, err
	}

	// Check if the model uses soft delete.
	_, hasDeletedAt := s.mStruct.DeletedAt()
	if hasDeletedAt {
		return s.softDeleteFiltered(ctx)
	}

	// If the model doesn't implement any of the hooks and doesn't use soft delete just Delete given query scope.
	deleter, isDeleter := s.repository().(Deleter)
	if !isDeleter {
		return 0, errors.Newf(repository.ClassNotImplements, "repository for model: '%s' doesn't implement Deleter interface", s.mStruct)
	}
	return deleter.Delete(ctx, s)
}

func (s *Scope) deleteFilteredWithHooks(ctx context.Context) (int64, error) {
	query := s.db.QueryCtx(ctx, s.mStruct)
	for _, filter := range s.Filters {
		query.Filter(filter)
	}
	models, err := query.Find()
	if err != nil {
		return 0, err
	}
	s.Models = models
	s.clearFilters()

	return s.deleteModels(ctx)
}

func (s *Scope) deleteModels(ctx context.Context) (modelsAffected int64, err error) {
	if len(s.Models) == 0 {
		return 0, errors.Newf(ClassNoModels, "no models provided to delete")
	}

	// If the model has 'DeletedAt' timestamp - do the soft delete - updates models with the 'DeletedAt' timestamp.
	_, hasDeletedAt := s.mStruct.DeletedAt()
	if hasDeletedAt {
		return s.softDeleteModels(ctx)
	}

	deleter, isDeleter := s.repository().(Deleter)
	if !isDeleter {
		return 0, errors.Newf(repository.ClassNotImplements, "repository for model: '%s' doesn't implement Deleter interface", s.mStruct)
	}

	primaries := make([]interface{}, len(s.Models))
	// Otherwise get all models primary keys and set it as the filter for the Delete method.
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.New(ClassInvalidModels, "one of the models have primary key with zero value")
		}
		primaries = append(primaries, model.GetPrimaryKeyValue())
	}
	// AddModel primary key filter that matches all primaries from the models
	s.Filters = append(s.Filters, NewFilterField(s.mStruct.Primary(), OpIn, primaries...))

	// Execute BeforeDelete hook if the model implements BeforeDeleter interface.
	if err = s.hookBeforeDelete(ctx); err != nil {
		return 0, err
	}

	modelsAffected, err = deleter.Delete(ctx, s)
	if err != nil {
		return 0, err
	}

	// Execute AfterDelete hook if the model implements AfterDeleter interface.
	if err = s.hookAfterDelete(ctx); err != nil {
		return modelsAffected, err
	}

	return modelsAffected, nil
}

// softDeleteFiltered is the function that updates all the models defined by provided filters and sets the
// 'DeletedAt' timestamp to current time.
func (s *Scope) softDeleteFiltered(ctx context.Context) (int64, error) {
	updater, isUpdater := s.repository().(Updater)
	if isUpdater {
		return 0, errors.Newf(repository.ClassNotImplements, "repository for model: '%s' doesn't implement Updater interface", s.mStruct)
	}

	// Create update model with 'DeletedAt' field set with current timestamp.
	updateModel := mapping.NewModel(s.mStruct)
	deletedAt, _ := s.mStruct.DeletedAt()
	fielder, isFielder := updateModel.(mapping.Fielder)
	if !isFielder {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.mStruct)
	}
	if err := fielder.SetFieldValue(deletedAt, s.db.Controller().Now()); err != nil {
		return 0, err
	}

	// Only the Primary Key and DeletedAt fields should be selected.
	s.Fieldset = mapping.FieldSet{s.mStruct.Primary(), deletedAt}
	s.Models = []mapping.Model{updateModel}
	modelsAffected, err := updater.Update(ctx, s)
	if err != nil {
		return 0, err
	}
	return modelsAffected, nil
}

func (s *Scope) softDeleteModels(ctx context.Context) (modelsAffected int64, err error) {
	updater, isUpdater := s.repository().(Updater)
	if !isUpdater {
		return 0, errors.Newf(repository.ClassNotImplements, "repository for model: '%s' doesn't implement Updater interface", s.mStruct)
	}

	// Soft delete should update only the DeletedAt timestamp field.
	deletedAt, _ := s.mStruct.DeletedAt()
	deletedAtTS := s.db.Controller().Now()

	// Get all primary key values and set it as the filter.
	primaries := make([]interface{}, len(s.Models))
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.New(ClassInvalidModels, "one of the models have primary key with zero value")
		}
		primaries = append(primaries, model.GetPrimaryKeyValue())
	}
	s.Filters = append(s.Filters, NewFilterField(s.mStruct.Primary(), OpIn, primaries...))

	// Create a draft update model which would be used for updating filtered soft deletes.
	updateModel := mapping.NewModel(s.mStruct)
	fielder, isFielder := updateModel.(mapping.Fielder)
	if !isFielder {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.mStruct)
	}
	if err = fielder.SetFieldValue(deletedAt, deletedAtTS); err != nil {
		return 0, err
	}

	// If the model implements any of the before or after delete hooks set the timestamps for all models.
	_, isBeforeDeleter := updateModel.(BeforeDeleter)
	_, isAfterDeleter := updateModel.(AfterDeleter)
	if isBeforeDeleter || isAfterDeleter {
		// Iterate over all models and set the 'DeletedAt' timestamp to current value.
		for _, model := range s.Models {
			fielder, ok := model.(mapping.Fielder)
			if !ok {
				return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.mStruct)
			}
			if err = fielder.SetFieldValue(deletedAt, deletedAtTS); err != nil {
				return 0, err
			}
		}
	}

	// Execute BeforeDelete hook if the model implements BeforeDeleter interface.
	if isBeforeDeleter {
		if err = s.hookBeforeDelete(ctx); err != nil {
			return 0, err
		}
	}

	// Get the models into temporary slice variable and set the models to the draft update model.
	models := s.Models
	s.Models = []mapping.Model{updateModel}
	// Only the Primary Key and DeletedAt fields should be selected.
	s.Fieldset = mapping.FieldSet{s.mStruct.Primary(), deletedAt}
	modelsAffected, err = updater.Update(ctx, s)
	if err != nil {
		return 0, err
	}

	// Execute AfterDelete hook if the model implements AfterDeleter interface.
	if isAfterDeleter {
		// Reset the models so that the hookAfterDelete method could use it.
		s.Models = models
		if err = s.hookAfterDelete(ctx); err != nil {
			return 0, err
		}
	}
	return modelsAffected, nil
}

func (s *Scope) hookBeforeDelete(ctx context.Context) (err error) {
	for _, model := range s.Models {
		beforeDeleter, isBeforeDeleter := model.(BeforeDeleter)
		if !isBeforeDeleter {
			break
		}
		if err = beforeDeleter.BeforeDelete(ctx, s.db); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func (s *Scope) hookAfterDelete(ctx context.Context) (err error) {
	for _, model := range s.Models {
		afterDeleter, isAfterDeleter := model.(AfterDeleter)
		if !isAfterDeleter {
			break
		}
		if err = afterDeleter.AfterDelete(ctx, s.db); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
