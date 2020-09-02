package database

import (
	"context"
	"sync"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
)

// New creates new DB for given controller.
func New(options ...Option) (*base, error) {
	o := &Options{
		RepositoryModels: map[repository.Repository][]mapping.Model{},
		TimeFunc:         time.Now,
	}
	for _, option := range options {
		option(o)
	}
	if o.ModelMap == nil {
		o.ModelMap = mapping.New()
	}
	if len(o.RepositoryModels) == 0 && o.DefaultRepository == nil {
		return nil, errors.Wrap(ErrRepository, "no repositories registered")
	}
	d := &base{
		repositories: &RepositoryMapper{
			Repositories:      map[string]repository.Repository{},
			ModelRepositories: map[*mapping.ModelStruct]repository.Repository{},
			ModelMap:          o.ModelMap,
		},
		options: o,
	}
	if o.DefaultRepository != nil {
		d.repositories.DefaultRepository = o.DefaultRepository
	}
	for repo, models := range o.RepositoryModels {
		d.repositories.RegisterRepositories(repo)
		for _, model := range models {
			mStruct, err := o.ModelMap.ModelStruct(model)
			if err != nil {
				return nil, err
			}
			d.repositories.ModelRepositories[mStruct] = repo
		}
	}
	return d, nil
}

// Compile time check for the DB interface implementations.
var _ DB = &base{}

// Composer is the default query composer that implements DB interface.
type base struct {
	options      *Options
	repositories *RepositoryMapper
	closed       bool
}

// Begin starts new transaction with respect to the transaction context and transaction options with controller 'c'.
func (b *base) Begin(ctx context.Context, options *query.TxOptions) *Tx {
	return begin(ctx, b, options)
}

// Dial initialize connection with the database.
func (b *base) Dial(ctx context.Context) error {
	var cancelFunc context.CancelFunc
	if _, deadlineSet := ctx.Deadline(); !deadlineSet {
		// if no default timeout is already set - try with 30 second timeout.
		ctx, cancelFunc = context.WithTimeout(ctx, time.Second*30)
	} else {
		// otherwise create a cancel function.
		ctx, cancelFunc = context.WithCancel(ctx)
	}
	defer cancelFunc()

	// Dial repositories.
	if err := b.dialRepositories(ctx); err != nil {
		return err
	}
	// Migrate all marked models.
	if err := b.migrateModels(ctx); err != nil {
		// If an error occurred close the connections.
		if er := b.Close(ctx); er != nil {
			log.Errorf("Closing connection failed: %v", er)
		}
		return err
	}
	return nil
}

func (b *base) migrateModels(ctx context.Context) error {
	for _, model := range b.options.MigrateModels {
		mStruct, err := b.repositories.ModelMap.ModelStruct(model)
		if err != nil {
			return err
		}
		repo, err := b.repositories.GetRepositoryByModelStruct(mStruct)
		if err != nil {
			return err
		}
		migrator, ok := repo.(repository.Migrator)
		if !ok {
			return errors.Wrapf(ErrRepository, "cannot migrate model: '%s' to repository: '%v' - repository doesn't implement Migrator interface", mStruct, repo)
		}
		if err := migrator.MigrateModels(ctx, mStruct); err != nil {
			return err
		}
	}
	return nil
}

func (b *base) dialRepositories(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	waitChan := make(chan struct{})

	jobs := b.dialJobsCreator(ctx, wg)
	// Create error channel.
	errChan := make(chan error)
	// Dial to all repositories.
	for job := range jobs {
		b.dial(ctx, job, wg, errChan)
	}
	// Create wait group channel finish function.
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Errorf("Dial - context deadline exceeded: %v", ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Errorf("Dial error: %v", e)
		return e
	case <-waitChan:
		log.Debug("Successful dial to all repositories")
	}
	return nil
}

