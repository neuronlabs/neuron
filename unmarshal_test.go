package jsonapi

import (
	"strings"
	"testing"
)

func TestUnmarshalScope(t *testing.T) {

	in := strings.NewReader("{\"data\": {\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}}")
	clearMap()
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	scope, err := unmarshalScopeOne(in, c)
	assertNil(t, err)

	assertNotNil(t, scope)
	t.Log(scope.Value)
}
