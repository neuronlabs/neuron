package scope

import (
	"context"
	"github.com/kucjac/uni-db"

	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

var (
	// ProcessDelete is the process that does the Repository Delete method
	ProcessDelete = &Process{
		Name: "neuron:delete",
		Func: deleteFunc,
	}

	// ProcessBeforeDelete is the Process that does the HBeforeDelete hook
	ProcessBeforeDelete = &Process{
		Name: "neuron:hook_before_delete",
		Func: beforeDeleteFunc,
	}

	// ProcessAfterDelete is the Process that does the HAfterDelete hook
	ProcessAfterDelete = &Process{
		Name: "neuron:hook_after_delete",
		Func: afterDeleteFunc,
	}

	// ProcessDeleteForeignRelationships is the Process that deletes the foreing relatioionships
	ProcessDeleteForeignRelationships = &Process{
		Name: "neuron:delete_foreign_relationships",
		Func: deleteForeignRelationshipsFunc,
	}
)

func deleteFunc(ctx context.Context, s *Scope) error {

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Warningf("Repository not found for model: %v", s.Struct().Type().Name())
		return ErrNoRepositoryFound
	}

	dRepo, ok := repo.(Deleter)
	if !ok {
		log.Warningf("Repository for model: '%v' doesn't implement Deleter interface", s.Struct().Type().Name())
		return ErrNoDeleterFound
	}

	// do the delete operation
	if err := dRepo.Delete(ctx, s); err != nil {
		return err
	}

	return nil
}

func beforeDeleteFunc(ctx context.Context, s *Scope) error {
	beforeDeleter, ok := s.Value.(BeforeDeleter)
	if !ok {
		return nil
	}

	if err := beforeDeleter.HBeforeDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func afterDeleteFunc(ctx context.Context, s *Scope) error {
	afterDeleter, ok := s.Value.(AfterDeleter)
	if !ok {
		return nil
	}

	if err := afterDeleter.HAfterDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

func deleteForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	iScope := (*scope.Scope)(s)

	// get only relationship fields
	var relationships []*models.StructField
	for _, field := range iScope.Struct().Fields() {
		if field.IsRelationship() {
			relationships = append(relationships, field)
		}
	}

	if len(relationships) > 0 {

		var results = make(chan interface{}, len(relationships))

		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Builder.RepositoryTimeout
		for _, rel := range relationships {
			if rel.Relationship().Struct().Config() == nil {
				continue
			}
			if modelRepo := rel.Relationship().Struct().Config().Repository; modelRepo != nil {
				if tm := modelRepo.MaxTimeout; tm != nil {
					if *tm > maxTimeout {
						maxTimeout = *tm
					}
				}
			}
		}

		ctx, cancel := context.WithTimeout(ctx, maxTimeout)
		defer cancel()

		// delete foreign relationships
		for _, field := range relationships {
			go deleteForeignRelationships(ctx, iScope, field, results)
		}

		var ctr int

		for {
			select {
			case <-ctx.Done():
			case v, ok := <-results:
				if !ok {
					return nil
				}
				if err, ok := v.(error); ok {
					return err
				}
				ctr++
				if ctr == len(relationships) {
					return nil
				}
			}
		}
	}
	return nil

}

func deleteForeignRelationships(ctx context.Context, iScope *scope.Scope, field *models.StructField, results chan<- interface{}) {
	var err error
	defer func() {
		if err != nil {
			results <- err
		} else {
			results <- struct{}{}
		}
	}()

	rel := field.Relationship()
	switch rel.Kind() {
	case models.RelBelongsTo:
		return
	case models.RelHasOne, models.RelHasMany:
		if rel.Sync() != nil && !*rel.Sync() {
			return
		}

		var isMany bool
		if rel.Kind() == models.RelHasMany {
			isMany = true
		}

		// clearScope clears the foreign key values for the relationships
		clearScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), isMany)
		clearScope.AddSelectedField(rel.ForeignKey())

		for _, prim := range iScope.PrimaryFilters() {

			err = clearScope.AddFilterField(filters.NewFilter(rel.ForeignKey(), prim.Values()...))
			if err != nil {
				log.Debugf("Adding Relationship's foreign key failed: %v", err)
				return
			}
		}

		if tx := (*Scope)(iScope).tx(); tx != nil {
			iScope.AddChainSubscope(clearScope)
			if err = (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
				log.Debug("BeginTx for the related scope failed: %s", err)
				return
			}
		}

		if err = (*Scope)(clearScope).PatchContext(ctx); err != nil {
			switch e := err.(type) {
			case *unidb.Error:
				if e.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			return
		}
	case models.RelMany2Many:
		if rel.Sync() != nil && !*rel.Sync() {
			return
		}

		if rel.BackreferenceField() == nil {
			return
		}

		clearScope := newScopeWithModel((*Scope)(iScope).Controller(), rel.Struct(), false)
		clearScope.AddSelectedField(rel.ForeignKey())

		innerFilter := filters.NewFilter(rel.BackreferenceField().Relationship().Struct().PrimaryField())

		f := filters.NewFilter(rel.BackreferenceField())
		f.AddNestedField(innerFilter)

		for _, prim := range iScope.PrimaryFilters() {
			innerFilter.AddValues(prim.Values()...)
		}
		if err = clearScope.AddFilterField(f); err != nil {
			log.Debugf("Deleting relationship: '%s' AddFilterField failed: %v ", field.Name(), err)
			return
		}

		if tx := (*Scope)(iScope).tx(); tx != nil {
			iScope.AddChainSubscope(clearScope)
			if err = (*Scope)(clearScope).begin(ctx, &tx.Options, false); err != nil {
				log.Debug("BeginTx for the related scope failed: %s", err)
				return
			}
		}

		if err = (*Scope)(clearScope).PatchContext(ctx); err != nil {
			switch e := err.(type) {
			case *unidb.Error:
				if e.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			return
		}

	}
}
