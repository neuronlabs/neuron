package jsonapi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestControllerCreation(t *testing.T) {
	cNew := NewController()

	assertNotEmpty(t, cNew.Models)

	assertEqual(t, 1, cNew.IncludeNestedLimit)
	assertEqual(t, 1, cNew.FilterValueLimit)

	cDefault := DefaultController()
	assertNotEmpty(t, cDefault.Models)

	assertNotEqual(t, cNew.ErrorLimitMany, cDefault.ErrorLimitMany)
	assertNotEqual(t, cNew.ErrorLimitSingle, cDefault.ErrorLimitSingle)
}

func TestBuildScopeList(t *testing.T) {
	var (
		err   error
		req   *http.Request
		scope *Scope
		errs  []*ErrorObject
		c     *Controller
	)

	c = DefaultController()
	require.NoError(t, c.PrecomputeModels(&Blog{}, &Post{}, &Comment{}))

	// raw scope without query
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assert.Empty(t, errs)
	assert.Nil(t, err)
	if assert.NotNil(t, scope) {

		assertNotEmpty(t, scope.Fieldset)
		assertEmpty(t, scope.Sorts)
		assertNil(t, scope.Pagination)
		assertEqual(t, scope.Struct, c.MustGetModelStruct(&Blog{}))
	}

	// with include
	req = httptest.NewRequest("GET", "/api/v1/blogs?include=current_post", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertEmpty(t, errs)
	assertNil(t, err)
	if assert.NotNil(t, scope) {

		assertNotEmpty(t, scope.IncludedScopes)
		assertNotNil(t, scope.IncludedScopes[c.MustGetModelStruct(&Post{})])
	}

	assert.NoError(t, c.PrecomputeModels(&Blog{}, &Post{}, &Comment{}))

	// with sorts
	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=id,-title,posts.id", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)

	if assert.NotNil(t, scope) {

		assertEqual(t, 3, len(scope.Sorts))
		assertEqual(t, AscendingOrder, scope.Sorts[0].Order)
		assertEqual(t, "id", scope.Sorts[0].jsonAPIName)
		assertEqual(t, DescendingOrder, scope.Sorts[1].Order)
		assertEqual(t, "title", scope.Sorts[1].jsonAPIName)
		assertEqual(t, "posts", scope.Sorts[2].jsonAPIName)
		assertEqual(t, 1, len(scope.Sorts[2].SubFields))
		assertEqual(t, AscendingOrder, scope.Sorts[2].SubFields[0].Order)
		assertEqual(t, "id", scope.Sorts[2].SubFields[0].jsonAPIName)
	}
	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=posts.id,posts.title", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	if assert.NotNil(t, scope) {

		assertEqual(t, 1, len(scope.Sorts))
		assertEqual(t, 2, len(scope.Sorts[0].SubFields))
		assertEqual(t, "id", scope.Sorts[0].SubFields[0].jsonAPIName)
		assertEqual(t, "title", scope.Sorts[0].SubFields[1].jsonAPIName)
	}
	// paginations
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[size]=4&page[number]=5", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	if assert.NotNil(t, scope) {

		assertNotNil(t, scope.Pagination)
		assertEqual(t, 4, scope.Pagination.PageSize)
		assertEqual(t, 5, scope.Pagination.PageNumber)
	}
	// pagination limit, offset
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=10&page[offset]=5", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	if assert.NotNil(t, scope) {

		assertNotNil(t, scope.Pagination)
		assertEqual(t, 10, scope.Pagination.Limit)
		assertEqual(t, 5, scope.Pagination.Offset)
	}
	// pagination errors
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=have&page[offset]=a&page[size]=nice&page[number]=day", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)
	// t.Log(errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=2&page[number]=1", nil)
	_, errs, _ = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNotEmpty(t, errs)

	// filter
	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[blogs][id][eq]=12,55", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)

	assertEmpty(t, errs)

	assertNotNil(t, scope)

	assertEqual(t, 1, len(scope.PrimaryFilters))
	assertEqual(t, OpEqual, scope.PrimaryFilters[0].Values[0].Operator)
	assertEqual(t, 2, len(scope.PrimaryFilters[0].Values[0].Values))

	// invalid filter
	//	- invalid bracket
	//	- invalid operator
	//	- invalid value
	//	- not included collection - 'posts'
	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[[blogs][id][eq]=12,55&filter[blogs][id][invalid]=125&filter[blogs][id]=stringval&filter[posts][id]=12&fields[blogs]=id", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[blogs]=somethingnotid&filter[blogs][id]=againbad&filter[blogs][posts][id]=badid", nil)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})

	// fields
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[blogs]=title,posts", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)

	// title, posts
	assertEqual(t, 3, len(scope.Fieldset))
	// assertNotEqual(t, scope.Fieldset[0].fieldName, scope.Fieldset[1].fieldName)

	// fields error
	//	- bracket error
	//	- nested error
	//	- invalid collection name
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[[blogs]=title&fields[blogs][title]=now&fields[blog]=title&fields[blogs]=title&fields[blogs]=posts", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// field error too many
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[blogs]=title,id,posts,comments,this-comment,some-invalid,current_post", nil)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// sorterror
	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=posts.comments.id,current_post.itle,postes.comm", nil)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// unsupported parameter
	req = httptest.NewRequest("GET", "/api/v1/blogs?title=name", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// too many errors
	// after 5 errors the function stops
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[[blogs]=title&fields[blogs][title]=now&fields[blog]=title&sort=-itle&filter[blog][id]=1&filter[blogs][unknown]=123&filter[blogs][current_post][something]=123", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	//internal
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	c.Models.Set(reflect.TypeOf(Blog{}), nil)
	_, _, err = c.BuildScopeList(req, &Endpoint{Type: List}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertError(t, err)
}

func TestBuildScopeSingle(t *testing.T) {
	c := DefaultController()
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	req := httptest.NewRequest("GET", "/api/v1/blogs/55", nil)
	scope, errs, err := c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 55, scope.PrimaryFilters[0].Values[0].Values[0])

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[posts]=title", nil)
	scope, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 44, scope.PrimaryFilters[0].Values[0].Values[0])

	postsScope := scope.IncludedScopes[c.MustGetModelStruct(&Post{})]
	assertNotNil(t, postsScope)
	assertEqual(t, 2, len(postsScope.Fieldset))
	// assertNotNil(t, postsScope.Fieldset["title"])

	// errored
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertError(t, err)

	req = httptest.NewRequest("GET", "/api/v1/posts/1", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertError(t, err)

	req = httptest.NewRequest("GET", "/api/v1/blogs/bad-id", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=invalid", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]]=posts", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs][posts]=title", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[postis]=title", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?title=some-title", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?fields[postis]=title&fields[posts]=idss&fields[posts]=titles&title=sometitle&fields[blogs]=titles,current_posts", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?filter[posts][id]=1&include=current_post", nil)
	scope, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)

	// invalid form

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?filter[posts][", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?filter[postis]", nil)
	_, errs, err = c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?filter[comments]", nil)
	_, errs, err = c.buildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})}, 123)
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// CASE X:
	// Preset id
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	scope, errs, err = c.buildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})}, 123)
	assertNil(t, err)
	assertEmpty(t, errs)

	assertNotEmpty(t, scope.PrimaryFilters, failNow)
	assertNotEmpty(t, scope.PrimaryFilters[0].Values, failNow)
	assertEqual(t, scope.PrimaryFilters[0].Values[0].Values[0], 123)
}

