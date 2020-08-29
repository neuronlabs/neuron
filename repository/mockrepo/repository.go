package mockrepo

import (
	"context"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
)

var (
	_ repository.Transactioner = &Repository{}
	_ repository.Repository    = &Repository{}
)

// Repository is a mock repository implementation.
type Repository struct {
	IDValue                string
	Beginners              []*TransExecuter
	Committers             []*TransExecuter
	Rollbackers            []*TransExecuter
	Savepointers           []*SavepointExecuter
	RollbackerSavepointers []*SavepointExecuter
	Inserters              []*CommonExecuter
	Finders                []*CommonExecuter
	Counters               []*ResultExecuter
	Updaters               []*ResultExecuter
	ModelUpdaters          []*ResultExecuter
	Deleters               []*ResultExecuter
}

// ID implements repository.Repository.
func (r *Repository) ID() string {
	if r.IDValue != "" {
		return r.IDValue
	}
	return "test-repo"
}

// OnBegin adds begin executor.
func (r *Repository) OnBegin(transFunc TransFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Beginners = append(r.Beginners, &TransExecuter{Options: o, ExecuteFunc: transFunc})
}

// Begin implements repository.Transactioner interface.
func (r *Repository) Begin(ctx context.Context, tx *query.Transaction) error {
	if len(r.Beginners) == 0 {
		log.Panicf("no beginners found")
	}
	beginner := r.Beginners[0]
	if beginner.Options.Count > 0 {
		beginner.Options.Count--
	}
	if beginner.Options.Count == 0 && !beginner.Options.Permanent {
		r.Beginners = r.Beginners[1:]
	}
	return beginner.ExecuteFunc(ctx, tx)
}

// OnCommit adds the committer executer.
func (r *Repository) OnCommit(transFunc TransFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Committers = append(r.Committers, &TransExecuter{Options: o, ExecuteFunc: transFunc})
}

// Commit implements repository.Transactioner.
func (r *Repository) Commit(ctx context.Context, tx *query.Transaction) error {
	if len(r.Committers) == 0 {
		log.Panicf("no committers found")
	}
	committer := r.Committers[0]
	if committer.Options.Count > 0 {
		committer.Options.Count--
	}
	if committer.Options.Count == 0 && !committer.Options.Permanent {
		r.Committers = r.Committers[1:]
	}
	return committer.ExecuteFunc(ctx, tx)
}

// OnRollback adds the committer executer.
func (r *Repository) OnRollback(transFunc TransFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Rollbackers = append(r.Rollbackers, &TransExecuter{Options: o, ExecuteFunc: transFunc})
}

// Rollback implements repository.Transactioner.
func (r *Repository) Rollback(ctx context.Context, tx *query.Transaction) error {
	if len(r.Rollbackers) == 0 {
		log.Panicf("no committers found")
	}
	rollBacker := r.Rollbackers[0]
	if rollBacker.Options.Count > 0 {
		rollBacker.Options.Count--
	}
	if rollBacker.Options.Count == 0 && !rollBacker.Options.Permanent {
		r.Rollbackers = r.Rollbackers[1:]
	}
	return rollBacker.ExecuteFunc(ctx, tx)
}

// OnCount adds the count function executioner.
func (r *Repository) OnCount(resultFunc ResultFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Counters = append(r.Counters, &ResultExecuter{Options: o, ExecuteFunc: resultFunc})
}

// Count implements repository.Repository interface.
func (r *Repository) Count(ctx context.Context, s *query.Scope) (int64, error) {
	if len(r.Counters) == 0 {
		log.Panicf("no counters found: %+v", s)
	}
	counter := r.Counters[0]
	if counter.Options.Count > 0 {
		counter.Options.Count--
	}
	if counter.Options.Count == 0 && !counter.Options.Permanent {
		r.Counters = r.Counters[1:]
	}
	return counter.ExecuteFunc(ctx, s)
}

// OnInsert adds the insert executioner function.
func (r *Repository) OnInsert(insertFunc CommonFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Inserters = append(r.Inserters, &CommonExecuter{Options: o, ExecuteFunc: insertFunc})
}

// Insert implements repository.Repository interface.
func (r *Repository) Insert(ctx context.Context, s *query.Scope) error {
	if len(r.Inserters) == 0 {
		log.Panicf("no inserters found: %+v", s)
	}
	inserter := r.Inserters[0]
	if inserter.Options.Count > 0 {
		inserter.Options.Count--
	}
	if inserter.Options.Count == 0 && !inserter.Options.Permanent {
		r.Inserters = r.Inserters[1:]
	}
	return inserter.ExecuteFunc(ctx, s)
}

