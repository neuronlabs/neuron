package jsonapi

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestUnmarshalScopeOne(t *testing.T) {

	clearMap()
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	require.Nil(t, err)

	// Case 1:
	// Correct with  attributes
	t.Run("valid_attributes", func(t *testing.T) {
		in := strings.NewReader("{\"data\": {\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}}")
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.NoError(t, err)
		assert.NotNil(t, scope)
	})

	// Case 2
	// Walid with relationships and attributes

	t.Run("valid_rel_attrs", func(t *testing.T) {
		in := strings.NewReader(`{
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

		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assertNoError(t, err)
		assertNotNil(t, scope)
	})

	// Case 3:
	// Invalid document - no opening bracket.
	t.Run("invalid_document", func(t *testing.T) {
		in := strings.NewReader(`"data":{"type":"blogs","id":"1"}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, strconv.Itoa(http.StatusBadRequest), errObj.Status)
				assert.Equal(t, ErrInvalidJSONDocument.Title, errObj.Title)
			}
		}

	})

	// Case 3 :
	// Invalid collection - unrecognized collection
	t.Run("invalid_collection", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"unrecognized","id":"1"}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidResourceName.ID, errObj.ID)
			}
		}

	})

	// Case 4:
	// Invalid syntax - syntax error
	t.Run("invalid_syntax", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1",}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONDocument.ID, errObj.ID)
			}
		}

	})

	// Case 5:
	// Invalid Field - unrecognized field
	t.Run("invalid_field_value", func(t *testing.T) {
		// number instead of string
		in := strings.NewReader(`{"data":{"type":"blogs","id":1.03}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}
		}

	})

	t.Run("invalid_relationship_type", func(t *testing.T) {
		// string instead of object
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":"invalid"}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}

		}
	})

	// array
	t.Run("invalid_id_value_array", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":{"1":"2"}}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}
		}
	})

	// array
	t.Run("invalid_relationship_value_array", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":["invalid"]}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}
		}
	})

	// bool
	t.Run("invalid_relationship_value_bool", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":true}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}
		}
	})

	// Case 6:
	// invalid field value within i.e. for attribute
	t.Run("invalid_attribute_value", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "attributes":{"title":1.02}}}`)
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONFieldValue.ID, errObj.ID)
			}
		}
	})

	t.Run("invalid_field_strict_mode", func(t *testing.T) {
		// title attribute is missspelled as 'Atitle'
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "attributes":{"Atitle":1.02}}}`)
		c.StrictUnmarshalMode = true
		defer func() {
			c.StrictUnmarshalMode = false
		}()
		scope, err := c.UnmarshalScopeOne(in, &Blog{}, false)
		assert.Nil(t, scope)
		if assert.NotNil(t, err) {
			errObj, ok := err.(*ErrorObject)
			if assert.True(t, ok) {
				assert.Equal(t, ErrInvalidJSONDocument.ID, errObj.ID)
			}
		}
	})
}

func TestUnmarshalScopeMany(t *testing.T) {
	clearMap()

	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	require.Nil(t, err)

	// Case 1:
	// Correct with  attributes
	t.Run("valid_attributes", func(t *testing.T) {
		in := strings.NewReader("{\"data\": [{\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}]}")
		scope, err := c.UnmarshalScopeMany(in, &Blog{})
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			assert.NotEmpty(t, scope.Value)
		}

	})

	// Case 2
	// Walid with relationships and attributes

	t.Run("valid_rel_attrs", func(t *testing.T) {
		in := strings.NewReader(`{
		"data":[
			{
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
		]
	}`)

		scope, err := c.UnmarshalScopeMany(in, &Blog{})
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			assert.NotEmpty(t, scope.Value)
			v := scope.Value.([]*Blog)
			for _, blog := range v {
				t.Logf("Scope value: %#v", blog)
				t.Logf("CurrentPost: %v", blog.CurrentPost)
			}

		}
	})

}

func TestUnmarshalUpdateFields(t *testing.T) {
	clearMap()
	assertNil(t, c.PrecomputeModels(&Blog{}, &Post{}, &Comment{}))

	buf := bytes.NewBuffer(nil)

	t.Run("attribute", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data":{"type":"blogs","id":"1", 	"attributes":{"title":"New title"}}}`)
		scope, err := c.UnmarshalScopeOne(buf, &Blog{}, true)
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			assert.Contains(t, scope.SelectedFields, scope.Struct.attributes["title"])
			assert.Len(t, scope.SelectedFields, 1)
		}

	})

	t.Run("multiple-attributes", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data":{"type":"blogs","id":"1", "attributes":{"title":"New title","view_count":16}}}`)

		scope, err := c.UnmarshalScopeOne(buf, &Blog{}, true)
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Contains(t, scope.SelectedFields, mStruct.attributes["title"])
				assert.Contains(t, scope.SelectedFields, mStruct.attributes["view_count"])
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
					"id": "3"
				}
			}
		}
	}
}`)

		scope, err := c.UnmarshalScopeOne(buf, &Blog{}, true)
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.SelectedFields, 1)
				assert.Contains(t, scope.SelectedFields, mStruct.relationships["current_post"])
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
						"id": "3"
					},
					{
						"type":"posts",
						"id": "4"
					}
				]
			}
		}
	}
}`)

		scope, err := c.UnmarshalScopeOne(buf, &Blog{}, true)
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.SelectedFields, 1)
				assert.Contains(t, scope.SelectedFields, mStruct.relationships["posts"])
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
					"id": "3"
				}
			},
			"posts":{
				"data": [
					{
						"type":"posts",
						"id": "3"
					}
				]
			}
		}
	}
}`)

		scope, err := c.UnmarshalScopeOne(buf, &Blog{}, true)
		assert.NoError(t, err)
		if assert.NotNil(t, scope) {
			if assert.Equal(t, "blogs", scope.Struct.collectionType) {
				mStruct := scope.Struct
				assert.Len(t, scope.SelectedFields, 3)
				assert.Contains(t, scope.SelectedFields, mStruct.attributes["title"])
				assert.Contains(t, scope.SelectedFields, mStruct.relationships["current_post"])
				assert.Contains(t, scope.SelectedFields, mStruct.relationships["posts"])
			}
		}
	})
}
