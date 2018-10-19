package jsonapi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
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

				h.log.Debugf("Selected Fields: %#v", scope.SelectedFields)
				assert.NotEmpty(t, scope.PrimaryFilters)
				h.log.Debugf("Patch post repo: %#v", scope.Value)
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
		},
		"RelationHasMany": func(t *testing.T) {
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

		},
		"RelationManyToMany": func(t *testing.T) {
			t.Run("Sync", func(t *testing.T) {
				models := []interface{}{&PetSDK{}, &HumanSDK{}}
				h := prepareHandler(defaultLanguages, models...)
				for _, model := range h.ModelHandlers {

					model.Repository = &MockRepository{}
					model.Create = &Endpoint{Type: Create}
				}

				var (
					petID1          int = 3
					petID2          int = 4
					humanID         int = 2
					humanID1Before1 int = 10
					humanID1Before2 int = 12
					humanID2Before1 int = 15
				)

				rw, req := getHttpPair(
					"POST",
					fmt.Sprintf("/humans/%d", humanID),
					h.getModelJSON(&HumanSDK{PetsSync: []*PetSDK{{ID: petID1}, {ID: petID2}}}),
				)

				humanModel := h.ModelHandlers[reflect.TypeOf(HumanSDK{})]
				require.NotNil(t, humanModel)
				humanRepo, ok := humanModel.Repository.(*MockRepository)
				require.True(t, ok)

				humanRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().
					Return(nil).Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)
					human, ok := scope.Value.(*HumanSDK)
					require.True(t, ok)

					if assert.Len(t, human.PetsSync, 2) {
						pet1 := human.PetsSync[0]
						if assert.NotNil(t, pet1) {
							assert.Equal(t, petID1, pet1.ID, pet1)
						}
						pet2 := human.PetsSync[1]
						if assert.NotNil(t, pet2) {
							assert.Equal(t, petID2, pet2.ID, pet2)
						}
					}
				})

				petModel := h.ModelHandlers[reflect.TypeOf(PetSDK{})]
				require.NotNil(t, petModel)
				require.NotNil(t, petModel.Repository)

				petRepo, ok := petModel.Repository.(*MockRepository)
				require.True(t, ok)

				petRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)
					_, ok := scope.Value.([]*PetSDK)
					require.True(t, ok)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						prim := scope.PrimaryFilters[0]
						if assert.NotNil(t, prim) {
							if assert.Len(t, prim.Values, 1) {
								fv := prim.Values[0]
								if assert.NotNil(t, fv) && assert.Len(t, fv.Values, 2) {
									assert.Equal(t, petID1, fv.Values[0])
									assert.Equal(t, petID2, fv.Values[1])
								}
							}
						}
					}
					scope.Value = []*PetSDK{
						{ID: petID1, HumansSync: []*HumanSDK{{ID: humanID1Before1}, {ID: humanID1Before2}}},
						{ID: petID2, HumansSync: []*HumanSDK{{ID: humanID2Before1}}},
					}
				})

				petRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Return(nil).Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)
					pet, ok := scope.Value.(*PetSDK)
					require.True(t, ok, reflect.TypeOf(scope.Value).String())

					if assert.NotNil(t, pet) {
						switch pet.ID {
						case petID1:
							if assert.Len(t, pet.HumansSync, 3) {
								if assert.NotNil(t, pet.HumansSync[0]) {
									assert.Equal(t, pet.HumansSync[0].ID, humanID1Before1)
								}
								if assert.NotNil(t, pet.HumansSync[1]) {
									assert.Equal(t, pet.HumansSync[1].ID, humanID1Before2)
								}
								if assert.NotNil(t, pet.HumansSync[2]) {
									assert.Equal(t, pet.HumansSync[2].ID, humanID)
								}
							}
						case petID2:
							if assert.Len(t, pet.HumansSync, 2) {
								if assert.NotNil(t, pet.HumansSync[0]) {
									assert.Equal(t, pet.HumansSync[0].ID, humanID2Before1)
								}
								if assert.NotNil(t, pet.HumansSync[1]) {
									assert.Equal(t, pet.HumansSync[1].ID, humanID)
								}
							}
						default:
							t.FailNow()
						}

					}
				})
				if assert.NotPanics(t, func() { h.Patch(humanModel, humanModel.Patch).ServeHTTP(rw, req) }) {
					assert.Equal(t, 200, rw.Code, rw.Body.String())
				}
				// })

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
