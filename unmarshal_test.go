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

	scope, errObj, err := unmarshalScopeOne(in, c)
	assertNil(t, err)
	assertNil(t, errObj)

	assertNotNil(t, scope)
	t.Log(scope.Value)

	in = strings.NewReader(`{"data":{"type":"materials","attributes":{"capacity":6102,"price":1234},"relationships":{"company":{"data":{"type":"material-companies","id":"52423045-80f1-468a-a6bd-763ce0f4c033"}},"storage":{"data":{"type":"material-storages","id":"1"}},"type":{"data":{"type":"material-types","id":"1"}}}}}`)
	scope, errObj, err = unmarshalScopeOne(in, c)
}
