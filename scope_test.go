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

func TestScopeCheckFields(t *testing.T) {
	clearMap()
	blogScope := getBlogScope()
	var errs []*ErrorObject
	var err error

	// filter attribute
	errs, err = blogScope.checkFilterField("title", "SomeValue")
	assertNil(t, err)
	assertEmpty(t, errs)

	// relationship (list type)
	errs, err = blogScope.checkFilterField("posts", "1")
	assertNil(t, err)
	assertEmpty(t, errs)

	errs, err = blogScope.checkFilterField("current-post", "2")
	assertNil(t, err)
	assertEmpty(t, errs)

	// blog does not have field 'name'
	errs, err = blogScope.checkFilterField("name", "some name")
	assertNil(t, err)
	assertNotEmpty(t, errs)

	// assign invalid value for relationships
	errs, err = blogScope.checkFilterField("posts", "something not int")
	assertNil(t, err)
	// assertNotEmpty(t, errs)

	errs, err = blogScope.checkFilterField("current-post", "not an int")
	assertNil(t, err)
	// assertNotEmpty(t, errs)

	// assign invalid value for some attribute
	errs, err = blogScope.checkFilterField("view_count", "too many")
	assertNil(t, err)
	// assertNotEmpty(t, errs)

}

func getBlogScope() *Scope {
	PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	return newRootScope(MustGetModelStruct(&Blog{}))

}
