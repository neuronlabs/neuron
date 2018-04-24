package jsonapi

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMarshalScope(t *testing.T) {
	scope := getBlogScope()
	buf := bytes.NewBufferString("")
	errs := scope.buildIncludedScopes("posts")
	assertEmpty(t, errs)
	scope.Value = &Blog{ID: 3, Title: "My own title.", CreatedAt: time.Now(), Posts: []*Post{{ID: 1}}}

	scope.Fields = append(scope.Fields, scope.Struct.attributes["title"], scope.Struct.attributes["created_at"], scope.Struct.relationships["posts"])
	postScope := scope.SubScopes[0]
	postScope.Value = []*Post{{ID: 1, Title: "Post title", Body: "Post body."}}
	postScope.setWorkingFields("title", "body", "comments", "latest_comment")

	payload, err := marshalScope(scope, c)
	assertNoError(t, err)

	err = MarshalPayload(buf, payload)
	assertNoError(t, err)
	// even if included, there is no
	assertTrue(t, strings.Contains(buf.String(), "\"included\":[{\"type\":\"posts\",\"id\":\"1\",\"attributes\":{"))
	assertTrue(t, strings.Contains(buf.String(), "\"title\":\"My own title.\""))
	// assertTrue(t, strings.Contains(buf.String(), "\"relationships\":{\"posts\":{\"data\":[{\"type\":\"posts\",\"id\":\"1\"}]}}"))
	assertTrue(t, strings.Contains(buf.String(), "\"id\":\"3\""))
	clearMap()
	buf.Reset()

	scope = getBlogScope()
	errs = scope.buildIncludedScopes("current-post")
	assertEmpty(t, errs)
	scope.Value = &Blog{ID: 4, Title: "The title.", CreatedAt: time.Now(), CurrentPost: &Post{ID: 3}}
	errs = scope.setWorkingFields("title", "created_at", "current-post")
	assertEmpty(t, errs)

	scope.SubScopes[0].Value = &Post{ID: 3, Title: "Breaking News!", Body: "Some body"}
	errs = scope.SubScopes[0].setWorkingFields("title", "body")
	assertEmpty(t, errs)

	payload, err = marshalScope(scope, c)
	assertNil(t, err)
	err = MarshalPayload(buf, payload)
	assertNil(t, err)

	// assertTrue(t, strings.Contains(buf.String(),
	// "\"relationships\":{\"current-post\":{\"data\":{\"type\":\"posts\",\"id\":\"3\"}}}"))

	t.Log(buf.String())
	assertTrue(t, strings.Contains(buf.String(),
		"\"type\":\"blogs\",\"id\":\"4\",\"attributes\":{\"created_at\":"))
	// assertTrue(t, strings.Contains(buf.String(),
	// "\"included\":[{\"type\":\"posts\",\"id\":\"3\",\"attributes\":{\"body\":\"Some body\",\"title\":\"Breaking News!\"}}]"))

	clearMap()
	buf.Reset()
	scope = getBlogScope()
	scope.Value = []*Blog{{ID: 4, Title: "The title one."}, {ID: 5, Title: "The title two"}}
	errs = scope.setWorkingFields("title")
	assertEmpty(t, errs)

	payload, err = marshalScope(scope, c)
	assertNil(t, err)

	err = MarshalPayload(buf, payload)
	assertNil(t, err)

	// assertTrue(t, strings.Contains(buf.String(), `{"data":[{"type":"blogs","id":"4","attributes":{"title":"The title one."}},{"type":"blogs","id":"5","attributes":{"title":"The title two"}}]}`))

	// t.Log(buf.String())

	// scope with no value
	clearMap()
	buf.Reset()
	scope = getBlogScope()

	payload, err = marshalScope(scope, c)
	assertNil(t, err)

	err = MarshalPayload(buf, payload)
	assertNil(t, err)

	t.Log(buf.String())

}
