package jsonapi

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var c *Controller

func init() {
	c = Default()
}

func TestScopeSetSortScopes(t *testing.T) {
	// having some scope on the model struct
	clearMap()
	err := c.PrecomputeModels(&User{}, &Pet{})
	assertNil(t, err)
	// t.Log(err)

	mStruct := c.MustGetModelStruct(&User{})
	assertNotNil(t, mStruct)

	userRootScope := newScope(mStruct)

	assertNotNil(t, userRootScope)

	errs := userRootScope.buildSortFields("-name")
	assertEmpty(t, errs)

	errs = userRootScope.buildSortFields("name", "-name")
	assertNotEmpty(t, errs)

	// get too many sortfields
	errs = userRootScope.buildSortFields("name", "surname", "somethingelse")
	assertNotEmpty(t, errs)
	t.Log(errs)

	// check multiple with multiple sortable fields
	c.PrecomputeModels(&Driver{}, &Car{})
	driverModelStruct := c.MustGetModelStruct(&Driver{})

	assertNotNil(t, driverModelStruct)

	driverRootScope := newScope(driverModelStruct)
	assertNotNil(t, driverRootScope)

	// let's check duplicates
	errs = driverRootScope.buildSortFields("name", "-name")
	assertNotEmpty(t, errs)

	driverRootScope = newScope(driverModelStruct)
	// if duplicate is typed more than or equal to three times no more fields are being checked
	errs = driverRootScope.buildSortFields("name", "-name", "name")
	assertNotEmpty(t, errs)

	errs = driverRootScope.buildSortFields("invalid")
	assertNotEmpty(t, errs)
	fmt.Println(errs)
}

func TestBuildIncludedScopes(t *testing.T) {
	// having some scope for possible model
	clearMap()
	err := c.PrecomputeModels(&Driver{}, &Car{})
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&Driver{})
	assertNotNil(t, mStruct)

	driverRootScope := newScope(mStruct)
	assertNotNil(t, driverRootScope)

	// having some included parameter that is valid for given model
	var included []string
	var errs []*ErrorObject

	included = []string{"favorite-car"}
	errs = driverRootScope.buildIncludeList(included...)
	assertEmpty(t, errs)

	// if checked again for the same included an ErrorObject should return
	included = append(included, "favorite-car")
	errs = driverRootScope.buildIncludeList(included...)
	assertNotEmpty(t, errs)

	clearMap()

	blogScope := getBlogScope()

	// let's try too many possible includes - blog has max of 6.
	errs = blogScope.buildIncludeList("some", "thing", "that", "is", "too", "long", "for", "this")

	assertNotEmpty(t, errs)

	// let's use too many nested includes
	blogScope = newScope(c.MustGetModelStruct(&Blog{}))
	errs = blogScope.buildIncludeList("too.many.nesteds")
	assertNotEmpty(t, errs)

	// spam with the same include too many times
	blogScope = getBlogScope()
	errs = blogScope.buildIncludeList("posts", "posts", "posts", "posts")

	assertNotEmpty(t, errs)
	clearMap()

	blogScope = getBlogScope()
	errs = blogScope.buildIncludeList("posts.comments")
	assertEmpty(t, errs)
	clearMap()

	// misspelled or invalid nested
	blogScope = getBlogScope()
	errs = blogScope.buildIncludeList("posts.commentes")

	assertNotEmpty(t, errs)

	// misspeled first include in nested
	blogScope = getBlogScope()
	errs = blogScope.buildIncludeList("postes.comments")

	assertNotEmpty(t, errs)

	// check created scopes
	blogScope = getBlogScope()
	errs = blogScope.buildIncludeList("posts.comments", "posts.latest_comment")

	assertEmpty(t, errs)
	assertNotEmpty(t, blogScope.IncludedScopes)
	// t.Log(blogScope.SubScopes)
	postScope := blogScope.IncludedScopes[c.MustGetModelStruct(&Post{})]
	assertNotNil(t, postScope)
	assertTrue(t, postScope.Struct.GetCollectionType() == "posts")

	commentsScope := blogScope.IncludedScopes[c.MustGetModelStruct(&Comment{})]
	assertNotNil(t, commentsScope)
	assertTrue(t, commentsScope.Struct.GetCollectionType() == "comments")
	assertEmpty(t, commentsScope.IncludedScopes)
}

