package jsonapi

import (
	"fmt"
	"reflect"
	"testing"
)

func TestScopeSetSortFields(t *testing.T) {
	// having some scope on the model struct
	PrecomputeModels(&User{}, &Pet{})
	mStruct := MustGetModelStruct(&User{})
	assertNotNil(t, mStruct)

	userRootScope := newRootScope(mStruct)

	assertNotNil(t, userRootScope)

	errs := userRootScope.setSortFields("-name")
	assertEmpty(t, errs)

	errs = userRootScope.setSortFields("name", "-name")
	assertNotEmpty(t, errs)
	fmt.Println(errs)

	// get too many sortfields
	errs = userRootScope.setSortFields("name", "surname", "somethingelse")
	assertNotEmpty(t, errs)
	t.Log(errs)

	// check multiple with multiple sortable fields
	PrecomputeModels(&Driver{}, &Car{})
	driverModelStruct := MustGetModelStruct(&Driver{})
	assertNotNil(t, driverModelStruct)

	driverRootScope := newRootScope(driverModelStruct)
	assertNotNil(t, driverRootScope)

	// let's check duplicates
	errs = driverRootScope.setSortFields("name", "-name")
	assertNotEmpty(t, errs)

	driverRootScope = newRootScope(driverModelStruct)
	// if duplicate is typed more than or equal to three times no more fields are being checked
	errs = driverRootScope.setSortFields("name", "-name", "name")
	assertNotEmpty(t, errs)

	errs = driverRootScope.setSortFields("invalid")
	assertNotEmpty(t, errs)
	fmt.Println(errs)
}

func TestBuildIncludedScopes(t *testing.T) {
	// having some scope for possible model
	PrecomputeModels(&Driver{}, &Car{})

	mStruct := MustGetModelStruct(&Driver{})
	assertNotNil(t, mStruct)

	driverRootScope := newRootScope(mStruct)
	assertNotNil(t, driverRootScope)

	// having some included parameter that is valid for given model
	var included []string
	var errs []*ErrorObject
	var err error

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

	postScope := blogScope.SubScopes[0]
	assertNotNil(t, postScope)
	assertTrue(t, postScope.Struct.GetCollectionType() == "posts")
	assertTrue(t, len(postScope.SubScopes) == 2)

	commentsScope := postScope.SubScopes[0]
	assertNotNil(t, commentsScope)
	assertTrue(t, commentsScope.Struct.GetCollectionType() == "comments")
	assertEmpty(t, commentsScope.SubScopes)
}

func TestNewFilterField(t *testing.T) {

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
		_, errs, err = blogScope.newFilterField("blogs", correctValues[i], blogScope.Struct, correctParams[i]...)
		if err != nil {
			t.Log(i)
			t.Log(correctParams[i])
		}
		assertNil(t, err)
		assertEmpty(t, errs)

	}
	for k, v := range blogScope.Filters {
		t.Logf("Key: %v, FieldName: %v", k, v.fieldName)

		for _, f := range v.PrimFilters {
			t.Logf("Primary field filter: %v", f)
		}

		for _, f := range v.AttrFilters {
			t.Logf("AttrFilters: %v", f)
		}

		for _, f := range v.RelFilters {
			t.Logf("RelFilter FieldName: %s", f.fieldName)
			for _, fv := range f.PrimFilters {
				t.Logf("Primary filters:%v", fv)
			}

			for _, fv := range f.AttrFilters {
				t.Logf("Attrs: %v", fv)
			}
			t.Logf("\n")
		}
		t.Logf("\n")

	}

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
		t.Log(i)
		_, errs, err = blogScope.newFilterField("blogs", invValues[i], blogScope.Struct, invParams[i]...)
		assertNil(t, err)
		assertNotEmpty(t, errs)
		t.Logf("%d: %s", i, errs)
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
		_, errs, err = blgoScope.newFilterField("blogs", internalValues[i], blgoScope.Struct, internalParameters[i]...)
		assertError(t, err)
		assertNotEmpty(t, errs)
	}
}

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
	PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	return newRootScope(MustGetModelStruct(&Blog{}))

}
