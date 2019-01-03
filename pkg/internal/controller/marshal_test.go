package controller

import (
	"bytes"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	buf := bytes.Buffer{}

	prepare := func(t *testing.T, models ...interface{}) *Controller {
		t.Helper()
		c := DefaultTesting()

		buf.Reset()
		require.NoError(t, c.RegisterModels(models...))
		return c
	}

	prepareBlogs := func(t *testing.T) *Controller {
		return prepare(t, &internal.Blog{}, &internal.Post{}, &internal.Comment{})
	}

	tests := map[string]func(*testing.T){
		"single": func(t *testing.T) {
			c := prepareBlogs(t)

			value := &internal.Blog{ID: 5, Title: "My title", ViewCount: 14}
			if assert.NoError(t, c.Marshal(&buf, value)) {
				marshaled := buf.String()
				assert.Contains(t, marshaled, `"title":"My title"`)
				assert.Contains(t, marshaled, `"view_count":14`)
				assert.Contains(t, marshaled, `"id":"5"`)
			}
		},
		"Time": func(t *testing.T) {
			type ModelPtrTime struct {
				ID   int        `jsonapi:"type=primary"`
				Time *time.Time `jsonapi:"type=attr"`
			}

			type ModelTime struct {
				ID   int       `jsonapi:"type=primary"`
				Time time.Time `jsonapi:"type=attr"`
			}

			t.Run("NoPtr", func(t *testing.T) {
				c := prepare(t, &ModelTime{})
				now := time.Now()
				v := &ModelTime{ID: 5, Time: now}
				if assert.NoError(t, c.Marshal(&buf, v)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
				}
			})

			t.Run("Ptr", func(t *testing.T) {
				c := prepare(t, &ModelPtrTime{})
				now := time.Now()
				v := &ModelPtrTime{ID: 5, Time: &now}
				if assert.NoError(t, c.Marshal(&buf, v)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
				}
			})

		},
		"singleWithMap": func(t *testing.T) {
			t.Run("PtrString", func(t *testing.T) {
				type MpString struct {
					ID  int                `jsonapi:"type=primary"`
					Map map[string]*string `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpString{})

				kv := "some"
				value := &MpString{ID: 5, Map: map[string]*string{"key": &kv}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":"some"}`)
				}
			})

			t.Run("NilString", func(t *testing.T) {
				type MpString struct {
					ID  int                `jsonapi:"type=primary"`
					Map map[string]*string `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpString{})
				value := &MpString{ID: 5, Map: map[string]*string{"key": nil}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})

			t.Run("PtrInt", func(t *testing.T) {
				type MpInt struct {
					ID  int             `jsonapi:"type=primary"`
					Map map[string]*int `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpInt{})

				kv := 5
				value := &MpInt{ID: 5, Map: map[string]*int{"key": &kv}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":5}`)
				}
			})
			t.Run("NilPtrInt", func(t *testing.T) {
				type MpInt struct {
					ID  int             `jsonapi:"type=primary"`
					Map map[string]*int `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpInt{})

				value := &MpInt{ID: 5, Map: map[string]*int{"key": nil}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})
			t.Run("PtrFloat", func(t *testing.T) {
				type MpFloat struct {
					ID  int                 `jsonapi:"type=primary"`
					Map map[string]*float64 `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpFloat{})

				fv := 1.214
				value := &MpFloat{ID: 5, Map: map[string]*float64{"key": &fv}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":1.214}`)
				}
			})
			t.Run("NilPtrFloat", func(t *testing.T) {
				type MpFloat struct {
					ID  int                 `jsonapi:"type=primary"`
					Map map[string]*float64 `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpFloat{})

				value := &MpFloat{ID: 5, Map: map[string]*float64{"key": nil}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})

			t.Run("SliceInt", func(t *testing.T) {
				type MpSliceInt struct {
					ID  int              `jsonapi:"type=primary"`
					Map map[string][]int `jsonapi:"type=attr"`
				}
				c := prepare(t, &MpSliceInt{})

				value := &MpSliceInt{ID: 5, Map: map[string][]int{"key": {1, 5}}}
				if assert.NoError(t, c.Marshal(&buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":[1,5]}`)
				}
			})

		},
		"many": func(t *testing.T) {
			c := prepareBlogs(t)

			values := []*internal.Blog{{ID: 5, Title: "First"}, {ID: 2, Title: "Second"}}
			if assert.NoError(t, c.Marshal(&buf, values)) {
				marshaled := buf.String()
				assert.Contains(t, marshaled, `"title":"First"`)
				assert.Contains(t, marshaled, `"title":"Second"`)

				assert.Contains(t, marshaled, `"id":"5"`)
				assert.Contains(t, marshaled, `"id":"2"`)
				t.Log(marshaled)
			}
		},
		"Nested": func(t *testing.T) {

			t.Run("Simple", func(t *testing.T) {
				type NestedSub struct {
					First int
				}

				type Simple struct {
					ID     int        `jsonapi:"type=primary"`
					Nested *NestedSub `jsonapi:"type=attr"`
				}

				c := prepare(t, &Simple{})

				err := c.Marshal(&buf, &Simple{ID: 2, Nested: &NestedSub{First: 1}})
				if assert.NoError(t, err) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"nested":{"first":1}`)
				}
			})

			t.Run("DoubleNested", func(t *testing.T) {

				type NestedSub struct {
					First int
				}

				type DoubleNested struct {
					Nested *NestedSub
				}

				type Simple struct {
					ID     int           `jsonapi:"type=primary"`
					Double *DoubleNested `jsonapi:"type=attr"`
				}

				c := prepare(t, &Simple{})

				err := c.Marshal(&buf, &Simple{ID: 2, Double: &DoubleNested{Nested: &NestedSub{First: 1}}})
				if assert.NoError(t, err) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"nested":{"first":1}`)
					assert.Contains(t, marshaled, `"double":{"nested"`)
				}
			})

		},
	}

	for name, testFunc := range tests {

		t.Run(name, testFunc)
	}

}

