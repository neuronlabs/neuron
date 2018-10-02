package jsonapi

import (
	"bytes"
	"github.com/stretchr/testify/assert"
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

func TestUnmarshalUpdate(t *testing.T) {
	clearMap()
	assertNil(t, c.PrecomputeModels(&Blog{}, &Post{}, &Comment{}))

	buf := bytes.NewBuffer(nil)

	t.Run("attribute", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data":{"type":"blogs","id":"1", 	"attributes":{"title":"New title"}}}`)
		scope, errObj, err := c.unmarshalUpdate(buf)
		assert.NoError(t, err)
		assert.Nil(t, errObj)
		if assert.NotNil(t, scope) {
			assert.Contains(t, scope.UpdatedFields, scope.Struct.attributes["title"])
			assert.Len(t, scope.UpdatedFields, 1)
		}

	})

	t.Run("multiple-attributes", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data":{"type":"blogs","id":"1", "attributes":{"title":"New title","view_count":16}}}`)

		scope, errObj, err := c.unmarshalUpdate(buf)
		assert.NoError(t, err)
		assert.Nil(t, errObj)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Contains(t, scope.UpdatedFields, mStruct.attributes["title"])
				assert.Contains(t, scope.UpdatedFields, mStruct.attributes["view_count"])
			}
		}
	})

	t.Run("relationship-to-one", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`
{
	"data":	{
		"type":"blogs",
		"id":"1",
		"relationships":{
			"current_post":{
				"data": {
					"type":"posts",
					"id": 3
				}
			}
		}
	}
}`)

		scope, errObj, err := c.unmarshalUpdate(buf)
		assert.NoError(t, err)
		assert.Nil(t, errObj)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.UpdatedFields, 1)
				assert.Contains(t, scope.UpdatedFields, mStruct.relationships["current_post"])
			}
		}
	})

	t.Run("relationship-to-many", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`
{
	"data":	{
		"type":"blogs",
		"id":"1",
		"relationships":{
			"posts":{
				"data": [
					{
						"type":"posts",
						"id": 3
					},
					{
						"type":"posts",
						"id": 4
					}
				]
			}
		}
	}
}`)

		scope, errObj, err := c.unmarshalUpdate(buf)
		assert.NoError(t, err)
		assert.Nil(t, errObj)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.UpdatedFields, 1)
				assert.Contains(t, scope.UpdatedFields, mStruct.relationships["posts"])
			}
		}
	})

	t.Run("mixed", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`
{
	"data":	{
		"type":"blogs",
		"id":"1",
		"attributes":{
			"title":"mixed"			
		},
		"relationships":{		
			"current_post":{
				"data": {
					"type":"posts",
					"id": 3
				}
			},
			"posts":{
				"data": [
					{
						"type":"posts",
						"id": 3
					}
				]
			}
		}
	}
}`)

		scope, errObj, err := c.unmarshalUpdate(buf)
		assert.NoError(t, err)
		assert.Nil(t, errObj)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.UpdatedFields, 3)
				assert.Contains(t, scope.UpdatedFields, mStruct.attributes["title"])
				assert.Contains(t, scope.UpdatedFields, mStruct.relationships["current_post"])
				assert.Contains(t, scope.UpdatedFields, mStruct.relationships["posts"])
			}
		}
	})
}
