package jsonapi

import (
	"bytes"
	"context"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

import (
	// mocks import and register mock repository
	_ "github.com/neuronlabs/neuron/query/scope/mocks"
)

func TestMarshal(t *testing.T) {
	buf := bytes.Buffer{}

	prepare := func(t *testing.T, models ...interface{}) *ctrl.Controller {
		t.Helper()
		c := controller.DefaultTesting(t, nil)

		buf.Reset()
		require.NoError(t, c.RegisterModels(models...))
		return (*ctrl.Controller)(c)
	}

	prepareBlogs := func(t *testing.T) *ctrl.Controller {
		return prepare(t, &internal.Blog{}, &internal.Post{}, &internal.Comment{})
	}

	tests := map[string]func(*testing.T){
		"single": func(t *testing.T) {
			c := prepareBlogs(t)

			value := &internal.Blog{ID: 5, Title: "My title", ViewCount: 14}
			if assert.NoError(t, MarshalC(c, &buf, value)) {
				marshaled := buf.String()
				assert.Contains(t, marshaled, `"title":"My title"`)
				assert.Contains(t, marshaled, `"view_count":14`)
				assert.Contains(t, marshaled, `"id":"5"`)
			}
		},
		"Time": func(t *testing.T) {
			type ModelPtrTime struct {
				ID   int        `neuron:"type=primary"`
				Time *time.Time `neuron:"type=attr"`
			}

			type ModelTime struct {
				ID   int       `neuron:"type=primary"`
				Time time.Time `neuron:"type=attr"`
			}

			t.Run("NoPtr", func(t *testing.T) {
				c := prepare(t, &ModelTime{})
				now := time.Now()
				v := &ModelTime{ID: 5, Time: now}
				if assert.NoError(t, MarshalC(c, &buf, v)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
				}
			})

			t.Run("Ptr", func(t *testing.T) {
				c := prepare(t, &ModelPtrTime{})
				now := time.Now()
				v := &ModelPtrTime{ID: 5, Time: &now}
				if assert.NoError(t, MarshalC(c, &buf, v)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
				}
			})

		},
		"singleWithMap": func(t *testing.T) {
			t.Run("PtrString", func(t *testing.T) {
				type MpString struct {
					ID  int                `neuron:"type=primary"`
					Map map[string]*string `neuron:"type=attr"`
				}
				c := prepare(t, &MpString{})

				kv := "some"
				value := &MpString{ID: 5, Map: map[string]*string{"key": &kv}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":"some"}`)
				}
			})

			t.Run("NilString", func(t *testing.T) {
				type MpString struct {
					ID  int                `neuron:"type=primary"`
					Map map[string]*string `neuron:"type=attr"`
				}
				c := prepare(t, &MpString{})
				value := &MpString{ID: 5, Map: map[string]*string{"key": nil}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})

			t.Run("PtrInt", func(t *testing.T) {
				type MpInt struct {
					ID  int             `neuron:"type=primary"`
					Map map[string]*int `neuron:"type=attr"`
				}
				c := prepare(t, &MpInt{})

				kv := 5
				value := &MpInt{ID: 5, Map: map[string]*int{"key": &kv}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":5}`)
				}
			})
			t.Run("NilPtrInt", func(t *testing.T) {
				type MpInt struct {
					ID  int             `neuron:"type=primary"`
					Map map[string]*int `neuron:"type=attr"`
				}
				c := prepare(t, &MpInt{})

				value := &MpInt{ID: 5, Map: map[string]*int{"key": nil}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})
			t.Run("PtrFloat", func(t *testing.T) {
				type MpFloat struct {
					ID  int                 `neuron:"type=primary"`
					Map map[string]*float64 `neuron:"type=attr"`
				}
				c := prepare(t, &MpFloat{})

				fv := 1.214
				value := &MpFloat{ID: 5, Map: map[string]*float64{"key": &fv}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":1.214}`)
				}
			})
			t.Run("NilPtrFloat", func(t *testing.T) {
				type MpFloat struct {
					ID  int                 `neuron:"type=primary"`
					Map map[string]*float64 `neuron:"type=attr"`
				}
				c := prepare(t, &MpFloat{})

				value := &MpFloat{ID: 5, Map: map[string]*float64{"key": nil}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":null}`)
				}
			})

			t.Run("SliceInt", func(t *testing.T) {
				type MpSliceInt struct {
					ID  int              `neuron:"type=primary"`
					Map map[string][]int `neuron:"type=attr"`
				}
				c := prepare(t, &MpSliceInt{})

				value := &MpSliceInt{ID: 5, Map: map[string][]int{"key": {1, 5}}}
				if assert.NoError(t, MarshalC(c, &buf, value)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, `"map":{"key":[1,5]}`)
				}
			})

		},
		"many": func(t *testing.T) {
			c := prepareBlogs(t)

			values := []*internal.Blog{{ID: 5, Title: "First"}, {ID: 2, Title: "Second"}}
			if assert.NoError(t, MarshalC(c, &buf, &values)) {
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
					ID     int        `neuron:"type=primary"`
					Nested *NestedSub `neuron:"type=attr"`
				}

				c := prepare(t, &Simple{})

				err := MarshalC(c, &buf, &Simple{ID: 2, Nested: &NestedSub{First: 1}})
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
					ID     int           `neuron:"type=primary"`
					Double *DoubleNested `neuron:"type=attr"`
				}

				c := prepare(t, &Simple{})

				err := MarshalC(c, &buf, &Simple{ID: 2, Double: &DoubleNested{Nested: &NestedSub{First: 1}}})
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

	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}
	buf := bytes.NewBufferString("")

	ctx := context.Background()
	c := BlogController(t)

	t.Run("Included", func(t *testing.T) {

		u, err := url.Parse("/blogs/3?include=posts,current_post.latest_comment&fields[blogs]=title,created_at,posts&fields[posts]=title,body,comments")
		require.NoError(t, err)

		s, errs, err := (*controller.Controller)(c).QueryBuilder().BuildScopeSingle(ctx, &internal.Blog{}, u, 3)

		if assert.NoError(t, err) && assert.Empty(t, errs) && assert.NotNil(t, s) {

			s.Value = &internal.Blog{ID: 3, Title: "My own title.", CreatedAt: time.Now(), Posts: []*internal.Post{{ID: 1}}, CurrentPost: &internal.Post{ID: 2}}

			includes := s.IncludedFields()

			if assert.Len(t, includes, 2) {
				// assertEqual(t, 1, len(scope.IncludedFields))
				postInclude := s.IncludedFields()[0]
				postScope := postInclude.Scope
				postScope.Value = &([]*internal.Post{{ID: 1, Title: "Post title", Body: "Post body."}})

				currentPost := s.IncludedFields()[1]
				currentPost.Scope.Value = &internal.Post{ID: 2, Title: "Current One", Body: "This is current post", LatestComment: &internal.Comment{ID: 1}}

				latestComment := currentPost.Scope.IncludedFields()[0]
				latestComment.Scope.Value = &internal.Comment{ID: 1, Body: "This is such a great post", PostID: 2}

				log.Debugf("SettingCollectionValues")
				err = s.SetCollectionValues()
				if assert.Nil(t, err) {

					log.Debugf("Setting collection values")

					for s.NextIncludedField() {
						includedField, err := s.CurrentIncludedField()
						if assert.Nil(t, err) {
							_, err = includedField.GetMissingPrimaries()
							if assert.Nil(t, err) {
								err = includedField.Scope.SetCollectionValues()
								if assert.Nil(t, err) {

									log.Debugf("Outer")
									for includedField.Scope.NextIncludedField() {
										nestedIncluded, err := includedField.Scope.CurrentIncludedField()
										log.Debugf("Inner")
										assert.Nil(t, err)
										_, err = nestedIncluded.GetMissingPrimaries()
										assert.Nil(t, err)

										err = nestedIncluded.Scope.SetCollectionValues()
										assert.Nil(t, err)
									}
								}
							}
						}

					}
				}
			}
		}

		log.Debugf("Preparing to marshal")
		payload, err := marshalScope((*controller.Controller)(c), s)
		assert.NoError(t, err)

		err = marshalPayload(buf, payload)
		assert.NoError(t, err)
		// even if included, there is no

		assert.True(t, strings.Contains(buf.String(), "{\"type\":\"posts\",\"id\":\"1"))
		assert.True(t, strings.Contains(buf.String(), "\"title\":\"My own title.\""))
		assert.True(t, strings.Contains(buf.String(), "{\"type\":\"comments\",\"id\":\"1\",\"attributes\":{\"body\""))
		// assert.True(t, strings.Contains(buf.String(), "\"relationships\":{\"posts\":{\"data\":[{\"type\":\"posts\",\"id\":\"1\"}]}}"))
		assert.True(t, strings.Contains(buf.String(), "\"type\":\"blogs\",\"id\":\"3\""))

	})

	t.Run("MarshalToManyRelationship", func(t *testing.T) {
		c := controller.DefaultTesting(t, nil)
		require.NoError(t, c.RegisterModels(&internal.Pet{}, &internal.User{}))

		s, err := (*controller.Controller)(c).QueryBuilder().NewScope(&internal.Pet{})
		require.NoError(t, err)

		s.Value = &internal.Pet{ID: 5, Owners: []*internal.User{{ID: 2}, {ID: 3}}}

		err = s.SetFields("Owners")
		assert.NoError(t, err)

		payload, err := marshalScope((*controller.Controller)(c), s)
		if assert.NoError(t, err) {
			single, ok := payload.(*onePayload)
			if assert.True(t, ok) {
				if assert.NotNil(t, single.Data) {
					if assert.NotEmpty(t, single.Data.Relationships) {
						if assert.NotNil(t, single.Data.Relationships["owners"]) {
							owners, ok := single.Data.Relationships["owners"].(*relationshipManyNode)
							if assert.True(t, ok) {
								var count int
								for _, owner := range owners.Data {
									if assert.NotNil(t, owner) {
										switch owner.ID {
										case "2", "3":
											count++
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
		c := controller.DefaultTesting(t, nil)

		require.NoError(t, c.RegisterModels(&internal.Pet{}, &internal.User{}))

		s, err := (*controller.Controller)(c).QueryBuilder().NewScope(&internal.Pet{})
		require.NoError(t, err)

		s.Value = &internal.Pet{ID: 5, Owners: []*internal.User{}}
		s.SetFields("Owners")

		payload, err := marshalScope(c, s)
		if assert.NoError(t, err) {
			single, ok := payload.(*onePayload)
			if assert.True(t, ok) {
				if assert.NotNil(t, single.Data) {
					if assert.NotEmpty(t, single.Data.Relationships) {
						if assert.NotNil(t, single.Data.Relationships["owners"]) {
							owners, ok := single.Data.Relationships["owners"].(*relationshipManyNode)
							if assert.True(t, ok, reflect.TypeOf(single.Data.Relationships["owners"]).String()) {
								if assert.NotNil(t, owners) {
									assert.Empty(t, owners.Data)
								}
							}

						}
					}
				}
				buf := bytes.Buffer{}
				assert.NoError(t, marshalPayload(&buf, single))
				assert.Contains(t, buf.String(), "owners")
			}
		}

	})
}

func BlogController(t *testing.T) *controller.Controller {
	v := testing.Verbose()
	internal.Verbose = &v

	c := controller.DefaultTesting(t, nil)

	err := c.RegisterModels(&internal.Blog{}, &internal.Post{}, &internal.Comment{})
	require.NoError(t, err)
	return c
}

func BlogScope(t *testing.T, c *controller.Controller) *scope.Scope {
	scope, err := c.QueryBuilder().NewScope(&internal.Blog{})
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
	ID          int    `neuron:"type=primary;flags=hidden"`
	Visibile    string `neuron:"type=attr"`
	HiddenField string `neuron:"type=attr;flags=hidden"`
}

func (h *HiddenModel) CollectionName() string {
	return "hiddens"
}

func TestMarshalHiddenScope(t *testing.T) {

	c := controller.DefaultTesting(t, nil)
	assert.NoError(t, c.RegisterModels(&HiddenModel{}))

	scope, err := c.QueryBuilder().NewScope(&HiddenModel{})
	assert.NoError(t, err)

	scope.Value = &HiddenModel{ID: 1, Visibile: "Visible", HiddenField: "Invisible"}

	payload, err := marshalScope(c, scope)
	assert.NoError(t, err)

	buffer := bytes.NewBufferString("")
	err = marshalPayload(buffer, payload)
	assert.NoError(t, err)

}
