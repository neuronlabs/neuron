package query

import (
	"context"
	"fmt"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
)

// processes contains registered processes mapped by their names.
var processes = make(map[string]ProcessFunc)

// RegisterProcessFunc registers the process with it's unique name.
// If the process is already registered the function panics.
func RegisterProcessFunc(name string, p ProcessFunc) {
	_, ok := processes[name]
	if ok {
		panic(fmt.Errorf("Process: '%s' already registered", name))
	}
	log.Debugf("Registered process: '%s'.", name)
	processes[name] = p
	// store the process name internally, so that config might check the process names.
	internal.Processes[name] = struct{}{}
}

func init() {
	// create processes
	RegisterProcessFunc(ProcessHookBeforeCreate, beforeCreateFunc)
	RegisterProcessFunc(ProcessSetBelongsToRelations, setBelongsToRelationshipsFunc)
	RegisterProcessFunc(ProcessCreate, createFunc)
	RegisterProcessFunc(ProcessStoreScopePrimaries, storeScopePrimaries)
	RegisterProcessFunc(ProcessPatchForeignRelations, patchForeignRelationshipsFunc)
	RegisterProcessFunc(ProcessPatchForeignRelationsSafe, patchForeignRelationshipsSafeFunc)
	RegisterProcessFunc(ProcessHookAfterCreate, afterCreateFunc)

	// get processes
	RegisterProcessFunc(ProcessFillEmptyFieldset, fillEmptyFieldset)
	RegisterProcessFunc(ProcessConvertRelationFilters, convertRelationshipFiltersFunc)
	RegisterProcessFunc(ProcessConvertRelationFiltersSafe, convertRelationshipFiltersSafeFunc)
	RegisterProcessFunc(ProcessHookBeforeGet, beforeGetFunc)
	RegisterProcessFunc(ProcessGet, getFunc)
	RegisterProcessFunc(ProcessGetForeignRelations, getForeignRelationshipsFunc)
	RegisterProcessFunc(ProcessGetForeignRelationsSafe, getForeignRelationshipsSafeFunc)
	RegisterProcessFunc(ProcessHookAfterGet, afterGetFunc)

	// List
	RegisterProcessFunc(ProcessCheckPagination, checkPaginationFunc)
	RegisterProcessFunc(ProcessHookBeforeList, beforeListFunc)
	RegisterProcessFunc(ProcessList, listFunc)
	RegisterProcessFunc(ProcessHookAfterList, afterListFunc)
	RegisterProcessFunc(ProcessGetIncluded, getIncludedFunc)
	RegisterProcessFunc(ProcessGetIncludedSafe, getIncludedSafeFunc)

	// Patch
	RegisterProcessFunc(ProcessHookBeforePatch, beforePatchFunc)
	RegisterProcessFunc(ProcessPatch, patchFunc)
	RegisterProcessFunc(ProcessHookAfterPatch, afterPatchFunc)
	RegisterProcessFunc(ProcessPatchBelongsToRelations, patchBelongsToRelationshipsFunc)

	// Delete
	RegisterProcessFunc(ProcessReducePrimaryFilters, reducePrimaryFilters)
	RegisterProcessFunc(ProcessHookBeforeDelete, beforeDeleteFunc)
	RegisterProcessFunc(ProcessDelete, deleteFunc)
	RegisterProcessFunc(ProcessHookAfterDelete, afterDeleteFunc)
	RegisterProcessFunc(ProcessDeleteForeignRelations, deleteForeignRelationshipsFunc)
	RegisterProcessFunc(ProcessDeleteForeignRelationsSafe, deleteForeignRelationshipsSafeFunc)

	// Transactions
	RegisterProcessFunc(ProcessTxBegin, beginTransactionFunc)
	RegisterProcessFunc(ProcessTxCommitOrRollback, commitOrRollbackFunc)
}

// ProcessFunc is the function that modifies or changes the scope value
type ProcessFunc func(ctx context.Context, s *Scope) error

// Processor is the wrapper over config processor that allows to start the queries.
type Processor config.Processor

// Create initializes the Create Process Chain for the Scope.
func (p *Processor) Create(ctx context.Context, s *Scope) error {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("Processor.Create: %s. Value: %+v", s.String(), s.Value)
	}
	var processError error

	for _, processName := range p.CreateProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Creating failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.StoreSet(processErrorKey, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s] Create process finished", s.ID())
	}
	return processError
}

// Get initializes the Get Process chain for the scope.
func (p *Processor) Get(ctx context.Context, s *Scope) error {
	var processError error
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("Processor.Get: %s", s.String())
	}
	for _, processName := range p.GetProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Getting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.StoreSet(processErrorKey, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s] Get process finished", s.ID())
	}
	return processError
}

// List initializes the List Process Chain for the scope.
func (p *Processor) List(ctx context.Context, s *Scope) error {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("Processor.List: %s", s.String())
	}
	var processError error
	for _, processName := range p.ListProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Listing failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.StoreSet(processErrorKey, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s] List process finished", s.ID())
	}
	return processError
}

// Patch initializes the Patch Process Chain for the scope 's'.
func (p *Processor) Patch(ctx context.Context, s *Scope) error {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("Processor.Patch: %s with value: %+v", s.String(), s.Value)
	}
	var processError error
	for _, processName := range p.PatchProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Patching failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.StoreSet(processErrorKey, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s] Patch process finished", s.ID())
	}
	return processError
}

// Delete initializes the Delete Process Chain for the scope 's'.
func (p *Processor) Delete(ctx context.Context, s *Scope) error {
	if log.Level() == log.LDEBUG3 {
		log.Debug3f("Processor.Delete: %s", s.String())
	}
	var processError error
	for _, processName := range p.DeleteProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Deleting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.StoreSet(processErrorKey, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("SCOPE[%s] Delete process finished", s.ID())
	}
	return processError
}

// process error key instance
var processErrorKey = processError{}

type processError struct{}
