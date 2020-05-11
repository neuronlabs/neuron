package query

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// IncludedRelation is the includes information scope
// it contains the field to include from the root scope
// related subScope, and subfields to include.
type IncludedRelation struct {
	StructField       *mapping.StructField
	Fieldset          mapping.FieldSet
	IncludedRelations []*IncludedRelation
}

func (i *IncludedRelation) copy() *IncludedRelation {
	copiedIncludedField := &IncludedRelation{StructField: i.StructField}
	if i.Fieldset != nil {
		copiedIncludedField.Fieldset = make([]*mapping.StructField, len(i.Fieldset))
		for index, v := range i.Fieldset {
			copiedIncludedField.Fieldset[index] = v
		}
	}
	if i.IncludedRelations != nil {
		copiedIncludedField.IncludedRelations = make([]*IncludedRelation, len(i.IncludedRelations))
		for index, v := range i.IncludedRelations {
			copiedIncludedField.IncludedRelations[index] = v.copy()
		}
	}
	return copiedIncludedField
}

// Select sets the fieldset for given included
func (i *IncludedRelation) setFieldset(fields ...*mapping.StructField) error {
	model := i.StructField.Relationship().Struct()
	for _, field := range fields {
		// Check if the field belongs to the relationship's model.
		if field.Struct() != model {
			return errors.NewDetf(ClassInvalidField, "provided field: '%s' doesn't belong to the relationship model: '%s'", field, model)
		}
		// Check if provided field is not a relationship.
		if !field.IsField() {
			return errors.NewDetf(ClassInvalidField,
				"provided invalid field: '%s' in the fieldset of included field: '%s'", field, i.StructField.Name())
		}

		// Check if given fieldset doesn't contain this field already.
		if i.Fieldset.Contains(field) {
			return errors.NewDetf(ClassInvalidField,
				"provided field: '%s' in the fieldset of included field: '%s' is already in the included fieldset",
				field, i.StructField.Name())
		}
		i.Fieldset = append(i.Fieldset, field)
	}
	return nil
}

// Include includes 'relation' field in the scope's query results.
func (s *Scope) Include(relation *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	if !relation.IsRelationship() {
		return errors.NewDetf(ClassInvalidField,
			"included relation: '%s' is not found for the model: '%s'", relation, s.mStruct.String())
	}

	includedField := &IncludedRelation{StructField: relation}
	if len(relationFieldset) == 0 {
		includedField.Fieldset = append(mapping.FieldSet{}, relation.Relationship().Struct().Fields()...)
	} else if err := includedField.setFieldset(relationFieldset...); err != nil {
		return err
	}
	s.IncludedRelations = append(s.IncludedRelations, includedField)
	return nil
}

func (s *Scope) findIncludedRelations(ctx context.Context) (err error) {
	if len(s.IncludedRelations) == 0 {
		if log.Level().IsAllowed(log.LevelDebug3) {
			log.Debug3(s.logFormat("no included fields found"))
		}
		return nil
	}
	ts := time.Now()
	defer func() {
		log.Debugf(s.logFormat("Finding included relations taken: '%s'"), time.Since(ts))
	}()

	if s.db.Controller().Config.AsynchronousIncludes {
		return s.findIncludedRelationsAsynchronous(ctx)
	}
	return s.findIncludedRelationsSynchronous(ctx)
}

func (s *Scope) findIncludedRelationsSynchronous(ctx context.Context) (err error) {
	for _, included := range s.IncludedRelations {
		if err = s.findIncludedRelation(ctx, included); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) findIncludedRelation(ctx context.Context, included *IncludedRelation) (err error) {
	switch included.StructField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		err = s.findBelongsToRelation(ctx, included)
	case mapping.RelHasOne, mapping.RelHasMany:
		err = s.findHasRelation(ctx, included)
	case mapping.RelMany2Many:
		err = s.findManyToManyRelation(ctx, included)
	default:
		return errors.Newf(ClassInternal, "invalid relationship: '%s' kind: '%s'", included.StructField, included.StructField.Relationship().Kind())
	}
	return err
}

func (s *Scope) findIncludedRelationsAsynchronous(ctx context.Context) (err error) {
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
	jobs := s.includedJobCreator(ctx, wg)

	errChan := make(chan error, 1)
	// Execute jobs with respect to given wait group and error channel.
	for job := range jobs {
		s.includedJobFind(ctx, job, wg, errChan)
	}

	// Wait channel waits until all jobs marks wait group done.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Debug2(s.logFormat("Getting includes - context done"))
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Find include error: %v", e)
		return e
	case <-waitChan:
		if log.Level().IsAllowed(log.LevelDebug2) {
			log.Debug2f(s.logFormat("Finished finding asynchronous included relations"))
		}
	}
	return nil
}

func (s *Scope) includedJobCreator(ctx context.Context, wg *sync.WaitGroup) (jobs <-chan *IncludedRelation) {
	out := make(chan *IncludedRelation)
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

func (s *Scope) includedJobFind(ctx context.Context, include *IncludedRelation, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := s.findIncludedRelation(ctx, include); err != nil {
			errChan <- err
		}
	}()
}
