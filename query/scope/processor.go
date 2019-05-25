package scope

import (
	"context"
	"github.com/google/uuid"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

// ProcessFunc is the function that modifies or changes the scope value
type ProcessFunc func(ctx context.Context, s *Scope) error

// DefaultQueryProcessor is the default Query Processor set for scopes
var DefaultQueryProcessor *QueryProcessor

func init() {
	DefaultQueryProcessor = newQueryProcessor()
}

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
			ProcessSetBelongsToRelationships,
			ProcessCreate,
			ProcessPatchForeignRelationships,
			ProcessAfterCreate,
		},
		GetChain: ProcessChain{
			ProcessConvertRelationshipFilters,
			ProcessBeforeGet,
			ProcessConvertRelationshipFilters,
			ProcessGet,
			ProcessGetForeignRelationships,
			ProcessAfterGet,
		},
		ListChain: ProcessChain{
			ProcessConvertRelationshipFilters, // convert the relationship filters that before the 'before' hook
			ProcessBeforeList,
			ProcessConvertRelationshipFilters, // convert the relationship filters after 'before' hook
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

// CreateTxID creates the new transaction ID
func (q *QueryProcessor) CreateTxID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

// DoCreate is the initializes the Create Process Chain for the Scope
func (q *QueryProcessor) DoCreate(ctx context.Context, s *Scope) error {

	for _, f := range q.CreateChain {
		if err := f.Func(ctx, s); err != nil {
			if s.tx() != nil {
				s.Rollback()
			}
			return err
		}
	}

	return nil
}

// DoGet initializes the Get Process chain for the scope
func (q *QueryProcessor) DoGet(ctx context.Context, s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.GetChain {
		if err := f.Func(ctx, s); err != nil {
			return err
		}
	}

	return nil
}

// DoList initializes the List Process Chain for the scope
func (q *QueryProcessor) DoList(ctx context.Context, s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.ListChain {
		if err := f.Func(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// DoPatch does the Patch Process Chain
func (q *QueryProcessor) DoPatch(ctx context.Context, s *Scope) error {
	for _, f := range q.PatchChain {
		if err := f.Func(ctx, s); err != nil {
			if s.tx() != nil {
				s.Rollback()
			}
			return err
		}
	}

	return nil
}

// DoDelete does the Delete process chain
func (q *QueryProcessor) DoDelete(ctx context.Context, s *Scope) error {
	for _, f := range q.DeleteChain {
		if err := f.Func(ctx, s); err != nil {
			if s.tx() != nil {
				s.Rollback()
			}
			return err
		}
	}

	return nil
}
