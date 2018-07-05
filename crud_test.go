package jsonapi

import (
	"bytes"
	"context"
	"github.com/kucjac/uni-db"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type funcScopeMatcher func(*Scope) bool

func TestHandlerCreate(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	rw, req := getHttpPair("POST", "/blogs", h.getModelJSON(&BlogSDK{ID: 1, Lang: "pl", CurrentPost: &PostSDK{ID: 1}}))

	// Case 1:
	// Succesful create.
	mockRepo.On("Create", mock.MatchedBy(
		matchScopeByTypeAndID(BlogSDK{}, 1),
	)).Return(nil)
	mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	mh.Create = &Endpoint{Type: Create}
	h.Create(mh, mh.Create).ServeHTTP(rw, req)
	assert.Equal(t, 201, rw.Result().StatusCode)

	// Case 2:
	// Duplicated value
	rw, req = getHttpPair("POST", "/blogs",
		h.getModelJSON(&BlogSDK{ID: 2, Lang: "pl", CurrentPost: &PostSDK{ID: 1}}))

	mockRepo.On("Create", mock.MatchedBy(
		matchScopeByTypeAndID(BlogSDK{}, 2),
	)).Return(unidb.ErrUniqueViolation.New())

	h.Create(mh, mh.Create).ServeHTTP(rw, req)
	assert.Equal(t, 409, rw.Result().StatusCode)

	// Case 3:
	// No language provided error
	// rw, req = getHttpPair("POST", "/blogs",
	// 	h.getModelJSON(&BlogSDK{ID: 3, CurrentPost: &PostSDK{ID: 1}}))

	// h.Create(mh, mh.Create).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 4:
	rw, req = getHttpPair("POST", "/blogs", strings.NewReader(`{"data":{"type":"unknown_collection"}}`))
	h.Create(mh, mh.Create).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Preset Values
	commentModel, ok := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
	assert.True(t, ok, "Model not valid")

	key := "my_key"
	presetPair := h.Controller.BuildPresetPair("preset=blogs.current_post", "filter[comments][post][id][in]").WithKey(key)

	assert.NoError(t, commentModel.AddPresetPair(presetPair, Create))

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 6}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*PostSDK{{ID: 6}}
		})

	mockRepo.On("Create", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			t.Logf("Create value: \"%+v\"", arg.Value)
			comment := arg.Value.(*CommentSDK)
			t.Logf("PostSDK: %v", comment.Post)
			comment.ID = 3

		})

	rw, req = getHttpPair("POST", "/comments", strings.NewReader(`{"data":{"type":"comments","attributes":{"body":"Some title"}}}`))

	req = req.WithContext(context.WithValue(req.Context(), key, 6))
	h.Create(commentModel, commentModel.Create).ServeHTTP(rw, req)
	assert.Equal(t, 201, rw.Result().StatusCode, rw.Body.String())

}

func TestHandlerGet(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	// Case 1:
	// Getting an object correctly without accept-language header
	// mockRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		assert.NotNil(t, arg.LanguageFilters)

	// 		arg.Value = &BlogSDK{ID: 1, Lang: arg.LanguageFilters.Values[0].Values[0].(string),
	// 			CurrentPost: &PostSDK{ID: 1}}
	// 	})

	// rw, req := getHttpPair("GET", "/blogs/1", nil)

	blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogHandler.Get = &Endpoint{Type: Get}
	// h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)

	// // assert content-language is the same
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 2:
	// Getting a non-existing object
	mockRepo.On("Get",
		mock.Anything).Once().Return(unidb.ErrUniqueViolation.New())
	rw, req := getHttpPair("GET", "/blogs/123", nil)
	h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)

	assert.Equal(t, 409, rw.Result().StatusCode)

	// Case 3:
	// assigning bad url - internal error
	rw, req = getHttpPair("GET", "/blogs", nil)
	h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 4:
	// User input error (i.e. invalid query)
	rw, req = getHttpPair("GET", "/blogs/1?include=nonexisting", nil)
	h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)

	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// User provided unsupported language
	rw, req = getHttpPair("GET", "/blogs/1?language=nonsupportedlang", nil)
	h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 6:
	// Getting Included values with nil Values set.
	// Usually the value is inited to be non nil.
	// Otherwise like here it sends internal.
	rw, req = getHttpPair("GET", "/blogs/1?include=current_post", nil)

	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &BlogSDK{ID: 1, Lang: h.SupportedLanguages[0].String(),
				CurrentPost: &PostSDK{ID: 3}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = nil
		})

	h.Get(blogHandler, blogHandler.Get).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	precheckPair := h.Controller.BuildPrecheckPair("preset=blogs.current_post&filter[blogs][id][eq]=1", "filter[comments][post][id]")

	h.ModelHandlers[reflect.TypeOf(CommentSDK{})].AddPrecheckPair(precheckPair, Get)

	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 1, CurrentPost: &PostSDK{ID: 3}}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			t.Logf("%v", arg.Fieldset)
			arg.Value = []*PostSDK{{ID: 3}}
		})

	mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*CommentSDK{{ID: 1, Body: "Some body"}, {ID: 3, Body: "Other body"}}
		})

	rw, req = getHttpPair("GET", "/comments/1", nil)
	commentModel := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]

	h.Get(commentModel, commentModel.Get).ServeHTTP(rw, req)

}

