package query

import (
	"context"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/filters"

	"github.com/neuronlabs/neuron/internal"
	internalController "github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	internalFilters "github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/paginations"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

// Scope is the query's structure that contains all information required to
// get, create, patch or delete the data in the repository.
type Scope scope.Scope

// MustNewC creates new scope with given 'model' for the given internalController 'c'.
func MustNewC(c *controller.Controller, model interface{}) *Scope {
	s, err := newScope((*internalController.Controller)(c), model)
	if err != nil {
		panic(err)
	}
	return s
}

// NewC creates the scope for the provided model with respect to the provided internalController 'c'.
func NewC(c *controller.Controller, model interface{}) (*Scope, error) {
	return newScope((*internalController.Controller)(c), model)
}

// NewModelC creates new scope on the base of the provided model struct with the value.
// The value might be a slice of instances if 'isMany' is true.
func NewModelC(c *controller.Controller, mStruct *mapping.ModelStruct, isMany bool) *Scope {
	return (*Scope)(newScopeWithModel(c, (*models.ModelStruct)(mStruct), isMany))
}

// New creates the scope on the base of the given 'model' it uses default internalController.
func New(model interface{}) (*Scope, error) {
	return newScope(internalController.Default(), model)
}

// AddTxChain adds a transaction subscope to the given query's scope transaction chain.
// By default scopes created by the 'New' method are added to the transaction chain and
// should not be added again.
func (s *Scope) AddTxChain(sub *Scope) {
	s.internal().AddChainSubscope(sub.internal())
}

// AddFilter adds the filter field to the given query.
func (s *Scope) AddFilter(filter *filters.FilterField) error {
	return s.internal().AddFilterField((*internalFilters.FilterField)(filter))
}

// AddStringFilter parses the filter into the filters.FilterField and adds
// it to the given scope.
func (s *Scope) AddStringFilter(rawFilter string, values ...interface{}) error {
	filter, err := filters.NewStringFilter(s.Controller(), rawFilter, s.Struct().SchemaName(), values...)
	if err != nil {
		log.Debugf("BuildRawFilter: '%s' with values: %v failed. %v", rawFilter, values, err)
		return err
	}
	return s.AddFilter(filter)
}

// AddToFieldset adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (s *Scope) AddToFieldset(fields ...interface{}) error {
	for i, field := range fields {
		mField, ok := field.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(mField)
		}
	}
	return s.internal().AddToFieldset(fields...)
}

// AddSelectedFields adds provided fields into the scope's selected fields.
// This would affect the Create or Patch processes where the SelectedFields are taken
// as the unmarshaled fields.
func (s *Scope) AddSelectedFields(fields ...interface{}) error {
	for i, f := range fields {
		// cast all *mapping.StructFields into models.StructField
		field, ok := f.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(field)
		}
	}
	return s.internal().AddSelectedFields(fields...)
}

// AddStringSortFields adds the sort fields in a string form.
// The string form may look like: '[-field_1, field_2]' which means:
//	 - Descending 'field_1'
//	 - Ascending 'field_2'.
func (s *Scope) AddStringSortFields(fields ...string) error {
	errs := s.internal().BuildSortFields(fields...)
	if len(errs) > 0 {
		return errors.MultiError(errs)
	}
	return nil
}

// AttributeFilters returns scope's attribute filters.
func (s *Scope) AttributeFilters() []*filters.FilterField {
	var attributeFilters []*filters.FilterField
	for _, filter := range s.internal().AttributeFilters() {
		attributeFilters = append(attributeFilters, (*filters.FilterField)(filter))
	}
	return attributeFilters
}

// Begin begins the transaction for the provided scope with the default Options.
func (s *Scope) Begin() (*Tx, error) {
	return s.begin(context.Background(), nil, true)
}

// BeginTx begins the transaction with the given context.Context 'ctx' and Options 'opts'.
func (s *Scope) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error) {
	return s.begin(ctx, opts, true)
}

// Commit commits the transaction for the scope and it's subscopes.
func (s *Scope) Commit() error {
	return s.commit(context.Background())
}

// CommitContext commits the given transaction for the scope with the context.Context 'ctx'.
func (s *Scope) CommitContext(ctx context.Context) error {
	return s.commit(ctx)
}

