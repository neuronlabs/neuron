package query

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/repository"

	"github.com/neuronlabs/neuron-core/internal"
)

// Scope is the query's structure that contains information required
// for the processor to operate.
// The scope has its unique 'ID', contains predefined model, operational value, fieldset, filters, sorts and pagination.
// It also contains the mapping of the included scopes.
type Scope struct {
	// id is the unique identification of the scope.
	id uuid.UUID
	// controller defines the controller for given scope
	c *controller.Controller
	// tx defines the transaction related to this scope.
	tx *Tx
	// Struct is a modelStruct this scope is based on
	mStruct *mapping.ModelStruct

	// Value is the values or / value of the queried object / objects.
	Value interface{}
	// Fieldset represents fields used specified to get / update for this query scope.
	Fieldset map[string]*mapping.StructField
	// PrimaryFilters contain filter for the primary field.
	PrimaryFilters Filters
	// RelationFilters contain relationship field filters.
	RelationFilters Filters
	// AttributeFilters contain filter for the attribute fields.
	AttributeFilters Filters
	// ForeignFilters contain the foreign key filter fields.
	ForeignFilters Filters
	// SortFields are the query sort fields.
	SortFields []*SortField
	// IncludedFields contain fields to include. If the included field is a relationship type, then
	// specific included field contains information about it
	IncludedFields []*IncludedField
	// Pagination is the query pagination.
	Pagination *Pagination
	// Processor is current query processor.
	Processor *Processor
	// Err defines the process error.
	Err error

	// store stores the scope's related key values
	store map[interface{}]interface{}
	// isMany defines if the scope is of
	isMany bool
	// autoSelectedFields is the flag that defines if the query had automatically selected fieldset.
	autoSelectedFields bool
	processMethod      processMethod
}

// MustNewC creates new scope with given 'model' for the given controller 'c'.
// Panics on error.
func MustNewC(c *controller.Controller, model interface{}) *Scope {
	s, err := newQueryScope(c, model)
	if err != nil {
		panic(err)
	}
	return s
}

// MustNew creates new scope with given 'model' for the default controller.
// Panics on error.
func MustNew(model interface{}) *Scope {
	s, err := newQueryScope(controller.Default(), model)
	if err != nil {
		panic(err)
	}
	return s
}

// NewC creates the scope for the provided model with respect to the provided internalController 'c'.
func NewC(c *controller.Controller, model interface{}) (*Scope, error) {
	return newQueryScope(c, model)
}

// NewModelC creates new scope on the base of the provided model struct with the value.
// The value might be a slice of instances if 'isMany' is true.
func NewModelC(c *controller.Controller, mStruct *mapping.ModelStruct, isMany bool) *Scope {
	return newScopeWithModel(c, mStruct, isMany)
}

// New creates the scope on the base of the given 'model' it uses default internalController.
func New(model interface{}) (*Scope, error) {
	return newQueryScope(controller.Default(), model)
}

// AutoSelectedFields checks if the scope fieldset was set automatically.
// This function returns false if a user had defined any field in the Fieldset.
func (s *Scope) AutoSelectedFields() bool {
	return s.autoSelectedFields
}

// Controller gets the scope's internalController.
func (s *Scope) Controller() *controller.Controller {
	return s.c
}

// Copy creates a copy of the given scope.
func (s *Scope) Copy() *Scope {
	return s.copy(nil)
}

// Count returns the number of the values for the provided query scope.
func (s *Scope) Count() (int64, error) {
	return s.count(context.Background())
}

// CountContext returns the number of the values for the provided query scope, with the provided 'ctx' context.
func (s *Scope) CountContext(ctx context.Context) (int64, error) {
	return s.count(ctx)
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
	if err := s.processor().Delete(context.Background(), s); err != nil {
		return err
	}
	return nil
}

// DeleteContext deletes the values provided in the scope's value with the context.
func (s *Scope) DeleteContext(ctx context.Context) error {
	if err := s.processor().Delete(ctx, s); err != nil {
		return err
	}
	return nil
}

// FormatQuery formats the scope's query into the url.Values.
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (s *Scope) Get() error {
	if err := s.processor().Get(context.Background(), s); err != nil {
		return err
	}
	return nil
}

