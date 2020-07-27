package query

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrQuery is the major error classification for the query package.
	MjrQuery errors.Major

	// ClassInternal is the internal error classification.
	ClassInternal errors.Class
	// ClassNoResult is the error classification when for the query that returns no result.
	ClassNoResult errors.Class

	// MnrInput is the minor error classification related to the query input.
	MnrInput errors.Minor
	// ClassInvalidInput is the error classification for invalid query input.
	ClassInvalidInput errors.Class
	// ClassInvalidParameter is the error classification for invalid query parameter.
	ClassInvalidParameter errors.Class
	// ClassInvalidSort is the error classification for invalid sort input.
	ClassInvalidSort errors.Class
	// ClassInvalidModels is the error classification for the invalid models input.
	ClassInvalidModels errors.Class
	// ClassInvalidField is the error classification for the invalid field input
	ClassInvalidField errors.Class
	// ClassInvalidFieldSet is the error classification for the invalid fieldset input
	ClassInvalidFieldSet errors.Class
	// ClassFieldValue is the error classification for the invalid field input
	ClassFieldValue errors.Class
	// ClassNoModels is the error classification when there is models provided in the input.
	ClassNoModels errors.Class
	// ClassNoFieldsInFieldset is the error classification when no fields are present in the fieldset.
	ClassNoFieldsInFieldSet errors.Class

	// MnrTransaction is minor error classification for the query transactions.
	MnrTransaction errors.Minor

	// ClassTxDone is the classification for finished transactions.
	ClassTxDone errors.Class
	// ClassTxState is the classification for the transaction state.
	ClassTxState errors.Class
	// ClassTxInvalid is the classification for the invalid transaction.
	ClassTxInvalid errors.Class

	// MnrViolation is the minor error classification when query violates some restrictions.
	MnrViolation errors.Minor
	// ClassViolationIntegrityConstraint is the violation error classification.
	ClassViolationIntegrityConstraint errors.Class
	// ClassViolationRestrict is the violation error classification.
	ClassViolationRestrict errors.Class
	// ClassViolationNotNull is the violation error classification.
	ClassViolationNotNull errors.Class
	// ClassViolationForeignKey is the violation error classification.
	ClassViolationForeignKey errors.Class
	// ClassViolationUnique is the violation error classification.
	ClassViolationUnique errors.Class
	// ClassViolationCheck is the violation error classification.
	ClassViolationCheck errors.Class
	// ClassViolationDataType is the violation error classification.
	ClassViolationDataType errors.Class
)

func init() {
	MjrQuery = errors.MustNewMajor()

	// Major Classes
	ClassNoResult = errors.MustNewMajorClass(MjrQuery)
	ClassInternal = errors.MustNewMajorClass(MjrQuery)

	// Input
	MnrInput = errors.MustNewMinor(MjrQuery)
	ClassInvalidInput = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidParameter = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidField = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidFieldSet = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassFieldValue = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidSort = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidModels = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassNoModels = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassNoFieldsInFieldSet = errors.MustNewClassWIndex(MjrQuery, MnrInput)

	// MnrTransaction
	MnrTransaction = errors.MustNewMinor(MjrQuery)

	ClassTxState = errors.MustNewClassWIndex(MjrQuery, MnrTransaction)
	ClassTxDone = errors.MustNewClassWIndex(MjrQuery, MnrTransaction)
	ClassTxInvalid = errors.MustNewClassWIndex(MjrQuery, MnrTransaction)

	// Violation
	MnrViolation = errors.MustNewMinor(MjrQuery)
	ClassViolationIntegrityConstraint = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationIntegrityConstraint = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationRestrict = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationNotNull = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationForeignKey = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationUnique = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationCheck = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
	ClassViolationDataType = errors.MustNewClassWIndex(MjrQuery, MnrViolation)
}
