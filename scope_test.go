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

	driverRootScope := newRootScope(mStruct)
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
		{"current_post", "id"},
		{"current_post", "id", "eq"},
		{"current_post", "title", "lt"},
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
		{"current_post"},
		{"invalid"},
		{"current_post", "invalid"},
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

func TestScopeGetIncludePrimaryFields(t *testing.T) {
	clearMap()
	getBlogScope()

	req := httptest.NewRequest("GET", "/api/v1/blogs/1?include=posts,current_post.latest_comment", nil)
	scope, errs, err := c.BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	scope.Value = &Blog{
		ID:          1,
		Posts:       []*Post{{ID: 1}, {ID: 3}, {ID: 4}},
		CurrentPost: &Post{ID: 1},
	}

	assertEqual(t, 3, scope.GetTotalIncludeFieldCount())

	scope.NextIncludedField()
	includedField, err := scope.CurrentIncludedField()
	assertNil(t, err)
	assertEqual(t, "posts", includedField.jsonAPIName)

	nonUsed, err := includedField.GetMissingPrimaries()
	assertNil(t, err)

	assertEqual(t, 3, len(nonUsed))

	scope.NextIncludedField()
	includedField, err = scope.CurrentIncludedField()
	assertNil(t, err)
	assertEqual(t, "current_post", includedField.jsonAPIName)

	nonUsed, err = includedField.GetMissingPrimaries()
	assertNil(t, err)
	assertEmpty(t, nonUsed)

	assertTrue(t, includedField.Scope.NextIncludedField())

	includedField.Scope.Value = &Post{
		ID:            1,
		Title:         "This post",
		LatestComment: &Comment{ID: 2, Body: "This is latest comment"},
	}
	nestedIncluded, err := includedField.Scope.CurrentIncludedField()
	assertNil(t, err)
	assertEqual(t, "latest_comment", nestedIncluded.jsonAPIName)

	nonUsed, err = nestedIncluded.GetMissingPrimaries()
	assertNil(t, err)
	assertEqual(t, 1, len(nonUsed))

	assertEmpty(t, nestedIncluded.Scope.PrimaryFilters)
	nestedIncluded.Scope.SetIDFilters(nonUsed...)
	assertNotEmpty(t, nestedIncluded.Scope.PrimaryFilters)

	// let's check if empty values are entered.
	nestedIncluded.Scope.PrimaryFilters = []*FilterField{}

	assertFalse(t, nestedIncluded.Scope.NextIncludedField())
	assertFalse(t, includedField.Scope.NextIncludedField())

	// After Resetting the pointer goes back to it's position
	includedField.Scope.ResetIncludedField()
	assertTrue(t, includedField.Scope.NextIncludedField())

	req = httptest.NewRequest("GET", "/api/v1/blogs?include=posts&filter[posts][id][gt]=0&filter[posts][title][startswith]=this&filter[posts][latest_comment][id][gt]=3", nil)
	scope, errs, err = c.BuildScopeList(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)
	scope.Value = []*Blog{
		{ID: 1, Posts: []*Post{{ID: 1}, {ID: 2}}},
		{ID: 2, Posts: []*Post{{ID: 2}, {ID: 3}}},
		nil,
	}

	includedField = scope.IncludedFields[0]
	includedField.copyScopeBoundaries()
	nonUsed, err = includedField.GetMissingPrimaries()
	assertNil(t, err)
	assertEqual(t, 3, len(nonUsed))
}

func TestScopeBuildFieldset(t *testing.T) {
	clearMap()
	scope := getBlogScope()

	errs := scope.buildFieldset("title", "title", "title", "title", "title")
	assertNotEmpty(t, errs)

}

func TestScopeSetCollectionValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs?include=posts.comments,current_post&fieldset[posts]=title", nil)
	clearMap()
	getBlogScope()
	scope, errs, err := c.BuildScopeList(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	assertNotNil(t, scope)

	scope.Value = []*Blog{
		{ID: 1, Posts: []*Post{{ID: 1}, {ID: 2}}, CurrentPost: &Post{ID: 2}},
		{ID: 2, Posts: []*Post{{ID: 2}, {ID: 3}}, CurrentPost: &Post{ID: 3}},
	}
	err = scope.SetCollectionValues()

	for scope.NextIncludedField() {
		iField, _ := scope.CurrentIncludedField()
		missing, _ := iField.GetMissingPrimaries()
		if iField.jsonAPIName == "posts" {
			assertNotEmpty(t, missing)
			assertNotEqual(t, len(iField.Scope.Fieldset), len(iField.Scope.collectionScope.Fieldset))
			_, ok := iField.Scope.Fieldset["comments"]
			assertTrue(t, ok)
			// get from repo
			iField.Scope.Value = []*Post{
				{ID: 1, Title: "Title", Comments: []*Comment{{ID: 1}}},
				{ID: 2, Title: "Another Title", Comments: []*Comment{{ID: 2}}},
			}
			err := iField.Scope.SetCollectionValues()
			assertNil(t, err)

			for iField.Scope.NextIncludedField() {

			}
		}
	}

	req = httptest.NewRequest("GET", "/blogs/1?include=current_post.comments&fieldset[blogs]=title", nil)
	scope, errs, err = c.BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)

	scope.Value = &Blog{ID: 1, Title: "This post"}
	if scope.IncludedValues == nil {
		scope.IncludedValues = NewSafeHashMap()
	}
	scope.IncludedValues.Add(1, &Blog{ID: 1, Title: "This post", CurrentPost: &Post{ID: 1}})

	err = scope.SetCollectionValues()
	assertNil(t, err)

	blogValue, ok := scope.Value.(*Blog)
	assertTrue(t, ok)
	assertEqual(t, uint64(1), blogValue.CurrentPost.ID)

	if scope.NextIncludedField() {
		currentPost, err := scope.CurrentIncludedField()
		assertNil(t, err)
		missing, err := currentPost.GetMissingPrimaries()
		assertNil(t, err)
		assertEqual(t, uint64(1), missing[0])

		currentPost.Scope.Value = &Post{ID: 1, Title: "Post Title"}
		assertNil(t, currentPost.Scope.SetCollectionValues())

		if currentPost.Scope.NextIncludedField() {
			comment, err := currentPost.Scope.CurrentIncludedField()
			assertNil(t, err)
			assertEqual(t, reflect.TypeOf(Comment{}), comment.relatedStruct.modelType)

		}
	}
}

