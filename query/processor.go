package query

import (
	"context"
	"fmt"

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
	RegisterProcess(ProcessAfterCreate)

	// get processes
	RegisterProcess(ProcessFillEmptyFieldset)
	RegisterProcess(ProcessConvertRelationshipFilters)
	RegisterProcess(ProcessBeforeGet)
	RegisterProcess(ProcessGet)
	RegisterProcess(ProcessGetForeignRelationships)
	RegisterProcess(ProcessAfterGet)

	// List
	RegisterProcess(ProcessBeforeList)
	RegisterProcess(ProcessList)
	RegisterProcess(ProcessAfterList)
	RegisterProcess(ProcessGetIncluded)

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
	for _, f := range p.CreateChain {
		log.Debug2f("Scope[%s] %s", s.ID(), f.Name)
		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debugf("Scope[%s] Creating failed on process: %s. %v", s.ID(), f.Name, err)
			if ts.tx() != nil {
				ts.Rollback()
			}
			return err
		}

		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return nil
}

// Get initializes the Get Process chain for the scope
func (p *Processor) Get(ctx context.Context, s *scope.Scope) error {
	for _, f := range p.GetChain {
		log.Debug2f("Scope[%s] %s", s.ID(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debugf("Scope[%s] Getting failed on process: %s. %v", s.ID(), f.Name, err)
			if ts.tx() != nil {
				ts.Rollback()
			}
			return err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return nil
}

// List initializes the List Process Chain for the scope
func (p *Processor) List(ctx context.Context, s *scope.Scope) error {
	for _, f := range p.ListChain {
		log.Debug2f("Scope[%s] %s", s.ID(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debugf("Scope[%s] Listing failed on process: %s. %v", s.ID(), f.Name, err)
			if ts.tx() != nil {
				ts.Rollback()
			}
			return err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}
	return nil
}

// Patch does the Patch Process Chain
func (p *Processor) Patch(ctx context.Context, s *scope.Scope) error {
	for _, f := range p.PatchChain {
		log.Debug2f("Scope[%s] %s", s.ID(), f.Name)

		ts := (*Scope)(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debugf("Scope[%s] Patching failed on process: %s. %v", s.ID(), f.Name, err)
			if ts.tx() != nil {
				ts.Rollback()
			}
			return err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return nil
}

// Delete does the Delete process chain
func (p *Processor) Delete(ctx context.Context, s *scope.Scope) error {
	for _, f := range p.DeleteChain {
		log.Debug2f("Scope[%s] %s", s.ID(), f.Name)

		ts := queryS(s)
		if err := f.Func(ctx, ts); err != nil {
			log.Debugf("Scope[%s] Deleting failed on process: %s. %v", s.ID(), f.Name, err)
			if ts.tx() != nil {
				ts.Rollback()
			}
			return err
		}
		s.StoreSet(internal.PreviousProcessStoreKey, f)
	}

	return nil
}
