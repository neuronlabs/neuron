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
	// ClassInvalidSort is the error classification for invalid sort input.
	ClassInvalidSort errors.Class
	// ClassInvalidModels is the error classification for the invalid models input.
	ClassInvalidModels errors.Class
	// ClassInvalidField is the error classification for the invalid field input
	ClassInvalidField errors.Class
	// ClassNoModels is the error classification when there is models provided in the input.
	ClassNoModels errors.Class

	// MnrFilter is the minor error classification related to the filters.
	MnrFilter errors.Minor
	// ClassFilterField is the error classification for the filter fields.
	ClassFilterField errors.Class
	// ClassFilterFormat is the error classification for the filter format.
	ClassFilterFormat errors.Class
	// ClassFilterCollection is the error classification for the filter collection.
	ClassFilterCollection errors.Class

	// MnrTransaction is minor error classification for the query transactions.
	MnrTransaction errors.Minor

	// ClassTxDone is the classification for finished transactions.
	ClassTxDone errors.Class
	// ClassTxState is the classification for the transaction state.
	ClassTxState errors.Class
)

func init() {
	MjrQuery = errors.MustNewMajor()

	// Major Classes
	ClassNoResult = errors.MustNewMajorClass(MjrQuery)
	ClassInternal = errors.MustNewMajorClass(MjrQuery)

	// Input
	MnrInput = errors.MustNewMinor(MjrQuery)
	ClassInvalidInput = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidField = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidSort = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassInvalidModels = errors.MustNewClassWIndex(MjrQuery, MnrInput)
	ClassNoModels = errors.MustNewClassWIndex(MjrQuery, MnrInput)

	// Filter
	MnrFilter = errors.MustNewMinor(MjrQuery)
	ClassFilterField = errors.MustNewClassWIndex(MjrQuery, MnrFilter)
	ClassFilterFormat = errors.MustNewClassWIndex(MjrQuery, MnrFilter)
	ClassFilterCollection = errors.MustNewClassWIndex(MjrQuery, MnrFilter)

	// MnrTransaction
	MnrTransaction = errors.MustNewMinor(MjrQuery)

	ClassTxState = errors.MustNewClassWIndex(MjrQuery, MnrTransaction)
	ClassTxDone = errors.MustNewClassWIndex(MjrQuery, MnrTransaction)
}