func TestMarshalScope(t *testing.T) {

	buf := bytes.NewBufferString("")

	c := BlogController(t)

	// req := httptest.NewRequest("GET", `/blogs/3?include=posts,current_post.latest_comment&fields[blogs]=title,created_at,posts&fields[posts]=title,body,comments`, nil)
	// scope, errs, err := c.BuildScopeSingle(req, &Endpoint{Type: Get}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	// assert.Nil(t, err)
	// assertEmpty(t, errs)

	t.Log("Creating RootScope")
	scope, err := c.NewRootScope(&internal.Blog{})
	require.NoError(t, err)

	c.log().Debugf("Prepared")

	scope.Value = &internal.Blog{ID: 3, Title: "My own title.", CreatedAt: time.Now(), Posts: []*internal.Post{{ID: 1}}, CurrentPost: &internal.Post{ID: 2}}

	errs := scope.BuildIncludeList("posts", "current_post.latest_comment")
	require.Empty(t, errs)

	c.log().Debugf("Built include list")

	// assertEqual(t, 1, len(scope.IncludedFields))
	postInclude := scope.IncludedFields()[0]
	postScope := postInclude.Scope
	postScope.Value = []*internal.Post{{ID: 1, Title: "Post title", Body: "Post body."}}

	currentPost := scope.IncludedFields()[1]
	currentPost.Scope.Value = &internal.Post{ID: 2, Title: "Current One", Body: "This is current post", LatestComment: &internal.Comment{ID: 1}}

	latestComment := currentPost.Scope.IncludedFields()[0]
	latestComment.Scope.Value = &internal.Comment{ID: 1, Body: "This is such a great post", PostID: 2}

	c.log().Debugf("SettingCollectionValues")
	err = scope.SetCollectionValues()
	assert.Nil(t, err)

	c.log().Debugf("Setting collection values")

	for scope.NextIncludedField() {
		includedField, err := scope.CurrentIncludedField()
		assert.Nil(t, err)
		_, err = includedField.GetMissingPrimaries()
		assert.Nil(t, err)
		err = includedField.Scope.SetCollectionValues()
		assert.Nil(t, err)

		c.log().Debugf("Outer")
		for includedField.Scope.NextIncludedField() {
			nestedIncluded, err := includedField.Scope.CurrentIncludedField()
			c.log().Debugf("Inner")
			assert.Nil(t, err)
			_, err = nestedIncluded.GetMissingPrimaries()
			assert.Nil(t, err)

			err = nestedIncluded.Scope.SetCollectionValues()
			assert.Nil(t, err)
		}
	}

	c.log().Debugf("Preparing to marshal")
	payload, err := marshalScope(scope, c)
	assert.NoError(t, err)

	err = MarshalPayload(buf, payload)
	assert.NoError(t, err)
	// even if included, there is no

	assert.True(t, strings.Contains(buf.String(), "{\"type\":\"posts\",\"id\":\"1"))
	assert.True(t, strings.Contains(buf.String(), "\"title\":\"My own title.\""))
	assert.True(t, strings.Contains(buf.String(), "{\"type\":\"comments\",\"id\":\"1\",\"attributes\":{\"body\""))
	// assert.True(t, strings.Contains(buf.String(), "\"relationships\":{\"posts\":{\"data\":[{\"type\":\"posts\",\"id\":\"1\"}]}}"))
	assert.True(t, strings.Contains(buf.String(), "\"type\":\"blogs\",\"id\":\"3\""))

	c = BlogController(t)
	buf.Reset()

	scope = BlogScope(t, c)

	errs = scope.BuildIncludeList("current_post")
	assert.Empty(t, errs)
	scope.Value = &internal.Blog{ID: 4, Title: "The title.", CreatedAt: time.Now(), CurrentPost: &internal.Post{ID: 3}}
	errs = scope.BuildFieldset("title", "created_at", "current_post")
	assert.Empty(t, errs)

	postScope, ok := scope.IncludeScopeByStruct(c.MustGetModelStruct(&internal.Post{}))
	if assert.True(t, ok) {
		postScope.Value = &internal.Post{ID: 3, Title: "Breaking News!", Body: "Some body"}
	}

	errs = postScope.BuildFieldset("title", "body")
	assert.Empty(t, errs)

	payload, err = marshalScope(scope, c)
	assert.Nil(t, err)
	err = MarshalPayload(buf, payload)
	assert.Nil(t, err)

	// assert.True(t, strings.Contains(buf.String(),
	// "\"relationships\":{\"current_post\":{\"data\":{\"type\":\"posts\",\"id\":\"3\"}}}"))

	// t.Log(buf.String())
	assert.True(t, strings.Contains(buf.String(),
		"\"type\":\"blogs\",\"id\":\"4\",\"attributes\":{\"created_at\":"))
	// assert.True(t, strings.Contains(buf.String(),
	// "\"included\":[{\"type\":\"posts\",\"id\":\"3\",\"attributes\":{\"body\":\"Some body\",\"title\":\"Breaking News!\"}}]"))

	c = BlogController(t)
	buf.Reset()

	scope = BlogScope(t, c)
	scope.Value = []*internal.Blog{{ID: 4, Title: "The title one."}, {ID: 5, Title: "The title two"}}
	errs = scope.BuildFieldset("title")
	assert.Empty(t, errs)

	payload, err = marshalScope(scope, c)
	assert.Nil(t, err)

	err = MarshalPayload(buf, payload)
	assert.Nil(t, err)

	assert.True(t, strings.Contains(buf.String(), `"type":"blogs","id":"4","attributes":{"title":"The title one."}`))
	assert.True(t, strings.Contains(buf.String(), `"type":"blogs","id":"5","attributes":{"title":"The title two"}`))

	// scope with no value
	c = BlogController(t)
	buf.Reset()
	scope = BlogScope(t, c)

	payload, err = marshalScope(scope, c)
	assert.Error(t, err)

	err = MarshalPayload(buf, payload)
	assert.Nil(t, err)

	t.Run("MarshalToManyRelationship", func(t *testing.T) {
		c := DefaultTesting()
		require.NoError(t, c.RegisterModels(&internal.Pet{}, &internal.User{}))

		scope, err := c.NewScope(&internal.Pet{})
		require.NoError(t, err)

		scope.Value = &internal.Pet{ID: 5, Owners: []*internal.User{{ID: 2}, {ID: 3}}}

		err = scope.SetFields("Owners")
		assert.NoError(t, err)

		payload, err := c.MarshalScope(scope)
		if assert.NoError(t, err) {
			single, ok := payload.(*OnePayload)
			if assert.True(t, ok) {
				if assert.NotNil(t, single.Data) {
					if assert.NotEmpty(t, single.Data.Relationships) {
						if assert.NotNil(t, single.Data.Relationships["owners"]) {
							owners, ok := single.Data.Relationships["owners"].(*RelationshipManyNode)
							if assert.True(t, ok) {
								var count int
								for _, owner := range owners.Data {
									if assert.NotNil(t, owner) {
										switch owner.ID {
										case "2", "3":
											count += 1
										}
									}
								}
								assert.Equal(t, 2, count)
							}

						}
					}
				}
			}
		}

	})

	t.Run("MarshalToManyEmptyRelationship", func(t *testing.T) {
		c := DefaultTesting()
		require.NoError(t, c.RegisterModels(&internal.Pet{}, &internal.User{}))

		scope, err := c.NewScope(&internal.Pet{})
		require.NoError(t, err)

		scope.Value = &internal.Pet{ID: 5, Owners: []*internal.User{}}
		scope.SetFields("Owners")

		payload, err := c.MarshalScope(scope)
		if assert.NoError(t, err) {
			single, ok := payload.(*OnePayload)
			if assert.True(t, ok) {
				if assert.NotNil(t, single.Data) {
					if assert.NotEmpty(t, single.Data.Relationships) {
						if assert.NotNil(t, single.Data.Relationships["owners"]) {
							owners, ok := single.Data.Relationships["owners"].(*RelationshipManyNode)
							if assert.True(t, ok, reflect.TypeOf(single.Data.Relationships["owners"]).String()) {
								if assert.NotNil(t, owners) {
									assert.Empty(t, owners.Data)
								}
							}

						}
					}
				}
				buf := bytes.Buffer{}
				assert.NoError(t, MarshalPayload(&buf, single))
				assert.Contains(t, buf.String(), "owners")
			}
		}

	})
}