func TestHandlerGetRelated(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	// Case 1:
	// Correct related field
	rw, req := getHttpPair("GET", "/blogs/1/current_post", nil)

	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &BlogSDK{ID: 1, CurrentPost: &PostSDK{ID: 1}}
		})
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &PostSDK{ID: 1, Title: "This title", Comments: []*CommentSDK{{ID: 1}, {ID: 2}}}
		})

	blogModel := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogModel.GetRelated = &Endpoint{Type: GetRelated}
	h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 2:
	// Invalid field name
	rw, req = getHttpPair("GET", "/blogs/1/current_invalid_post", nil)
	h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 3:
	// Invalid model's url. - internal
	rw, req = getHttpPair("GET", "/blogs/current_invalid_post", nil)
	h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 4:
	// invalid language
	// rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)
	// req.Header.Add(headerAcceptLanguage, "invalid_language")
	// h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Root repo dberr
	rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)

	mockRepo.On("Get", mock.Anything).Once().Return(unidb.ErrUniqueViolation.New())
	h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 409, rw.Result().StatusCode)

	// Case 6:
	// No primary filter for scope
	rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &BlogSDK{ID: 1}
		})
	h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	postsHandler := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
	postsHandler.GetRelated = &Endpoint{Type: GetRelated}

	// Case 7:
	// Get many relateds, correct
	rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &PostSDK{ID: 1, Comments: []*CommentSDK{{ID: 1}, {ID: 2}}}

		})
	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*CommentSDK{{ID: 1, Body: "First comment"}, {ID: 2, Body: "Second CommentSDK"}}
		})
	h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 8:
	// Get many relateds, with empty map
	rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &PostSDK{ID: 1}
		})
	h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 9:
	// Provided nil value after getting root from repository - Internal Error
	rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = nil
		})
	h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 9:
	// dbErr for related
	rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &PostSDK{ID: 1, Comments: []*CommentSDK{{ID: 1}}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(unidb.ErrInternalError.New())
	h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 10:
	// Related field that use language has language filter
	rw, req = getHttpPair("GET", "/authors/1/blogs", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &AuthorSDK{ID: 1, Blogs: []*BlogSDK{{ID: 1}, {ID: 2}, {ID: 5}}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 1}, {ID: 2}, {ID: 5}}
		})

	authHandler := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
	authHandler.GetRelated = &Endpoint{Type: GetRelated}
	h.GetRelated(authHandler, authHandler.GetRelated).ServeHTTP(rw, req)
}

func TestGetRelationship(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogHandler.GetRelationship = &Endpoint{Type: GetRelationship}

	// Case 1:
	// Correct Relationship
	rw, req := getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = &BlogSDK{ID: 1, CurrentPost: &PostSDK{ID: 1}}
		})

	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 2:
	// Invalid url - Internal Error
	rw, req = getHttpPair("GET", "/blogs/relationships/current_post", nil)
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 3:
	// Invalid field name
	rw, req = getHttpPair("GET", "/blogs/1/relationships/different_post", nil)
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)
	t.Log(rw.Body)

	// Case 4:
	// Bad languge provided
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// req.Header.Add(headerAcceptLanguage, "invalid language name")
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Error while getting from root repo
	rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().
		Return(unidb.ErrNoResult.New())
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 404, rw.Result().StatusCode)

	// Case 6:
	// Error while getting relationship scope. I.e. assigned bad value type - Internal error
	rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		arg := args[0].(*Scope)
		arg.Value = &AuthorSDK{ID: 1}
	})
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 7:
	// Provided empty value for the scope.Value
	rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		arg := args[0].(*Scope)
		arg.Value = nil
	})
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 8:
	// Provided nil relationship value within scope
	rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		arg := args[0].(*Scope)
		arg.Value = &BlogSDK{ID: 1}
	})
	h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	authorHandler := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
	authorHandler.GetRelationship = &Endpoint{Type: GetRelationship}

	// Case 9:
	// Provided empty slice in hasMany relationship within scope
	rw, req = getHttpPair("GET", "/authors/1/relationships/blogs", nil)
	mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		arg := args[0].(*Scope)
		arg.Value = &AuthorSDK{ID: 1}
	})

	h.GetRelationship(authorHandler, authorHandler.GetRelationship).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

}

