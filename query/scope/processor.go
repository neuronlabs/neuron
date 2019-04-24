package scope

import (
	"github.com/neuronlabs/neuron/internal/query/scope"
)

// ProcessFunc is the function that modifies or changes the scope value
type ProcessFunc func(s *Scope) error

// DefaultQueryProcessor is the default Query Processor set for scopes
var DefaultQueryProcessor *QueryProcessor

func init() {
	DefaultQueryProcessor = newQueryProcessor()
}

type transactionState int8

const (
	transUnset transactionState = iota
	transBegan
	transFailed
)

// Process is the pair of the name and the ProcessFunction
type Process struct {
	Name string
	Func ProcessFunc
}

// QueryProcessor is the struct that allows to query over the gateway's model's
type QueryProcessor struct {
	CreateChain ProcessChain
	GetChain    ProcessChain
	ListChain   ProcessChain
	PatchChain  ProcessChain
	DeleteChain ProcessChain
}

// New creates the query processor
func newQueryProcessor() *QueryProcessor {
	return &QueryProcessor{
		CreateChain: ProcessChain{
			ProcessBeforeCreate,
			ProcessCreate,
			ProcessAfterCreate,
		},
		GetChain: ProcessChain{
			ProcessBeforeGet,
			ProcessSetBelongsToRelationships,
			ProcessGet,
			ProcessGetForeignRelationships,
			ProcessAfterGet,
		},
		ListChain: ProcessChain{
			ProcessBeforeList,
			ProcessSetBelongsToRelationships,
			ProcessList,
			ProcessAfterList,
			ProcessGetForeignRelationships,
			ProcessGetIncluded,
		},
		PatchChain: ProcessChain{
			ProcessBeforePatch,
			ProcessPatchBelongsToRelationships,
			ProcessPatch,
			ProcessPatchForeignRelationships,
			ProcessAfterPatch,
		},
		DeleteChain: ProcessChain{
			ProcessBeforeDelete,
			ProcessDelete,
			ProcessAfterDelete,
			ProcessDeleteForeignRelationships,
		},
	}
}

// DoCreate is the initializes the Create Process Chain for the Scope
func (q *QueryProcessor) DoCreate(s *Scope) error {

	for _, f := range q.CreateChain {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

// DoGet initializes the Get Process chain for the scope
func (q *QueryProcessor) DoGet(s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.GetChain {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

// DoList initializes the List Process Chain for the scope
func (q *QueryProcessor) DoList(s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.ListChain {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

// DoPatch does the Patch Process Chain
func (q *QueryProcessor) DoPatch(s *Scope) error {
	for _, f := range q.PatchChain {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

// DoDelete does the Delete process chain
func (q *QueryProcessor) DoDelete(s *Scope) error {
	for _, f := range q.DeleteChain {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}
