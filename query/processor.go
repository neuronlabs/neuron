package query

import (
	"context"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
)

// processes contains registered processes mapped by their names.
var processes = make(map[string]ProcessFunc)

// RegisterProcess registers the process with it's unique name.
// If the process is already registered the function panics.
func RegisterProcess(name string, p ProcessFunc) {
	registerQueryProcess(name, p)
	// store the process name internally, so that config might check the process names.
	internal.Processes[name] = struct{}{}
}

func registerQueryProcess(name string, p ProcessFunc) {
	_, ok := processes[name]
	if ok {
		log.Panicf("Process: '%s' already registered", name)
	}
	log.Debugf("Registered process: '%s'.", name)
	processes[name] = p
}

func init() {
	// create processes
	registerQueryProcess(ProcessHookBeforeCreate, beforeCreateFunc)
	registerQueryProcess(ProcessSetBelongsToRelations, setBelongsToRelationshipsFunc)
	registerQueryProcess(ProcessSetCreatedAt, setCreatedAtField)
	registerQueryProcess(ProcessCreate, createFunc)
	registerQueryProcess(ProcessStoreScopePrimaries, storeScopePrimaries)
	registerQueryProcess(ProcessPatchForeignRelations, patchForeignRelationshipsFunc)
	registerQueryProcess(ProcessPatchForeignRelationsSafe, patchForeignRelationshipsSafeFunc)
	registerQueryProcess(ProcessHookAfterCreate, afterCreateFunc)

	// get processes
	registerQueryProcess(ProcessFillEmptyFieldset, fillEmptyFieldset)
	registerQueryProcess(ProcessDeletedAtFilter, getNotDeletedFilter)
	registerQueryProcess(ProcessConvertRelationFilters, convertRelationshipFiltersFunc)
	registerQueryProcess(ProcessConvertRelationFiltersSafe, convertRelationshipFiltersSafeFunc)
	registerQueryProcess(ProcessHookBeforeGet, beforeGetFunc)
	registerQueryProcess(ProcessGet, getFunc)
	registerQueryProcess(ProcessGetForeignRelations, getForeignRelationshipsFunc)
	registerQueryProcess(ProcessGetForeignRelationsSafe, getForeignRelationshipsSafeFunc)
	registerQueryProcess(ProcessHookAfterGet, afterGetFunc)

	// List
	registerQueryProcess(ProcessCheckPagination, checkPaginationFunc)
	registerQueryProcess(ProcessHookBeforeList, beforeListFunc)
	registerQueryProcess(ProcessList, listFunc)
	registerQueryProcess(ProcessHookAfterList, afterListFunc)
	registerQueryProcess(ProcessGetIncluded, getIncludedFunc)
	registerQueryProcess(ProcessGetIncludedSafe, getIncludedSafeFunc)

	// Patch
	registerQueryProcess(ProcessHookBeforePatch, beforePatchFunc)
	registerQueryProcess(ProcessSetUpdatedAt, setUpdatedAtField)
	registerQueryProcess(ProcessPatch, patchFunc)
	registerQueryProcess(ProcessHookAfterPatch, afterPatchFunc)
	registerQueryProcess(ProcessPatchBelongsToRelations, patchBelongsToRelationshipsFunc)

	// Delete
	registerQueryProcess(ProcessReducePrimaryFilters, reducePrimaryFilters)
	registerQueryProcess(ProcessHookBeforeDelete, beforeDeleteFunc)
	registerQueryProcess(ProcessSetDeletedAt, setDeletedAtField)
	registerQueryProcess(ProcessDelete, deleteFunc)
	registerQueryProcess(ProcessHookAfterDelete, afterDeleteFunc)
	registerQueryProcess(ProcessDeleteForeignRelations, deleteForeignRelationshipsFunc)
	registerQueryProcess(ProcessDeleteForeignRelationsSafe, deleteForeignRelationshipsSafeFunc)

	// Transactions
	registerQueryProcess(ProcessTxBegin, beginTransactionFunc)
	registerQueryProcess(ProcessTxCommitOrRollback, commitOrRollbackFunc)

	// Count
	registerQueryProcess(ProcessHookBeforeCount, beforeCountProcessFunc)
	registerQueryProcess(ProcessCount, countProcessFunc)
	registerQueryProcess(ProcessHookAfterCount, afterCountProcessFunc)
}

// ProcessFunc is the function that modifies or changes the scope value
type ProcessFunc func(ctx context.Context, s *Scope) error

// Processor is the wrapper over config processor that allows to start the queries.
type Processor config.Processor

// Count initializes the Count Process Chain for the Scope.
func (p *Processor) Count(ctx context.Context, s *Scope) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.Count: %s.", s.String(), s.Value)
	}
	var processError error
	s.processMethod = pmCount

	for _, processName := range p.CountProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Count failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] Count process finished", s.ID())
	}
	return processError
}

// Create initializes the Create Process Chain for the Scope.
func (p *Processor) Create(ctx context.Context, s *Scope) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.Create: %s. Value: %+v", s.String(), s.Value)
	}
	var processError error
	s.processMethod = pmCreate

	for _, processName := range p.CreateProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Creating failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] Create process finished", s.ID())
	}
	return processError
}

// Get initializes the Get Process chain for the scope.
func (p *Processor) Get(ctx context.Context, s *Scope) error {
	var processError error
	s.processMethod = pmGet

	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.Get: %s", s.String())
	}
	for _, processName := range p.GetProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Getting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] Get process finished", s.ID())
	}
	return processError
}

// List initializes the List Process Chain for the scope.
func (p *Processor) List(ctx context.Context, s *Scope) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.List: %s", s.String())
	}
	var processError error
	s.processMethod = pmList

	for _, processName := range p.ListProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Listing failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] List process finished", s.ID())
	}
	return processError
}

// Patch initializes the Patch Process Chain for the scope 's'.
func (p *Processor) Patch(ctx context.Context, s *Scope) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.Patch: %s with value: %+v", s.String(), s.Value)
	}
	var processError error
	s.processMethod = pmPatch

	for _, processName := range p.PatchProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Patching failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] Patch process finished", s.ID())
	}
	return processError
}

// Delete initializes the Delete Process Chain for the scope 's'.
func (p *Processor) Delete(ctx context.Context, s *Scope) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("Processor.Delete: %s", s.String())
	}
	var processError error
	s.processMethod = pmDelete

	for _, processName := range p.DeleteProcesses {
		processFunc := processes[processName]
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), processName)

		if err := processFunc(ctx, s); err != nil {
			log.Debug2f("Scope[%s][%s] Deleting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), processName, err)
			s.Err = err
			processError = err
		}
		select {
		case <-ctx.Done():
			processError = ctx.Err()
		default:
		}
		s.StoreSet(internal.PreviousProcessStoreKey, processName)
	}
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s] Delete process finished", s.ID())
	}
	return processError
}

type processMethod int8

const (
	_ processMethod = iota
	pmCreate
	pmCount
	pmDelete
	pmGet
	pmList
	pmPatch
)
