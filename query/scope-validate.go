package query

import (
	"context"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

// CreateValidator is an interface implemented by the neuron models used for validating create values.
type CreateValidator interface {
	ValidateCreate(s *Scope) error
}

// PatchValidator is an interface implemented by the neuron models used for validating patch values.
type PatchValidator interface {
	ValidatePatch(s *Scope) error
}

// Validator is common validator, used if a model doesn't implement CreateValidator or PatchValidator.
// A model should implement this interface if it's validation is done in the same way for both create and patch
// processes.
type Validator interface {
	Validate(s *Scope) error
}

// this should be done just at the beginning of the processor.
func (s *Scope) validateInitialQuery(ctx context.Context) error {
	switch s.processMethod {
	case pmCount:
		return s.validateCountQuery(ctx)
	case pmCreate:
		return s.validateInitialCreateQuery(ctx)
	case pmDelete:
		return s.validateDeleteQuery(ctx)
	case pmGet:
		return s.validateGetQuery(ctx)
	case pmList:
		return s.validateListQuery(ctx)
	case pmPatch:
		return s.validateInitialPatchQuery(ctx)
	default:
		return errors.Newf(class.InternalQueryValidation, "provided unknown process method: '%d' for given query", s.processMethod)
	}
}

// this should be done just before executing the main process function.
func (s *Scope) validateQuery(ctx context.Context) error {
	switch s.processMethod {
	case pmCount:
		return s.validateCountQuery(ctx)
	case pmCreate:
		return s.validateCreateQuery(ctx)
	case pmDelete:
		return s.validateDeleteQuery(ctx)
	case pmGet:
		return s.validateGetQuery(ctx)
	case pmList:
		return s.validateListQuery(ctx)
	case pmPatch:
		return s.validatePatchQuery(ctx)
	default:
		return errors.Newf(class.InternalQueryValidation, "provided unknown process method: '%d' for given query", s.processMethod)
	}
}

func (s *Scope) validateCountQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validatePagination(); err != nil {
		return err
	}
	if err = s.validateScopeFilters(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateInitialCreateQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if s.isMany {
		return errors.NewDet(class.QueryValueType, "creating with multiple values in are not supported yet")
	}
	if err = s.validateCreateFieldset(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateCreateQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if err = s.validateCreateValue(); err != nil {
		return err
	}
	if err = s.validateCreateFieldset(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateInitialPatchQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if err = s.validateFieldset(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validatePatchQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if err = s.validatePatchValue(); err != nil {
		return err
	}
	if err = s.validateFieldset(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateDeleteQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeFilters(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateGetQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validateScopeFilters(); err != nil {
		return err
	}
	if err = s.validateFieldset(); err != nil {
		return err
	}
	if err = s.validateIncludes(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateListQuery(ctx context.Context) (err error) {
	if err = s.validateContext(ctx); err != nil {
		return err
	}
	if err = s.validateScopeNilValue(); err != nil {
		return err
	}
	if err = s.validateTransaction(); err != nil {
		return err
	}
	if err = s.validatePagination(); err != nil {
		return err
	}
	if err = s.validateFieldset(); err != nil {
		return err
	}
	if err = s.validateIncludes(); err != nil {
		return err
	}
	if err = s.validateScopeFilters(); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateTransaction() error {
	if s.tx == nil {
		return nil
	}
	if s.tx.state.Done() {
		return errors.Newf(class.QueryTxDone, "transaction is already finished")
	}
	return nil
}

func (s *Scope) validateCreateFieldset() error {
	return s.validateFieldsInFieldset()
}

func (s *Scope) validateFieldset() error {
	if len(s.Fieldset) == 0 {
		return errors.Newf(class.QueryFieldsetEmpty, "provided empty fieldset")
	}
	return s.validateFieldsInFieldset()
}

func (s *Scope) validateFieldsInFieldset() error {
	for _, field := range s.Fieldset {
		if field.Struct() != s.mStruct {
			return errors.Newf(class.QueryFieldsetInvalidModel,
				"field: '%s' in the fieldset is mapped to different model: '%s' should be: '%s",
				field.Name(), field.Struct().String(), s.mStruct.String())
		}
	}
	return nil
}

func (s *Scope) validatePagination() error {
	if s.Pagination == nil {
		return nil
	}
	return s.Pagination.Validate()
}

func (s *Scope) validateFilters(filters Filters) error {
	// validate if all filters are well formed and are mapped to the right model.
	for _, filter := range filters {
		if filter.StructField.Struct() != s.mStruct {
			return errors.Newf(class.QueryFilterInvalidField, "provided filter is not related to given model")
		}
	}
	return nil
}

func (s *Scope) validateScopeFilters() (err error) {
	if err = s.validateFilters(s.PrimaryFilters); err != nil {
		return err
	}
	if err = s.validateFilters(s.AttributeFilters); err != nil {
		return err
	}
	if err = s.validateFilters(s.RelationFilters); err != nil {
		return err
	}
	if err = s.validateFilters(s.ForeignFilters); err != nil {
		return err
	}
	return nil
}

func (s *Scope) validateIncludes() error {
	if len(s.IncludedFields) == 0 {
		return nil
	}
	// check if all included fields are in the fieldset.
	for _, includedField := range s.IncludedFields {
		_, ok := s.Fieldset[includedField.StructField.NeuronName()]
		if !ok {
			return errors.NewDetf(class.QueryFieldsetInvalid, "included field: '%s' is not in the fieldset", includedField.StructField.NeuronName())
		}
	}
	return nil
}

func (s *Scope) validateContext(ctx context.Context) error {
	return ctx.Err()
}

func (s *Scope) validateScopeNilValue() error {
	if s.Value == nil {
		return errors.Newf(class.QueryNilValue, "provided nil scope value")
	}
	return nil
}

func (s *Scope) validateCreateValue() error {
	switch validator := s.Value.(type) {
	case CreateValidator:
		return validator.ValidateCreate(s)
	case Validator:
		return validator.Validate(s)
	}
	return nil
}

func (s *Scope) validatePatchValue() error {
	switch validator := s.Value.(type) {
	case PatchValidator:
		return validator.ValidatePatch(s)
	case Validator:
		return validator.Validate(s)
	}
	return nil
}
