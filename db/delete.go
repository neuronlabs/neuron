package db

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// Delete deletes the values provided in the query's scope.
func Delete(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	if len(s.Filters) > 0 {
		return deleteFiltered(ctx, db, s)
	}
	return deleteModels(ctx, db, s)
}

func DeleteAll(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	return deleteFiltered(ctx, db, s)
}

func deleteFiltered(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	// If the model has 'DeletedAt' timestamp - do the soft delete - updates models with the 'DeletedAt' timestamp.
	model := mapping.NewModel(s.ModelStruct)
	// For models that implements delete hooks the function needs to find all models that matches given filters so
	// that the hook could be executed.
	_, hasHooks := model.(BeforeDeleter)
	if !hasHooks {
		_, hasHooks = model.(AfterDeleter)
	}
	if hasHooks {
		return deleteFilteredWithHooks(ctx, db, s)
	}

	// Reduce relationship filters into scope models filters.
	if err := reduceRelationshipFilters(ctx, db, s); err != nil {
		return 0, err
	}

	// Check if the model uses soft delete.
	_, hasDeletedAt := s.ModelStruct.DeletedAt()
	if hasDeletedAt {
		return softDeleteFiltered(ctx, db, s)
	}

	// If the model doesn't implement any of the hooks and doesn't use soft delete just Delete given query scope.
	repository := getRepository(db.Controller(), s)
	return repository.Delete(ctx, s)
}

func deleteFilteredWithHooks(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	q := db.QueryCtx(ctx, s.ModelStruct)
	for _, filter := range s.Filters {
		q.Filter(filter)
	}
	models, err := q.Find()
	if err != nil {
		return 0, err
	}
	s.Models = models
	s.ClearFilters()

	return deleteModels(ctx, db, s)
}

func deleteModels(ctx context.Context, db DB, s *query.Scope) (modelsAffected int64, err error) {
	if len(s.Models) == 0 {
		return 0, errors.Newf(query.ClassNoModels, "no models provided to delete")
	}

	// If the model has 'DeletedAt' timestamp - do the soft delete - updates models with the 'DeletedAt' timestamp.
	_, hasDeletedAt := s.ModelStruct.DeletedAt()
	if hasDeletedAt {
		return softDeleteModels(ctx, db, s)
	}

	primaries := make([]interface{}, len(s.Models))
	// Otherwise get all models primary keys and set it as the filter for the Delete method.
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.New(query.ClassInvalidModels, "one of the models have primary key with zero value")
		}
		primaries = append(primaries, model.GetPrimaryKeyValue())
	}
	// addModel primary key filter that matches all primaries from the models
	s.Filters = append(s.Filters, filter.New(s.ModelStruct.Primary(), filter.OpIn, primaries...))

	// Execute BeforeDelete hook if the model implements BeforeDeleter interface.
	if err = hookBeforeDelete(ctx, db, s); err != nil {
		return 0, err
	}

	modelsAffected, err = getRepository(db.Controller(), s).Delete(ctx, s)
	if err != nil {
		return 0, err
	}

	// Execute AfterDelete hook if the model implements AfterDeleter interface.
	if err = hookAfterDelete(ctx, db, s); err != nil {
		return modelsAffected, err
	}

	return modelsAffected, nil
}

// softDeleteFiltered is the function that updates all the models defined by provided filters and sets the
// 'DeletedAt' timestamp to current time.
func softDeleteFiltered(ctx context.Context, db DB, s *query.Scope) (int64, error) {
	repository := getRepository(db.Controller(), s)

	// Create update model with 'DeletedAt' field set with current timestamp.
	updateModel := mapping.NewModel(s.ModelStruct)
	deletedAt, _ := s.ModelStruct.DeletedAt()
	fielder, isFielder := updateModel.(mapping.Fielder)
	if !isFielder {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.ModelStruct)
	}
	if err := fielder.SetFieldValue(deletedAt, db.Controller().Now()); err != nil {
		return 0, err
	}

	// Only the Primary Key and DeletedAt fields should be selected.
	s.FieldSets = []mapping.FieldSet{{s.ModelStruct.Primary(), deletedAt}}
	s.Models = []mapping.Model{updateModel}
	modelsAffected, err := repository.Update(ctx, s)
	if err != nil {
		return 0, err
	}
	return modelsAffected, nil
}

func softDeleteModels(ctx context.Context, db DB, s *query.Scope) (modelsAffected int64, err error) {
	repository := getRepository(db.Controller(), s)

	// Soft delete should update only the DeletedAt timestamp field.
	deletedAt, _ := s.ModelStruct.DeletedAt()
	deletedAtTS := db.Controller().Now()

	// Get all primary key values and set it as the filter.
	primaries := make([]interface{}, len(s.Models))
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.New(query.ClassInvalidModels, "one of the models have primary key with zero value")
		}
		primaries = append(primaries, model.GetPrimaryKeyValue())
	}
	s.Filters = append(s.Filters, filter.New(s.ModelStruct.Primary(), filter.OpIn, primaries...))

	// Create a draft update model which would be used for updating filtered soft deletes.
	updateModel := mapping.NewModel(s.ModelStruct)
	fielder, isFielder := updateModel.(mapping.Fielder)
	if !isFielder {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.ModelStruct)
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
				return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", s.ModelStruct)
			}
			if err = fielder.SetFieldValue(deletedAt, deletedAtTS); err != nil {
				return 0, err
			}
		}
	}

	// Execute BeforeDelete hook if the model implements BeforeDeleter interface.
	if isBeforeDeleter {
		if err = hookBeforeDelete(ctx, db, s); err != nil {
			return 0, err
		}
	}

	// Get the models into temporary slice variable and set the models to the draft update model.
	models := s.Models
	s.Models = []mapping.Model{updateModel}
	// Only the Primary Key and DeletedAt fields should be selected.
	s.FieldSets = []mapping.FieldSet{{s.ModelStruct.Primary(), deletedAt}}
	modelsAffected, err = repository.Update(ctx, s)
	if err != nil {
		return 0, err
	}

	// Execute AfterDelete hook if the model implements AfterDeleter interface.
	if isAfterDeleter {
		// Reset the models so that the hookAfterDelete method could use it.
		s.Models = models
		if err = hookAfterDelete(ctx, db, s); err != nil {
			return 0, err
		}
	}
	return modelsAffected, nil
}

func hookBeforeDelete(ctx context.Context, db DB, s *query.Scope) (err error) {
	for _, model := range s.Models {
		beforeDeleter, isBeforeDeleter := model.(BeforeDeleter)
		if !isBeforeDeleter {
			break
		}
		if err = beforeDeleter.BeforeDelete(ctx, db); err != nil {
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

func hookAfterDelete(ctx context.Context, db DB, s *query.Scope) (err error) {
	for _, model := range s.Models {
		afterDeleter, isAfterDeleter := model.(AfterDeleter)
		if !isAfterDeleter {
			break
		}
		if err = afterDeleter.AfterDelete(ctx, db); err != nil {
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