func TestSetLanguageFilter(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	scope.SetLanguageFilter("en", "pl")

	assertNil(t, scope.LanguageFilters)

	err := c.PrecomputeModels(&Modeli18n{})
	assertNil(t, err)
	scope = newRootScope(c.MustGetModelStruct(&Modeli18n{}))
	scope.SetLanguageFilter("en", "pl")

	assertNotNil(t, scope.LanguageFilters)

	assertEqual(t, 1, len(scope.LanguageFilters.Values))
	scope.SetLanguageFilter("de")
	assertEqual(t, 2, len(scope.LanguageFilters.Values))
}

func TestUseI18n(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	assertFalse(t, scope.UseI18n())

	err := c.PrecomputeModels(&Modeli18n{})
	assertNil(t, err)
	scope = newRootScope(c.MustGetModelStruct(&Modeli18n{}))

	assertTrue(t, scope.UseI18n())
}

func TestScopeGetLangtag(t *testing.T) {
	clearMap()
	err := c.PrecomputeModels(&Modeli18n{})
	assertNil(t, err)

	scope := newRootScope(c.MustGetModelStruct(&Modeli18n{}))
	preSetLang := "pl"
	scope.Value = &Modeli18n{Lang: preSetLang}
	langtag, err := scope.GetLangtagValue()
	assertNil(t, err)
	assertEqual(t, preSetLang, langtag)

	var NilModelVal *Modeli18n
	scope.Value = NilModelVal
	_, err = scope.GetLangtagValue()
	assertError(t, err)

	scope.Value = nil
	_, err = scope.GetLangtagValue()
	assertError(t, err)

	scope.Value = &Blog{ID: 1}
	_, err = scope.GetLangtagValue()
	assertError(t, err)

	scope.Value = []*Modeli18n{}
	_, err = scope.GetLangtagValue()
	assertError(t, err)

	scope = getBlogScope()
	_, err = scope.GetLangtagValue()
	assertError(t, err)
}

func TestScopeGetCollection(t *testing.T) {
	clearMap()
	scope := getBlogScope()

	coll := scope.GetCollection()
	assertEqual(t, coll, scope.Struct.collectionType)
}

func TestScopeSetLangtag(t *testing.T) {
	clearMap()

	err := c.PrecomputeModels(&Modeli18n{})
	assertNil(t, err)

	scope := newRootScope(c.MustGetModelStruct(&Modeli18n{}))

	plLangtag := "pl"

	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope.Value = &Modeli18n{ID: 1}
	err = scope.SetLangtagValue(plLangtag)
	assertNil(t, err)
	assertEqual(t, plLangtag, scope.Value.(*Modeli18n).Lang)

	scope.Value = []*Modeli18n{{ID: 1}, {ID: 2}}
	err = scope.SetLangtagValue(plLangtag)
	assertNil(t, err)

	var nilPtr *Modeli18n
	scope.Value = nilPtr
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope.Value = &Blog{}
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope.Value = []Modeli18n{}
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope.Value = []*Blog{}
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope.Value = []*Modeli18n{{ID: 1}, nilPtr}
	err = scope.SetLangtagValue(plLangtag)
	assertNil(t, err)

	scope.Value = Modeli18n{}
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

	scope = getBlogScope()
	err = scope.SetLangtagValue(plLangtag)
	assertError(t, err)

}

func TestScopeSetValueFromAddressable(t *testing.T) {
	clearMap()
	scope := getBlogScope()

	err := scope.SetValueFromAddressable()
	assertError(t, err)

	scope.valueAddress = &Blog{ID: 1}
	err = scope.SetValueFromAddressable()
	assertNil(t, err)
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
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	if err != nil {
		panic(err)
	}
	scope := newScope(c.MustGetModelStruct(&Blog{}))
	scope.maxNestedLevel = c.IncludeNestedLimit
	scope.collectionScope = scope
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
