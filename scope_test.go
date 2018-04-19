package jsonapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestScopeSetSortScopes(t *testing.T) {
	// having some scope on the model struct
	clearMap()
	err := PrecomputeModels(&User{}, &Pet{})
	assertNil(t, err)
	// t.Log(err)

	mStruct := MustGetModelStruct(&User{})
	assertNotNil(t, mStruct)

	userRootScope := newRootScope(mStruct)

	assertNotNil(t, userRootScope)

	errs := userRootScope.setSortScopes("-name")
	assertEmpty(t, errs)

	errs = userRootScope.setSortScopes("name", "-name")
	assertNotEmpty(t, errs)

	// get too many sortfields
	errs = userRootScope.setSortScopes("name", "surname", "somethingelse")
	assertNotEmpty(t, errs)
	t.Log(errs)

	// check multiple with multiple sortable fields
	PrecomputeModels(&Driver{}, &Car{})
	driverModelStruct := MustGetModelStruct(&Driver{})

	assertNotNil(t, driverModelStruct)

	driverRootScope := newRootScope(driverModelStruct)
	assertNotNil(t, driverRootScope)

	// let's check duplicates
	errs = driverRootScope.setSortScopes("name", "-name")
	assertNotEmpty(t, errs)

	driverRootScope = newRootScope(driverModelStruct)
	// if duplicate is typed more than or equal to three times no more fields are being checked
	errs = driverRootScope.setSortScopes("name", "-name", "name")
	assertNotEmpty(t, errs)

	errs = driverRootScope.setSortScopes("invalid")
	assertNotEmpty(t, errs)
	fmt.Println(errs)
}

func TestBuildIncludedScopes(t *testing.T) {
	// having some scope for possible model
	clearMap()
	err := PrecomputeModels(&Driver{}, &Car{})
	assertNil(t, err)

	mStruct := MustGetModelStruct(&Driver{})
	assertNotNil(t, mStruct)

	driverRootScope := newRootScope(mStruct)
	assertNotNil(t, driverRootScope)

	// having some included parameter that is valid for given model
	var included []string
	var errs []*ErrorObject

	included = []string{"favorite-car"}
	errs, err = driverRootScope.buildIncludedScopes(included...)
	assertNil(t, err)
	assertEmpty(t, errs)

	// if checked again for the same included an ErrorObject should return
	included = append(included, "favorite-car")
	errs, err = driverRootScope.buildIncludedScopes(included...)
	assertNil(t, err)
	assertNotEmpty(t, errs)
	fmt.Println(errs)

	clearMap()

	blogScope := getBlogScope()
	// let's try too many possible includes - blog has max of 6.
	errs, err = blogScope.buildIncludedScopes("some", "thing", "that", "is", "too", "long", "for", "this")
	assertNil(t, err)
	assertNotEmpty(t, errs)
	fmt.Println(errs)

	// let's use too many nested includes
	blogScope = newRootScope(MustGetModelStruct(&Blog{}))
	errs, err = blogScope.buildIncludedScopes("too.many.nesteds")
	assertNil(t, err)
	assertNotEmpty(t, errs)
	fmt.Println(errs)

	// spam with the same include too many times
	blogScope = getBlogScope()
	errs, err = blogScope.buildIncludedScopes("posts", "posts", "posts", "posts")
	assertNil(t, err)
	assertNotEmpty(t, errs)
	clearMap()

	blogScope = getBlogScope()
	errs, err = blogScope.buildIncludedScopes("posts.comments")
	assertNil(t, err)
	assertEmpty(t, errs)
	clearMap()

	// suppose that one of the models is somehow nil - internal should return
	blogScope = getBlogScope()
	cacheModelMap.Set(reflect.TypeOf(Post{}), nil)
	errs, err = blogScope.buildIncludedScopes("posts.comments")
	assertError(t, err)
	assertNotEmpty(t, errs)
	clearMap()

	blogScope = getBlogScope()
	cacheModelMap.Set(reflect.TypeOf(Post{}), nil)
	errs, err = blogScope.buildIncludedScopes("posts")
	assertError(t, err)
	assertNotEmpty(t, errs)
	clearMap()

	// misspelled or invalid nested
	blogScope = getBlogScope()
	errs, err = blogScope.buildIncludedScopes("posts.commentes")
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// misspeled first include in nested
	blogScope = getBlogScope()
	errs, err = blogScope.buildIncludedScopes("postes.comments")
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// check created scopes
	blogScope = getBlogScope()
	errs, err = blogScope.buildIncludedScopes("posts.comments", "posts.latest_comment")
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotEmpty(t, blogScope.SubScopes)
	// t.Log(blogScope.SubScopes)
	postScope := blogScope.SubScopes[0]
	assertNotNil(t, postScope)
	assertTrue(t, postScope.Struct.GetCollectionType() == "posts")
	assertTrue(t, len(postScope.SubScopes) == 2)

	commentsScope := postScope.SubScopes[0]
	assertNotNil(t, commentsScope)
	assertTrue(t, commentsScope.Struct.GetCollectionType() == "comments")
	assertEmpty(t, commentsScope.SubScopes)
}

