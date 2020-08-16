package query

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrQuery is the major error classification for the query package.
	ErrQuery = errors.New("query")

	// ErrInternal is the internal error classification.
	ErrInternal = errors.Wrap(errors.ErrInternal, "query")
	// ErrNoResult is the error classification when for the query that returns no result.
	ErrNoResult = errors.Wrap(ErrQuery, "no results")

	// ErrInput is the minor error classification related to the query input.
	ErrInput = errors.Wrap(ErrQuery, "input")
	// ErrInvalidInput is the error classification for invalid query input.
	ErrInvalidInput = errors.Wrap(ErrInput, "invalid")
	// ErrInvalidParameter is the error classification for invalid query parameter.
	ErrInvalidParameter = errors.Wrap(ErrInput, "invalid parameter")
	// ErrInvalidSort is the error classification for invalid sort input.
	ErrInvalidSort = errors.Wrap(ErrInput, "invalid sort")
	// ErrInvalidModels is the error classification for the invalid models input.
	ErrInvalidModels = errors.Wrap(ErrInput, "invalid models")
	// ErrInvalidField is the error classification for the invalid field input
	ErrInvalidField = errors.Wrap(ErrInput, "invalid field")
	// ErrInvalidFieldSet is the error classification for the invalid fieldset input
	ErrInvalidFieldSet = errors.Wrap(ErrInput, "invalid field set")
	// ErrFieldValue is the error classification for the invalid field input
	ErrFieldValue = errors.Wrap(ErrInput, "field value")
	// ErrNoModels is the error classification when there is models provided in the input.
	ErrNoModels = errors.Wrap(ErrInput, "no models")
	// ErrNoFieldsInFieldSet is the error classification when no fields are present in the fieldset.
	ErrNoFieldsInFieldSet = errors.Wrap(ErrInput, "no fields in field set")

	// ErrTransaction is minor error classification for the query transactions.
	ErrTransaction = errors.Wrap(ErrQuery, "transaction")

	// ErrTxDone is the classification for finished transactions.
	ErrTxDone = errors.Wrap(ErrTransaction, "done")
	// ErrTxState is the classification for the transaction state.
	ErrTxState = errors.Wrap(ErrTransaction, "state")
	// ErrTxInvalid is the classification for the invalid transaction.
	ErrTxInvalid = errors.Wrap(ErrTransaction, "invalid")

	// ErrViolation is the minor error classification when query violates some restrictions.
	ErrViolation = errors.Wrap(ErrQuery, "violation")
	// ErrViolationIntegrityConstraint is the violation error classification.
	ErrViolationIntegrityConstraint = errors.Wrap(ErrViolation, "integrity constraint")
	// ErrViolationRestrict is the violation error classification.
	ErrViolationRestrict = errors.Wrap(ErrViolation, "restrict")
	// ErrViolationNotNull is the violation error classification.
	ErrViolationNotNull = errors.Wrap(ErrViolation, "not null")
	// ErrViolationForeignKey is the violation error classification.
	ErrViolationForeignKey = errors.Wrap(ErrViolation, "foreign key")
	// ErrViolationUnique is the violation error classification.
	ErrViolationUnique = errors.Wrap(ErrViolation, "unique")
	// ErrViolationCheck is the violation error classification.
	ErrViolationCheck = errors.Wrap(ErrViolation, "check")
	// ErrViolationDataType is the violation error classification.
	ErrViolationDataType = errors.Wrap(ErrViolation, "data type")
)