func TestHandlerList(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogHandler.List = &Endpoint{Type: List}

	// Case 1:
	// Correct with no Accept-Language header
	rw, req := getHttpPair("GET", "/blogs", nil)

	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 1, CurrentPost: &PostSDK{ID: 1}}}
		})

	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 2:
	// Correct with Accept-Language header

	rw, req = getHttpPair("GET", "/blogs", nil)
	req.Header.Add(headerAcceptLanguage, "pl;q=0.9, en")

	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{}
		})

	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 3:
	// User input error on query
	rw, req = getHttpPair("GET", "/blogs?include=invalid", nil)
	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 4:
	// Getting incorrect language

	rw, req = getHttpPair("GET", "/blogs?language=polski", nil)

	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Internal Error - model not precomputed

	// rw, req = getHttpPair("GET", "/models", nil)
	// h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 6:
	// repository error occurred

	rw, req = getHttpPair("GET", "/blogs", nil)

	mockRepo.On("List", mock.Anything).Once().Return(unidb.ErrInternalError.New())
	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 7:
	// Getting includes correctly

	rw, req = getHttpPair("GET", "/blogs?include=current_post", nil)

	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 1, CurrentPost: &PostSDK{ID: 1}}}
		})
	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*PostSDK{{ID: 1}}
		})

	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Result().StatusCode)

	// Case 8:
	// Getting includes with an error
	rw, req = getHttpPair("GET", "/blogs?include=current_post", nil)
	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 1, CurrentPost: &PostSDK{ID: 1}}}
		})
	mockRepo.On("List", mock.Anything).Once().Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = nil
		})
	h.List(blogHandler, blogHandler.List).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)
}

func TestHandlerPatch(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogHandler.Patch = &Endpoint{Type: Patch}

	// Case 1:
	// Correctly Patched
	rw, req := getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{Lang: "en", CurrentPost: &PostSDK{ID: 2}}))

	mockRepo.On("Patch", mock.Anything).Once().Return(nil)
	h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	assert.Equal(t, 204, rw.Result().StatusCode)

	// Case 2:
	// Patch with preset value
	presetPair := h.Controller.BuildPrecheckPair("preset=blogs.current_post&filter[blogs][id][eq]=3", "filter[comments][post][id][in]")

	commentModel := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
	commentModel.Patch = &Endpoint{Type: Patch, PresetPairs: []*PresetPair{presetPair}}
	assert.NotNil(t, commentModel, "Nil comment model.")

	// commentModel.AddPrecheckPair(presetPair, Patch)

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			arg.Value = []*BlogSDK{{ID: 3, CurrentPost: &PostSDK{ID: 4}}}
			t.Log("First")
		})

	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = []*PostSDK{{ID: 4}}
	// 		t.Log("Second")
	// 	})

	mockRepo.On("Patch", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*Scope)
			t.Logf("%+v", arg)
			t.Logf("Scope value: %+v\n\n", arg.Value)
			// t.Log(arg.PrimaryFilters[0].Values[0].Values)
			// assert.NotEmpty(t, arg.RelationshipFilters)
			// assert.NotEmpty(t, arg.RelationshipFilters[0].Relationships)
			// assert.NotEmpty(t, arg.RelationshipFilters[0].Relationships[0].Values)
			// t.Log(arg.RelationshipFilters[0].Relationships[0].Values[0].Values)
			// t.Log(arg.RelationshipFilters[0].Relationships[0].Values[0].Operator)
		})

	rw, req = getHttpPair("PATCH", "/comments/1", h.getModelJSON(&CommentSDK{Body: "Some body."}))
	h.Patch(commentModel, commentModel.Patch).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNoContent, rw.Result().StatusCode)
	commentModel.Patch.PrecheckPairs = nil

	// Case 3:
	// Incorrect URL for ID provided - internal
	rw, req = getHttpPair("PATCH", "/blogs", h.getModelJSON(&BlogSDK{}))
	h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 4:
	// // No language provided - user error
	// rw, req = getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{CurrentPost: &PostSDK{ID: 2}}))
	// h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Repository error
	rw, req = getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{Lang: "pl", CurrentPost: &PostSDK{ID: 2}}))
	mockRepo.On("Patch", mock.Anything).Once().Return(unidb.ErrForeignKeyViolation.New())
	h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 6:
	// Preset some hidden value

	presetPair = h.Controller.BuildPresetPair("preset=blogs.current_post&filter[blogs][id][eq]=4", "filter[comments][post][id]")
	commentModel.AddPresetPair(presetPair, Patch)
	rw, req = getHttpPair("PATCH", "/comments/1", h.getModelJSON(&CommentSDK{Body: "Preset post"}))

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			scope := args.Get(0).(*Scope)
			scope.Value = []*BlogSDK{{ID: 4}}
		})

	mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			scope := args.Get(0).(*Scope)
			scope.Value = []*PostSDK{{ID: 3}}
		})

	mockRepo.On("Patch", mock.Anything).Once().Return(nil).Run(
		func(args mock.Arguments) {
			scope := args.Get(0).(*Scope)
			t.Log(scope.Value)
		})
	h.Patch(commentModel, commentModel.Patch).ServeHTTP(rw, req)

}

