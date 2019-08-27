package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/query/scope"
)

var (
	// ProcessGetIncluded is the process that gets the included scope values.
	ProcessGetIncluded = &Process{
		Name: "neuron:get_included",
		Func: getIncludedFunc,
	}

	// ProcessGetIncludedSafe is the gouroutine safe process that gets the included scope values.
	ProcessGetIncludedSafe = &Process{
		Name: "neuron:get_included_safe",
		Func: getIncludedSafeFunc,
	}
)

// processGetIncluded gets the included fields for the
// how it should look like:
// - get the primaries from the current included fields that are in the fieldset
// - prepare the included collections scopes for a single list function:
//	* add current included fields primaries to their collection scope included primaries map
// 	* get missing primaries (related primaries excluding current collection scope primaries) from the non fieldset included relationships and store them in their collection scope
//
// 	NOTE: this should take related values with respect to the collection fieldset.
func getIncludedFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	if s.internal().IsRoot() && len(s.internal().IncludedScopes()) == 0 {
		return nil
	}

	if err := s.internal().SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, incScope := range s.internal().IncludedScopes() {
		if incScope.Struct().Config() == nil {
			continue
		}

		if modelRepo := incScope.Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	includedFields := s.internal().IncludedFields()
	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) {
		// get missing primaries from the included field.
		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			results <- err
			return
		}

		if len(missing) > 0 {
			includedScope := includedField.Scope
			includedScope.SetIDFilters(missing...)
			includedScope.NewValueMany()

			if err = (*Scope)(includedScope).ListContext(ctx); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				results <- err
				return
			}
		}
		results <- struct{}{}
	}

	// send the jobs
	for _, includedField := range includedFields {
		go getInclude(includedField, results)
	}

	// collect the results
	var ctr int
	for {
		select {
		case <-ctx.Done():
		case v, ok := <-results:
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(includedFields) {
				break
			}
		}
	}
}

func getIncludedSafeFunc(ctx context.Context, s *Scope) error {
	if _, ok := s.StoreGet(processErrorKey); ok {
		return nil
	}

	if s.internal().IsRoot() && len(s.internal().IncludedScopes()) == 0 {
		return nil
	}

	if err := s.internal().SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Processor.DefaultTimeout
	for _, incScope := range s.internal().IncludedScopes() {
		if incScope.Struct().Config() == nil {
			continue
		}

		if modelRepo := incScope.Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	includedFields := s.internal().IncludedFields()
	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) error {
		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			return err
		}

		if len(missing) > 0 {
			includedScope := includedField.Scope
			includedScope.SetIDFilters(missing...)
			includedScope.NewValueMany()

			log.Debug2f("Included scope collection: %v", includedScope.Struct().Collection())
			if err = (*Scope)(includedScope).ListContext(ctx); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				return err
			}
		}
		return nil
	}

	// send the jobs
	for _, includedField := range includedFields {
		err := getInclude(includedField, results)
		if err != nil {
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