func TestNewFilterScope(t *testing.T) {

	var errs []*ErrorObject

	var correctParams [][]string = [][]string{
		{"id"},
		{"id", "eq"},
		{"title"},
		{"title", "gt"},
		{"current-post", "id"},
		{"current-post", "id", "eq"},
		{"current-post", "title", "lt"},
		{"posts", "id", "gt"},
		{"posts", "body", "contains"},
	}
	var correctValues [][]string = [][]string{
		{"1", "2"},
		{"3", "256"},
		{"maciek", "21-mietek"},
		{"mincek"},
		{"11", "124"},
		{"15", "634"},
		{"the tile of post", "this title"},
		{"333"},
		{"This is in contain"},
	}

	blogScope := getBlogScope()
	for i := range correctParams {
		_, errs = blogScope.buildFilterfield("blogs", correctValues[i], blogScope.Struct, correctParams[i]...)
		assertEmpty(t, errs)

	}
	// for k, v := range blogScope.Filters {
	// 	t.Logf("Key: %v, FieldName: %v", k, v.fieldName)

	// 	for _, f := range v.PrimFilters {
	// 		// t.Logf("Primary field filter: %v", f)
	// 	}

	// 	for _, f := range v.AttrFilters {
	// 		// t.Logf("AttrFilters: %v", f)
	// 	}

	// 	for _, f := range v.Relationships {
	// 		// t.Logf("RelFilter FieldName: %s", f.fieldName)
	// 		for _, fv := range f.PrimFilters {
	// 			// t.Logf("Primary filters:%v", fv)
	// 		}

	// 		for _, fv := range f.AttrFilters {
	// 			// t.Logf("Attrs: %v", fv)
	// 		}
	// 		// t.Logf("\n")
	// 	}
	// 	// t.Logf("\n")

	// }

	var invParams [][]string = [][]string{
		{},
		{"current-post"},
		{"invalid"},
		{"current-post", "invalid"},
		{"invalid", "subfield"},
		{"title", "nosuchoperator"},
		{"invalid-field", "with", "operator"},
		{"title", "with", "operator"},
		{"so", "many", "parameters", "here"},
	}

	var invValues [][]string = [][]string{
		{},
		{},
		{},
		{},
		{},
		{},
		{},
		{},
		{},
	}

	for i := range invParams {
		blogScope := getBlogScope()
		// t.Log(i)
		_, errs = blogScope.buildFilterfield("blogs", invValues[i], blogScope.Struct, invParams[i]...)
		assertNotEmpty(t, errs)
		// t.Logf("%d: %s", i, errs)
	}

}

func TestScopeNewValue(t *testing.T) {
	clearMap()
	scope := getBlogScope()

	scope.NewValueSingle()
	assertEqual(t, reflect.TypeOf(scope.Value), reflect.TypeOf(&Blog{}))

	scope = getBlogScope()
	scope.NewValueMany()
	assertEqual(t, reflect.TypeOf(scope.Value), reflect.TypeOf([]*Blog{}))
}

