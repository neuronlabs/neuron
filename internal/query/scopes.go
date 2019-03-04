package query

import (
	"context"
	"fmt"
	aerrors "github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query/paginations"
	"github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/log"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// BuildScopeMany builds the scope for many type
func (b *Builder) BuildScopeMany(
	ctx context.Context,
	model interface{},
	query *url.URL,
	flgs ...*flags.Container,
) (s *scope.Scope, errs []*aerrors.ApiError, err error) {
	var mStruct *models.ModelStruct
	mStruct, err = b.modelStruct(model)
	if err != nil {
		err = errors.Wrapf(err, "BuildScopeMany failed for model: %T", model)
		return
	}

	schema, ok := b.schemas.Schema(mStruct.SchemaName())
	if !ok {
		err = errors.Wrap(err, "BuildScopeMany get model's schema failed.")
		return
	}

	// overloadPreventer - is a warden upon invalid query parameters
	var (
		errObj       *aerrors.ApiError
		errorObjects []*aerrors.ApiError
		addErrors    = func(errObjects ...*aerrors.ApiError) {
			errs = append(errs, errObjects...)
			s.IncreaseErrorCount(len(errObjects))
		}
	)

	s = scope.NewRootScopeWithCtx(ctx, mStruct)
	s.NewValueMany()

	// Initialize Includes
	s.InitializeIncluded(b.schemas.NestedIncludeLimit)

	// Set Flags
	s.SetFlagsFrom(flgs...)

	q := query.Query()

	// Get Includes
	included, ok := q[internal.QueryParamInclude]
	if ok {
		// build included scopes
		includedFields := strings.Split(included[0], internal.AnnotationSeperator)
		errorObjects = s.BuildIncludeList(includedFields...)
		addErrors(errorObjects...)
		if len(errs) > 0 {
			return
		}
	}

	languages, ok := q[internal.QueryParamLanguage]
	if ok {
		if b.I18n == nil {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.WithDetail("I18n is not acceptable by the server.")
			addErrors(errObj)
			return
		}

		errorObjects, err = b.setIncludedLangaugeFilters(s, languages[0])
		if err != nil {
			return
		}
		if len(errorObjects) > 0 {
			addErrors(errorObjects...)
			return
		}
	}

	for key, value := range q {
		if len(value) > 1 {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' set more than once.", key)
			addErrors(errObj)
			continue
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
		}

		switch {
		case key == internal.QueryParamInclude, key == internal.QueryParamLanguage:
			continue
		case key == internal.QueryParamPageLimit:
			errObj = s.PreparePaginatedValue(key, value[0], paginations.ParamLimit)
			if errObj != nil {
				addErrors(errObj)
				break
			}
		case key == internal.QueryParamPageOffset:
			errObj = s.PreparePaginatedValue(key, value[0], paginations.ParamOffset)
			if errObj != nil {
				addErrors(errObj)
			}
		case key == internal.QueryParamPageNumber:
			errObj = s.PreparePaginatedValue(key, value[0], paginations.ParamNumber)
			if errObj != nil {
				addErrors(errObj)
			}
		case key == internal.QueryParamPageSize:
			errObj = s.PreparePaginatedValue(key, value[0], paginations.ParamSize)
			if errObj != nil {
				addErrors(errObj)
			}
		case strings.HasPrefix(key, internal.QueryParamFilter):
			_, er := b.buildFilterField(s, key, value[0])
			if errObj, ok = er.(*aerrors.ApiError); ok {
				addErrors(errObj)
			} else {
				err = er
			}
		case key == internal.QueryParamSort:
			splitted := strings.Split(value[0], internal.AnnotationSeperator)
			errorObjects = s.BuildSortFields(splitted...)
			addErrors(errorObjects...)

		case strings.HasPrefix(key, internal.QueryParamFields):
			// fields[collection]
			var splitted []string
			var er error
			splitted, er = internal.SplitBracketParameter(key[len(internal.QueryParamFields):])
			if er != nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", er)
				addErrors(errObj)
				continue
			}

			if len(splitted) != 1 {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
				addErrors(errObj)
				continue
			}

			collection := splitted[0]

			fieldModel := schema.ModelByCollection(collection)
			if fieldModel == nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' for the fields query.", collection)
				addErrors(errObj)
				continue
			}

			fieldsScope := s.GetModelsRootScope(fieldModel)
			if fieldsScope == nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter collection: '%s' is not included in the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(value[0], internal.AnnotationSeperator)
			errorObjects = fieldsScope.BuildFieldset(splitValues...)
			addErrors(errorObjects...)
		case key == internal.QueryParamPageTotal:
			s.Flags().Set(flags.AddMetaCountList, true)
		case key == internal.QueryParamLinks:
			var er error
			var links bool
			links, er = strconv.ParseBool(value[0])
			if er != nil {
				addErrors(aerrors.ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			}
			s.Flags().Set(flags.UseLinks, links)
		default:
			errObj = aerrors.ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
		}

		if s.CurrentErrorCount() >= b.Config.ErrorLimits {
			return
		}
	}
	if paginate := s.Pagination(); paginate != nil {
		er := paginations.CheckPagination(paginate)
		if er != nil {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Pagination parameters are not valid. %s", er)
			addErrors(errObj)
		}
	}

	s.CopyIncludedBoundaries()

	return
}