func BlogController(t *testing.T) *Controller {
	c := DefaultTesting()
	err := c.RegisterModels(&internal.Blog{}, &internal.Post{}, &internal.Comment{})
	require.NoError(t, err)
	return c
}

func BlogScope(t *testing.T, c *Controller) *query.Scope {
	scope, err := c.NewScope(&internal.Blog{})
	require.NoError(t, err)
	return scope
}

// func TestMarshalScopeRelationship(t *testing.T) {
// 	c := BlogController(t)

// 	scope := BlogScope(t, c)

// 	req := httptest.NewRequest("GET", "/blogs/1/relationships/posts", nil)
// 	scope, errs, err := c.BuildScopeRelationship(req, &Endpoint{Type: GetRelationship}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})

// 	assert.Nil(t, err)
// 	assert.Empty(t, errs)

// 	scope.Value = &internal.Blog{ID: 1, Posts: []*Post{{ID: 1}, {ID: 3}}}

// 	postsScope, err := scope.GetRelationshipScope()
// 	assert.Nil(t, err)

// 	payload, err := c.MarshalScope(postsScope)
// 	assert.Nil(t, err)

// 	buffer := bytes.NewBufferString("")

// 	err = MarshalPayload(buffer, payload)
// 	assert.Nil(t, err)

// }

type HiddenModel struct {
	ID          int    `jsonapi:"type=primary;flags=hidden"`
	Visibile    string `jsonapi:"type=attr"`
	HiddenField string `jsonapi:"type=attr;flags=hidden"`
}

func (h *HiddenModel) CollectionName() string {
	return "hiddens"
}

func TestMarshalHiddenScope(t *testing.T) {

	c := DefaultTesting()
	assert.NoError(t, c.RegisterModels(&HiddenModel{}))

	scope, err := c.NewScope(&HiddenModel{})
	assert.NoError(t, err)

	scope.Value = &HiddenModel{ID: 1, Visibile: "Visible", HiddenField: "Invisible"}

	payload, err := c.MarshalScope(scope)
	assert.NoError(t, err)

	buffer := bytes.NewBufferString("")
	err = MarshalPayload(buffer, payload)
	assert.NoError(t, err)

}
