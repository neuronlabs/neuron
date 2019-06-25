package query

import (
	"context"
	"fmt"

	"github.com/neuronlabs/neuron/common"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/log"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

// processes contains registered processes by their name
var processes = make(map[string]*Process)

// RegisterProcess registers the process. If the process is already registered the function panics
func RegisterProcess(p *Process) error {
	_, ok := processes[p.Name]
	if ok {
		panic(fmt.Errorf("Process: '%s' already registered", p.Name))
	}
	log.Debugf("Registered process: '%s'.", p.Name)
	processes[p.Name] = p
	internal.Processes[p.Name] = struct{}{}

	return nil
}

func init() {
	// create processes
	RegisterProcess(ProcessBeforeCreate)
	RegisterProcess(ProcessSetBelongsToRelationships)
	RegisterProcess(ProcessCreate)
	RegisterProcess(ProcessStoreScopePrimaries)
	RegisterProcess(ProcessPatchForeignRelationships)
	RegisterProcess(ProcessPatchForeignRelationshipsSafe)
	RegisterProcess(ProcessAfterCreate)

	// get processes
	RegisterProcess(ProcessFillEmptyFieldset)
	RegisterProcess(ProcessConvertRelationshipFilters)
	RegisterProcess(ProcessConvertRelationshipFiltersSafe)
	RegisterProcess(ProcessBeforeGet)
	RegisterProcess(ProcessGet)
	RegisterProcess(ProcessGetForeignRelationships)
	RegisterProcess(ProcessGetForeignRelationshipsSafe)
	RegisterProcess(ProcessAfterGet)

	// List
	RegisterProcess(ProcessBeforeList)
	RegisterProcess(ProcessList)
	RegisterProcess(ProcessAfterList)
	RegisterProcess(ProcessGetIncluded)
	RegisterProcess(ProcessGetIncludedSafe)

	// Patch
	RegisterProcess(ProcessBeforePatch)
	RegisterProcess(ProcessPatch)
	RegisterProcess(ProcessAfterPatch)
	RegisterProcess(ProcessPatchBelongsToRelationships)

	// Delete
	RegisterProcess(ProcessReducePrimaryFilters)
	RegisterProcess(ProcessBeforeDelete)
	RegisterProcess(ProcessAfterDelete)
	RegisterProcess(ProcessDelete)
	RegisterProcess(ProcessDeleteForeignRelationships)
	RegisterProcess(ProcessDeleteForeignRelationshipsSafe)

	// Transactions
	RegisterProcess(ProcessTransactionBegin)
	RegisterProcess(ProcessTransactionCommitOrRollback)
}

// ProcessFunc is the function that modifies or changes the scope value
type ProcessFunc func(ctx context.Context, s *Scope) error

// Process is the pair of the name and the ProcessFunction
type Process struct {
	Name string
	Func ProcessFunc
}

// Processor is the struct that allows to query over the gateway's model's
type Processor struct {
	CreateChain ProcessChain
	GetChain    ProcessChain
	ListChain   ProcessChain
	PatchChain  ProcessChain
	DeleteChain ProcessChain
}

// New creates the query processor
func newProcessor(cfg *config.Processor) *Processor {
	p := &Processor{}

	for _, processName := range cfg.CreateProcesses {
		process, ok := processes[processName]
		if !ok {
			panic(fmt.Sprintf("Process: '%s' is not registered", processName))
		}

		p.CreateChain = append(p.CreateChain, process)
	}

	for _, processName := range cfg.DeleteProcesses {
		process, ok := processes[processName]
		if !ok {
			panic(fmt.Sprintf("Process: '%s' is not registered", processName))
		}

		p.DeleteChain = append(p.DeleteChain, process)
	}

	for _, processName := range cfg.GetProcesses {
		process, ok := processes[processName]
		if !ok {
			panic(fmt.Sprintf("Process: '%s' is not registered", processName))
		}

		p.GetChain = append(p.GetChain, process)
	}

	for _, processName := range cfg.ListProcesses {
		process, ok := processes[processName]
		if !ok {
			panic(fmt.Sprintf("Process: '%s' is not registered", processName))
		}

		p.ListChain = append(p.ListChain, process)
	}

	for _, processName := range cfg.PatchProcesses {
		process, ok := processes[processName]
		if !ok {
			panic(fmt.Sprintf("Process: '%s' is not registered", processName))
		}

		p.PatchChain = append(p.PatchChain, process)
	}

	return p
}

var _ scope.Processor = &Processor{}

// Create is the initializes the Create Process Chain for the Scope
func (p *Processor) Create(ctx context.Context, s *scope.Scope) error {
	var processError error
	for _, f := range p.CreateChain {
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debug2f("Scope[%s][%s] Creating failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), f.Name, err)
			s.StoreSet(common.ProcessError, err)
			processError = err
		}

		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return processError
}

// Get initializes the Get Process chain for the scope
func (p *Processor) Get(ctx context.Context, s *scope.Scope) error {
	var processError error
	for _, f := range p.GetChain {
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debug2f("Scope[%s][%s] Getting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), f.Name, err)
			s.StoreSet(common.ProcessError, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return processError
}

// List initializes the List Process Chain for the scope
func (p *Processor) List(ctx context.Context, s *scope.Scope) error {
	var processError error
	for _, f := range p.ListChain {
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debug2f("Scope[%s][%s] Listing failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), f.Name, err)
			s.StoreSet(common.ProcessError, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}
	return processError
}

// Patch does the Patch Process Chain
func (p *Processor) Patch(ctx context.Context, s *scope.Scope) error {
	var processError error
	for _, f := range p.PatchChain {
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), f.Name)

		ts := (*Scope)(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debug2f("Scope[%s][%s] Patching failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), f.Name, err)
			s.StoreSet(common.ProcessError, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}
	return processError
}

// Delete does the Delete process chain
func (p *Processor) Delete(ctx context.Context, s *scope.Scope) error {
	var processError error
	for _, f := range p.DeleteChain {
		log.Debug3f("Scope[%s][%s] %s", s.ID(), s.Struct().Collection(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debug2f("Scope[%s][%s] Deleting failed on process: '%s'. %v", s.ID(), s.Struct().Collection(), f.Name, err)
			s.StoreSet(common.ProcessError, err)
			processError = err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}
	return processError
}
