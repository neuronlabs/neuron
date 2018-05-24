package jsonapi

import (
	"strings"
	"testing"
)

func TestUnmarshalScope(t *testing.T) {

	clearMap()
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	// Case 1:
	// Correct input
	in := strings.NewReader("{\"data\": {\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}}")
	scope, errObj, err := unmarshalScopeOne(in, c)
	assertNil(t, err)
	assertNil(t, errObj)

	assertNotNil(t, scope)

	// Case 1:
	// Correct with relationship and attributes
	in = strings.NewReader(`{
		"data":{
			"type":"blogs",
			"id":"2",
			"attributes": {
				"title":"Correct Unmarshal"
			},
			"relationships":{
				"current_post":{
					"data":{
						"type":"posts",
						"id":"2"
					}					
				}
			}
		}
	}`)
	scope, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNil(t, errObj)
	assertNotNil(t, scope)

	// Case 2:
	// Invalid document - no opening bracket.
	in = strings.NewReader(`"data":{"type":"blogs","id":"1"}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNil(t, err)
	assertNotNil(t, errObj)

	// Case 3 :
	// Invalid input - unrecognized collection
	in = strings.NewReader(`{"data":{"type":"unrecognized","id":"1"}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNil(t, err)
	assertNotNil(t, errObj)

	// Case 4:
	// Invalid syntax - syntax error
	in = strings.NewReader(`{"data":{"type":"blogs","id":"1",}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNil(t, err)
	assertNotNil(t, errObj)

	// Case 5:
	// Invalid Field - unrecognized field

	// number instead of string
	in = strings.NewReader(`{"data":{"type":"blogs","id":1.03}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)

	// string instead of object
	in = strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":"invalid"}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)

	t.Log(errObj)

	// array
	in = strings.NewReader(`{"data":{"type":"blogs","id":{"1":"2"}}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)
	t.Log(errObj)

	// array
	in = strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":["invalid"]}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)

	// bool
	in = strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":true}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)

	// Case 6:
	// invalid field value within i.e. for attribute
	in = strings.NewReader(`{"data":{"type":"blogs","id":"1", "attributes":{"title":1.02}}}`)
	_, errObj, err = unmarshalScopeOne(in, c)
	assertNoError(t, err)
	assertNotNil(t, errObj)

}
