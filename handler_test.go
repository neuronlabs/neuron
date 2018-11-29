package jsonapi

import (
	"github.com/kucjac/uni-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandlerMarshalScope(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)

	rw, req := getHttpPair("GET", "/blogs/1/current_post", nil)
	scope, err := h.Controller.NewScope(&BlogSDK{})
	assert.NoError(t, err)

	h.MarshalScope(scope, rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)
}

func TestHandlerUnmarshalScope(t *testing.T) {
	blogModels := []interface{}{&BlogSDK{}, &PostSDK{}, &CommentSDK{}, &AuthorSDK{}}
	h := prepareHandler(defaultLanguages, blogModels...)

	rw, req := getHttpPair("GET", "/blogs", nil)
	h.UnmarshalScope(reflect.TypeOf(&BlogSDK{}), rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)
}

func TestMarshalErrors(t *testing.T) {
	// Case 1:
	// Marshal custom error with invalid status (non http.Status style)
	customError := &ErrorObject{ID: "My custom ID", Status: "Invalid status"}
	h := prepareHandler(defaultLanguages)

	rw := httptest.NewRecorder()
	h.MarshalErrors(rw, customError)

	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 2:
	// no errors provided
	rw = httptest.NewRecorder()
	h.MarshalErrors(rw)
	assert.Equal(t, 400, rw.Result().StatusCode)
}

func TestManageDBError(t *testing.T) {
	h := prepareHandler(defaultLanguages)
	// Case 1:
	// Having custom error not registered within the error manager.
	customError := &unidb.Error{ID: 30, Title: "My custom DBError"}

	// error not registered in the manager
	rw := httptest.NewRecorder()
	h.manageDBError(rw, customError)
	assert.Equal(t, 500, rw.Result().StatusCode)

}

func TestGetRelationshipsFilter(t *testing.T) {
	h := prepareHandler([]language.Tag{language.Polish}, &HumanSDK{}, &PetSDK{})

	rw, req := getHttpPair("GET", "/pets?filter[pets][humans][id][$lt]=2", nil)
	scope, errs, err := h.Controller.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(PetSDK{})})
	assert.NoError(t, err)
	assert.Empty(t, errs)

	mockRepo := &MockRepository{}
	h.DefaultRepository = mockRepo

	err = h.GetRelationshipFilters(scope, req, rw)
	assert.NoError(t, err)

	rw, req = getHttpPair("GET", "/humans?filter[humans][pets][legs][$gt]=3&filter[humans][pets][id][$in]=3,4,5", nil)
	scope, errs, err = h.Controller.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(HumanSDK{})})
	assert.NoError(t, err)
	assert.Empty(t, errs)

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			relScope := args.Get(0).(*Scope)
			assert.NotEmpty(t, relScope.PrimaryFilters)
			assert.NotEmpty(t, relScope.AttributeFilters)
			assert.Empty(t, relScope.RelationshipFilters)
			assert.Empty(t, relScope.Value)

			v := reflect.ValueOf(relScope.Value)
			assert.Equal(t, reflect.Slice, v.Type().Kind())
			relScope.Value = []*PetSDK{{ID: 3}, {ID: 4}}
		})
	err = h.GetRelationshipFilters(scope, req, rw)
	assert.NoError(t, err)

	assert.NotEmpty(t, scope.RelationshipFilters)
	assert.NotEmpty(t, scope.RelationshipFilters[0])

	assert.Contains(t, scope.RelationshipFilters[0].Nested[0].Values[0].Values, 3, 4)
}
