package scope

import (
	"context"
	stdErrors "errors"
	"fmt"
	"github.com/google/uuid"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	iFilters "github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/paginations"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/filters"
	"github.com/neuronlabs/neuron/query/pagination"
	"github.com/neuronlabs/neuron/query/sorts"
	"github.com/neuronlabs/neuron/query/tx"
	"gopkg.in/go-playground/validator.v9"
	"net/url"
	"reflect"
	"strings"
)

var (
	// ErrFieldNotFound is an error thrown when the provided Field is not found wihtin the scope
	ErrFieldNotFound error = stdErrors.New("Field not found")

	// ErrModelNotIncluded is an error that is thrown when the provided model is not included into given scope
	ErrModelNotIncluded error = stdErrors.New("Model were not included into scope")
)

// Scope is the Queries heart and soul which keeps all possible information
// Within it's structure
type Scope scope.Scope

// MustC creates the scope's model for the provided controller
func MustC(c *ctrl.Controller, model interface{}) *Scope {
	s, err := newScope((*controller.Controller)(c), model)
	if err != nil {
		panic(err)
	}
	return s
}

// NewC creates the scope for the provided model with respect to the provided
// controller 'c'
func NewC(c *ctrl.Controller, model interface{}) (*Scope, error) {
	return newScope((*controller.Controller)(c), model)
}

// NewModelC creates new scope on the base of the provided model struct with the new single
// value
func NewModelC(c *ctrl.Controller, mStruct *mapping.ModelStruct, isMany bool) *Scope {

	return (*Scope)(newScopeWithModel(c, (*models.ModelStruct)(mStruct), isMany))
}

// New creates the scope on the base of the given model
func New(model interface{}) (*Scope, error) {
	return newScope(controller.Default(), model)
}

// AddTxChain adds the 'sub' Scope to the 's' transaction chain
// Use on scope not created with 's'.New()
func (s *Scope) AddTxChain(sub *Scope) {
	(*scope.Scope)(s).AddChainSubscope((*scope.Scope)(sub))
}

// AddFilter adds the given scope's filter field
func (s *Scope) AddFilter(filter *filters.FilterField) error {
	return scope.AddFilterField((*scope.Scope)(s), (*iFilters.FilterField)(filter))
}

// AddStringFilter parses the filter into the FilterField
// and adds to the provided scope's filters
func (s *Scope) AddStringFilter(rawFilter string, values ...interface{}) error {
	filter, err := filters.NewStringFilter(s.Controller(), rawFilter, s.Struct().SchemaName(), values...)
	if err != nil {
		log.Debugf("BuildRawFilter: '%s' with values: %v failed. %v", rawFilter, values, err)
		return err
	}
	return s.AddFilter(filter)

}

// AddToFieldset adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as the string - which might be
// the 'api name' or structFields name.
func (s *Scope) AddToFieldset(fields ...interface{}) error {
	for i, field := range fields {
		mField, ok := field.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(mField)
		}
	}
	return (*scope.Scope)(s).AddToFieldset(fields...)
}

// AddToSelectedFields adds provided fields into the scope's selected fields
// This would affect the Create or Patch processes where the SelectedFields are taken
// as the unmarshaled fields.
func (s *Scope) AddToSelectedFields(fields ...interface{}) error {
	for i, f := range fields {
		// cast all *mapping.StructFields into models.StructField
		field, ok := f.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(field)
		}
	}

	return (*scope.Scope)(s).AddToSelectedFields(fields...)
}

// AddStringSortFields adds the sort fields in a string form
// i.e. [-field_1, field_2] -> Descending Field1 and Ascending Field2
func (s *Scope) AddStringSortFields(fields ...string) error {
	errs := (*scope.Scope)(s).BuildSortFields(fields...)
	if len(errs) > 0 {
		return errors.MultipleErrors(errs)
	}
	return nil
}