func TestPrecomputeModels(t *testing.T) {
	// input valid models
	validModels := []interface{}{&User{}, &Pet{}, &Driver{}, &Car{}, &WithPointer{}, &Blog{}, &Post{}, &Comment{}}
	err := c.PrecomputeModels(validModels...)
	if err != nil {
		t.Error(err)
	}
	clearMap()

	// if somehow map is nil
	c.Models = nil
	err = c.PrecomputeModels(&Timestamp{})
	if err != nil {
		t.Error(err)
	}

	// if one of the relationship is not precomputed
	clearMap()
	// User has relationship with Pet
	err = c.PrecomputeModels(&User{})
	if err == nil {
		t.Error("The User is related to Pets and so that should be an error")
	}
	clearMap()

	// if one of the models is invalid
	err = c.PrecomputeModels(&Timestamp{}, &BadModel{})
	if err == nil {
		t.Error("BadModel should not be accepted in precomputation.")
	}

	// provided Struct type to precompute models
	err = c.PrecomputeModels(Timestamp{})
	if err == nil {
		t.Error("A pointer to the model should be provided.")
	}

	// provided ptr to basic type
	basic := "value"
	err = c.PrecomputeModels(&basic)
	if err == nil {
		t.Error("Only structs should be accepted!")
	}

	// provided slice
	err = c.PrecomputeModels(&[]*Timestamp{})
	if err == nil {
		t.Error("Slice should not be accepted in precomputedModels")
	}

	// if no tagged fields are provided an error would be thrown
	err = c.PrecomputeModels(&ModelNonTagged{})
	if err == nil {
		t.Error("Non tagged models are not allowed.")
	}
	clearMap()

	// models without primary are not allowed.
	err = c.PrecomputeModels(&NoPrimaryModel{})
	if err == nil {
		t.Error("No primary field provided.")
	}
	clearMap()

	type InvalidPrimaryField struct {
		ID float64 `jsonapi:"primary,invalids"`
	}

	err = c.PrecomputeModels(&InvalidPrimaryField{})
	assertError(t, err)
}

