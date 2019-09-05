package query

import (
	"strings"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

// MaxPermissibleDuplicates is the maximum permissible dupliactes value used for errors
// TODO: get the value from config
var MaxPermissibleDuplicates = 3

// IncludeFields adds the included fields into query scope.
func (s *Scope) IncludeFields(fields ...string) error {
	if err := s.buildIncludedFields(fields...); len(err) > 0 {
		if len(err) == 1 {
			return err[0]
		}

		var multi errors.MultiError

		for _, e := range err {
			multi = append(multi, e.(errors.ClassError))
		}
		return multi
	}
	return nil
}

// IncludedScope gets included scope if exists for the 'model'.
// NOTE: included scope contains no resultant values. This function is used only to set the included models fieldset and filters.
func (s *Scope) IncludedScope(model interface{}) (*Scope, error) {
	var err error
	mStruct, ok := model.(*mapping.ModelStruct)
	if !ok {
		mStruct, err = s.Controller().ModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	included, ok := s.includedScopes[mStruct]
	if !ok {
		log.Debugf("Model: '%s' is not included into scope of: '%s'", mStruct.Collection(), s.Struct().Collection())
		return nil, errors.NewDet(class.QueryNotIncluded, "provided model is not included within query's scope")
	}
	return included, nil
}

// IncludedScopes returns the slice of included scopes if exists.
func (s *Scope) IncludedScopes() (scopes []*Scope) {
	for _, scope := range s.includedScopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// IncludedValues returns included scope unique values, where the
// map key is unique primary field value, and the map value a model instance.
func (s *Scope) IncludedValues() map[interface{}]interface{} {
	if s.includedValues != nil {
		return s.includedValues.Values()
	}
	return nil
}

// IncludedModelValues gets the scope's included values for the given 'model'.
// Returns the map of primary keys to the model values.
func (s *Scope) IncludedModelValues(model interface{}) (map[interface{}]interface{}, error) {
	var (
		mStruct *mapping.ModelStruct
		ok      bool
		err     error
	)

	if mStruct, ok = model.(*mapping.ModelStruct); !ok {
		mStruct, err = s.Controller().ModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	included, ok := s.includedScopes[mStruct]
	if !ok {
		log.Info("Model: '%s' is not included into scope of: '%s'", mStruct.Collection(), s.Struct().Collection())
		return nil, errors.NewDet(class.QueryNotIncluded, "provided model is not included within query's scope")
	}
	return included.includedValues.Map(), nil
}

// buildIncludedFields builds the included fields for the given scope.
func (s *Scope) buildIncludedFields(includedList ...string) []errors.DetailedError {
	var (
		errorObjects []errors.DetailedError
		errs         []errors.DetailedError
		errObj       errors.DetailedError
	)

	// check if the number of included fields is possible
	if len(includedList) > s.mStruct.MaxIncludedCount() {
		errObj = errors.NewDet(class.QueryIncludeTooMany, "too many included fields provided")
		errObj.SetDetailsf("Too many included parameter values for: '%s' collection.", s.mStruct.Collection())
		errs = append(errs, errObj)
		return errs
	}

	// includedScopes for root are always set
	s.includedScopes = make(map[*mapping.ModelStruct]*Scope)
	var includedMap map[string]int

	// many includes flag if there is more than one include
	var manyIncludes = len(includedList) > 1
	if manyIncludes {
		includedMap = make(map[string]int)
	}

	// having multiple included in the query
	for _, included := range includedList {
		// check the nested level of every included
		annotCount := strings.Count(included, annotation.NestedSeparator)
		if annotCount > s.Struct().MaxIncludedDepth() {
			errObj = errors.NewDetf(class.QueryIncludeTooMany, "reached the maximum nested include limit")
			errObj.SetDetails("Maximum nested include limit reached for the given query.")
			errs = append(errs, errObj)
			continue
		}

		// check if there are more than one include.
		if manyIncludes {
			// assert no duplicates are provided in the include list.
			includedCount := includedMap[included]
			includedCount++
			includedMap[included] = includedCount

			if annotCount == 0 && includedCount > 1 {
				if includedCount == 2 {
					errObj = errors.NewDet(class.QueryIncludeTooMany, "included fields duplicated")
					errObj.SetDetailsf("Included parameter '%s' used more than once.", included)
					errs = append(errs, errObj)
					continue
				} else if includedCount >= MaxPermissibleDuplicates {
					break
				}
			}
		}
		errorObjects = s.buildInclude(included)
		errs = append(errs, errorObjects...)
	}
	return errs
}

// buildInclude searches for the relationship field within given scope
// if not found, then tries to separate the 'included' argument
// by the 'annotation.NestedSeparator'. If separated correctly
// it tries to create nested fields.
// adds IncludeScope for given field.
func (s *Scope) buildInclude(included string) []errors.DetailedError {
	var (
		includedField *IncludeField
		errs          []errors.DetailedError
	)
	// search for the 'included' in the model's
	relationField, ok := s.mStruct.RelationField(included)
	if !ok {
		// no relationship found check nesteds
		index := strings.Index(included, annotation.NestedSeparator)
		if index == -1 {
			err := errors.NewDet(class.QueryFilterUnknownField, "no relationship field found")
			err.SetDetailsf("Object: '%v', has no relationship named: '%v'.", s.mStruct.Collection(), included)
			errs = append(errs, err)
			return errs
		}

		// root part of included (root.subfield)
		field := included[:index]
		relationField, ok := s.mStruct.RelationField(field)
		if !ok {
			err := errors.NewDet(class.QueryFilterUnknownField, "no relationship field found")
			err.SetDetailsf("Object: '%v', has no relationship named: '%v'.", s.mStruct.Collection(), included)
			errs = append(errs, err)
			return errs
		}

		// create new included field
		includedField = s.getOrCreateIncludeField(relationField)

		errs = includedField.Scope.buildInclude(included[index+1:])
		if len(errs) > 0 {
			return errs
		}
	} else {
		// create new includedField if the field was not already created during nested process.
		includedField = s.getOrCreateIncludeField(relationField)
	}

	includedField.Scope.kind = IncludedKind
	return errs
}

// createOrGetIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field
func (s *Scope) getOrCreateIncludeField(field *mapping.StructField) (includeField *IncludeField) {
	for _, included := range s.includedFields {
		if included.StructField == field {
			return included
		}
	}
	return s.createIncludedField(field)
}

func (s *Scope) createIncludedField(field *mapping.StructField) (includeField *IncludeField) {
	includeField = newIncludeField(field, s)
	if s.includedFields == nil {
		s.includedFields = make([]*IncludeField, 0)
	}

	s.includedFields = append(s.includedFields, includeField)
	return
}
