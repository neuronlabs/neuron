package query

import (
	"github.com/neuronlabs/neuron-core/internal"
)

const (
	// ParamInclude is the url.Query parameter name for the included fields.
	ParamInclude string = "include"
	// ParamFields is the url.Query parameter name for the fieldset.
	ParamFields string = "fields"
)

const (
	// ParamLanguage is the language query parameter used in the url values.
	ParamLanguage = "lang"
	// ParamFilter is the filter query parameter used as the key in the url values.
	ParamFilter = "filter"
)

// Processes constant names.
const (
	ProcessHookBeforeCreate          = internal.ProcessHookBeforeCreate
	ProcessSetBelongsToRelations     = internal.ProcessSetBelongsToRelations
	ProcessSetCreatedAt              = internal.ProcessSetCreatedAt
	ProcessCreate                    = internal.ProcessCreate
	ProcessStoreScopePrimaries       = internal.ProcessStoreScopePrimaries
	ProcessPatchForeignRelations     = internal.ProcessPatchForeignRelations
	ProcessPatchForeignRelationsSafe = internal.ProcessPatchForeignRelationsSafe
	ProcessHookAfterCreate           = internal.ProcessHookAfterCreate

	ProcessDeletedAtFilter            = internal.ProcessDeletedAtFilter
	ProcessFillEmptyFieldset          = internal.ProcessFillEmptyFieldset
	ProcessConvertRelationFilters     = internal.ProcessConvertRelationFilters
	ProcessConvertRelationFiltersSafe = internal.ProcessConvertRelationFiltersSafe
	ProcessHookBeforeGet              = internal.ProcessHookBeforeGet
	ProcessGet                        = internal.ProcessGet
	ProcessGetForeignRelations        = internal.ProcessGetForeignRelations
	ProcessGetForeignRelationsSafe    = internal.ProcessGetForeignRelationsSafe
	ProcessHookAfterGet               = internal.ProcessHookAfterGet

	ProcessCheckPagination = internal.ProcessCheckPagination
	ProcessHookBeforeList  = internal.ProcessHookBeforeList
	ProcessList            = internal.ProcessList
	ProcessHookAfterList   = internal.ProcessHookAfterList

	ProcessHookBeforePatch         = internal.ProcessHookBeforePatch
	ProcessSetUpdatedAt            = internal.ProcessSetUpdatedAt
	ProcessPatch                   = internal.ProcessPatch
	ProcessHookAfterPatch          = internal.ProcessHookAfterPatch
	ProcessPatchBelongsToRelations = internal.ProcessPatchBelongsToRelations

	ProcessReducePrimaryFilters       = internal.ProcessReducePrimaryFilters
	ProcessHookBeforeDelete           = internal.ProcessHookBeforeDelete
	ProcessSetDeletedAt               = internal.ProcessSetDeletedAt
	ProcessDelete                     = internal.ProcessDelete
	ProcessHookAfterDelete            = internal.ProcessHookAfterDelete
	ProcessDeleteForeignRelations     = internal.ProcessDeleteForeignRelations
	ProcessDeleteForeignRelationsSafe = internal.ProcessDeleteForeignRelationsSafe

	ProcessTxBegin            = internal.ProcessTxBegin
	ProcessTxCommitOrRollback = internal.ProcessTxCommitOrRollback

	ProcessCount           = internal.ProcessCount
	ProcessHookBeforeCount = internal.ProcessHookBeforeCount
	ProcessHookAfterCount  = internal.ProcessHookAfterCount

	ProcessValidateScope        = internal.ProcessValidateScope
	ProcessInitialValidateScope = internal.ProcessInitialValidateScope
)
