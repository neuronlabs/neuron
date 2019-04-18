package scope

import (
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
)

type processFunc func(s *Scope) error

var DefaultQueryProcessor *queryProcessor

func init() {
	DefaultQueryProcessor = newqueryProcessor()
}

type transactionState int8

const (
	transUnset transactionState = iota
	transBegan
	transFailed
)

type Process struct {
	Name string
	Func processFunc
}

type processes []Process

func (p processes) Register(s Process) {
	p = append(p, s)
}

// queryProcessor is the struct that allows to query over the gateway's model's
type queryProcessor struct {
	creates processes
	gets    processes
	lists   processes
	patches processes
	deletes processes
}

// New creates the query processor
func newqueryProcessor() *queryProcessor {
	return &queryProcessor{
		creates: processes{
			beforeCreate,
			create,
			afterCreate,
		},
		gets: []Process{
			beforeGet,
			setBelongsToRelationships,
			get,
			getForeignRelationships,
			afterGet,
		},
		lists: []Process{
			beforeList,
			setBelongsToRelationships,
			list,
			afterList,
			getForeignRelationships,
			getIncluded,
		},
		patches: []Process{
			beforePatch,
			patchBelongsToRelationships,
			patch,
			patchForeignRelationships,
			afterPatch,
		},
		deletes: []Process{
			beforeDelete,
			deleteProcess,
			afterDelete,
			deleteForeignRelationships,
		},
	}
}

func (q *queryProcessor) RegisterCreateProcess(p Process) {
	log.Debugf("Registering Create Process: %s", p.Name)
	q.creates = append(q.creates, p)
}

// Create is the query create
func (q *queryProcessor) doCreate(s *Scope) error {

	for _, f := range q.creates {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryProcessor) doGet(s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.gets {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryProcessor) doList(s *Scope) error {
	(*scope.Scope)(s).FillFieldsetIfNotSet()

	for _, f := range q.lists {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryProcessor) doPatch(s *Scope) error {
	for _, f := range q.patches {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryProcessor) doDelete(s *Scope) error {
	for _, f := range q.deletes {
		if err := f.Func(s); err != nil {
			return err
		}
	}
	return nil
}
