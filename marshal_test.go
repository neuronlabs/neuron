package jsonapi

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMarshalScope(t *testing.T) {

	buf := bytes.NewBufferString("")

	c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})

	req := httptest.NewRequest("GET", `/blogs/3?include=posts,current_post.latest_comment&fields[blogs]=title,created_at,posts&fields[posts]=title,body,comments`, nil)
	scope, errs, err := c.BuildScopeSingle(req, &Blog{})
	assertNil(t, err)
	assertEmpty(t, errs)
	scope.Value = &Blog{ID: 3, Title: "My own title.", CreatedAt: time.Now(), Posts: []*Post{{ID: 1}}, CurrentPost: &Post{ID: 2}}

	// assertEqual(t, 1, len(scope.IncludedFields))
	postInclude := scope.IncludedFields[0]
	postScope := postInclude.Scope
	postScope.Value = []*Post{{ID: 1, Title: "Post title", Body: "Post body."}}

	currentPost := scope.IncludedFields[1]
	currentPost.Scope.Value = &Post{ID: 2, Title: "Current One", Body: "This is current post", LatestComment: &Comment{ID: 1}}

	latestComment := currentPost.Scope.IncludedFields[0]
	latestComment.Scope.Value = &Comment{ID: 1, Body: "This is such a great post", PostID: 2}

	err = scope.SetCollectionValues()
	assertNil(t, err)
	for scope.NextIncludedField() {
		includedField, err := scope.CurrentIncludedField()
		assertNil(t, err)
		_, err = includedField.GetMissingPrimaries()
		assertNil(t, err)
		err = includedField.Scope.SetCollectionValues()
		assertNil(t, err)
		t.Log(includedField.fieldName)
		for includedField.Scope.NextIncludedField() {
			includedField.Scope.CurrentIncludedField()
			nestedIncluded, err := includedField.Scope.CurrentIncludedField()
			assertNil(t, err)
			_, err = nestedIncluded.GetMissingPrimaries()
			assertNil(t, err)
			t.Log(nestedIncluded.fieldName)
			err = nestedIncluded.Scope.SetCollectionValues()
			assertNil(t, err)
		}
	}

	payload, err := marshalScope(scope, c)
	assertNoError(t, err)

	err = MarshalPayload(buf, payload)
	assertNoError(t, err)
	// even if included, there is no
	t.Log(buf.String())
	assertTrue(t, strings.Contains(buf.String(), "{\"type\":\"posts\",\"id\":\"1"))
	assertTrue(t, strings.Contains(buf.String(), "\"title\":\"My own title.\""))
	assertTrue(t, strings.Contains(buf.String(), "{\"type\":\"comments\",\"id\":\"1\",\"attributes\":{\"body\""))
	// assertTrue(t, strings.Contains(buf.String(), "\"relationships\":{\"posts\":{\"data\":[{\"type\":\"posts\",\"id\":\"1\"}]}}"))
	assertTrue(t, strings.Contains(buf.String(), "\"type\":\"blogs\",\"id\":\"3\""))
	clearMap()
	buf.Reset()

	scope = getBlogScope()
	errs = scope.buildIncludeList("current_post")
	assertEmpty(t, errs)
	scope.Value = &Blog{ID: 4, Title: "The title.", CreatedAt: time.Now(), CurrentPost: &Post{ID: 3}}
	errs = scope.buildFieldset("title", "created_at", "current_post")
	assertEmpty(t, errs)

	scope.IncludedScopes[c.MustGetModelStruct(&Post{})].Value = &Post{ID: 3, Title: "Breaking News!", Body: "Some body"}
	errs = scope.IncludedScopes[c.MustGetModelStruct(&Post{})].buildFieldset("title", "body")
	assertEmpty(t, errs)

	payload, err = marshalScope(scope, c)
	assertNil(t, err)
	err = MarshalPayload(buf, payload)
	assertNil(t, err)

	// assertTrue(t, strings.Contains(buf.String(),
	// "\"relationships\":{\"current_post\":{\"data\":{\"type\":\"posts\",\"id\":\"3\"}}}"))

	// t.Log(buf.String())
	assertTrue(t, strings.Contains(buf.String(),
		"\"type\":\"blogs\",\"id\":\"4\",\"attributes\":{\"created_at\":"))
	// assertTrue(t, strings.Contains(buf.String(),
	// "\"included\":[{\"type\":\"posts\",\"id\":\"3\",\"attributes\":{\"body\":\"Some body\",\"title\":\"Breaking News!\"}}]"))

	clearMap()
	buf.Reset()
	scope = getBlogScope()
	scope.Value = []*Blog{{ID: 4, Title: "The title one."}, {ID: 5, Title: "The title two"}}
	errs = scope.buildFieldset("title")
	assertEmpty(t, errs)

	payload, err = marshalScope(scope, c)
	assertNil(t, err)

	err = MarshalPayload(buf, payload)
	assertNil(t, err)

	assertTrue(t, strings.Contains(buf.String(), `"type":"blogs","id":"4","attributes":{"title":"The title one."}`))
	assertTrue(t, strings.Contains(buf.String(), `"type":"blogs","id":"5","attributes":{"title":"The title two"}`))

	// scope with no value
	clearMap()
	buf.Reset()
	scope = getBlogScope()

	payload, err = marshalScope(scope, c)
	assertError(t, err)

	err = MarshalPayload(buf, payload)
	assertNil(t, err)
}

func TestMarshalScopeRelationship(t *testing.T) {
	clearMap()
	getBlogScope()
	req := httptest.NewRequest("GET", "/blogs/1/relationships/posts", nil)
	scope, errs, err := c.BuildScopeRelationship(req, &Blog{})

	assertNil(t, err)
	assertEmpty(t, errs)

	scope.Value = &Blog{ID: 1, Posts: []*Post{{ID: 1}, {ID: 3}}}

	postsScope, err := scope.GetRelationshipScope()
	assertNil(t, err)

	payload, err := c.MarshalScope(postsScope)
	assertNil(t, err)

	buffer := bytes.NewBufferString("")

	err = MarshalPayload(buffer, payload)
	assertNil(t, err)

	t.Log(buffer)

}