// GetContext gets single value from repository taking into account the scope
// filters, parameters and the context.
func (s *Scope) GetContext(ctx context.Context) error {
	if err := s.processor().Get(ctx, s); err != nil {
		return err
	}
	return nil
}

// ID returns the scope's identity number stored as the UUID.
func (s *Scope) ID() uuid.UUID {
	return s.id
}

// IsMany checks if the scope's value is of slice type.
func (s *Scope) IsMany() bool {
	return s.isMany
}

// List gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values.
func (s *Scope) List() error {
	return s.list(context.Background())
}

// ListContext gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values. Provided context.Context 'ctx'
// would be used while querying the repositories.
func (s *Scope) ListContext(ctx context.Context) error {
	return s.list(ctx)
}

// Query creates new query with the provided 'model'.
func (s *Scope) Query(model interface{}) Builder {
	return s.query(context.Background(), s.c, model)
}

// QueryCtx creates new Query with the provided 'model' with the context.Context 'ctx'.
func (s *Scope) QueryCtx(ctx context.Context, model interface{}) Builder {
	return s.query(ctx, s.c, model)
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

// SetFields adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as field's NeuronName (string) or
// the StructField Name (string).
func (s *Scope) SetFields(fields ...interface{}) error {
	if len(fields) == 0 {
		log.Debug("Fields - provided no fields")
		return nil
	}
	return s.addToFieldset(fields...)
}

// SetFieldset sets the fieldset for the 'fields'.
// A field may be a field's name (string), NeuronName (string) or *mapping.StructField.
func (s *Scope) SetFieldset(fields ...interface{}) error {
	return s.setFields(fields...)
}

// Struct returns scope's model's structure - *mapping.ModelStruct.
func (s *Scope) Struct() *mapping.ModelStruct {
	return s.mStruct
}

// StoreGet gets the value from the scope's Store for given 'key'.
func (s *Scope) StoreGet(key interface{}) (value interface{}, ok bool) {
	value, ok = s.store[key]
	return value, ok
}

// StoreSet sets the 'key' and 'value' in the given scope's store.
func (s *Scope) StoreSet(key, value interface{}) {
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s][%s] Store Set key: '%v', value: '%v'", s.ID(), s.mStruct.Collection(), key, value)
	}
	s.store[key] = value
}

// String implements fmt.Stringer interface.
func (s *Scope) String() string {
	sb := &strings.Builder{}

	// Scope ID
	sb.WriteString("SCOPE[" + s.ID().String() + "][" + s.Struct().Collection() + "]")

	// Fieldset
	sb.WriteString(" Fieldset")
	if s.isDefaultFieldset() {
		sb.WriteString("(default)")
	}
	sb.WriteString(": [")
	var i int
	for field := range s.Fieldset {
		sb.WriteString(field)
		if i != len(s.Fieldset)-1 {
			sb.WriteRune(',')
		}
		i++
	}
	sb.WriteRune(']')

	// Primary Filters
	if len(s.PrimaryFilters) > 0 {
		sb.WriteString(" Primary Filters: ")
		sb.WriteString(s.PrimaryFilters.String())
	}

	// Attribute Filters
	if len(s.AttributeFilters) > 0 {
		sb.WriteString(" Attribute Filters: ")
		sb.WriteString(s.AttributeFilters.String())
	}

	// Relationship Filters
	if len(s.RelationFilters) > 0 {
		sb.WriteString(" Relationship Filters: ")
		sb.WriteString(s.RelationFilters.String())
	}

	// ForeignKey Filters
	if len(s.ForeignFilters) > 0 {
		sb.WriteString(" ForeignKey Filters: ")
		sb.WriteString(s.ForeignFilters.String())
	}

	if s.Pagination != nil {
		sb.WriteString(" Pagination: ")
		sb.WriteString(s.Pagination.String())
	}

	if len(s.SortFields) > 0 {
		sb.WriteString(" SortFields: ")
		for j, field := range s.SortFields {
			sb.WriteString(field.StructField.NeuronName())
			if j != len(s.SortFields)-1 {
				sb.WriteRune(',')
			}
		}
	}
	return sb.String()
}