// AttributeFilters returns scope's attribute iFilters
func (s *Scope) AttributeFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersAttributes((*scope.Scope)(s)) {

		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// Begin begins the transaction for the provided scope
func (s *Scope) Begin() error {
	return s.begin(context.Background(), nil, true)
}

// BeginTx begins the transaction with the given tx.Options
func (s *Scope) BeginTx(ctx context.Context, opts *tx.Options) error {
	return s.begin(ctx, opts, true)
}

// Commit commits the given transaction for the scope
func (s *Scope) Commit() error {
	return s.commit(context.Background())
}

// CommitContext commits the given transaction for the scope with the context
func (s *Scope) CommitContext(ctx context.Context) error {
	return s.commit(ctx)
}

// Controller getsthe scope's predefined controller
func (s *Scope) Controller() *ctrl.Controller {
	c := s.Store[internal.ControllerCtxKey].(*ctrl.Controller)
	return c
}

// Create creates the scope values
func (s *Scope) Create() error {
	return s.createContext(context.Background())
}

// CreateContext creates the scope values with the context.Context
func (s *Scope) CreateContext(ctx context.Context) error {
	return s.createContext(ctx)
}

// Delete deletes the values provided in the scope's value
func (s *Scope) Delete() error {
	if err := DefaultQueryProcessor.DoDelete(context.Background(), s); err != nil {
		return err
	}
	return nil
}

// DeleteContext deletes the values provided in the scope's value with the context
func (s *Scope) DeleteContext(ctx context.Context) error {
	if err := DefaultQueryProcessor.DoDelete(ctx, s); err != nil {
		return err
	}
	return nil
}

// Fieldset returns the fields in the scope's Fieldset
func (s *Scope) Fieldset() (fs []*mapping.StructField) {
	for _, f := range (*scope.Scope)(s).Fieldset() {
		fs = append(fs, (*mapping.StructField)(f))
	}

	return fs
}

// FilterKeyFilters returns scope's primary iFilters
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersKeys((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// ForeignFilters returns scope's foreign key iFilters
func (s *Scope) ForeignFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersForeigns((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// FormatQuery formats the scope's query into the url.Values
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
}

// Get gets single value from the repository taking into account the scope
// filters and parameters
func (s *Scope) Get() error {
	if err := DefaultQueryProcessor.DoGet(context.Background(), s); err != nil {
		return err
	}
	return nil
}

// GetContext gets single value from repository taking into account the scope
// filters, parameters and the context.
func (s *Scope) GetContext(ctx context.Context) error {
	if err := DefaultQueryProcessor.DoGet(ctx, s); err != nil {
		return err
	}
	return nil
}

// ID returns the scope's identity number represented as
// github.com/google/uuid
func (s *Scope) ID() uuid.UUID {
	return (*scope.Scope)(s).ID()
}

// IncludeFields adds the included fields into the root scope
func (s *Scope) IncludeFields(fields ...string) error {
	iscope := (*scope.Scope)(s)
	iscope.InitializeIncluded((*controller.Controller)(s.Controller()).QueryBuilder().Config.IncludeNestedLimit)
	if errs := iscope.BuildIncludeList(fields...); len(errs) > 0 {
		return errors.MultipleErrors(errs)
	}
	return nil
}

// IncludedValue getst the scope's included values for provided model's
// the returning value would be pointer to slice of pointer to models
// i.e.: type Model struct {}, the result would be returned as a *[]*Model{}
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

	iscope := (*scope.Scope)(s)
	included, ok := iscope.IncludeScopeByStruct((*models.ModelStruct)(mStruct))
	if !ok {
		log.Info("Model: '%s' is not included into scope of: '%s'", mStruct.Collection(), s.Struct().Collection())
		return nil, ErrModelNotIncluded
	}

	return included.Value, nil
}

// InFieldset checks if the provided field is in the fieldset
func (s *Scope) InFieldset(field string) (*mapping.StructField, bool) {
	f, ok := (*scope.Scope)(s).InFieldset(field)
	if ok {
		return (*mapping.StructField)(f), true
	}
	return nil, false
}

// LanguageFilter returns language filter for given scope
func (s *Scope) LanguageFilter() *filters.FilterField {
	return (*filters.FilterField)((*scope.Scope)(s).LanguageFilter())
}

// List gets the values from the repository taking into account the scope
// filters and parameters
func (s *Scope) List() error {
	if err := DefaultQueryProcessor.DoList(context.Background(), s); err != nil {
		return err
	}

	return nil
}

// ListContext gets the values from the repository taking into account the scope
// filters and parameters
func (s *Scope) ListContext(ctx context.Context) error {
	if err := DefaultQueryProcessor.DoList(ctx, s); err != nil {
		return err
	}

	return nil
}

// New creates new scope for the provided model value.
// Created scope is a subscope for the scope. If the root scope 's' is on the transacation
// the new one will be created with the current transaction.
// It allows to commit or rollback a chain of scopes within a single method usage
func (s *Scope) New(value interface{}) (*Scope, error) {
	return s.newSubscope(context.Background(), value)
}

// NewContext creates new scope for the provided model value.
// Created scope is a subscope for the scope. If the root scope 's' is on the transacation
// the new one will be created with the current transaction.
// It allows to commit or rollback a chain of scopes within a single method usage
func (s *Scope) NewContext(ctx context.Context, value interface{}) (*Scope, error) {
	return s.newSubscope(ctx, value)
}

// NotSelectedFields returns all the fields that are not selected
func (s *Scope) NotSelectedFields(withForeigns ...bool) (notSelected []*mapping.StructField) {
	for _, field := range (*scope.Scope)(s).NotSelectedFields(withForeigns...) {
		notSelected = append(notSelected, (*mapping.StructField)(field))
	}
	return
}

// Patch updates the scope's attribute and relationship values with the restrictions provided
// in the scope's parameters
func (s *Scope) Patch() error {
	return s.patch(context.Background())
}

// PatchContext updates the scope's attribute and relationship values with the restrictions provided
// in the scope's parameters
func (s *Scope) PatchContext(ctx context.Context) error {
	return s.patch(ctx)
}

// Pagination returns the pagination for given scope
func (s *Scope) Pagination() *pagination.Pagination {
	return (*pagination.Pagination)((*scope.Scope)(s).Pagination())
}

// PrimaryFilters returns scope's primary iFilters
func (s *Scope) PrimaryFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersPrimary((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// RelationFilters returns scope's relation fields iFilters
func (s *Scope) RelationFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersRelationFields((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// Rollback rollsback the transaction for given scope
func (s *Scope) Rollback() error {
	return s.rollback(context.Background())
}

// RollbackContext rollsback the transaction for given scope
func (s *Scope) RollbackContext(ctx context.Context) error {
	return s.rollback(ctx)
}

// SetPagination sets the Pagination for the scope.
func (s *Scope) SetPagination(p *pagination.Pagination) error {
	return scope.SetPagination((*scope.Scope)(s), (*paginations.Pagination)(p))
}

// SelectField selects the field by the name.
// Selected fields are used in the patching process.
// By default the selected fields are all non zero valued fields in the struct.
func (s *Scope) SelectField(name string) error {
	field, ok := s.Struct().FieldByName(name)
	if !ok {
		log.Debug("Field not found: '%s'", name)
		return ErrFieldNotFound
	}

	(*scope.Scope)(s).AddSelectedField((*models.StructField)(field))

	return nil
}

// SelectedFields returns fields selected during
func (s *Scope) SelectedFields() (selected []*mapping.StructField) {
	for _, field := range (*scope.Scope)(s).SelectedFields() {
		selected = append(selected, (*mapping.StructField)(field))
	}
	return
}

// SetFieldset sets the fieldset for the provided scope
func (s *Scope) SetFieldset(fields ...interface{}) error {
	for i, field := range fields {
		mField, ok := field.(*mapping.StructField)
		if ok {
			fields[i] = (*models.StructField)(mField)
		}
	}
	return (*scope.Scope)(s).SetFields(fields...)
}

// SortBy adds the sort fields into given scope
func (s *Scope) SortBy(fields ...string) error {
	errs := (*scope.Scope)(s).BuildSortFields(fields...)
	if len(errs) > 0 {
		return errors.MultipleErrors(errs)
	}
	return nil
}

// SortFields returns the sorts used in the scope
func (s *Scope) SortFields() []*sorts.SortField {
	var sortFields []*sorts.SortField
	sfs := (*scope.Scope)(s).SortFields()
	for _, sf := range sfs {
		sortFields = append(sortFields, (*sorts.SortField)(sf))
	}
	return sortFields
}

// Struct returns scope's ModelStruct
func (s *Scope) Struct() *mapping.ModelStruct {
	mStruct := (*scope.Scope)(s).Struct()
	return (*mapping.ModelStruct)(mStruct)
}

// ValidateCreate validates the scope's value with respect to the 'create validator' and
// the 'create' validation tags
func (s *Scope) ValidateCreate() []*errors.ApiError {
	v := s.Controller().CreateValidator
	return s.validate(v, "CreateValidator")
}

// ValidatePatch validates the scope's value with respect to the 'Patch Validator'
func (s *Scope) ValidatePatch() []*errors.ApiError {
	v := s.Controller().PatchValidator
	return s.validate(v, "PatchValidator")
}

func (s *Scope) begin(ctx context.Context, opts *tx.Options, checkError bool) (err error) {

	// check if the context contains the transaction
	if v := s.Store[internal.TxStateCtxKey]; v != nil {
		txn := v.(*tx.Tx)
		if checkError {
			if txn.State != tx.Done {
				err = tx.ErrTxAlreadyBegan
				return
			}
		}
	}

	// for nil options set it to default value
	if opts == nil {
		opts = &tx.Options{}
	}

	txn := &tx.Tx{
		Options: *opts,
	}

	// generate the id
	// TODO: enterprise set the uuid based on the namespace of the gateway
	// so that the node name can be taken from the UUID v3 or v5 namespace
	txn.ID, err = uuid.NewRandom()
	if err != nil {
		return
	}

	txn.State = tx.Begin

	// set the transaction to the context
	s.Store[internal.TxStateCtxKey] = txn

	c := s.Controller()

	repo, ok := (*controller.Controller)(c).RepositoryByModel((*scope.Scope)(s).Struct())
	if !ok {
		log.Errorf("No repository found for the %s model.", s.Struct().Collection())
		err = ErrNoRepositoryFound
		return
	}

	transactioner, ok := repo.(Transactioner)
	if !ok {
		log.Errorf("The repository deosn't implement Creater interface for model: %s", (*scope.Scope)(s).Struct().Collection())
		err = ErrRepsitoryNotATransactioner
	}

	if err = transactioner.Begin(ctx, s); err != nil {
		return
	}

	return nil
}

func (s *Scope) commit(ctx context.Context) error {

	txV, ok := s.Store[internal.TxStateCtxKey].(*tx.Tx)
	if txV == nil || !ok {
		log.Debugf("COMMIT: No transaction found for the scope")
		return stdErrors.New("Nothing to commit")
	}

	if txV != nil && ok && txV.State != tx.Begin {
		log.Debugf("COMMIT: Transaction already resolved: %s", txV.State)
		return stdErrors.New("Transaction already resolved")
	}

	chain := (*scope.Scope)(s).Chain()

	if len(chain) > 0 {

		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Builder.RepositoryTimeout
		for _, sub := range chain {
			if sub.Struct().Config() == nil {
				continue
			}
			if conn := sub.Struct().Config().Connection; conn != nil {
				if tm := conn.MaxTimeout; tm != nil {
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
				if err, ok := v.(error); ok && err != ErrRepositoryNotACommiter && err != ErrTransactionAlreadyResolved {
					return err
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
	// if no fields were selected set automatically non zero
	if err := (*scope.Scope)(s).AutoSelectFields(); err != nil {
		return err
	}

	if err := DefaultQueryProcessor.DoCreate(ctx, s); err != nil {
		return err
	}
	return nil
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

	for _, rel := range s.RelationFilters() {
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

func (s *Scope) newSubscope(ctx context.Context, value interface{}) (*Scope, error) {
	sub, err := newScope((*controller.Controller)(s.Controller()), value)
	if err != nil {
		return nil, err
	}

	(*scope.Scope)(sub).SetKind(scope.SubscopeKind)

	if txn := s.tx(); txn != nil {

		if err := sub.begin(ctx, &txn.Options, false); err != nil {
			log.Debug("Begin the subscope failed.")
			return nil, err
		}

		(*scope.Scope)(s).AddChainSubscope((*scope.Scope)(sub))
	}

	return sub, nil
}

func (s *Scope) patch(ctx context.Context) error {
	// if no fields were selected set automatically non zero
	if err := (*scope.Scope)(s).AutoSelectFields(); err != nil {
		return err
	}

	if err := DefaultQueryProcessor.DoPatch(ctx, s); err != nil {
		return err
	}
	return nil
}

func (s *Scope) rollback(ctx context.Context) error {

	txV, ok := s.Store[internal.TxStateCtxKey].(*tx.Tx)
	if txV == nil || !ok {
		log.Debugf("ROLLBACK: No transaction found for the scope")
		return stdErrors.New("Nothing to commit")
	}

	if txV != nil && ok && txV.State != tx.Begin {
		log.Debugf("ROLLBACK: Transaction already resolved: %s", txV.State)
		return ErrTransactionAlreadyResolved
	}

	log.Debugf("s.Rollback: %s for model: %s rolling back", s.ID().String(), s.Struct().Collection())

	chain := (*scope.Scope)(s).Chain()

	if len(chain) > 0 {

		results := make(chan interface{}, len(chain))

		// get initial time out from the controller builder config
		maxTimeout := s.Controller().Config.Builder.RepositoryTimeout

		// check if any model has a preset timeout greater than the maxTimeout
		for _, sub := range chain {
			if sub.Struct().Config() == nil {
				continue
			}
			if conn := sub.Struct().Config().Connection; conn != nil {
				if tm := conn.MaxTimeout; tm != nil {
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

				if err, ok := v.(error); ok && err != ErrRepositoryNotARollbacker && err != ErrTransactionAlreadyResolved {
					return err
				}

				rescount++
				if rescount == len(chain) {
					break fl
				}

			}

		}
	}

	log.Debugf("Rolling back root.")
	single := make(chan interface{}, 1)

	go s.rollbackSingle(ctx, single)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case v := <-single:
		if err, ok := v.(error); ok && err != ErrTransactionAlreadyResolved {
			return err
		}
	}

	return nil
}

func (s *Scope) validate(v *validator.Validate, validatorName string) []*errors.ApiError {
	if err := v.Struct(s.Value); err != nil {
		switch er := err.(type) {
		case *validator.InvalidValidationError:
			// Invalid argument passed to validator
			log.Errorf("[%s] %s-> Invalid Validation Error: %v", s.ID().String(), validatorName, er)
			e := errors.ErrInternalError.Copy()
			e.Detail = fmt.Sprintf("%v", er)
			return []*errors.ApiError{e}
		case validator.ValidationErrors:
			var errs []*errors.ApiError
			for _, verr := range er {
				tag := verr.Tag()

				var errObj *errors.ApiError
				if tag == "required" {

					// if field is required and the field tag is empty
					if verr.Field() == "" {
						log.Errorf("[%s] Model: '%v'. '%s' failed. Field is required and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
						errObj = errors.ErrInternalError.Copy()
						errObj.Detail = fmt.Sprintf("[%s] Model: '%v'. '%s' failed. Field is 'required' and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
						return append(errs, errObj)
					}
					errObj = errors.ErrMissingRequiredJSONField.Copy()
					errObj.Detail = fmt.Sprintf("The field: %s, is required.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if tag == "isdefault" {

					if verr.Field() == "" {
						if verr.Field() == "" {
							log.Errorf("[%s] Model: '%v'. %s failed. Field is 'isdefault' and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
							errObj = errors.ErrInternalError.Copy()
							errObj.Detail = fmt.Sprintf("[%s] Model: '%v'. %s failed. Field is 'isdefault' and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
							return append(errs, errObj)
						}

						errObj = errors.ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = fmt.Sprintf("The field: '%s' must be of zero value.", verr.Field())
						errs = append(errs, errObj)
						continue

					} else if strings.HasPrefix(tag, "len") {
						// length
						if verr.Field() == "" {
							log.Errorf("[%s] Model: '%v'. %s failed. Field must have specific length and the field tag is empty.", s.ID().String(),
								validatorName, s.Struct().Type().String())
							errObj = errors.ErrInternalError.Copy()
							errObj.Detail = fmt.Sprintf("[%s] Model: '%v'. %s failed. Field must have specific len and the field tag is empty.", s.ID().String(), validatorName, s.Struct().Type().String())
							return append(errs, errObj)
						}
						errObj = errors.ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = fmt.Sprintf("The value of the field: %s is of invalid length.", verr.Field())
						errs = append(errs, errObj)
						continue
					} else {
						errObj = errors.ErrInvalidJSONFieldValue.Copy()
						if verr.Field() != "" {
							errObj.Detail = fmt.Sprintf("Invalid value for the field: '%s'.", verr.Field())
						}
						errs = append(errs, errObj)
						continue
					}
				}
			}
			return errs
		default:
			return []*errors.ApiError{errors.ErrInternalError.Copy()}
		}
	}
	return nil
}

func (s *Scope) tx() *tx.Tx {
	txn, _ := s.Store[internal.TxStateCtxKey].(*tx.Tx)
	return txn
}

func newScope(c *controller.Controller, model interface{}) (*Scope, error) {
	mStruct, err := c.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	s := scope.New(mStruct)
	s.Store[internal.ControllerCtxKey] = (*ctrl.Controller)(c)

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, internal.ErrInvalidType
	}

	s.Value = model

	switch t.Elem().Kind() {
	case reflect.Struct:
	case reflect.Array, reflect.Slice:
		s.SetIsMany(true)
	default:
		return nil, internal.ErrInvalidType
	}

	return (*Scope)(s), nil
}

func newScopeWithModel(c *ctrl.Controller, m *models.ModelStruct, isMany bool) *scope.Scope {
	s := scope.NewRootScope(m)

	s.Store[internal.ControllerCtxKey] = (*ctrl.Controller)(c)

	if isMany {
		s.Value = m.NewValueMany()
		s.SetIsMany(true)
	} else {
		s.Value = m.NewValueSingle()
	}

	return s
}

func copyStore(s *Scope, store map[interface{}]interface{}) {
	for k, v := range store {
		// TODO:
		s.Store[k] = v
	}
}