// BuildScopeSingle builds the scope for the single query with id
func (b *Builder) BuildScopeSingle(
	ctx context.Context,
	model interface{},
	query *url.URL,
	id interface{},
	flgs ...*flags.Container,
) (s *scope.Scope, errs []*aerrors.ApiError, err error) {
	mStruct, err := b.modelStruct(model)
	if err != nil {
		err = errors.Wrap(err, "BuildScopeSingle modelstruct failed.")
		return
	}

	q := query.Query()

	s = scope.NewRootScopeWithCtx(ctx, mStruct)
	s.NewValueSingle()
	s.InitializeIncluded(b.Config.IncludeNestedLimit)

	if id == nil {
		var stringId string
		stringId, err = getID(query, s.Struct())
		if err != nil {
			return
		}
		_, er := b.buildFilterField(s, "filter["+s.Struct().Collection()+"][id][$eq]", stringId)
		if er != nil {
			if errObj, ok := er.(*aerrors.ApiError); ok {
				errs = append(errs, errObj)
			} else {
				err = er
			}
			return
		}
	} else {
		s.SetPrimaryFilters(id)
	}

	errs, err = b.buildQueryParametersSingle(s, q)
	return
}

/**

Related Value

*/
//BuildScopeRelated builds the Related scope
func (b *Builder) BuildScopeRelated(
	ctx context.Context,
	model interface{},
	query *url.URL,
	flgs ...*flags.Container,
) (s *scope.Scope, errs []*aerrors.ApiError, err error) {
	var mStruct *models.ModelStruct
	mStruct, err = b.modelStruct(model)
	if err != nil {
		err = errors.Wrapf(err, "BuildScopeRelated model: '%T' q:%v", model, query.RawQuery)
		return
	}

	// schema, ok := b.schemas.Schema(mStruct.SchemaName())
	// if !ok {
	// 	err = errors.Wrapf(internal.IErrModelNotMapped, "Model: %T, q: %v", model, query.RawQuery)
	// 	return
	// }

	id, related, err := getIDAndRelated(query, mStruct)
	if err != nil {
		return
	}

	s = scope.NewWithCtx(ctx, mStruct)

	s.SetKind(scope.RootKind)

	s.NewValueSingle()

	scopeVal := reflect.ValueOf(s.Value).Elem()
	prim := scopeVal.FieldByIndex(mStruct.PrimaryField().ReflectField().Index)

	if er := setPrimaryField(id, prim); er != nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided 'id' is of invalid type.")
		errs = append(errs, errObj)
		return
	}

	s.SetCollectionScope(s)

	relationField, ok := mStruct.RelationshipField(related)
	if !ok {
		// invalid query parameter
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid related field name: '%s', for the collection: '%s'", related, mStruct.Collection())
		errs = append(errs, errObj)
		return
	}

	_, er := b.buildFilterField(s, "filter["+mStruct.Collection()+"][id][$eq]", id)
	if er != nil {
		if errObj, ok := er.(*aerrors.ApiError); ok {
			errs = append(errs, errObj)
		} else {
			err = er
		}
		return
	}

	s.SetFieldsetNoCheck(relationField)

	if relationField.Relationship().Kind() == models.RelBelongsTo {
		fk := relationField.Relationship().ForeignKey()
		if fk != nil {
			s.SetFieldsetNoCheck(fk)
		}
	}

	s.InitializeIncluded(b.Config.IncludeNestedLimit)

	// preset related scope
	includedField := s.GetOrCreateIncludeField(relationField)
	includedField.Scope.SetKind(scope.RelatedKind)

	q := query.Query()
	languages, ok := q[internal.QueryParamLanguage]
	if ok {
		if b.I18n == nil {
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("I18n is not supported.")
			errs = append(errs, errObj)
			return
		}

		errs, err = b.setIncludedLangaugeFilters(includedField.Scope, languages[0])
		if err != nil || len(errs) > 0 {
			return
		}
	}

	qLinks, ok := q[internal.QueryParamLinks]
	if ok {
		var er error
		var links bool
		links, er = strconv.ParseBool(qLinks[0])
		if er != nil {
			errs = append(errs, aerrors.ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			return
		}
		s.Flags().Set(flags.UseLinks, links)
	}

	s.CopyIncludedBoundaries()

	return
}