/**
 *
 * Transactions
 *
 */

// Begin starts new transactions for given scope.
func (s *Scope) Begin() (*Tx, error) {
	tx := Begin(context.Background(), s.c, nil)
	if err := s.WithTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// BeginCtx starts new transactions for given scope with respect to the given context 'ctx'.
func (s *Scope) BeginCtx(ctx context.Context, opts *TxOptions) (*Tx, error) {
	tx := Begin(ctx, s.c, opts)
	if err := s.WithTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// Tx returns the transaction for the given scope if exists.
func (s *Scope) Tx() *Tx {
	return s.tx
}

// WithTransaction applies transaction 't' to the query scope. The transactions are divided by the repositories.
// It begins a sub-transaction for the scope's repository if it is not already started yet,
// Returns error if given scope is already connected with other transaction
func (s *Scope) WithTransaction(t *Tx) error {
	if s.tx != nil {
		if s.tx == t {
			log.Debugf("Scope: '%s' is already included in the transaction: '%s'", s.id, t.id)
			return nil
		}
		log.Debugf("Scope: '%s' is already included in another transaction: '%s'", s.id, s.tx.id)
		return errors.Newf(class.QueryTxAlreadyBegin, "scope '%s' already have a transaction: '%s'", s.id, s.tx.id)
	}

	// check if the model of given scope implements repository interface.
	transactioner, ok := s.repository().(Transactioner)
	if !ok {
		log.Debugf("Factory '%s' repository does not implement Transactioner interface", s.repository().FactoryName())
		return errors.Newf(class.RepositoryNotImplementsTransactioner, "repository for model: '%s' "+
			"doesn't allow transactions", s.mStruct)
	}
	log.Debug2f("Scope: '%s' is applied to transaction: '%s'", s.id, t.id)
	s.tx = t
	return t.beginUniqueTransaction(transactioner, s.mStruct)
}

/**

Private scope methods

*/

func (s *Scope) copy(from *Scope) *Scope {
	copiedScope := &Scope{
		id:      uuid.New(),
		mStruct: s.mStruct,
		store:   map[interface{}]interface{}{internal.ControllerStoreKey: s.Controller()},
		isMany:  s.isMany,
	}

	if s.isMany {
		copiedScope.Value = mapping.NewValueMany(s.mStruct)
	} else {
		copiedScope.Value = mapping.NewValueSingle(s.mStruct)
	}

	if s.Fieldset != nil {
		copiedScope.Fieldset = make(map[string]*mapping.StructField)
		for k, v := range s.Fieldset {
			copiedScope.Fieldset[k] = v
		}
	}

	if s.PrimaryFilters != nil {
		copiedScope.PrimaryFilters = make([]*FilterField, len(s.PrimaryFilters))
		for i, v := range s.PrimaryFilters {
			copiedScope.PrimaryFilters[i] = v.Copy()
		}
	}

	if s.AttributeFilters != nil {
		copiedScope.AttributeFilters = make([]*FilterField, len(s.AttributeFilters))
		for i, v := range s.AttributeFilters {
			copiedScope.AttributeFilters[i] = v.Copy()
		}
	}

	if s.RelationFilters != nil {
		for i, v := range s.RelationFilters {
			copiedScope.RelationFilters[i] = v.Copy()
		}
	}

	if s.SortFields != nil {
		copiedScope.SortFields = make([]*SortField, len(s.SortFields))
		for i, v := range s.SortFields {
			copiedScope.SortFields[i] = v.Copy()
		}

	}

	if s.IncludedFields != nil {
		copiedScope.IncludedFields = make([]*IncludedField, len(s.IncludedFields))
		for i, v := range s.IncludedFields {
			copiedScope.IncludedFields[i] = v.copy()
		}
	}

	return copiedScope
}

func (s *Scope) count(ctx context.Context) (int64, error) {
	if err := s.processor().Count(ctx, s); err != nil {
		return 0, err
	}
	i, ok := s.Value.(int64)
	if !ok {
		return 0, errors.NewDetf(class.InternalQueryInvalidValue, "scope count value is not 'int64' type, but: '%T'", s.Value)
	}
	return i, nil
}

func (s *Scope) createContext(ctx context.Context) error {
	if s.isMany {
		return errors.NewDet(class.QueryValueType, "creating with multiple values in are not supported yet")
	}

	// if no fields were selected set automatically all fields for the model.
	if err := s.autoSelectFields(); err != nil {
		return err
	}
	if err := s.processor().Create(ctx, s); err != nil {
		return err
	}
	return nil
}

func (s *Scope) processor() *Processor {
	// at first try the scope's processor
	p := s.Processor
	if p != nil {
		return p
	}
	// then try internalController's processor
	p = (*Processor)(s.Controller().Config.Processor)
	if p == nil {
		// if nil create new processor for given config
		s.Controller().Config.Processor = config.ThreadSafeProcessor()
		p = (*Processor)(s.Controller().Config.Processor)
	}
	return p
}

func (s *Scope) formatQuery() url.Values {
	q := url.Values{}
	s.formatQueryFilters(q)
	s.formatQuerySorts(q)
	s.formatQueryPagination(q)
	s.formatQueryFieldset(q)
	s.formatQueryIncludes(q)
	return q
}

func (s *Scope) formatQuerySorts(q url.Values) {
	for _, sort := range s.SortFields {
		sort.FormatQuery(q)
	}
}

func (s *Scope) formatQueryPagination(q url.Values) {
	if s.Pagination != nil {
		s.Pagination.FormatQuery(q)
	}
}

func (s *Scope) formatQueryFilters(q url.Values) {
	for _, prim := range s.PrimaryFilters {
		prim.FormatQuery(q)
	}
	for _, fk := range s.ForeignFilters {
		fk.FormatQuery(q)
	}
	for _, attr := range s.AttributeFilters {
		attr.FormatQuery(q)
	}
	for _, rel := range s.RelationFilters {
		rel.FormatQuery(q)
	}
}

func (s *Scope) formatQueryFieldset(q url.Values) {
	if s.Fieldset != nil {
		fieldsKey := fmt.Sprintf("%s[%s]", ParamFields, s.Struct().Collection())
		var values string
		var i int
		for _, field := range s.Fieldset {
			values += field.NeuronName()
			if i != len(s.Fieldset)-1 {
				values += ","
			}
			i++
		}
		q.Add(fieldsKey, values)
	}
}

func (s *Scope) formatQueryIncludes(q url.Values) {
	var includes []string
	for _, included := range s.IncludedFields {
		includes = append(includes, included.StructField.NeuronName())
		fieldsKey := fmt.Sprintf("%s[%s]", ParamFields, included.StructField.Relationship().Struct().Collection())
		var values string
		var i int
		for _, field := range included.Fieldset {
			values += field.NeuronName()
			if i != len(s.Fieldset)-1 {
				values += ","
			}
			i++
		}
		q.Add(fieldsKey, values)
	}
	if len(includes) > 0 {
		q.Add("include", strings.Join(includes, ","))
	}
}

func (s *Scope) list(ctx context.Context) error {
	if !s.isMany {
		err := errors.NewDet(class.QueryValueType, "provided non slice value for the list query")
		err.SetDetails("Single value provided for the List process")
		return err
	}
	if log.Level() <= log.LevelDebug2 {
		log.Debug2f("[SCOPE][%s] process list", s.ID())
	}
	// list from the processor
	if err := s.processor().List(ctx, s); err != nil {
		return err
	}
	return nil
}

func (s *Scope) patch(ctx context.Context) error {
	// check if scope's value is single
	if s.isMany {
		return errors.NewDet(class.QueryValueType, "patching multiple values are not supported yet")
	}

	// if no fields were selected set automatically non zero
	if err := s.autoSelectFields(); err != nil {
		return err
	}

	if err := s.processor().Patch(ctx, s); err != nil {
		return err
	}
	return nil
}

func (s *Scope) query(ctx context.Context, c *controller.Controller, model interface{}) Builder {
	if s.tx != nil {
		return s.tx.query(c, model)
	}
	return newQuery(ctx, c, model)
}

func (s *Scope) repository() repository.Repository {
	repo, err := s.Controller().GetRepository(s.mStruct)
	if err != nil {
		log.Panicf("Can't find repository for model: %s", s.mStruct.String())
	}
	return repo
}

func (s *Scope) convertValidateError(err error) error {
	switch er := err.(type) {
	case *validator.InvalidValidationError:
		// Invalid argument passed to validator
		log.Errorf("[%s] Invalid Validation Error: %v", s.ID().String(), er)
		e := errors.NewDet(class.InternalQueryValidation, "invalid validation error")
		return e
	case validator.ValidationErrors:
		var errs []errors.ClassError
		for _, valueError := range er {
			tag := valueError.Tag()

			var errObj errors.DetailedError
			if tag == "required" {
				// if field is required and the field tag is empty
				if valueError.Field() == "" {
					log.Errorf("[%s] Model: '%v' validation failed. Field is required and the field tag is empty.", s.ID().String(), s.Struct().Type().String())
					errObj = errors.NewDet(class.InternalQueryValidation, "empty field tag")
					return errObj
				}

				errObj = errors.NewDet(class.QueryValueMissingRequired, "missing required field")
				errObj.SetDetailsf("The field: %s, is required.", valueError.Field())
				errs = append(errs, errObj)
				continue
			} else if tag == "isdefault" {
				switch {
				case valueError.Field() == "":
					errObj = errors.NewDet(class.QueryValueValidation, "non default field value")
					errObj.SetDetailsf("The field: '%s' must be of zero value.", valueError.Field())
					errs = append(errs, errObj)
					continue
				case strings.HasPrefix(tag, "len"):
					errObj = errors.NewDet(class.QueryValueValidation, "validation failed - field of invalid length")
					errObj.SetDetailsf("The value of the field: %s is of invalid length.", valueError.Field())
					errs = append(errs, errObj)
					continue
				default:
					errObj = errors.NewDet(class.QueryValueValidation, "validation failed - invalid field value")
					if valueError.Field() != "" {
						errObj.SetDetailsf("Invalid value for the field: '%s'.", valueError.Field())
					}

					errs = append(errs, errObj)
					continue
				}
			}
		}
		if errs != nil {
			return nil
		}
		return errors.MultiError(errs)
	default:
		return errors.NewDetf(class.InternalQueryValidation, "invalid error type: '%T'", er)
	}
}

func newQueryScope(c *controller.Controller, model interface{}) (*Scope, error) {
	var (
		err          error
		mStruct      *mapping.ModelStruct
		noModelValue bool
	)

	switch mt := model.(type) {
	case *mapping.ModelStruct:
		mStruct = mt
		noModelValue = true
	default:
		mStruct, err = c.ModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	if noModelValue {
		return newScopeWithModel(c, mStruct, false), nil
	}

	s := newScope(c, mStruct)

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryValueUnaddressable, "unaddressable query value provided")
	}

	s.Value = model
	switch t.Elem().Kind() {
	case reflect.Struct:
	case reflect.Array, reflect.Slice:
		s.isMany = true
	default:
		return nil, errors.NewDet(class.QueryValueType, "invalid query value type")
	}
	return s, nil
}

func newScopeWithModel(c *controller.Controller, m *mapping.ModelStruct, isMany bool) *Scope {
	s := newScope(c, m)
	if isMany {
		s.Value = mapping.NewValueMany(m)
		s.isMany = isMany
	} else {
		s.Value = mapping.NewValueSingle(m)
	}
	return s
}

// initialize new scope with added primary field to fieldset
func newScope(c *controller.Controller, modelStruct *mapping.ModelStruct) *Scope {
	s := &Scope{
		id:       uuid.New(),
		c:        c,
		mStruct:  modelStruct,
		Fieldset: make(map[string]*mapping.StructField),
		store:    map[interface{}]interface{}{},
	}

	if log.Level() <= log.LevelDebug2 {
		log.Debug2f("[SCOPE][%s][%s] query new scope", s.id.String(), modelStruct.Collection())
	}
	return s
}