func TestBuildScopeRelationship(t *testing.T) {
	clearMap()
	c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	req := httptest.NewRequest("GET", "/api/v1/blogs/1/relationships/posts", nil)
	scope, errs, err := c.BuildScopeRelationship(req, &Endpoint{Type: GetRelationship}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 1, len(scope.Fieldset))

	scope.Value = &Blog{ID: 1, Posts: []*Post{{ID: 1}, {ID: 2}}}
	postsScope, err := scope.GetRelationshipScope()
	assertNil(t, err)

	assertEqual(t, relationshipKind, postsScope.kind)
	assertEqual(t, reflect.TypeOf([]*Post{}), reflect.TypeOf(postsScope.Value))
	posts, ok := postsScope.Value.([]*Post)
	assertTrue(t, ok)

	for _, val := range posts {
		assertTrue(t, val.ID == 1 || val.ID == 2)
	}

	req = httptest.NewRequest("GET", "/api/v1/blogs/1/relationships/current_post", nil)
	scope, errs, err = c.BuildScopeRelationship(req, &Endpoint{Type: GetRelationship}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	scope.Value = &Blog{ID: 2, CurrentPost: &Post{ID: 1}}
	postScope, err := scope.GetRelationshipScope()
	assertNil(t, err)
	assertNil(t, postScope.Fieldset)

	assertEqual(t, relationshipKind, postScope.kind)
	assertEqual(t, reflect.TypeOf(&Post{}), reflect.TypeOf(postScope.Value))

	req = httptest.NewRequest("GET", "/api/v1/blogs/1/relationships/invalid_field", nil)
	_, errs, err = c.BuildScopeRelationship(req, &Endpoint{Type: GetRelationship}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

}

func TestBuildScopeRelated(t *testing.T) {
	clearMap()
	c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})

	// Case 1:
	// Valid field name
	req := httptest.NewRequest("GET", "/api/v1/blogs/1/posts", nil)
	scope, errs, err := c.BuildScopeRelated(req, &Endpoint{Type: GetRelated}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)

	assertEqual(t, 1, len(scope.Fieldset))

	scope.Value = &Blog{}

	assertTrue(t, scope.NextIncludedField(), failNow)
	inc, err := scope.CurrentIncludedField()
	assertNoError(t, err, failNow)

	assertNotNil(t, inc, failNow)

	// Case 2:
	// invalid field name
	req = httptest.NewRequest("GET", "/api/v1/blogs/1/invalid_field", nil)
	_, errs, err = c.BuildScopeRelated(req, &Endpoint{Type: GetRelated}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// Case 3:
	// Valid field with hasOne relationship
	req = httptest.NewRequest("GET", "/api/v1/blogs/1/current_post", nil)
	scope, errs, err = c.BuildScopeRelated(req, &Endpoint{Type: GetRelated}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNil(t, err)
	assertEmpty(t, errs)

	scope.Value = &Blog{ID: 1, CurrentPost: &Post{ID: 5}}
	relatedScope, err := scope.GetRelatedScope()
	assertNoError(t, err)
	assertEqual(t, 1, len(relatedScope.PrimaryFilters))

	assertEqual(t, uint64(5), relatedScope.PrimaryFilters[0].Values[0].Values[0])

	// Case 4:
	// Valid field with hasMany relationship
	req = httptest.NewRequest("GET", "/api/v1/blogs/1/posts", nil)
	scope, errs, err = c.BuildScopeRelated(req, &Endpoint{Type: GetRelated}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assertNoError(t, err)
	assertEmpty(t, errs)

	scope.Value = &Blog{ID: 1, Posts: []*Post{{ID: 1}, {ID: 5}}}
	relatedScope, err = scope.GetRelatedScope()
	assertNoError(t, err)

	assertEqual(t, 1, len(relatedScope.PrimaryFilters))
	assertEqual(t, uint64(1), relatedScope.PrimaryFilters[0].Values[0].Values[0])
	assertEqual(t, uint64(5), relatedScope.PrimaryFilters[0].Values[0].Values[1])

}

func TestGetModelStruct(t *testing.T) {
	// MustGetModelStruct
	// if the model is not in the cache map
	clearMap()
	assertPanic(t, func() {
		c.MustGetModelStruct(Timestamp{})
	})

	c.Models.Set(reflect.TypeOf(Timestamp{}), &ModelStruct{})
	mStruct := c.MustGetModelStruct(Timestamp{})
	if mStruct == nil {
		t.Error("The model struct shoud not be nil.")
	}

	// GetModelStruct
	// providing ptr should return mStruct
	var err error
	_, err = c.GetModelStruct(&Timestamp{})
	if err != nil {
		t.Error(err)
	}

	// nil model
	_, err = c.GetModelStruct(nil)
	if err == nil {
		t.Error(err)
	}
}

func TestNewScope(t *testing.T) {
	clearMap()
	getBlogScope()

	scope, err := c.NewScope(&Blog{})
	assertNil(t, err)
	assertNotNil(t, scope)
	assertEqual(t, scope.Struct.modelType, reflect.TypeOf(Blog{}))

	scope, err = c.NewScope(&Book{})
	assertError(t, err)
	assertNil(t, scope)

	scope, err = c.NewScope(Book{})
	assertError(t, err)
	assertNil(t, scope)
}

func TestPresetScope(t *testing.T) {
	clearMap()
	assertNoError(t, c.PrecomputeModels(&Blog{}, &Post{}, &Comment{}), failNow)

	// Case 1:
	// Valid query

	// Select all possible comments for blog with id 1 where only last 10 post are taken into
	// account
	query := "preset=blogs.posts&filter[blogs][id][eq]=1&page[limit][posts]=10&sort[posts]=-id&fields[posts]=comments"
	filter := "filter[comments][id][eq]"
	var (
		presetPair *PresetPair
	)

	assertNoPanic(t, func() {
		presetPair = c.BuildPresetPair(query, filter)
	}, failNow)

	presetScope := presetPair.Scope
	filterField := presetPair.Filter

	// The filter field should be the primary field of comments...
	assertEqual(t, c.MustGetModelStruct(&Comment{}).primary, filterField.StructField, failNow)

	// ...with filter operator - eq
	assertEqual(t, OpEqual, filterField.Values[0].Operator)

	// The presetScope should be of type Blogs
	assertEqual(t, presetScope.Struct.modelType, reflect.TypeOf(Blog{}))
	// It should contain filterField for ID equal to 1
	assertNotNil(t, presetScope.PrimaryFilters, failNow)

	assertNotEmpty(t, presetScope.PrimaryFilters, failNow)

	assertEqual(t, Primary, presetScope.PrimaryFilters[0].fieldType)
	assertEqual(t, 1, presetScope.PrimaryFilters[0].Values[0].Values[0].(int))

	// The preset scope should include posts and comments
	assertNotNil(t, presetScope.IncludedScopes, failNow)

	assertNotNil(t, presetScope.IncludedScopes[c.MustGetModelStruct(&Post{})], failNow)

	// The preset scope should contain Posts include field
	assertNotEmpty(t, presetScope.IncludedFields, failNow)

	assertEqual(t, presetScope.IncludedFields[0].jsonAPIName, "posts")

	postsScope := presetScope.IncludedFields[0].Scope
	assertNotNil(t, postsScope, failNow)

	assertEqual(t, reflect.TypeOf(Post{}), postsScope.Struct.modelType)

	assertNotEmpty(t, postsScope.Sorts, failNow)
	assertEqual(t, Primary, postsScope.Sorts[0].fieldType)
	assertEqual(t, DescendingOrder, postsScope.Sorts[0].Order)

	assertNotNil(t, postsScope.Pagination, failNow)
	assertEqual(t, OffsetPaginate, postsScope.Pagination.Type)

	// // The PostIncludeFieldScope should contain Comment Include Field
	// assertNotEmpty(t, postsScope.IncludedFields, failNow)
	// assertEqual(t, "comments", postsScope.IncludedFields[0].jsonAPIName)
	// commentsScope := postsScope.IncludedFields[0].Scope
	// assertNotNil(t, commentsScope, failNow)

	// // The comment scope should be of Comment type
	// assertEqual(t, reflect.TypeOf(Comment{}), commentsScope.Struct.modelType)

	// Case 2:
	// Bad collection name in query

	query = "preset=blog.posts.comments"
	assertPanic(t, func() { c.BuildPresetPair(query, filter) })

	// Case 3:
	// invalid field name
	query = "preset=blogs.posts.comms"
	assertPanic(t, func() { c.BuildPresetPair(query, filter) })

	// Case 4:
	// No fieldset found within the query
	// query = "preset=blogs.posts"
	// assertPanic(t, func() { c.BuildPresetPair(query, filter) }, printPanic)

	// Case 5:
	// Invalid filter field
	query = "preset=blogs.posts&fieldset[posts]=comments&filter[posts][nofield]=3"
	assertPanic(t, func() { c.BuildPresetPair(query, filter) })

	// Case 6:
	// Invalid sort field
	query = "preset=blogs.current_post&fieldset[posts]=comments&sort[blogs]=-nofield&page[limit][blogs]=10"
	assertPanic(t, func() { c.BuildPresetPair(query, filter) }, printPanic)

}

func TestControllerNewFilterField(t *testing.T) {
	filter := "filter[blogs][posts][id][notin]"
	values := []interface{}{uint64(1), uint64(2), uint64(3), uint64(4)}

	clearMap()
	getBlogScope()

	filterField, err := c.NewFilterField(filter, values...)
	assertNoError(t, err, failNow)

	mStruct := c.MustGetModelStruct(&Blog{})
	assertEqual(t, mStruct, filterField.mStruct)
	assertEqual(t, mStruct.relationships["posts"], filterField.StructField)
	assertEqual(t, c.MustGetModelStruct(&Post{}), filterField.relatedStruct, failNow)

	assertNotEmpty(t, filterField.Nested, failNow)
	assertEqual(t, c.MustGetModelStruct(&Post{}).primary, filterField.Nested[0].StructField)

	filter = "filter[posts][title][eq]"
	values = []interface{}{"myTitle", "yourTitle"}

	filterField, err = c.NewFilterField(filter, values...)
	assertNoError(t, err, failNow)

	assertEqual(t, c.MustGetModelStruct(&Post{}), filterField.mStruct)
	assertEqual(t, c.MustGetModelStruct(&Post{}).attributes["title"], filterField.StructField)

	assertEqual(t, OpEqual, filterField.Values[0].Operator)
	assertEqual(t, len(values), len(filterField.Values[0].Values))

}

func TestControllerLanguageQuery(t *testing.T) {
	languages := []interface{}{[]language.Tag{language.Polish, language.English}}
	c := DefaultController(languages...)
	assertNoError(t, c.PrecomputeModels(&Modeli18n{}))

	// Case 1:
	// Check lang
	_, req := getHttpPair("GET", "/translateable?language=pl", nil)

	scope, errs, err := c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err, failNow)
	assertEmpty(t, errs)
	assertNotNil(t, scope.LanguageFilters)
	// t.Logf("%+v", scope.LanguageFilters.Values[0].Values[0])

	// Case 2:
	// individual filter
	_, req = getHttpPair("GET", "/translateable?filter[translateable][langcode]=en", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err, failNow)
	assertEmpty(t, errs)

	assertNotNil(t, scope.LanguageFilters)

	// Case 3:
	// invidividual filter with operator
	_, req = getHttpPair("GET", "/translateable?filter[translateable][langcode][in]=pl,en", nil)
	scope, errs, err = c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err)
	assertEmpty(t, errs)

	assertNotNil(t, scope.LanguageFilters)

	// Case 4:
	// provide not supported language
	_, req = getHttpPair("GET", "/translateable?filter[translateable][langcode][eq]="+language.Turkish.String(), nil)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err)
	assertNotEmpty(t, errs)

	// Case 5:
	// provided unsupported languages and syntetically invalid
	_, req = getHttpPair("GET",
		"/translateable?filter[translateable][langcode][in]=pl-EN,123",
		nil,
	)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err)
	assertNotEmpty(t, errs)

	// Case 6:
	// unsupported langueage in global query
	_, req = getHttpPair("GET", "/translateable?language=tr,de-GB", nil)
	_, errs, err = c.BuildScopeList(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Modeli18n{})})
	assertNoError(t, err)
	assertNotEmpty(t, errs)
}
