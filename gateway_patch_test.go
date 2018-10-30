package jsonapi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestHandlerPatch(t *testing.T) {
	getHandler := func() *Handler {
		h := prepareHandler(defaultLanguages, blogModels...)
		for _, model := range h.ModelHandlers {
			model.Patch = &Endpoint{}
			model.Repository = &MockRepository{}
		}
		return h
	}

	getModelAndRepo := func(
		t *testing.T,
		h *Handler,
		model reflect.Type,
	) (*ModelHandler, *MockRepository) {
		t.Helper()
		mh, ok := h.ModelHandlers[model]
		require.True(t, ok)

		require.NotNil(t, mh.Repository)
		mhRepo, ok := mh.Repository.(*MockRepository)
		require.True(t, ok)
		return mh, mhRepo
	}

	tests := map[string]func(*testing.T){
		"NoRelations": func(t *testing.T) {
			h := getHandler()

			blog := &BlogSDK{ID: 1, SomeAttr: "Some attribute"}
			rw, req := getHttpPair("PATCH", "/blogs/1", h.getModelJSON(blog))

			model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

			repo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
				scope, ok := args.Get(0).(*Scope)
				require.True(t, ok)

				blogPatched, ok := scope.Value.(*BlogSDK)
				require.True(t, ok)

				assert.Len(t, scope.SelectedFields, 2)

				assert.Equal(t, blog.ID, blogPatched.ID)
				assert.Equal(t, blog.SomeAttr, blogPatched.SomeAttr)
			})

			h.Patch(model, model.Patch).ServeHTTP(rw, req)

			if assert.Equal(t, 200, rw.Code, rw.Body.String()) {

				h.log.Debugf("Success: %v", rw.Body.String())
				payload, err := h.Controller.unmarshalPayload(rw.Body)
				if assert.NoError(t, err) {

					if assert.NotNil(t, payload.Data) {
						assert.Equal(t, strconv.Itoa(blog.ID), payload.Data.ID)
					}

				}
			}
		},
		"RelationBelongsTo": func(t *testing.T) {
			h := getHandler()

			post := &PostSDK{ID: 4}
			comment := &CommentSDK{ID: 1, Post: post}
			rw, req := getHttpPair("PATCH", "/comments/1", h.getModelJSON(comment))

			model, repo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))

			repo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
				scope, ok := args.Get(0).(*Scope)
				require.True(t, ok)

				commentPatched, ok := scope.Value.(*CommentSDK)
				require.True(t, ok)

				assert.Len(t, scope.SelectedFields, 2)

				assert.Equal(t, comment.ID, commentPatched.ID)
				if assert.NotNil(t, commentPatched.Post) {
					assert.Equal(t, post.ID, commentPatched.PostID)
				}
			})

			h.Patch(model, model.Patch).ServeHTTP(rw, req)

			if assert.Equal(t, 200, rw.Code, rw.Body.String()) {

				h.log.Debugf("Success: %v", rw.Body.String())
				commentUnmarshaled := &CommentSDK{}
				err := h.Controller.Unmarshal(rw.Body, commentUnmarshaled)
				if assert.NoError(t, err) {
					assert.Equal(t, comment.ID, commentUnmarshaled.ID)
					if assert.NotNil(t, commentUnmarshaled.Post) {
						assert.Equal(t, post.ID, commentUnmarshaled.Post.ID)
					}
				}
			}
		},
		"RelationHasOne": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Run("SetNew", func(t *testing.T) {
					h := getHandler()

					post := &PostSDK{ID: 4}
					blog := &BlogSDK{ID: 2, CurrentPost: post}
					rw, req := getHttpPair("PATCH", "/blogs/2", h.getModelJSON(blog))

					model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

					repo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						blogPatched, ok := scope.Value.(*BlogSDK)
						require.True(t, ok)

						assert.Len(t, scope.SelectedFields, 2)

						assert.Equal(t, blog.ID, blogPatched.ID)
					})

					_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						post, ok := scope.Value.(*PostSDK)
						require.True(t, ok)
						require.NotNil(t, post)

						// primary should be as the root.relation.id
						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv := scope.PrimaryFilters[0].Values[0]
								assert.Contains(t, fv.Values, blog.CurrentPost.ID)
								assert.Equal(t, fv.Operator, OpIn)
							}
						}

						// Foreign Key should be set to the root id
						assert.Equal(t, post.BlogID, blog.ID)

						// Selected Fields should contain the foreign key
						blogID, ok := scope.Struct.FieldByName("BlogID")
						require.True(t, ok)

						assert.Contains(t, scope.SelectedFields, blogID)
						assert.Len(t, scope.SelectedFields, 1)

					})

					h.Patch(model, model.Patch).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						blogUnm := &BlogSDK{}
						err := h.Controller.Unmarshal(rw.Body, blogUnm)
						if assert.NoError(t, err) {
							assert.Equal(t, blog.ID, blogUnm.ID)
							if assert.NotNil(t, blogUnm.CurrentPost) {
								assert.Equal(t, post.ID, blogUnm.CurrentPost.ID)
							}
						}
					}
				})

				t.Run("SetNil", func(t *testing.T) {
					// t.Logf("Sets the synced HasOne relation value to nil.")
					h := getHandler()

					blogID := 2
					rw, req := getHttpPair("PATCH", "/blogs/2", strings.NewReader(`{"data":{"type":"blogs","id":"2","relationships":{"current_post": {"data": null}}}}`))

					model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

					repo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						blogPatched, ok := scope.Value.(*BlogSDK)
						require.True(t, ok)

						assert.Len(t, scope.SelectedFields, 2)

						assert.Equal(t, blogID, blogPatched.ID)
						assert.Nil(t, blogPatched.CurrentPost)
					})

					_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					var postPatched bool
					postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						postPatched = true

						blogIDField, ok := scope.Struct.FieldByName("BlogID")
						require.True(t, ok)

						assert.Contains(t, scope.SelectedFields, blogIDField)
						assert.Len(t, scope.SelectedFields, 1)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.Equal(t, scope.ForeignKeyFilters[0].StructField, blogIDField) {
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									fv := scope.ForeignKeyFilters[0].Values[0]
									assert.Contains(t, fv.Values, blogID)
									assert.Equal(t, fv.Operator, OpEqual)
								}
							}

						}
						h.log.Debugf("Patch post repo: %#v", scope.Value)
					})

					h.Patch(model, model.Patch).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						blogUnm := &BlogSDK{}
						assert.True(t, postPatched)
						err := h.Controller.Unmarshal(rw.Body, blogUnm)
						if assert.NoError(t, err) {
							assert.Equal(t, blogID, blogUnm.ID)
							assert.Nil(t, blogUnm.CurrentPost)

						}
					}
				})

			})

			t.Run("NonSynced", func(t *testing.T) {
				t.Run("AddRelation", func(t *testing.T) {
					h := getHandler()

					post := &PostSDK{ID: 4}
					blog := &BlogSDK{ID: 2, CurrentPostNoSync: post}
					rw, req := getHttpPair("PATCH", "/blogs/2", h.getModelJSON(blog))

					model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))
					repo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						blogPatched, ok := scope.Value.(*BlogSDK)
						require.True(t, ok)

						assert.Len(t, scope.SelectedFields, 2)
						assert.Equal(t, blog.ID, blogPatched.ID)
					})

					h.Patch(model, model.Patch).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						blogUnm := &BlogSDK{}
						err := h.Controller.Unmarshal(rw.Body, blogUnm)
						if assert.NoError(t, err) {
							assert.Equal(t, blog.ID, blogUnm.ID)
							if assert.NotNil(t, blogUnm.CurrentPostNoSync) {
								assert.Equal(t, post.ID, blogUnm.CurrentPostNoSync.ID)
							}
						}
					}
				})

			})

		},
		"RelationHasMany": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Run("SetNew", func(t *testing.T) {
					h := getHandler()
					var commentID1, commentID2 int = 5, 16
					post := &PostSDK{ID: 2, Comments: []*CommentSDK{{ID: commentID1}, {ID: commentID2}}}
					rw, req := getHttpPair("PATCH", "/posts/2", h.getModelJSON(post))

					postModel, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						postInside, ok := scope.Value.(*PostSDK)
						require.True(t, ok)

						assert.Equal(t, post.ID, postInside.ID)
					})

					_, commentsRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))

					// at first commentsRepo should delete all existing comments containing
					// foreign key equals to the post.ID
					commentsRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						comment, ok := scope.Value.(*CommentSDK)
						require.True(t, ok)

						assert.Zero(t, comment.PostID)

						postIDField, ok := scope.Struct.FieldByName("PostID")
						require.True(t, ok)

						// Check Filter
						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							assert.Equal(t, postIDField, scope.ForeignKeyFilters[0].StructField)
							if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
								fv := scope.ForeignKeyFilters[0].Values[0]
								assert.Equal(t, OpEqual, fv.Operator)
								assert.Contains(t, fv.Values, post.ID)
							}
						}

						// Check Selected Fields
						assert.Contains(t, scope.SelectedFields, postIDField)
					})

					// then the commentsRepo should set the foreign key for all comments
					// with id as in Comments field.
					commentsRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().
						Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, commentID1)
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, commentID2)
									}
								}
							}
						}
					})
					h.Patch(postModel, postModel.Patch).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						h.log.Debugf("%s", rw.Body.String())
						postUnm := &PostSDK{}
						err := h.Controller.Unmarshal(rw.Body, postUnm)
						if assert.NoError(t, err) {
							assert.Equal(t, post.ID, postUnm.ID)
							ctr := 0
							for _, com := range postUnm.Comments {
								switch com.ID {
								case commentID1, commentID2:
									ctr++
								}
							}

							assert.Equal(t, 2, ctr)
						}
					}
				})

				t.Run("SetEmpty", func(t *testing.T) {
					h := getHandler()
					post := &PostSDK{ID: 2}
					rw, req := getHttpPair("PATCH", "/posts/2", strings.NewReader(`{"data":{"type":"posts","id":"2","relationships":{"comments":{"data":[]}}}}`))

					postModel, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						// postPatch
						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv := scope.PrimaryFilters[0].Values[0]
								assert.Equal(t, fv.Operator, OpEqual)
								assert.Contains(t, fv.Values, 2)
							}
						}

						postInside, ok := scope.Value.(*PostSDK)
						require.True(t, ok)

						assert.Equal(t, post.ID, postInside.ID)
					})

					_, commentsRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))

					// Patch on the relationship repository shuold find all relationships for
					// the foreign keys equal to the root scope primary
					commentsRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().
						Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						postID, ok := scope.Struct.FieldByName("PostID")
						require.True(t, ok)

						// comments should be patched by it's foreign key
						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							assert.Equal(t, scope.ForeignKeyFilters[0].StructField, postID)
							if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
								fv := scope.ForeignKeyFilters[0].Values[0]
								assert.Equal(t, OpEqual, fv.Operator)
								assert.Contains(t, fv.Values, post.ID)
							}
						}
					})

					h.Patch(postModel, postModel.Patch).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						h.log.Debugf("%s", rw.Body.String())
						postUnm := &PostSDK{}
						err := h.Controller.Unmarshal(rw.Body, postUnm)
						if assert.NoError(t, err) {
							assert.Equal(t, post.ID, postUnm.ID)
							assert.Empty(t, postUnm.Comments)
						}
					}
				})
			})

		},
		"RelationManyToMany": func(t *testing.T) {
			t.Run("Sync", func(t *testing.T) {
				t.Run("SetNew", func(t *testing.T) {
					models := []interface{}{&PetSDK{}, &HumanSDK{}}
					h := prepareHandler(defaultLanguages, models...)
					for _, model := range h.ModelHandlers {

						model.Repository = &MockRepository{}
						model.Patch = &Endpoint{Type: Patch}
					}

					var (
						petID         int = 3
						humanID1      int = 2
						humanID2      int = 5
						petID1Before1     = 4
					)

					rw, req := getHttpPair(
						"PATCH",
						fmt.Sprintf("/pets/%d", petID),
						h.getModelJSON(&PetSDK{ID: petID, HumansSync: []*HumanSDK{{ID: humanID1}, {ID: humanID2}}}),
					)

					petModel, petRepo := getModelAndRepo(t, h, reflect.TypeOf(PetSDK{}))

					// While patching the model with synced many2many relationship
					// The related field should be patched with
					petRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).
						Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						pet, ok := scope.Value.(*PetSDK)
						require.True(t, ok)

						humansSyncField, ok := scope.Struct.FieldByName("HumansSync")
						require.True(t, ok)

						assert.Contains(t, scope.SelectedFields, humansSyncField)
						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv := scope.PrimaryFilters[0].Values[0]
								assert.Equal(t, OpEqual, fv.Operator)
								assert.Contains(t, fv.Values, petID)
							}
						}

						if assert.Len(t, pet.HumansSync, 2) {
							human1 := pet.HumansSync[0]
							if assert.NotNil(t, human1) {
								assert.Equal(t, humanID1, human1.ID)
							}
							human2 := pet.HumansSync[1]
							if assert.NotNil(t, human2) {
								assert.Equal(t, humanID2, human2.ID)
							}
						}
					})

					// The related repository should patch the backreference field
					_, humanRepo := getModelAndRepo(t, h, reflect.TypeOf(HumanSDK{}))

					// At first clear the relationships not included into relation field.
					humanRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						human, ok := scope.Value.(*HumanSDK)
						require.True(t, ok)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv := scope.PrimaryFilters[0].Values[0]
								assert.Equal(t, OpNotIn, fv.Operator)
								assert.Contains(t, fv.Values, humanID1)
								assert.Contains(t, fv.Values, humanID2)
							}
						}

						petsField, ok := scope.Struct.FieldByName("Pets")
						require.True(t, ok)
						if assert.Len(t, scope.RelationshipFilters, 1) {
							assert.Equal(t, petsField, scope.RelationshipFilters[0].StructField)
							if assert.Len(t, scope.RelationshipFilters[0].Relationships, 1) {
								assert.Equal(t, petsField.relatedStruct.primary, scope.RelationshipFilters[0].Relationships[0].StructField)
								if assert.Len(t, scope.RelationshipFilters[0].Relationships[0].Values, 1) {
									filterValue := scope.RelationshipFilters[0].Relationships[0].Values[0]
									assert.Equal(t, OpEqual, filterValue.Operator)
									assert.Contains(t, filterValue.Values, petID)

								}
							}
						}
						assert.Contains(t, scope.SelectedFields, petsField)
						assert.Zero(t, human.Pets)
					})

					// At first it should get list of related pets and it's relationships
					humanRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						_, ok := scope.Value.([]*HumanSDK)
						require.True(t, ok)

						// petsField, ok := scope.Struct.FieldByName("Pets")
						// require.True(t, ok)

						assert.Contains(t, scope.Fieldset, "pets")
						if assert.Len(t, scope.PrimaryFilters, 1) {
							prim := scope.PrimaryFilters[0]
							if assert.NotNil(t, prim) {
								if assert.Len(t, prim.Values, 1) {
									fv := prim.Values[0]
									if assert.NotNil(t, fv) && assert.Len(t, fv.Values, 2) {
										assert.True(t,
											assert.Contains(t, fv.Values, humanID1) &&
												assert.Contains(t, fv.Values, humanID2),
										)
									}
								}
							}
						}

						scope.Value = []*HumanSDK{
							{ID: humanID1, Pets: []*PetSDK{{ID: petID1Before1}}},
							{ID: humanID2, Pets: []*PetSDK{}},
						}
					})

					var called int
					humanRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

						human, ok := scope.Value.(*HumanSDK)
						require.True(t, ok, reflect.TypeOf(scope.Value).String())

						petsField, ok := scope.Struct.FieldByName("Pets")
						require.True(t, ok)
						called += 1
						h.log.Debugf("Patched with value: %v", human)
						var fv *FilterValues
						assert.Contains(t, scope.SelectedFields, petsField)
						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv = scope.PrimaryFilters[0].Values[0]
							}
						}
						if assert.NotNil(t, human) {
							switch human.ID {
							case humanID1:
								if assert.NotNil(t, fv) {
									assert.Equal(t, OpEqual, fv.Operator)
									assert.Contains(t, fv.Values, humanID1)
								}
								if assert.Len(t, human.Pets, 2) {
									// assert contains the one before and the new one
									var count int
									for _, pet := range human.Pets {
										switch pet.ID {
										case petID, petID1Before1:
											count++
										}
									}
									assert.Equal(t, 2, count)
								}
							case humanID2:
								if assert.NotNil(t, fv) {
									assert.Equal(t, OpEqual, fv.Operator)
									assert.Contains(t, fv.Values, humanID2)
								}

								if assert.Len(t, human.Pets, 1) {
									// assert contains the one before and the new one
									var count int
									for _, pet := range human.Pets {
										switch pet.ID {
										case petID:
											count++
										}
									}
									assert.Equal(t, 1, count)
								}
							default:
								assert.Fail(t, "Invalid humanID")
							}
						}
					})

					// humanRepo.AssertNumberOfCalls(t, "Patch", 2)

					// if assert.NotPanics(t, func() { h.Patch(petModel, petModel.Patch).ServeHTTP(rw, req) }) {
					h.Patch(petModel, petModel.Patch).ServeHTTP(rw, req)
					assert.Equal(t, 200, rw.Code, rw.Body.String())
					assert.Equal(t, 2, called)
					// }
				})

				t.Run("SetEmpty", func(t *testing.T) {
					models := []interface{}{&PetSDK{}, &HumanSDK{}}
					h := prepareHandler(defaultLanguages, models...)
					for _, model := range h.ModelHandlers {

						model.Repository = &MockRepository{}
						model.Patch = &Endpoint{Type: Patch}
					}

					var (
						petID int = 3
					)

					rw, req := getHttpPair(
						"PATCH",
						fmt.Sprintf("/pets/%d", petID),
						strings.NewReader(`{"data":{"type":"pets","id":"3","relationships":{"humans_sync":{"data":[]}}}}`),
					)

					petModel, petRepo := getModelAndRepo(t, h, reflect.TypeOf(PetSDK{}))

					// While patching the model with synced many2many relationship
					// The related field should be patched with
					petRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).
						Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						pet, ok := scope.Value.(*PetSDK)
						require.True(t, ok)

						humansSyncField, ok := scope.Struct.FieldByName("HumansSync")
						require.True(t, ok)

						assert.Contains(t, scope.SelectedFields, humansSyncField)
						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								fv := scope.PrimaryFilters[0].Values[0]
								assert.Equal(t, OpEqual, fv.Operator)
								assert.Contains(t, fv.Values, petID)
							}
						}

						assert.Len(t, pet.HumansSync, 0)

					})

					// The related repository should patch the backreference field
					_, humanRepo := getModelAndRepo(t, h, reflect.TypeOf(HumanSDK{}))

					humanRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						human, ok := scope.Value.(*HumanSDK)
						require.True(t, ok)

						assert.Equal(t, "humans", scope.Struct.collectionType)
						petsField, ok := scope.Struct.FieldByName("Pets")
						require.True(t, ok)

						assert.Contains(t, scope.SelectedFields, petsField)

						if assert.Len(t, scope.RelationshipFilters, 1) {
							assert.Equal(t, petsField, scope.RelationshipFilters[0].StructField)
							if assert.Len(t, scope.RelationshipFilters[0].Relationships, 1) {
								relFilter := scope.RelationshipFilters[0].Relationships[0]
								if assert.Len(t, relFilter.Values, 1) {
									fv := relFilter.Values[0]
									assert.Equal(t, OpEqual, fv.Operator)
									assert.Contains(t, fv.Values, petID)
								}
							}
						}
						assert.Zero(t, human.Pets)

					})

					// if assert.NotPanics(t, func() { h.Patch(petModel, petModel.Patch).ServeHTTP(rw, req) }) {
					h.Patch(petModel, petModel.Patch).ServeHTTP(rw, req)
					assert.Equal(t, 200, rw.Code, rw.Body.String())

					humanRepo.AssertCalled(t, "Patch", mock.Anything)
					// }
				})
			})
		},
		"NonMatchedID": func(t *testing.T) {

			h := getHandler()
			post := &PostSDK{ID: 1, Title: "Something"}

			rw, req := getHttpPair("Patch", "/posts/4", h.getModelJSON(post))

			postModel, _ := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

			if assert.NotPanics(t, func() { h.Patch(postModel, postModel.Patch).ServeHTTP(rw, req) }) {
				assert.Equal(t, 400, rw.Code)
			}
		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}
	// h := prepareHandler(defaultLanguages, blogModels...)
	// mockRepo := &MockRepository{}
	// h.SetDefaultRepo(mockRepo)

	// blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	// blogHandler.Patch = &Endpoint{Type: Patch}

	// // Case 1:
	// // Correctly Patched
	// rw, req := getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{Lang: "en", CurrentPost: &PostSDK{ID: 2}}))

	// mockRepo.On("Patch", mock.Anything).Once().Return(nil)
	// h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// // Case 2:
	// // Patch with preset value
	// presetPair := h.Controller.BuildPrecheckPair("preset=blogs.current_post&filter[blogs][id][eq]=3", "filter[comments][post][id][in]")

	// commentModel := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
	// commentModel.Patch = &Endpoint{Type: Patch, PresetPairs: []*PresetPair{presetPair}}
	// assert.NotNil(t, commentModel, "Nil comment model.")

	// // commentModel.AddPrecheckPair(presetPair, Patch)

	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = []*BlogSDK{{ID: 3, CurrentPost: &PostSDK{ID: 4}}}
	// 		t.Log("First")
	// 	})

	// // mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// // 	func(args mock.Arguments) {
	// // 		arg := args.Get(0).(*Scope)
	// // 		arg.Value = []*PostSDK{{ID: 4}}
	// // 		t.Log("Second")
	// // 	})

	// mockRepo.On("Patch", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		t.Logf("%+v", arg)
	// 		t.Logf("Scope value: %+v\n\n", arg.Value)
	// 		// t.Log(arg.PrimaryFilters[0].Values[0].Values)
	// 		// assert.NotEmpty(t, arg.RelationshipFilters)
	// 		// assert.NotEmpty(t, arg.RelationshipFilters[0].Relationships)
	// 		// assert.NotEmpty(t, arg.RelationshipFilters[0].Relationships[0].Values)
	// 		// t.Log(arg.RelationshipFilters[0].Relationships[0].Values[0].Values)
	// 		// t.Log(arg.RelationshipFilters[0].Relationships[0].Values[0].Operator)
	// 	})

	// rw, req = getHttpPair("PATCH", "/comments/1", h.getModelJSON(&CommentSDK{Body: "Some body."}))
	// h.Patch(commentModel, commentModel.Patch).ServeHTTP(rw, req)

	// assert.Equal(t, http.StatusOK, rw.Result().StatusCode)

	// commentModel.Patch.PrecheckPairs = nil

	// // Case 3:
	// // Incorrect URL for ID provided - internal
	// rw, req = getHttpPair("PATCH", "/blogs", h.getModelJSON(&BlogSDK{}))
	// h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 4:
	// // // No language provided - user error
	// // rw, req = getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{CurrentPost: &PostSDK{ID: 2}}))
	// // h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	// // assert.Equal(t, 400, rw.Result().StatusCode)

	// // Case 5:
	// // Repository error
	// rw, req = getHttpPair("PATCH", "/blogs/1", h.getModelJSON(&BlogSDK{Lang: "pl", CurrentPost: &PostSDK{ID: 2}}))
	// mockRepo.On("Patch", mock.Anything).Once().Return(unidb.ErrForeignKeyViolation.New())
	// h.Patch(blogHandler, blogHandler.Patch).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// // Case 6:
	// // Preset some hidden value

	// presetPair = h.Controller.BuildPresetPair("preset=blogs.current_post&filter[blogs][id][eq]=4", "filter[comments][post][id]")
	// commentModel.AddPresetPair(presetPair, Patch)
	// rw, req = getHttpPair("PATCH", "/comments/1", h.getModelJSON(&CommentSDK{Body: "Preset post"}))

	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		scope := args.Get(0).(*Scope)
	// 		scope.Value = []*BlogSDK{{ID: 4}}
	// 	})

	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		scope := args.Get(0).(*Scope)
	// 		scope.Value = []*PostSDK{{ID: 3}}
	// 	})

	// mockRepo.On("Patch", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		scope := args.Get(0).(*Scope)
	// 		t.Log(scope.Value)
	// 	})
	// h.Patch(commentModel, commentModel.Patch).ServeHTTP(rw, req)

}