//BuildScopeRelated builds the Related scope
func (b *Builder) BuildScopeRelationship(
	ctx context.Context,
	model interface{},
	query *url.URL,
	flgs ...*flags.Container,
) (s *scope.Scope, errs []*aerrors.ApiError, err error) {
	var mStruct *models.ModelStruct
	mStruct, err = b.modelStruct(model)
	if err != nil {
		err = errors.Wrapf(err, "BuildScopeRelated model: '%T' q:%v", model, query.RawQuery)
		return
	}

	// schema, ok := b.schemas.Schema(mStruct.SchemaName())
	// if !ok {
	// 	err = errors.Wrapf(internal.IErrModelNotMapped, "Model: %T, q: %v", model, query.RawQuery)
	// 	return
	// }

	id, related, err := getIDAndRelationship(query, mStruct)
	if err != nil {
		return
	}

	s = scope.NewWithCtx(ctx, mStruct)

	s.SetKind(scope.RootKind)

	s.NewValueSingle()

	scopeVal := reflect.ValueOf(s.Value).Elem()
	prim := scopeVal.FieldByIndex(mStruct.PrimaryField().ReflectField().Index)

	if er := setPrimaryField(id, prim); er != nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided 'id' is of invalid type.")
		errs = append(errs, errObj)
		return
	}

	s.SetCollectionScope(s)

	_, er := b.buildFilterField(s, "filter["+mStruct.Collection()+"][id][$eq]", id)
	if er != nil {
		if errObj, ok := er.(*aerrors.ApiError); ok {
			errs = append(errs, errObj)
		} else {
			err = er
		}
		return
	}

	relationField, ok := mStruct.RelationshipField(related)
	if !ok {
		// invalid query parameter
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid related field name: '%s', for the collection: '%s'", related, mStruct.Collection())
		errs = append(errs, errObj)
		return
	}

	s.SetFieldsetNoCheck(relationField)

	if relationField.Relationship().Kind() == models.RelBelongsTo {
		fk := relationField.Relationship().ForeignKey()
		if fk != nil {
			s.SetFieldsetNoCheck(fk)
		}
	}

	s.InitializeIncluded(b.Config.IncludeNestedLimit)

	relStruct := models.FieldsRelatedModelStruct(relationField)
	s.CreateModelsRootScope(relStruct)

	// preset related scope
	includedField := s.GetOrCreateIncludeField(relationField)
	includedField.Scope.SetKind(scope.RelationshipKind)

	s.CopyIncludedBoundaries()

	includedField.Scope.SetFields(relStruct.PrimaryField())

	return
}

// NewRootScope creates new root scope for the given model struct
func (b *Builder) NewRootScope(mStruct *models.ModelStruct) *scope.Scope {
	s := scope.NewRootScope(mStruct)

	return s
}

// NewScope creates new scope for provided model
func (b *Builder) NewScope(model interface{}) (*scope.Scope, error) {
	mStruct, err := b.modelStruct(model)
	if err != nil {
		return nil, errors.Wrap(err, "modelStruct failed.")
	}

	s := scope.New(mStruct)
	return s, nil
}