func TestNewFilterScope(t *testing.T) {

	var errs []*ErrorObject
	var err error

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
		_, errs, err = blogScope.newFilterScope("blogs", correctValues[i], blogScope.Struct, correctParams[i]...)
		if err != nil {
			// t.Log(i)
			// t.Log(correctParams[i])
		}
		assertNil(t, err)
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
		_, errs, err = blogScope.newFilterScope("blogs", invValues[i], blogScope.Struct, invParams[i]...)
		assertNil(t, err)
		assertNotEmpty(t, errs)
		// t.Logf("%d: %s", i, errs)
	}

	internalParameters := [][]string{
		{"current-post", "id"},
		{"current-post", "title", "eq"},
	}

	// internal errors
	internalValues := [][]string{
		{"1"},
		{"title"},
	}

	for i := range internalParameters {
		clearMap()
		blgoScope := getBlogScope()
		cacheModelMap.Set(reflect.TypeOf(Post{}), nil)
		_, errs, err = blgoScope.newFilterScope("blogs", internalValues[i], blgoScope.Struct, internalParameters[i]...)
		assertError(t, err)
		assertNotEmpty(t, errs)
	}

}

func TestBuildMultiScope(t *testing.T) {
	var (
		err   error
		req   *http.Request
		scope *Scope
		errs  []*ErrorObject
	)

	err = PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	// raw scope without query
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertEmpty(t, errs)
	assertNil(t, err)
	assertNotNil(t, scope)

	assertNil(t, scope.Root)
	assertEmpty(t, scope.SubScopes)

	assertEmpty(t, scope.Fields)
	assertEmpty(t, scope.Filters)
	assertEmpty(t, scope.Sorts)
	assertNil(t, scope.PaginationScope)
	assertEqual(t, scope.Struct, MustGetModelStruct(&Blog{}))
	assertNil(t, scope.RelatedField)

	// with include
	req = httptest.NewRequest("GET", "/api/v1/blogs?include=current-post", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertEmpty(t, errs)
	assertNil(t, err)
	assertNotNil(t, scope)

	assertNotEmpty(t, scope.SubScopes)
	assertEqual(t, scope.SubScopes[0].Struct, MustGetModelStruct(&Post{}))

	// include internal error
	req = httptest.NewRequest("GET", "/api/v1/blogs?include=posts", nil)
	cacheModelMap.Set(reflect.TypeOf(Post{}), nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertError(t, err)
	assertNotEmpty(t, errs)

	PrecomputeModels(&Blog{}, &Post{}, &Comment{})

	// with sorts
	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=id,-title,posts.id", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)

	assertNotNil(t, scope)

	assertEqual(t, 3, len(scope.Sorts))
	assertEqual(t, AscendingOrder, scope.Sorts[0].Order)
	assertEqual(t, "id", scope.Sorts[0].Field.jsonAPIName)
	assertEqual(t, DescendingOrder, scope.Sorts[1].Order)
	assertEqual(t, "title", scope.Sorts[1].Field.jsonAPIName)
	assertEqual(t, "posts", scope.Sorts[2].Field.jsonAPIName)
	assertEqual(t, 1, len(scope.Sorts[2].RelScopes))
	assertEqual(t, AscendingOrder, scope.Sorts[2].RelScopes[0].Order)
	assertEqual(t, "id", scope.Sorts[2].RelScopes[0].Field.jsonAPIName)

	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=posts.id,posts.title", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)

	assertEqual(t, 1, len(scope.Sorts))
	assertEqual(t, 2, len(scope.Sorts[0].RelScopes))
	assertEqual(t, "id", scope.Sorts[0].RelScopes[0].Field.jsonAPIName)
	assertEqual(t, "title", scope.Sorts[0].RelScopes[1].Field.jsonAPIName)

	// paginations
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[size]=4&page[number]=5", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertNotNil(t, scope.PaginationScope)
	assertEqual(t, 4, scope.PaginationScope.PageSize)
	assertEqual(t, 5, scope.PaginationScope.PageNumber)

	// pagination limit, offset
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=10&page[offset]=5", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertNotNil(t, scope.PaginationScope)
	assertEqual(t, 10, scope.PaginationScope.Limit)
	assertEqual(t, 5, scope.PaginationScope.Offset)

	// pagination errors
	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=have&page[offset]=a&page[size]=nice&page[number]=day", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)
	// t.Log(errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs?page[limit]=2&page[number]=1", nil)
	_, errs, _ = BuildScopeMulti(req, &Blog{})
	assertNotEmpty(t, errs)

	// filter
	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[blogs][id][eq]=12,55", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 1, len(scope.Filters))
	assertEqual(t, OpEqual, scope.Filters[0].Values[0].Operator)
	assertEqual(t, 2, len(scope.Filters[0].Values[0].Values))

	// invalid filter
	//	- invalid bracket
	//	- invalid operator
	//	- invalid value
	//	- not included collection - 'posts'
	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[[blogs][id][eq]=12,55&filter[blogs][id][invalid]=125&filter[blogs][id]=stringval&filter[posts][id]=12&fields[blogs]=id", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[blogs]=somethingnotid&filter[blogs][id]=againbad&filter[blogs][posts][id]=badid", nil)
	_, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// internal error on filters
	req = httptest.NewRequest("GET", "/api/v1/blogs?filter[blogs][current-post][id][gt]=15", nil)
	cacheModelMap.Set(reflect.TypeOf(Post{}), nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertError(t, err)
	assertNotEmpty(t, errs)

	PrecomputeModels(&Blog{}, &Post{}, &Comment{})

	// fields
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[blogs]=title,posts", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)

	assertEqual(t, 2, len(scope.Fields))
	assertNotEqual(t, scope.Fields[0].fieldName, scope.Fields[1].fieldName)

	// fields error
	//	- bracket error
	//	- nested error
	//	- invalid collection name
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[[blogs]=title&fields[blogs][title]=now&fields[blog]=title&fields[blogs]=title&fields[blogs]=posts", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// field error too many
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[blogs]=title,id,posts,comments,this-comment,some-invalid,current-post", nil)
	_, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// sorterror
	req = httptest.NewRequest("GET", "/api/v1/blogs?sort=posts.comments.id,current-post.itle,postes.comm", nil)
	_, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// unsupported parameter
	req = httptest.NewRequest("GET", "/api/v1/blogs?title=name", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// too many errors
	// after 5 errors the function stops
	req = httptest.NewRequest("GET", "/api/v1/blogs?fields[[blogs]=title&fields[blogs][title]=now&fields[blog]=title&sort=-itle&filter[blog][id]=1&filter[blogs][unknown]=123&filter[blogs][current-post][something]=123", nil)
	scope, errs, err = BuildScopeMulti(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	//internal
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	cacheModelMap.Set(reflect.TypeOf(Blog{}), nil)
	_, _, err = BuildScopeMulti(req, &Blog{})
	assertError(t, err)
	_, _, err = BuildScopeSingle(req, &Blog{})
	assertError(t, err)
}

func TestBuildSingleScope(t *testing.T) {
	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	req := httptest.NewRequest("GET", "/api/v1/blogs/55", nil)
	scope, errs, err := BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 55, scope.Filters[0].Values[0].Values[0])

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[posts]=title", nil)
	scope, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	assertEqual(t, 44, scope.Filters[0].Values[0].Values[0])
	assertEqual(t, 1, len(scope.SubScopes))
	assertEqual(t, 1, len(scope.SubScopes[0].Fields))

	// errored
	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertError(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/posts/1", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertError(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/bad-id", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=invalid", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]]=posts", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs][posts]=title", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/44?include=posts&fields[postis]=title", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?title=some-title", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)

	req = httptest.NewRequest("GET", "/api/v1/blogs/123?fields[postis]=title&fields[posts]=idss&fields[posts]=titles&title=sometitle&fields[blogs]=titles,current-posts", nil)
	_, errs, err = BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertNotEmpty(t, errs)
	t.Log(len(errs))
	t.Log(errs)

}

func TestGetID(t *testing.T) {
	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	err = SetModelURL(&Blog{}, "/api/v1/blogs/")
	assertNil(t, err)

	_, err = url.Parse("/api/v1/blogs/")
	assertNil(t, err)

	// t.Log(u.Path)
	req := httptest.NewRequest("GET", "/api/v1/blogs/1", nil)
	mStruct := MustGetModelStruct(&Blog{})
	var id string
	id, err = getID(req, mStruct)
	assertNil(t, err)
	assertEqual(t, "1", id)

	mStruct.collectionURLIndex = -1
	id, err = getID(req, mStruct)
	assertNil(t, err)
	assertEqual(t, "1", id)

	req = httptest.NewRequest("GET", "/v1/blog/3", nil)
	id, err = getID(req, mStruct)
	assertError(t, err)

	req = httptest.NewRequest("GET", "/api/v1/blogs", nil)
	id, err = getID(req, mStruct)
	assertError(t, err)

	// t.Log(MustGetModelStruct(&Blog{}).collectionURLIndex)
}

func BenchmarkEmptyMap(b *testing.B) {
	var mp map[int]*FilterScope
	for i := 0; i < b.N; i++ {
		mp = make(map[int]*FilterScope)
	}
	if mp != nil {
	}
}

func BenchmarkEmptySlice(b *testing.B) {
	s := getBlogScope()
	l := s.Struct.modelType.NumField()

	var arr []*FilterScope

	for i := 0; i < b.N; i++ {
		for j := 0; j < l; j++ {
			arr = append(arr, &FilterScope{})
		}
		arr = nil
	}
}

func BenchmarkEmptyArr(b *testing.B) {
	s := getBlogScope()
	l := s.Struct.modelType.NumField()
	var arr []*FilterScope
	for i := 0; i < b.N; i++ {
		arr = make([]*FilterScope, l)
	}
	if len(arr) == 0 {

	}
}

func getBlogScope() *Scope {
	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	if err != nil {
		panic(err)
	}
	return newRootScope(MustGetModelStruct(&Blog{}))
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