func (b *base) dialJobsCreator(ctx context.Context, wg *sync.WaitGroup) <-chan repository.Dialer {
	out := make(chan repository.Dialer)
	go func() {
		defer close(out)

		for _, repo := range b.repositories.Repositories {
			dialer, isDialer := repo.(repository.Dialer)
			if !isDialer {
				continue
			}
			wg.Add(1)
			select {
			case out <- dialer:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func (b *base) dial(ctx context.Context, dialer repository.Dialer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := dialer.Dial(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}

//  Close closes the database connections.
func (b *base) Close(ctx context.Context) error {
	var cancelFunc context.CancelFunc
	if _, deadlineSet := ctx.Deadline(); !deadlineSet {
		// if no default timeout is already set - try with 30 second timeout.
		ctx, cancelFunc = context.WithTimeout(ctx, time.Second*30)
	} else {
		// otherwise create a cancel function.
		ctx, cancelFunc = context.WithCancel(ctx)
	}
	defer cancelFunc()

	wg := &sync.WaitGroup{}
	waitChan := make(chan struct{})
	jobs := b.closeJobsCreator(ctx, wg)

	errChan := make(chan error)
	for job := range jobs {
		log.Debugf("Closing: %T", job)
		b.closeCloser(ctx, job, wg, errChan)
	}

	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		log.Errorf("Close - context deadline exceeded: %v", ctx.Err())
		return ctx.Err()
	case e := <-errChan:
		log.Debugf("Close error: %v", e)
		return e
	case <-waitChan:
		log.Debug("Closed all repositories with success")
	}
	return nil
}

func (b *base) closeJobsCreator(ctx context.Context, wg *sync.WaitGroup) <-chan repository.Closer {
	out := make(chan repository.Closer)
	go func() {
		defer close(out)

		// Close all repositories.
		for _, repo := range b.repositories.Repositories {
			closer, isCloser := repo.(repository.Closer)
			if !isCloser {
				continue
			}
			wg.Add(1)
			select {
			case out <- closer:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func (b *base) closeCloser(ctx context.Context, closer repository.Closer, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		defer wg.Done()
		if err := closer.Close(ctx); err != nil {
			errChan <- err
			return
		}
	}()
}

// Repositories gets the repository mapping.
func (b *base) mapper() *RepositoryMapper {
	return b.repositories
}

// ModelMap returns model map.
func (b *base) ModelMap() *mapping.ModelMap {
	return b.repositories.ModelMap
}

// Now
func (b *base) Now() time.Time {
	return b.options.TimeFunc()
}

// Query creates new query builder for given 'model' and it's optional instances 'models'.
func (b *base) Query(model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return b.query(context.Background(), model, models...)
}

// QueryCtx creates new query builder for given 'model' and it's optional instances 'models'.
func (b *base) QueryCtx(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) Builder {
	return b.query(ctx, model, models...)
}

// QueryGet implements QueryGetter interface.
func (b *base) QueryGet(ctx context.Context, q *query.Scope) (mapping.Model, error) {
	return queryGet(ctx, b, q)
}

// QueryGet implements QueryGetter interface.
func (b *base) QueryFind(ctx context.Context, q *query.Scope) ([]mapping.Model, error) {
	return queryFind(ctx, b, q)
}

// Insert implements DB interface.
func (b *base) Insert(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if len(models) == 0 {
		return errors.Wrap(query.ErrNoModels, "nothing to insert")
	}
	s := query.NewScope(mStruct, models...)
	return queryInsert(ctx, b, s)
}

// InsertQuery implements QueryInserter interface.
func (b *base) InsertQuery(ctx context.Context, q *query.Scope) error {
	return queryInsert(ctx, b, q)
}

// Update implements DB interface.
func (b *base) Update(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.Wrap(query.ErrNoModels, "nothing to update")
	}
	s := query.NewScope(mStruct, models...)
	return queryUpdate(ctx, b, s)
}

// UpdateQuery implements QueryUpdater interface.
func (b *base) UpdateQuery(ctx context.Context, q *query.Scope) (int64, error) {
	return queryUpdate(ctx, b, q)
}

// deleteQuery implements DB interface.
func (b *base) Delete(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) (int64, error) {
	if len(models) == 0 {
		return 0, errors.Wrap(query.ErrNoModels, "nothing to delete")
	}
	s := query.NewScope(mStruct, models...)
	return deleteQuery(ctx, b, s)
}

// DeleteQuery implements QueryDeleter interface.
func (b *base) DeleteQuery(ctx context.Context, q *query.Scope) (int64, error) {
	return deleteQuery(ctx, b, q)
}

// Refresh implements DB interface.
func (b *base) Refresh(ctx context.Context, mStruct *mapping.ModelStruct, models ...mapping.Model) error {
	if len(models) == 0 {
		return nil
	}
	q := query.NewScope(mStruct, models...)
	return refreshQuery(ctx, b, q)
}

// QueryRefresh implements QueryRefresher interface.
func (b *base) QueryRefresh(ctx context.Context, q *query.Scope) error {
	return refreshQuery(ctx, b, q)
}

//
// Relations
//

func (b *base) AddRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	mStruct, err := b.mapper().ModelMap.ModelStruct(model)
	if err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	return queryAddRelations(ctx, b, q, relationField, relations...)
}

// QueryAddRelations implements QueryRelationAdder interface.
func (b *base) QueryAddRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relations ...mapping.Model) error {
	return queryAddRelations(ctx, b, s, relationField, relations...)
}

// querySetRelations clears all 'relationField' for the input models and set their values to the 'relations'.
// The relation's foreign key must be allowed to set to null.
func (b *base) SetRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField, relations ...mapping.Model) error {
	mStruct, err := b.mapper().ModelMap.ModelStruct(model)
	if err != nil {
		return err
	}
	q := query.NewScope(mStruct, model)
	return querySetRelations(ctx, b, q, relationField, relations...)
}

var _ QueryRelationSetter = &base{}

// QuerySetRelations implements QueryRelationSetter interface.
func (b *base) QuerySetRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField, relations ...mapping.Model) error {
	return querySetRelations(ctx, b, s, relationField, relations...)
}

// ClearRelations clears all 'relationField' relations for given input models.
// The relation's foreign key must be allowed to set to null.
func (b *base) ClearRelations(ctx context.Context, model mapping.Model, relationField *mapping.StructField) (int64, error) {
	// TODO(kucjac): allow to clear only selected relation models.
	mStruct, err := b.mapper().ModelMap.ModelStruct(model)
	if err != nil {
		return 0, err
	}
	q := query.NewScope(mStruct, model)
	return queryClearRelations(ctx, b, q, relationField)
}

var _ QueryRelationClearer = &base{}

// QueryClearRelations implements QueryRelationClearer interface.
func (b *base) QueryClearRelations(ctx context.Context, s *query.Scope, relationField *mapping.StructField) (int64, error) {
	return queryClearRelations(ctx, b, s, relationField)
}

// IncludeRelation gets the relations at the 'relationField' for provided models. An optional relationFieldset might be provided.
func (b *base) IncludeRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	return queryIncludeRelation(ctx, b, mStruct, models, relationField, relationFieldset...)
}

// GetRelations implements DB interface.
func (b *base) GetRelations(ctx context.Context, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	return queryGetRelations(ctx, b, mStruct, models, relationField, relationFieldset...)
}

func (b *base) query(ctx context.Context, model *mapping.ModelStruct, models ...mapping.Model) *dbQuery {
	q := &dbQuery{ctx: ctx, db: b}
	q.scope = query.NewScope(model, models...)
	return q
}

func (b *base) synchronousConnections() bool {
	return b.options.SynchronousConnections
}