// OnFind adds the find executioner function.
func (r *Repository) OnFind(findFunc CommonFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Finders = append(r.Finders, &CommonExecuter{Options: o, ExecuteFunc: findFunc})
}

// Find implements repository.Repository interface.
func (r *Repository) Find(ctx context.Context, s *query.Scope) error {
	if len(r.Finders) == 0 {
		log.Panicf("no finder found: %+v", s)
	}
	finder := r.Finders[0]
	if finder.Options.Count > 0 {
		finder.Options.Count--
	}
	if finder.Options.Count == 0 && !finder.Options.Permanent {
		r.Finders = r.Finders[1:]
	}
	return finder.ExecuteFunc(ctx, s)
}

// OnUpdate adds the count function executioner.
func (r *Repository) OnUpdate(updateFunc ResultFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Updaters = append(r.Updaters, &ResultExecuter{Options: o, ExecuteFunc: updateFunc})
}

// Update implements repository.Repository.
func (r *Repository) Update(ctx context.Context, s *query.Scope) (int64, error) {
	if len(r.Updaters) == 0 {
		log.Panicf("no updaters found: %+v", s)
	}
	updater := r.Updaters[0]
	if updater.Options.Count > 0 {
		updater.Options.Count--
	}
	if updater.Options.Count == 0 && !updater.Options.Permanent {
		r.Updaters = r.Updaters[1:]
	}
	return updater.ExecuteFunc(ctx, s)
}

// OnUpdateModels adds the count function executioner.
func (r *Repository) OnUpdateModels(updateFunc ResultFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.ModelUpdaters = append(r.ModelUpdaters, &ResultExecuter{Options: o, ExecuteFunc: updateFunc})
}

// UpdateModels implements repository.Repository.
func (r *Repository) UpdateModels(ctx context.Context, s *query.Scope) (int64, error) {
	if len(r.ModelUpdaters) == 0 {
		log.Panicf("no updaters found: %+v", s)
	}
	updater := r.ModelUpdaters[0]
	if updater.Options.Count > 0 {
		updater.Options.Count--
	}
	if updater.Options.Count == 0 && !updater.Options.Permanent {
		r.ModelUpdaters = r.ModelUpdaters[1:]
	}
	return updater.ExecuteFunc(ctx, s)
}

// OnDelete adds the delete function executioner.
func (r *Repository) OnDelete(deleteFunc ResultFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Deleters = append(r.Deleters, &ResultExecuter{Options: o, ExecuteFunc: deleteFunc})
}

// Delete implements repository.Repository.
func (r *Repository) Delete(ctx context.Context, s *query.Scope) (int64, error) {
	if len(r.Deleters) == 0 {
		log.Panicf("no deleter found: %+v", s)
	}
	deleter := r.Deleters[0]
	if deleter.Options.Count > 0 {
		deleter.Options.Count--
	}
	if deleter.Options.Count == 0 && !deleter.Options.Permanent {
		r.Deleters = r.Deleters[1:]
	}
	return deleter.ExecuteFunc(ctx, s)
}

var _ repository.Savepointer = &Repository{}

// OnSavepoint adds the delete function executioner.
func (r *Repository) OnSavepoint(savepointFunc SavepointFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.Savepointers = append(r.Savepointers, &SavepointExecuter{Options: o, ExecuteFunc: savepointFunc})
}

func (r *Repository) Savepoint(ctx context.Context, tx *query.Transaction, name string) error {
	if len(r.Savepointers) == 0 {
		log.Panicf("no deleter found: %+v", tx.ID)
	}
	savepointer := r.Savepointers[0]
	if savepointer.Options.Count > 0 {
		savepointer.Options.Count--
	}
	if savepointer.Options.Count == 0 && !savepointer.Options.Permanent {
		r.Savepointers = r.Savepointers[1:]
	}
	return savepointer.ExecuteFunc(ctx, tx, name)
}

// OnRollbackSavepoint adds the delete function executioner.
func (r *Repository) OnRollbackSavepoint(savepointFunc SavepointFunc, options ...Option) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	r.RollbackerSavepointers = append(r.RollbackerSavepointers, &SavepointExecuter{Options: o, ExecuteFunc: savepointFunc})
}

func (r *Repository) RollbackSavepoint(ctx context.Context, tx *query.Transaction, name string) error {
	if len(r.RollbackerSavepointers) == 0 {
		log.Panicf("no deleter found: %+v", tx.ID)
	}
	savepointer := r.RollbackerSavepointers[0]
	if savepointer.Options.Count > 0 {
		savepointer.Options.Count--
	}
	if savepointer.Options.Count == 0 && !savepointer.Options.Permanent {
		r.RollbackerSavepointers = r.RollbackerSavepointers[1:]
	}
	return savepointer.ExecuteFunc(ctx, tx, name)
}