// Controller gets the scope's internalController.
func (s *Scope) Controller() *controller.Controller {
	cval, _ := s.StoreGet(internal.ControllerStoreKey)
	return cval.(*controller.Controller)

}

// Create stores the values within the given scope's value repository, by starting
// the create process.
func (s *Scope) Create() error {
	return s.createContext(context.Background())
}

// CreateContext creates the scope values with the provided 'ctx' context.Context.
func (s *Scope) CreateContext(ctx context.Context) error {
	return s.createContext(ctx)
}

// Delete deletes the values provided in the query's scope.
func (s *Scope) Delete() error {
	if err := s.defaultProcessor().Delete(context.Background(), (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

// DeleteContext deletes the values provided in the scope's value with the context.
func (s *Scope) DeleteContext(ctx context.Context) error {
	if err := s.defaultProcessor().Delete(ctx, (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

// Fieldset returns the fields in the scope's Fieldset.
func (s *Scope) Fieldset() (fs []*mapping.StructField) {
	for _, f := range (*scope.Scope)(s).Fieldset() {
		fs = append(fs, (*mapping.StructField)(f))
	}
	return fs
}

// FilterKeyFilters returns scope's primary key filters.
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	filterKeys := make([]*filters.FilterField, len(s.internal().FilterKeyFilters()))
	for i, filter := range s.internal().FilterKeyFilters() {
		filterKeys[i] = (*filters.FilterField)(filter)
	}
	return filterKeys
}

// ForeignFilters returns scope's foreign key filters.
func (s *Scope) ForeignFilters() []*filters.FilterField {
	var foreignKeyFilters []*filters.FilterField
	for _, filter := range s.internal().ForeignKeyFilters() {
		foreignKeyFilters = append(foreignKeyFilters, (*filters.FilterField)(filter))
	}
	return foreignKeyFilters
}

// FormatQuery formats the scope's query into the url.Values.
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
}

// Get gets single value from the repository taking into account the scope
// filters and parameters
func (s *Scope) Get() error {
	if err := s.defaultProcessor().Get(context.Background(), (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

// GetContext gets single value from repository taking into account the scope
// filters, parameters and the context.
func (s *Scope) GetContext(ctx context.Context) error {
	if err := s.defaultProcessor().Get(ctx, (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

// ID returns the scope's identity number stored as the UUID.
func (s *Scope) ID() uuid.UUID {
	return (*scope.Scope)(s).ID()
}

// // IncludeFields adds the included fields into the root scope
// func (s *Scope) IncludeFields(fields ...string) error {
// 	iscope := (*scope.Scope)(s)
// 	iscope.InitializeIncluded((*internalController.Controller)(s.Controller()).QueryBuilder().Config.IncludeNestedLimit)
// 	if errs := iscope.BuildIncludeList(fields...); len(errs) > 0 {
// 		return errors.MultiAPIErrors(errs)
// 	}
// 	return nil
// }

// IncludedValue gets the scope's included values for the given 'model'.
// The returning value would be pointer to slice of pointer to models.
// i.e.: type Model struct {}, the result would be returned as a *[]*Model{}.
func (s *Scope) IncludedValue(model interface{}) (interface{}, error) {
	var (
		mStruct *mapping.ModelStruct
		ok      bool
		err     error
	)

	if mStruct, ok = model.(*mapping.ModelStruct); !ok {
		mStruct, err = s.Controller().ModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	included, ok := s.internal().IncludeScopeByStruct((*models.ModelStruct)(mStruct))
	if !ok {
		log.Info("Model: '%s' is not included into scope of: '%s'", mStruct.Collection(), s.Struct().Collection())
		return nil, errors.New(class.QueryNotIncluded, "provided model is not included within query's scope")
	}

	return included.Value, nil
}

// InFieldset checks if the provided field is in the scope's fieldset.
func (s *Scope) InFieldset(field string) (*mapping.StructField, bool) {
	f, ok := (*scope.Scope)(s).InFieldset(field)
	if ok {
		return (*mapping.StructField)(f), true
	}
	return nil, false
}

// IsSelected checks if the provided 'field' is selected within given query's scope.
func (s *Scope) IsSelected(field interface{}) (bool, error) {
	return s.internal().IsSelected(field)
}

// LanguageFilter returns language filter for given scope.
func (s *Scope) LanguageFilter() *filters.FilterField {
	return (*filters.FilterField)(s.internal().LanguageFilter())
}

// Limit sets the maximum number of objects returned by the List process,
// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as ommitting the 'Offset' clause.
func (s *Scope) Limit(limit, offset int) error {
	if s.internal().Pagination() != nil {
		return errors.New(class.QueryPaginationAlreadySet, "pagination already set")
	}

	p := newLimitOffset(limit, offset)
	if err := p.Check(); err != nil {
		return err
	}
	(*scope.Scope)(s).SetPaginationNoCheck((*paginations.Pagination)(p))

	return nil
}

// List gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (s *Scope) List() error {
	// Check the scope's values is an array
	if !s.internal().IsMany() {
		return errors.New(class.QueryValueType, "provided non slice value for the list query").SetDetail("Single value provided for the List process")
	}
	// list from the processor
	if err := s.defaultProcessor().List(context.Background(), (*scope.Scope)(s)); err != nil {
		return err
	}

	return nil
}

// ListContext gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values. Provided context.Context 'ctx'
// would be used while querying the repositories.
func (s *Scope) ListContext(ctx context.Context) error {
	if err := s.defaultProcessor().List(ctx, (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

// New creates new subscope with the provided model 'value'.
// If the root scope is on the transacation the new one will be
// added to the root's transaction chain.
// It is a recommended way to create new scope's within hooks if the given scope
// should be included in the given transaction.
func (s *Scope) New(value interface{}) (*Scope, error) {
	return s.newSubscope(context.Background(), value)
}

// NewContext creates new subscope with the provided model 'value' with the context.Context 'ctx'.
// If the root scope is on the transacation the new one will be
// added to the root's transaction chain.
// It is a recommended way to create new scope's within hooks if the given scope
// should be included in the given transaction.
func (s *Scope) NewContext(ctx context.Context, value interface{}) (*Scope, error) {
	return s.newSubscope(ctx, value)
}

// NotSelectedFields returns fields that are not selected in the query.
func (s *Scope) NotSelectedFields(withForeigns ...bool) (notSelected []*mapping.StructField) {
	for _, field := range (*scope.Scope)(s).NotSelectedFields(withForeigns...) {
		notSelected = append(notSelected, (*mapping.StructField)(field))
	}
	return notSelected
}

// Patch updates the scope's attribute and relationship values based on the scope's value and filters.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (s *Scope) Patch() error {
	return s.patch(context.Background())
}

// PatchContext updates the scope's attribute and relationship values based on the scope's value and filters
// with respect to the context.Context 'ctx'.
// In order to start patch process scope should contain a value with the non-zero primary field, or
// primary field filters.
func (s *Scope) PatchContext(ctx context.Context) error {
	return s.patch(ctx)
}

// Page sets the pagination of the type TpPage with the page 'number' and page 'size'.
func (s *Scope) Page(number, size int) error {
	if s.internal().Pagination() != nil {
		return errors.New(class.QueryPaginationAlreadySet, "pagination already set")
	}

	p := newPaged(number, size)
	if err := p.Check(); err != nil {
		return err
	}

	s.internal().SetPaginationNoCheck((*paginations.Pagination)(p))
	return nil
}

// Pagination returns the query pagination for given scope.
func (s *Scope) Pagination() *Pagination {
	return (*Pagination)(s.internal().Pagination())
}

// PrimaryFilters returns scope's primary filters.
func (s *Scope) PrimaryFilters() []*filters.FilterField {
	primaryFilters := make([]*filters.FilterField, len(s.internal().PrimaryFilters()))
	for i, filter := range s.internal().PrimaryFilters() {
		primaryFilters[i] = (*filters.FilterField)(filter)
	}
	return primaryFilters
}

// RelationshipFilters returns scope's relation fields filters.
func (s *Scope) RelationshipFilters() []*filters.FilterField {
	relationFilters := make([]*filters.FilterField, len(s.internal().RelationshipFilters()))
	for i, filter := range s.internal().RelationshipFilters() {
		relationFilters[i] = (*filters.FilterField)(filter)
	}
	return relationFilters
}

// Rollback does the transaction rollback process.
func (s *Scope) Rollback() error {
	return s.rollback(context.Background())
}

// RollbackContext does the transaction rollback process with the given context.Context 'ctx'.
func (s *Scope) RollbackContext(ctx context.Context) error {
	return s.rollback(ctx)
}

// SetPagination sets the Pagination for the scope.
func (s *Scope) SetPagination(p *Pagination) error {
	return s.internal().SetPagination((*paginations.Pagination)(p))
}

// SelectField selects the field by the name.
// Selected fields are used in the patching process.
// By default selected fields are all non zero valued fields in the struct.
func (s *Scope) SelectField(name string) error {
	field, ok := s.Struct().FieldByName(name)
	if !ok {
		log.Debug("Field not found: '%s'", name)
		return errors.Newf(class.QuerySelectedFieldsNotFound, "field: '%s' not found", name)
	}

	s.internal().AddSelectedField((*models.StructField)(field))

	return nil
}

// SelectedFields gets the fields selected to modify/create in the repository.
func (s *Scope) SelectedFields() (selected []*mapping.StructField) {
	for _, field := range s.internal().SelectedFields() {
		selected = append(selected, (*mapping.StructField)(field))
	}
	return selected
}

// SetFieldset sets the fieldset for the 'fields'.
// A field may be a field's name (string), NeuronName (string) or *mapping.StructField.
func (s *Scope) SetFieldset(fields ...interface{}) error {
	for i, field := range fields {
		mField, ok := field.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(mField)
		}
	}
	return s.internal().SetFields(fields...)
}

// SortBy adds the sort fields into given scope.
// If the scope already have sorted fields or the fields are duplicated returns error.
func (s *Scope) SortBy(fields ...string) error {
	if log.Level().IsAllowed(log.LDEBUG3) {
		log.Debug3f("[SCOPE][%s] Sorting by fields: %v ", s.ID(), fields)
	}
	if s.internal().HaveSortFields() {
		sortFields, err := s.internal().CreateSortFields(false, fields...)
		if err != nil {
			return err
		}
		s.internal().AppendSortFields(true, sortFields...)
		return nil
	}

	errs := s.internal().BuildSortFields(fields...)
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.MultiError(errs)
	}
}

// SortFields returns the sorts used by the query's scope.
func (s *Scope) SortFields() []*SortField {
	sfs := s.internal().SortFields()
	sortFields := make([]*SortField, len(sfs))

	for i, sf := range sfs {
		sortFields[i] = (*SortField)(sf)
	}
	return sortFields
}

// Struct returns scope's model's structure - *mapping.ModelStruct.
func (s *Scope) Struct() *mapping.ModelStruct {
	mStruct := s.internal().Struct()
	return (*mapping.ModelStruct)(mStruct)
}

// StoreGet gets the value from the scope's Store for given 'key'.
func (s *Scope) StoreGet(key interface{}) (value interface{}, ok bool) {
	return s.internal().StoreGet(key)
}

// StoreSet sets the 'key' and 'value' in the given scope's store.
func (s *Scope) StoreSet(key, value interface{}) {
	s.internal().StoreSet(key, value)
}

// Tx returns the transaction for the given scope if exists.
func (s *Scope) Tx() *Tx {
	return s.tx()
}

// ValidateCreate validates the scope's value with respect to the 'create validator' and
// the 'create' validation tags.
func (s *Scope) ValidateCreate() []*errors.Error {
	v := s.Controller().CreateValidator
	return s.validate(v, "CreateValidator")
}

// ValidatePatch validates the scope's value with respect to the 'Patch Validator'.
func (s *Scope) ValidatePatch() []*errors.Error {
	v := s.Controller().PatchValidator
	return s.validate(v, "PatchValidator")
}

func (s *Scope) begin(ctx context.Context, opts *TxOptions, checkError bool) (*Tx, error) {
	// check if the context contains the transaction
	if v, ok := s.StoreGet(internal.TxStateStoreKey); ok {
		txn := v.(*Tx)
		if checkError {
			if txn.State != TxDone {
				return nil, errors.New(class.QueryTxAlreadyBegin, "transaction had already began")
			}
		}
	}

	// for nil options set it to default value
	if opts == nil {
		opts = &TxOptions{}
	}

	txn := &Tx{
		Options: *opts,
		root:    s,
	}

	// generate the id
	// TODO: enterprise set the uuid based on the namespace of the gateway
	// so that the node name can be taken from the UUID v3 or v5 namespace
	var err error
	txn.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, errors.Newf(class.InternalCommon, "new uuid failed: '%s'", err.Error())
	}
	log.Debug3f("SCOPE[%s][%s] Begins transaction[%s]", s.ID(), s.Struct().Collection(), txn.ID)

	txn.State = TxBegin

	// set the transaction to the context
	s.StoreSet(internal.TxStateStoreKey, txn)

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		log.Errorf("No repository found for the %s model. %s", s.Struct().Collection(), err)
		return nil, err
	}

	transactioner, ok := repo.(Transactioner)
	if !ok {
		log.Errorf("The repository doesn't implement Creater interface for model: %s", s.Struct().Collection())
		err = errors.New(class.RepositoryNotImplementsTransactioner, "repository doesn't implement transactioner")
	}

	if err = transactioner.Begin(ctx, s); err != nil {
		return nil, err
	}

	return txn, nil
}

func (s *Scope) commit(ctx context.Context) error {
	txV, ok := s.StoreGet(internal.TxStateStoreKey)
	if txV == nil || !ok {
		log.Debugf("COMMIT: No transaction found for the scope")
		return errors.New(class.QueryTxNotFound, "transaction not found for the scope")
	}

	tx := txV.(*Tx)

	if tx != nil && ok && tx.State != TxBegin {
		log.Debugf("COMMIT: Transaction already resolved: %s", tx.State)
		return errors.New(class.QueryTxAlreadyResolved, "transaction already resolved")
	}

	chain := s.internal().Chain()

	if len(chain) > 0 {
		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Processor.DefaultTimeout
		for _, sub := range chain {
			if sub.Struct().Config() == nil {
				continue
			}
			if modelRepo := sub.Struct().Config().Repository; modelRepo != nil {
				if tm := modelRepo.MaxTimeout; tm != nil {
					if *tm > maxTimeout {
						maxTimeout = *tm
					}
				}
			}

		}
		ctx, cancel := context.WithTimeout(ctx, maxTimeout)
		defer cancel()

		results := make(chan interface{}, len(chain))
		for _, sub := range chain {
			// create goroutine for the given subscope that commits the query
			go (*Scope)(sub).commitSingle(ctx, results)
		}

		var resultCount int
	fl:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case v, ok := <-results:
				// break when the result count is equal to the length of the chain
				if !ok {
					break fl
				}

				//check if value is an error
				if err, ok := v.(*errors.Error); ok {
					if err.Class != class.RepositoryNotImplementsTransactioner &&
						err.Class != class.QueryTxAlreadyResolved {
						return err
					}
				}

				resultCount++
				if resultCount == len(chain) {
					break fl
				}
			}
		}
	}

	single := make(chan interface{}, 1)
	s.commitSingle(ctx, single)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case v := <-single:
		if err, ok := v.(error); ok {
			return err
		}
	}

	return nil
}

func (s *Scope) createContext(ctx context.Context) error {
	if s.internal().IsMany() {
		// TODO: add create many
		return errors.New(class.QueryValueType, "creating with multiple values in are not supported yet")
	}

	// if no fields were selected set automatically non zero
	if err := s.internal().AutoSelectFields(); err != nil {
		return err
	}

	if err := s.defaultProcessor().Create(ctx, s.internal()); err != nil {
		return err
	}
	return nil
}

func (s *Scope) defaultProcessor() scope.Processor {
	// at first try the scope's processor
	p := s.internal().Processor()
	if p == nil {
		// then try internalController's processor

		p = (*internalController.Controller)(s.Controller()).Processor()
		if p == nil {

			// if nil create new processor for given config
			p = newProcessor(s.Controller().Config.Processor)
			(*internalController.Controller)(s.Controller()).SetProcessor(p)
		}
	}

	return p

}

func (s *Scope) formatQuery() url.Values {
	q := url.Values{}

	for _, prim := range s.PrimaryFilters() {
		prim.FormatQuery(q)
	}

	for _, fk := range s.ForeignFilters() {
		fk.FormatQuery(q)
	}

	for _, attr := range s.AttributeFilters() {
		attr.FormatQuery(q)
	}

	for _, rel := range s.RelationshipFilters() {
		rel.FormatQuery(q)
	}

	if s.LanguageFilter() != nil {
		s.LanguageFilter().FormatQuery(q)
	}

	for _, fk := range s.FilterKeyFilters() {
		fk.FormatQuery(q)
	}

	for _, sort := range s.SortFields() {
		sort.FormatQuery(q)
	}

	if s.Pagination() != nil {
		s.Pagination().FormatQuery(q)
	}

	// TODO: add included fields into query formatting

	return q
}

func (s *Scope) internal() *scope.Scope {
	return (*scope.Scope)(s)
}

func (s *Scope) newSubscope(ctx context.Context, value interface{}) (*Scope, error) {
	sub, err := newScope((*internalController.Controller)(s.Controller()), value)
	if err != nil {
		return nil, err
	}

	sub.internal().SetKind(scope.SubscopeKind)

	if txn := s.tx(); txn != nil {
		if _, err := sub.begin(ctx, &txn.Options, false); err != nil {
			log.Debug("Begin subscope failed: %v", err)
			return nil, err
		}

		s.internal().AddChainSubscope(sub.internal())
	}

	return sub, nil
}

func (s *Scope) patch(ctx context.Context) error {
	// check if scope's value is single
	if s.internal().IsMany() {
		return errors.New(class.QueryValueType, "patching multiple values are not supported yet")
	}

	// if no fields were selected set automatically non zero
	if err := s.internal().AutoSelectFields(); err != nil {
		return err
	}

	if err := s.defaultProcessor().Patch(ctx, (*scope.Scope)(s)); err != nil {
		return err
	}
	return nil
}

func (s *Scope) rollback(ctx context.Context) error {
	txV, ok := s.StoreGet(internal.TxStateStoreKey)
	if txV == nil || !ok {
		log.Debugf("ROLLBACK: No transaction found for the scope")
		return errors.New(class.QueryTxNotFound, "transaction not found within the scope")
	}

	tx := txV.(*Tx)
	if tx != nil && ok && tx.State != TxBegin {
		log.Debugf("ROLLBACK: Transaction already resolved: %s", tx.State)
		return errors.New(class.QueryTxAlreadyResolved, "transaction already resolved")
	}

	chain := s.internal().Chain()

	if len(chain) > 0 {
		results := make(chan interface{}, len(chain))

		// get initial time out from the internalController builder config
		maxTimeout := s.Controller().Config.Processor.DefaultTimeout

		// check if any model has a preset timeout greater than the maxTimeout
		for _, sub := range chain {
			if sub.Struct().Config() == nil {
				continue
			}
			if modelRepo := sub.Struct().Config().Repository; modelRepo != nil {
				if tm := modelRepo.MaxTimeout; tm != nil {
					if *tm > maxTimeout {
						maxTimeout = *tm
					}
				}
			}
		}

		ctx, cancel := context.WithTimeout(ctx, maxTimeout)
		defer cancel()

		for _, sub := range chain {
			// get the cancel functions
			// create goroutine for the given subscope that commits the query
			go (*Scope)(sub).rollbackSingle(ctx, results)
		}

		var rescount int

	fl:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case v, ok := <-results:
				if !ok {
					break fl
				}

				if err, ok := v.(*errors.Error); ok {
					if err.Class != class.RepositoryNotImplementsTransactioner &&
						err.Class != class.QueryTxAlreadyResolved {
						return err
					}
				}

				rescount++
				if rescount == len(chain) {
					break fl
				}
			}
		}
	}

	single := make(chan interface{}, 1)

	go s.rollbackSingle(ctx, single)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case v := <-single:
		if err, ok := v.(*errors.Error); ok {
			if err.Class != class.QueryTxAlreadyResolved {
				return err
			}
		}
	}

	return nil
}

func (s *Scope) validate(v *validator.Validate, validatorName string) []*errors.Error {
	err := v.Struct(s.Value)
	if err == nil {
		return nil
	}

	switch er := err.(type) {
	case *validator.InvalidValidationError:
		// Invalid argument passed to validator
		log.Errorf("[%s] %s-> Invalid Validation Error: %v", s.ID().String(), validatorName, er)
		e := errors.New(class.InternalQueryValidation, "invalid validation error")
		return []*errors.Error{e}
	case validator.ValidationErrors:
		var errs []*errors.Error
		for _, verr := range er {
			tag := verr.Tag()

			var errObj *errors.Error
			if tag == "required" {
				// if field is required and the field tag is empty
				if verr.Field() == "" {
					log.Errorf("[%s] Model: '%v'. '%s' failed. Field is required and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
					errObj = errors.New(class.InternalQueryValidation, "empty field tag")
					return append(errs, errObj)
				}

				errObj = errors.New(class.QueryValueMissingRequired, "missing required field")
				errObj.SetDetailf("The field: %s, is required.", verr.Field())
				errs = append(errs, errObj)
				continue
			} else if tag == "isdefault" {
				if verr.Field() == "" {
					if verr.Field() == "" {
						log.Errorf("[%s] Model: '%v'. '%s' failed. Field is required and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
						errObj = errors.New(class.InternalQueryValidation, "empty field tag")
						return append(errs, errObj)
					}

					errObj = errors.New(class.QueryValueValidation, "non default field value")
					errObj.SetDetailf("The field: '%s' must be of zero value.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if strings.HasPrefix(tag, "len") {
					// length
					if verr.Field() == "" {
						log.Errorf("[%s] Model: '%v'. %s failed. Field must have specific length and the field tag is empty.", s.ID().String(),
							validatorName, s.Struct().Type().String())
						errObj = errors.New(class.InternalQueryValidation, "empty field tag")
						return append(errs, errObj)
					}

					errObj = errors.New(class.QueryValueValidation, "validation failed - field of invalid length")
					errObj.SetDetailf("The value of the field: %s is of invalid length.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else {
					errObj = errors.New(class.QueryValueValidation, "validation failed - invalid field value")
					if verr.Field() != "" {
						errObj.SetDetailf("Invalid value for the field: '%s'.", verr.Field())
					}

					errs = append(errs, errObj)
					continue
				}
			}
		}
		return errs
	default:
		return []*errors.Error{errors.Newf(class.InternalQueryValidation, "invalid error type: '%T'", er)}
	}
}

func (s *Scope) tx() *Tx {
	txV, ok := s.StoreGet(internal.TxStateStoreKey)
	if ok {
		return txV.(*Tx)
	}
	return nil
}

func newScope(c *internalController.Controller, model interface{}) (*Scope, error) {
	var (
		err     error
		mStruct *models.ModelStruct
	)
	switch mt := model.(type) {
	case *models.ModelStruct:
		mStruct = mt
	case *mapping.ModelStruct:
		mStruct = (*models.ModelStruct)(mt)
	default:
		mStruct, err = c.GetModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	s := scope.New(mStruct)
	s.StoreSet(internal.ControllerStoreKey, (*controller.Controller)(c))

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, errors.New(class.QueryValueUnaddressable, "unaddressable query value provided")
	}

	s.Value = model

	switch t.Elem().Kind() {
	case reflect.Struct:
	case reflect.Array, reflect.Slice:
		s.SetIsMany(true)
	default:
		return nil, errors.New(class.QueryValueType, "invalid query value type")
	}

	return (*Scope)(s), nil
}

func newScopeWithModel(c *controller.Controller, m *models.ModelStruct, isMany bool) *scope.Scope {
	s := scope.NewRootScope(m)

	s.StoreSet(internal.ControllerStoreKey, (*controller.Controller)(c))

	if isMany {
		s.Value = m.NewValueMany()
		s.SetIsMany(true)
	} else {
		s.Value = m.NewValueSingle()
	}

	return s
}

func queryS(s *scope.Scope) *Scope {
	return (*Scope)(s)
}
