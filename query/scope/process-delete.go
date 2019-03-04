package scope

import (
	"context"
	"github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/uni-db"
)

var (
	deleteProcess Process = Process{
		Name: "whiz:delete",
		Func: deleteFunc,
	}

	beforeDelete Process = Process{
		Name: "whiz:hook_before_delete",
		Func: beforeDeleteFunc,
	}

	afterDelete Process = Process{
		Name: "whiz:hook_after_delete",
		Func: afterDeleteFunc,
	}

	deleteForeignRelationships Process = Process{
		Name: "whiz:delete_foreign_relationships",
		Func: deleteForeignRelationshipsFunc,
	}
)

func deleteFunc(s *Scope) error {
	var c *controller.Controller = (*controller.Controller)(s.Controller())
	repo, ok := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))
	if !ok {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return ErrNoRepositoryFound
	}

	dRepo, ok := repo.(deleter)
	if !ok {
		log.Warningf("Repository for model: '%v' doesn't implement Deleter interface", s.Struct().Type().Name())
		return ErrNoDeleterFound
	}

	// do the delete operation
	if err := dRepo.Delete(s); err != nil {
		return err
	}

	return nil
}

// processExtractPrimaries extracts the primary field values and set as the primary filters for
// the current scope
func processExtractPrimaries(s *Scope) error {
	return nil
}

func beforeDeleteFunc(s *Scope) error {
	beforeDeleter, ok := s.Value.(wBeforeDeleter)
	if !ok {
		return nil
	}

	if err := beforeDeleter.HBeforeDelete(s); err != nil {
		return err
	}
	return nil
}

func afterDeleteFunc(s *Scope) error {
	afterDeleter, ok := s.Value.(wAfterDeleter)
	if !ok {
		return nil
	}

	if err := afterDeleter.HAfterDeleter(s); err != nil {
		return err
	}
	return nil
}

func deleteForeignRelationshipsFunc(s *Scope) error {
	iScope := (*scope.Scope)(s)

	// set cancel context
	ctx, cancel := context.WithCancel(s.Context())
	defer cancel()
	for _, field := range iScope.Struct().Fields() {
		if !field.IsRelationship() {
			continue
		}

		rel := field.Relationship()
		switch rel.Kind() {
		case models.RelBelongsTo:
			continue
		case models.RelHasOne, models.RelHasMany:
			if rel.Sync() != nil && !*rel.Sync() {
				continue
			}

			// clearScope clears the foreign key values for the relationships
			clearScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

			clearScope.AddSelectedField(rel.ForeignKey())

			for _, prim := range iScope.PrimaryFilters() {

				err := clearScope.AddFilterField(
					filters.NewFilter(
						rel.ForeignKey(),
						prim.Values()...,
					),
				)
				if err != nil {
					log.Debugf("Adding Relationship's foreign key failed: %v", err)
					return err
				}
			}

			err := (*Scope)(clearScope).Patch()
			if err != nil {
				switch e := err.(type) {
				case *unidb.Error:
					if e.Compare(unidb.ErrNoResult) {
						continue
					}
					return err
				default:
					return err
				}
			}
		case models.RelMany2Many:
			if rel.Sync() != nil && !*rel.Sync() {
				continue
			}

			if rel.BackreferenceField() == nil {
				continue
			}

			clearScope := scope.NewRootScopeWithCtx(ctx, rel.Struct())

			clearScope.NewValueSingle()
			clearScope.AddSelectedField(rel.ForeignKey())

			innerFilter := filters.NewFilter(rel.BackreferenceField().Relationship().Struct().PrimaryField())

			f := filters.NewFilter(
				rel.BackreferenceField(),
			)
			f.AddNestedField(innerFilter)

			for _, prim := range iScope.PrimaryFilters() {
				innerFilter.AddValues(prim.Values()...)
			}
			if err := clearScope.AddFilterField(f); err != nil {
				log.Debugf("Deleting relationship: '%s' AddFilterField failed: %v ", field.Name(), err)
				return err
			}

			if err := (*Scope)(clearScope).Patch(); err != nil {
				switch e := err.(type) {
				case *unidb.Error:
					if e.Compare(unidb.ErrNoResult) {
						continue
					}
					return err
				default:
					return err
				}
				return err
			}

		}

	}
	return nil
}
