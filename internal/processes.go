package internal

// Processes is a mapping of the already registered query processes name.
var Processes = make(map[string]struct{})

// Processes constant names.
const (
	// Create processes
	ProcessHookBeforeCreate          = "hook_before_create"
	ProcessSetBelongsToRelations     = "set_belongs_to_relations"
	ProcessCreate                    = "create"
	ProcessStoreScopePrimaries       = "store_scope_primaries"
	ProcessPatchForeignRelations     = "patch_foreign_relations"
	ProcessPatchForeignRelationsSafe = "patch_foreign_relationships_safe"
	ProcessHookAfterCreate           = "hook_after_create"

	// Get processes
	ProcessFillEmptyFieldset          = "fill_empty_fieldset"
	ProcessConvertRelationFilters     = "convert_relation_filters"
	ProcessConvertRelationFiltersSafe = "convert_relation_filters_safe"
	ProcessHookBeforeGet              = "hook_before_get"
	ProcessGet                        = "get"
	ProcessGetForeignRelations        = "get_foreign_relations"
	ProcessGetForeignRelationsSafe    = "get_foreign_relations_safe"
	ProcessHookAfterGet               = "hook_after_get"

	// List processes
	ProcessCheckPagination = "check_pagination"
	ProcessHookBeforeList  = "hook_before_list"
	ProcessList            = "list"
	ProcessHookAfterList   = "hook_after_list"
	ProcessGetIncluded     = "get_included"
	ProcessGetIncludedSafe = "get_included_safe"

	// Patch processes
	ProcessHookBeforePatch         = "hook_before_patch"
	ProcessPatch                   = "patch"
	ProcessHookAfterPatch          = "hook_after_patch"
	ProcessPatchBelongsToRelations = "patch_belongs_to_relations"

	// Delete processes
	ProcessReducePrimaryFilters       = "reduce_primary_filters"
	ProcessHookBeforeDelete           = "hook_before_delete"
	ProcessDelete                     = "delete"
	ProcessHookAfterDelete            = "hook_after_delete"
	ProcessDeleteForeignRelations     = "delete_foreign_relations"
	ProcessDeleteForeignRelationsSafe = "delete_foreign_relations_safe"

	// Transaction processes
	ProcessTxBegin            = "tx_begin"
	ProcessTxCommitOrRollback = "tx_commit_or_rollback"

	// Count Processess
	ProcessHookBeforeCount = "hook_before_ount"
	ProcessCount           = "count"
	ProcessHookAfterCount  = "hook_after_count"
)