func TestScopeSetPrimaryFields(t *testing.T) {
	clearMap()
	getBlogScope()

	req := httptest.NewRequest("GET", "/api/v1/blogs/1?include=posts,current-post", nil)
	scope, errs, err := c.BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	scope.Value = &Blog{ID: 1, Posts: []*Post{{ID: 3, Title: "Some title"}, {ID: 4, Title: "Other title"}}, CurrentPost: &Post{ID: 1, Title: "This post"}}
	err = scope.SetIncludedPrimaries()
	assertNil(t, err)

	assertTrue(t, len(scope.IncludedScopes) > 0)
	postScope, ok := scope.IncludedScopes[c.MustGetModelStruct(&Post{})]
	assertTrue(t, ok)

	req = httptest.NewRequest("GET", "/api/v1/blogs?include=current-post", nil)
	scope, errs, err = c.BuildScopeList(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	scope.Value = []*Blog{{ID: 2, CurrentPost: &Post{ID: 1}}, {ID: 3, CurrentPost: &Post{ID: 5}}}
	err = scope.SetIncludedPrimaries()
	assertNil(t, err)

	assertTrue(t, len(scope.IncludedScopes) > 0)

	postScope = scope.IncludedScopes[c.MustGetModelStruct(&Post{})]
	assertNotNil(t, postScope)

	// assertEqual(t, []interface{}{1, 5}, postScope.PrimaryFilters[0].Values[0].Values)

}

// func TestGetID(t *testing.T) {
// 	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
// 	assertNil(t, err)

// 	err = SetModelURL(&Blog{}, "/api/v1/blogs/")
// 	assertNil(t, err)

// 	_, err = url.Parse("/api/v1/blogs/")
// 	assertNil(t, err)

// 	// t.Log(u.Path)
// 	req := httptest.NewRequest("GET", "/api/v1/blogs/1", nil)
// 	mStruct := MustGetModelStruct(&Blog{})
// 	var id string
// 	id, err = getID(req, mStruct)
// 	assertNil(t, err)
// 	assertEqual(t, "1", id)

// 	mStruct.collectionURLIndex = -1
// 	id, err = getID(req, mStruct)
// 	assertNil(t, err)
// 	assertEqual(t, "1", id)

// 	req = httptest.NewRequest("GET", "/v1/blog/3", nil)
// 	id, err = getID(req, mStruct)
// 	assertError(t, err)

// 	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
// 	id, err = getID(req, mStruct)
// 	assertError(t, err)

// 	// t.Log(MustGetModelStruct(&Blog{}).collectionURLIndex)
// }

func BenchmarkEmptyMap(b *testing.B) {
	var mp map[int]*FilterField
	for i := 0; i < b.N; i++ {
		mp = make(map[int]*FilterField)
	}
	if mp != nil {
	}
}

func BenchmarkEmptySlice(b *testing.B) {
	s := getBlogScope()
	l := s.Struct.modelType.NumField()

	var arr []*FilterField

	for i := 0; i < b.N; i++ {
		for j := 0; j < l; j++ {
			arr = append(arr, &FilterField{})
		}
		arr = nil
	}
}

func BenchmarkEmptyArr(b *testing.B) {
	s := getBlogScope()
	l := s.Struct.modelType.NumField()
	var arr []*FilterField
	for i := 0; i < b.N; i++ {
		arr = make([]*FilterField, l)
	}
	if len(arr) == 0 {

	}
}

func getBlogScope() *Scope {
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	if err != nil {
		panic(err)
	}
	scope := newScope(c.MustGetModelStruct(&Blog{}))
	scope.maxNestedLevel = c.IncludeNestedLimit
	return scope
}

func BenchmarkCheckMapStrings(b *testing.B) {
	q := prepareValues()

	for i := 0; i < b.N; i++ {
		for k := range q {
			switch {
			case k == QueryParamPageSize:
			case k == QueryParamPageNumber:
			case k == QueryParamPageOffset:
			case k == QueryParamPageLimit:
			case k == QueryParamInclude:
			case k == QueryParamSort:
			case strings.HasPrefix(k, "fields"):
			case strings.HasPrefix(k, "filter"):
			}
		}
	}
}

func BenchmarkCheckMapStrings2(b *testing.B) {
	q := prepareValues()

	for i := 0; i < b.N; i++ {
		for k := range q {
			switch {
			case k == QueryParamPageSize:
			case k == QueryParamPageNumber:
			case k == QueryParamPageOffset:
			case k == QueryParamPageLimit:
			case k == QueryParamInclude:
			case k == QueryParamSort:
			case k == "fields[blogs]":
			case k == "fields[posts]":
			case strings.HasPrefix(k, "filter"):
			}
		}
	}
}

func prepareValues() url.Values {
	q := url.Values{}
	q.Add("page[size]", "3")
	q.Add("page[number]", "13")
	q.Add("sort", "+field")
	q.Add("include", "included")
	q.Add("fields[blogs]", "some fields")
	q.Add("fields[posts]", "some fields")
	q.Add("filter[blogs][id][eq]", "1")
	return q
}