func (b *Builder) buildQueryParametersSingle(
	s *scope.Scope, q url.Values,
) (errs []*aerrors.ApiError, err error) {

	var (
		addErrors = func(errObjects ...*aerrors.ApiError) {
			errs = append(errs, errObjects...)
			s.IncreaseErrorCount(len(errObjects))
		}
		errObj       *aerrors.ApiError
		errorObjects []*aerrors.ApiError
	)
	// Check first included in order to create subscopes
	included, ok := q[internal.QueryParamInclude]
	if ok {
		// build included scopes
		if len(included) != 1 {
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintln("Duplicated 'included' query parameter.")
			addErrors(errObj)
			return
		}

		includedFields := strings.Split(included[0], internal.AnnotationSeperator)
		errorObjects = s.BuildIncludeList(includedFields...)
		addErrors(errorObjects...)
		if len(errs) > 0 {
			return
		}
	}

	schema, ok := b.schemas.Schema(s.Struct().SchemaName())
	if !ok {
		err = errors.Errorf("Model: '%T' schema is not in the registered.", s.Value)
		return
	}

	languages, ok := q[internal.QueryParamLanguage]
	if ok {
		if b.I18n == nil {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("I18n is not supported.")
			addErrors(errObj)
			return
		}

		errorObjects, err = b.setIncludedLangaugeFilters(s, languages[0])
		if err != nil {
			return
		}
		if len(errorObjects) > 0 {
			addErrors(errorObjects...)
			return
		}
	}

	for key, values := range q {
		if len(values) > 1 {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' set more than once.", key)
			addErrors(errObj)
			continue
		}

		switch {
		case key == internal.QueryParamInclude, key == internal.QueryParamLanguage:
			continue
		case strings.HasPrefix(key, internal.QueryParamFields):
			// fields[collection]
			var splitted []string
			var er error
			splitted, er = internal.SplitBracketParameter(key[len(internal.QueryParamFields):])
			if er != nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", er)
				addErrors(errObj)
				continue
			}
			if len(splitted) != 1 {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
				addErrors(errObj)
				continue
			}
			collection := splitted[0]

			fieldsetModel := schema.ModelByCollection(collection)
			if fieldsetModel == nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' for the fields query.", collection)
				addErrors(errObj)
				continue
			}

			fieldsetScope := s.GetModelsRootScope(fieldsetModel)
			if fieldsetScope == nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter collection: '%s' is not included in the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(values[0], internal.AnnotationSeperator)

			errorObjects = fieldsetScope.BuildFieldset(splitValues...)
			addErrors(errorObjects...)
		case strings.HasPrefix(key, internal.QueryParamFilter):
			_, er := b.buildFilterField(s, key, values[0])
			if errObj, ok = er.(*aerrors.ApiError); ok {
				addErrors(errObj)
			} else {
				err = er
			}
		case key == internal.QueryParamLinks:
			var er error
			var links bool
			links, er = strconv.ParseBool(values[0])
			if er != nil {
				addErrors(aerrors.ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			}
			s.Flags().Set(flags.UseLinks, links)
		default:
			errObj = aerrors.ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
		}

		if s.CurrentErrorCount() >= b.Config.ErrorLimits {
			return
		}
	}
	s.CopyIncludedBoundaries()
	return
}

func (b *Builder) setIncludedLangaugeFilters(
	s *scope.Scope,
	languages string,
) ([]*aerrors.ApiError, error) {
	tags, errObjects := buildLanguageTags(languages)
	if len(errObjects) > 0 {
		return errObjects, nil
	}

	tag, _, confidence := b.I18n.Matcher.Match(tags...)
	if confidence <= language.Low {
		// language not supported
		errObj := aerrors.ErrLanguageNotAcceptable.Copy()
		errObj.Detail = fmt.Sprintf("Provided languages: '%s' are not supported. This document supports following languages: %s",
			languages,
			strings.Join(b.I18n.PrettyLanguages(), ","),
		)
		return []*aerrors.ApiError{errObj}, nil
	}

	s.SetQueryLanguage(tag)

	var setLanguage func(s *scope.Scope) error
	//setLanguage sets language filter field for all included fields
	setLanguage = func(s *scope.Scope) error {
		defer func() {
			s.ResetIncludedField()
		}()
		if s.UseI18n() {
			base, _ := tag.Base()
			s.SetLanguageFilter(base.String())
		}

		for s.NextIncludedField() {
			field, err := s.CurrentIncludedField()
			if err != nil {
				return err
			}

			if err = setLanguage(field.Scope); err != nil {
				return err
			}
		}
		return nil
	}

	if err := setLanguage(s); err != nil {
		return errObjects, err
	}
	return errObjects, nil
}

// modelStruct returns ModelStruct for provided model's interface
func (b *Builder) modelStruct(model interface{}) (mStruct *models.ModelStruct, err error) {
	mStruct, ok := model.(*models.ModelStruct)
	if !ok {
		var t reflect.Type
		switch mt := model.(type) {
		case reflect.Type:
			t = mt
		default:
			t = reflect.TypeOf(model)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
		}

		if t.Kind() != reflect.Struct {
			err = internal.IErrUnexpectedType
			return
		}

		mStruct, err = b.schemas.ModelByType(t)
		if err != nil {
			return
		}
	}

	if mStruct == nil {
		err = internal.IErrModelNotMapped
	}
	return
}