func TestHandlerDelete(t *testing.T) {
	h := prepareHandler(defaultLanguages, blogModels...)
	mockRepo := &MockRepository{}
	h.SetDefaultRepo(mockRepo)

	blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	blogHandler.Delete = &Endpoint{Type: Delete}

	// Case 1:
	// Correct delete.
	rw, req := getHttpPair("DELETE", "/blogs/1", nil)
	mockRepo.On("Delete", mock.Anything).Once().Return(nil)
	h.Delete(blogHandler, blogHandler.Delete).ServeHTTP(rw, req)
	assert.Equal(t, 204, rw.Result().StatusCode)

	// Case 2:
	// Invalid model provided
	// rw, req = getHttpPair("DELETE", "/models/1", nil)
	// h.Delete(h.ModelHandlers[reflect.TypeOf(Model{})]).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 3:
	// Invalid url for ID - internal
	rw, req = getHttpPair("DELETE", "/blogs", nil)
	h.Delete(blogHandler, blogHandler.Delete).ServeHTTP(rw, req)
	assert.Equal(t, 500, rw.Result().StatusCode)

	// Case 4:
	// Invalid ID - user error
	rw, req = getHttpPair("DELETE", "/blogs/stringtype-id", nil)
	h.Delete(blogHandler, blogHandler.Delete).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Repository error
	rw, req = getHttpPair("DELETE", "/blogs/1", nil)
	mockRepo.On("Delete", mock.Anything).Once().Return(unidb.ErrIntegrConstViolation.New())
	h.Delete(blogHandler, blogHandler.Delete).ServeHTTP(rw, req)
	assert.Equal(t, 400, rw.Result().StatusCode)

}

var (
	defaultLanguages = []language.Tag{language.English, language.Polish}
	blogModels       = []interface{}{&BlogSDK{}, &PostSDK{}, &CommentSDK{}, &AuthorSDK{}}
)

func getHttpPair(method, target string, body io.Reader,
) (rw *httptest.ResponseRecorder, req *http.Request) {
	req = httptest.NewRequest(method, target, body)
	req.Header.Add("Content-Type", MediaType)
	rw = httptest.NewRecorder()
	return
}

func prepareModelHandlers(models ...interface{}) (handlers []*ModelHandler) {
	for _, model := range models {
		handler, err := NewModelHandler(model, nil, []EndpointType{Get, GetRelated, GetRelationship, List, Create, Patch, Delete}...)
		if err != nil {
			panic(err)
		}
		handlers = append(handlers, handler)
	}
	return
}

func prepareHandler(languages []language.Tag, models ...interface{}) *JSONAPIHandler {
	c := DefaultController()

	logger := unilogger.MustGetLoggerWrapper(unilogger.NewBasicLogger(os.Stderr, "", log.Ldate))

	h := NewHandler(c, logger, NewDBErrorMgr())
	err := c.PrecomputeModels(models...)
	if err != nil {
		panic(err)
	}

	h.AddModelHandlers(prepareModelHandlers(models...)...)

	h.SetLanguages(languages...)

	return h
}

func matchScopeByType(model interface{}) funcScopeMatcher {
	return func(scope *Scope) bool {
		return isSameType(model, scope)
	}
}

func (h *JSONAPIHandler) getModelJSON(
	model interface{},
) *bytes.Buffer {
	scope, err := h.Controller.NewScope(model)
	if err != nil {
		panic(err)
	}

	scope.Value = model

	payload, err := h.Controller.MarshalScope(scope)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)

	if err = MarshalPayload(buf, payload); err != nil {
		panic(err)
	}
	return buf

}

func matchScopeByTypeAndID(model interface{}, id interface{}) funcScopeMatcher {
	return func(scope *Scope) bool {
		if matched := isSameType(model, scope); !matched {
			return false
		}

		if scope.Value == nil {
			return false
		}

		v := reflect.ValueOf(scope.Value)
		if v.Type().Kind() != reflect.Ptr {
			return false
		}

		idIndex := scope.Struct.GetPrimaryField().GetFieldIndex()
		return reflect.DeepEqual(id, v.Elem().Field(idIndex).Interface())
	}
}

func isSameType(model interface{}, scope *Scope) bool {
	return reflect.TypeOf(model) == scope.Struct.GetType()
}
