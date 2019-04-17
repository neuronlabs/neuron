package scope

import (
	"context"
	stdErrors "errors"
	"fmt"
	"github.com/google/uuid"
	ctrl "github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	iFilters "github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/internal/query/paginations"
	"github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/filters"
	"github.com/kucjac/jsonapi/query/pagination"
	"github.com/kucjac/jsonapi/query/sorts"
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

// MustWithC creates the scope's model for the provided controller
func MustWithC(c *ctrl.Controller, model interface{}) *Scope {
	s, err := newScope((*controller.Controller)(c), model)
	if err != nil {
		panic(err)
	}
	return s
}

// NewWithC creates the scope for the provided model with respect to the provided
// controller 'c'
func NewWithC(c *ctrl.Controller, model interface{}) (*Scope, error) {
	return newScope((*controller.Controller)(c), model)
}

// NewWithModelC creates new scope on the base of the provided model struct with the new single
// value
func NewWithModelC(c *ctrl.Controller, mStruct *mapping.ModelStruct, isMany bool) *Scope {
	ctx := context.WithValue(
		context.Background(),
		internal.ControllerIDCtxKey,
		(*ctrl.Controller)(c),
	)

	s := scope.NewWithCtx(ctx, (*models.ModelStruct)(mStruct))
	if isMany {
		s.NewValueMany()
	} else {
		s.NewValueSingle()
	}

	return (*Scope)(s)
}

// New creates the scope on the base of the given model
func New(model interface{}) (*Scope, error) {
	return newScope(controller.Default(), model)
}

func newScope(c *controller.Controller, model interface{}) (*Scope, error) {
	mStruct, err := c.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(context.Background(), internal.ControllerIDCtxKey, (*ctrl.Controller)(c))
	s := scope.NewWithCtx(ctx, mStruct)

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, internal.IErrInvalidType
	}

	s.Value = model

	switch t.Elem().Kind() {
	case reflect.Struct:
	case reflect.Array, reflect.Slice:
		s.SetIsMany(true)
	default:
		return nil, internal.IErrInvalidType
	}

	return (*Scope)(s), nil
}

// AddFilter adds the given scope's filter field
func (s *Scope) AddFilter(filter *filters.FilterField) error {
	return scope.AddFilterField((*scope.Scope)(s), (*iFilters.FilterField)(filter))
}

// AddStringFilter parses the filter into the FilterField
// and adds to the provided scope's filters
func (s *Scope) AddStringFilter(rawFilter string, values ...interface{}) error {
	_, err := (*controller.Controller)(s.Controller()).QueryBuilder().BuildRawFilter((*scope.Scope)(s), rawFilter, values...)
	if err != nil {
		log.Debugf("BuildRawFilter: '%s' with values: %v failed. %v", rawFilter, values, err)
		return err
	}
	return nil
}

// AddToFieldset adds the fields to the scope's fieldset.
// The fields may be a mapping.StructField as well as the string - which might be
// the 'api name' or structFields name.
func (s *Scope) AddToFieldset(fields ...interface{}) error {
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
		return errs[0]
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

// Controller getsthe scope's predefined controller
func (s *Scope) Controller() *ctrl.Controller {
	c := s.Context().Value(internal.ControllerIDCtxKey).(*ctrl.Controller)
	return c
}

// Context return's scope's context
func (s *Scope) Context() context.Context {
	return (*scope.Scope)(s).Context()
}

// Create creates the scope values
func (s *Scope) Create() error {
	if err := DefaultQueryProcessor.doCreate(s); err != nil {
		return err
	}
	return nil
}

// Delete deletes the values provided in the scope's value
func (s *Scope) Delete() error {
	if err := DefaultQueryProcessor.doDelete(s); err != nil {
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

// InFieldset checks if the provided field is in the fieldset
func (s *Scope) InFieldset(field string) (*mapping.StructField, bool) {
	f, ok := (*scope.Scope)(s).InFieldset(field)
	if ok {
		return (*mapping.StructField)(f), true
	}
	return nil, false
}

// ForeignFilters returns scope's foreign key iFilters
func (s *Scope) ForeignFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersForeigns((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// FilterKeyFilters returns scope's primary iFilters
func (s *Scope) FilterKeyFilters() []*filters.FilterField {
	var res []*filters.FilterField
	for _, filter := range scope.FiltersKeys((*scope.Scope)(s)) {
		res = append(res, (*filters.FilterField)(filter))
	}
	return res
}

// Get gets single value from the repository taking into account the scope
// filters and parameters
func (s *Scope) Get() error {
	if err := DefaultQueryProcessor.doGet(s); err != nil {
		return err
	}
	return nil
}

// ID returns the scope's identity number represented as
// github.com/google/uuid
func (s *Scope) ID() uuid.UUID {
	return s.Context().Value(internal.ScopeIDCtxKey).(uuid.UUID)
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

// LanguageFilter returns language filter for given scope
func (s *Scope) LanguageFilter() *filters.FilterField {
	return (*filters.FilterField)((*scope.Scope)(s).LanguageFilter())
}

// List gets the values from the repository taking into account the scope
// filters and parameters
func (s *Scope) List() error {
	if err := DefaultQueryProcessor.doList(s); err != nil {
		return err
	}

	return nil
}

// Patch updates the scope's attribute and relationship values with the restrictions provided
// in the scope's parameters
func (s *Scope) Patch() error {
	if err := DefaultQueryProcessor.doPatch(s); err != nil {
		return err
	}
	return nil
}

// FormatQuery formats the scope's query into the url.Values
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
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

// NotSelectedFields returns all the fields that are not selected
func (s *Scope) NotSelectedFields(withForeigns ...bool) (notSelected []*mapping.StructField) {
	for _, field := range (*scope.Scope)(s).NotSelectedFields(withForeigns...) {
		notSelected = append(notSelected, (*mapping.StructField)(field))
	}
	return
}

// SetContext sets the context for the provided scope
// Logic is the same as in the WithContext
func (s *Scope) SetContext(ctx context.Context) {
	s.WithContext(ctx)
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

// WithContext sets the context for given scope
func (s *Scope) WithContext(ctx context.Context) {
	(*scope.Scope)(s).WithContext(ctx)
	return
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