func buildLanguageTags(languages string) ([]language.Tag, []*aerrors.ApiError) {
	var (
		tags []language.Tag
		errs []*aerrors.ApiError
	)

	splitted := strings.SplitN(languages, internal.AnnotationSeperator, 5)
	for _, lang := range splitted {
		tag, err := language.Parse(lang)
		if err != nil {
			switch v := err.(type) {
			case language.ValueError:
				errObj := aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided language: '%s' contains unknown value: '%s'.", lang, v.Subtag())
				errs = append(errs, errObj)
			default:
				// syntax error
				errObj := aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided language: '%s' is not syntetically correct.", lang)
				errs = append(errs, errObj)
			}
			continue
		}
		if len(errs) == 0 {
			tags = append(tags, tag)
		}
	}
	return tags, errs
}

func getURLVariables(query *url.URL, mStruct *models.ModelStruct, indexFirst, indexSecond int,
) (valueFirst, valueSecond string, err error) {

	path := query.Path
	var invalidURL = func() error {
		return fmt.Errorf("Provided url is invalid for getting url variables: '%s' with indexes: '%d'/ '%d'", path, indexFirst, indexSecond)
	}

	pathSplitted := strings.Split(path, "/")
	if indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	var collectionIndex int = -1

	if cIndex := models.StructCollectionUrlIndex(mStruct); cIndex != -1 {
		collectionIndex = cIndex
	} else {
		for i, splitted := range pathSplitted {
			if splitted == mStruct.Collection() {
				collectionIndex = i
				break
			}
		}
		if collectionIndex == -1 {
			err = fmt.Errorf("The url for given request does not contain collection name: %s", mStruct.Collection())
			return
		}
	}

	if collectionIndex+indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	valueFirst = pathSplitted[collectionIndex+indexFirst]

	if indexSecond > 0 {
		if collectionIndex+indexSecond > len(pathSplitted)-1 {
			err = invalidURL()
			return
		}
		valueSecond = pathSplitted[collectionIndex+indexSecond]
	}
	return
}

// GetAndSetID gets the id from the provided request and sets the primary field value to the id
// If the scope's value is not zero the function returns an error
func GetAndSetID(
	req *http.Request,
	s *scope.Scope,
) (prim interface{}, err error) {
	id, err := getID(req.URL, s.Struct())
	if err != nil {
		return nil, err
	}

	if s.Value == nil {
		return nil, errors.New("Nil value provided")
	}

	val := reflect.ValueOf(s.Value)
	if val.IsNil() {
		return nil, errors.New("Nil value in scope's value")
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	} else {
		return nil, internal.IErrInvalidType
	}

	newPrim := reflect.New(s.Struct().PrimaryField().ReflectField().Type).Elem()

	err = setPrimaryField(id, newPrim)
	if err != nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = "Provided invalid id value within the url."
		return nil, errObj
	}

	primVal := val.FieldByIndex(s.Struct().PrimaryField().ReflectField().Index)

	if reflect.DeepEqual(
		primVal.Interface(),
		reflect.Zero(s.Struct().PrimaryField().ReflectField().Type).Interface(),
	) {
		primVal.Set(newPrim)
	} else {
		log.Debugf("Checking the values")
		if newPrim.Interface() != primVal.Interface() {
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = "Provided invalid id value within the url. The id value doesn't match the primary field within the root object."
			return nil, errObj
		}
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, internal.IErrInvalidType
	}
	return primVal.Interface(), nil
}

// GetID from the provided query
func GetID(query *url.URL, mStruct *models.ModelStruct) (id string, err error) {
	return getID(query, mStruct)
}

func getID(query *url.URL, mStruct *models.ModelStruct) (id string, err error) {
	id, _, err = getURLVariables(query, mStruct, 1, -1)
	return
}

func getIDAndRelationship(query *url.URL, mStruct *models.ModelStruct,
) (id, relationship string, err error) {
	return getURLVariables(query, mStruct, 1, 3)

}

func getIDAndRelated(query *url.URL, mStruct *models.ModelStruct,
) (id, related string, err error) {
	return getURLVariables(query, mStruct, 1, 2)
}
