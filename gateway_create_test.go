package jsonapi

import (
	"context"
	"github.com/google/uuid"
	"github.com/kucjac/uni-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestHandlerCreate(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {

			t.Errorf("Recovered in HandlerCreate: %v", r)
		}
	}()

	innerModels := []interface{}{&ModCliGenID{}, &PetSDK{}, &HumanSDK{}}
	h := prepareHandler(defaultLanguages, append(blogModels, innerModels...)...)
	mockRepo := &MockRepository{}
	for _, model := range h.ModelHandlers {

		model.Repository = &MockRepository{}
		model.Create = &Endpoint{Type: Create}
	}
	h.SetDefaultRepo(mockRepo)

	tests := map[string]func(t *testing.T){
		"Success": func(t *testing.T) {
			rw, req := getHttpPair("POST", "/blogs", strings.NewReader(`{"data":{"type":"blogs","attributes": {"lang":"pl"}}}`))

			mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			mockRepo, ok := mh.Repository.(*MockRepository)
			require.True(t, ok)
			mockRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
				func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)
					blog, ok := scope.Value.(*BlogSDK)
					require.True(t, ok)
					blog.ID = 5
				})
			h.Create(mh, mh.Create).ServeHTTP(rw, req)
			assert.Equal(t, 201, rw.Result().StatusCode, rw.Body.String())
		},
		"DuplicatedValue": func(t *testing.T) {
			// Case 2:
			// Duplicated value
			rw, req := getHttpPair("POST", "/blogs", h.getModelJSON(&BlogSDK{Lang: "pl", CurrentPost: &PostSDK{ID: 1}}))
			mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			mockRepo, ok := mh.Repository.(*MockRepository)
			require.True(t, ok)
			mockRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(unidb.ErrUniqueViolation.New())

			h.Create(mh, mh.Create).ServeHTTP(rw, req)
			assert.Equal(t, 409, rw.Result().StatusCode, rw.Body.String())
		},
		"InvalidCollectionName": func(t *testing.T) {
			rw, req := getHttpPair("POST", "/blogs", strings.NewReader(`{"data":{"type":"unknown_collection"}}`))
			mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			h.Create(mh, mh.Create).ServeHTTP(rw, req)
			assert.Equal(t, 400, rw.Result().StatusCode)
		},
		"PresetValue": func(t *testing.T) {
			t.Skip("Preset Values are not ready to test")
			commentModel, ok := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
			assert.True(t, ok, "Model not valid")

			key := "my_key"
			presetPair := h.Controller.BuildPresetPair("preset=blogs.current_post", "filter[comments][post][id][in]").WithKey(key)

			assert.NoError(t, commentModel.AddPresetPair(presetPair, Create))
			mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			mockRepo, ok := mh.Repository.(*MockRepository)
			require.True(t, ok)
			mockRepo.On("List", mock.Anything).Once().Return(nil)

			// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
			// 	func(args mock.Arguments) {
			// 		arg := args.Get(0).(*Scope)
			// 		arg.Value = []*BlogSDK{{ID: 6}}
			// 	})

			// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
			// 	func(args mock.Arguments) {
			// 		arg := args.Get(0).(*Scope)
			// 		arg.Value = []*PostSDK{{ID: 6}}
			// 	})

			// // mockRepo.On("Create", mock.Anything).Once().Return(nil).Run(
			// // 	func(args mock.Arguments) {
			// // 		arg := args.Get(0).(*Scope)
			// // 		t.Logf("Create value: \"%+v\"", arg.Value)
			// // 		comment := arg.Value.(*CommentSDK)
			// // 		t.Logf("PostSDK: %v", comment.Post)
			// // 		comment.ID = 3

			// // 	})

			rw, req := getHttpPair("POST", "/comments", strings.NewReader(`{"data":{"type":"comments","attributes":{"body":"Some title"}}}`))

			req = req.WithContext(context.WithValue(req.Context(), key, 6))
			h.Create(commentModel, commentModel.Create).ServeHTTP(rw, req)
			assert.Equal(t, 201, rw.Result().StatusCode, rw.Body.String())
		},

		"ClientID": func(t *testing.T) {
			mod := h.ModelHandlers[reflect.TypeOf(ModCliGenID{})]
			mockRepo, ok := mod.Repository.(*MockRepository)
			require.True(t, ok)
			t.Run("UUID", func(t *testing.T) {

				rw, req := getHttpPair("POST", "/client-generated", h.getModelJSON(&ModCliGenID{ID: uuid.New().String()}))
				mockRepo.On("Create", mock.Anything).Return(nil)
				h.Create(mod, mod.Create).ServeHTTP(rw, req)

				assert.Equal(t, 201, rw.Result().StatusCode, rw.Body.String())
			})

			t.Run("NonUUID", func(t *testing.T) {
				rw, req := getHttpPair("POST", "/client-generated", h.getModelJSON(&ModCliGenID{ID: "SomeID"}))
				mockRepo.On("Create", mock.Anything).Return(nil)
				h.Create(mod, mod.Create).ServeHTTP(rw, req)
				assert.Equal(t, 400, rw.Result().StatusCode, rw.Body.String())
			})

			t.Run("non-matched field value id", func(t *testing.T) {
				rw, req := getHttpPair("POST", "/client-generated", strings.NewReader(`{"data":{"type":"client-generated","id": 15}}`))
				h.Create(mod, mod.Create).ServeHTTP(rw, req)
				assert.Equal(t, ErrInvalidJSONFieldValue.Status, strconv.Itoa(rw.Result().StatusCode), rw.Body.String())
			})
		},
		"PatchForeignRepository": func(t *testing.T) {
			t.Run("HasMany", func(t *testing.T) {
				t.Run("synced", func(t *testing.T) {
					rw, req := getHttpPair(
						"POST",
						"/posts",
						h.getModelJSON(&PostSDK{Title: "Some title", Comments: []*CommentSDK{{ID: 1}}}))

					mh := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
					postRepo := mh.Repository.(*MockRepository)

					var idGenerated int = 5
					postRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
						Run(func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)
							post, ok := scope.Value.(*PostSDK)
							require.True(t, ok)

							post.ID = idGenerated
						})

					commentHandler := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
					commentRepo := commentHandler.Repository.(*MockRepository)
					commentRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						comment, ok := scope.Value.(*CommentSDK)
						require.True(t, ok)
						assert.Equal(t, idGenerated, comment.PostID)
					})

					h.Create(mh, mh.Create).ServeHTTP(rw, req)

					assert.Equal(t, 201, rw.Code, rw.Body.String())
				})

				t.Run("nosync", func(t *testing.T) {
					rw, req := getHttpPair(
						"POST",
						"/posts",
						h.getModelJSON(&PostSDK{Title: "Some title", CommentsNoSync: []*CommentSDK{{ID: 1}}}))

					mh := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
					postRepo := mh.Repository.(*MockRepository)

					var idGenerated int = 5
					postRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
						Run(func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)
							post, ok := scope.Value.(*PostSDK)
							require.True(t, ok)

							post.ID = idGenerated
						})
					assert.NotPanics(t, func() { h.Create(mh, mh.Create).ServeHTTP(rw, req) })

					assert.Equal(t, 201, rw.Code, rw.Body.String())
				})
			})
			t.Run("HasOne", func(t *testing.T) {
				t.Run("sync", func(t *testing.T) {
					var postID int = 1
					rw, req := getHttpPair("POST",
						"/blogs",
						h.getModelJSON(&BlogSDK{CurrentPost: &PostSDK{ID: postID}}),
					)

					mh := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					blogRepo := mh.Repository.(*MockRepository)

					var idGenerated int = 5
					blogRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
						Run(func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)
							blog, ok := scope.Value.(*BlogSDK)
							require.True(t, ok)

							blog.ID = idGenerated
						})

					postHandler := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
					postRepo := postHandler.Repository.(*MockRepository)
					postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						post, ok := scope.Value.(*PostSDK)
						require.True(t, ok)
						if assert.NotEmpty(t, scope.PrimaryFilters) {
							if assert.NotEmpty(t, scope.PrimaryFilters[0].Values) {
								assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
							}
						}
						assert.Equal(t, idGenerated, post.BlogID)
						assert.Contains(t, scope.SelectedFields, scope.Struct.foreignKeys["blog_id"])
					})

					h.Create(mh, mh.Create).ServeHTTP(rw, req)

					assert.Equal(t, 201, rw.Code, rw.Body.String())
				})
			})
			t.Run("BelongsToForeigns", func(t *testing.T) {
				var postID int = 3
				rw, req := getHttpPair("POST",
					"/comments",
					h.getModelJSON(&CommentSDK{Post: &PostSDK{ID: postID}}),
				)

				mh := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
				commentRepo := mh.Repository.(*MockRepository)

				var idGenerated int = 5
				commentRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						comment, ok := scope.Value.(*CommentSDK)
						require.True(t, ok)

						comment.ID = idGenerated
						assert.Equal(t, postID, comment.PostID)
					})

				h.Create(mh, mh.Create).ServeHTTP(rw, req)

				assert.Equal(t, 201, rw.Code, rw.Body.String())
			})
			t.Run("Many2Many", func(t *testing.T) {
				// Prepare handler
				getHandler := func() *Handler {
					t.Helper()
					models := []interface{}{&PetSDK{}, &HumanSDK{}}
					h := prepareHandler(defaultLanguages, models...)
					mockRepo := &MockRepository{}
					for _, model := range h.ModelHandlers {

						model.Repository = &MockRepository{}
						model.Create = &Endpoint{Type: Create}
					}
					h.SetDefaultRepo(mockRepo)
					return h
				}

				t.Run("sync", func(t *testing.T) {
					h := getHandler()
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
						"/humans",
						h.getModelJSON(&HumanSDK{PetsSync: []*PetSDK{{ID: petID1}, {ID: petID2}}}),
					)

					humanModel := h.ModelHandlers[reflect.TypeOf(HumanSDK{})]
					require.NotNil(t, humanModel)
					humanRepo, ok := humanModel.Repository.(*MockRepository)
					require.True(t, ok)

					humanRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().
						Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						human, ok := scope.Value.(*HumanSDK)
						require.True(t, ok)

						human.ID = humanID
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

					// assert.NotPanics(t, func() {
					h.Create(humanModel, humanModel.Create).ServeHTTP(rw, req)
					// })

					assert.Equal(t, 201, rw.Code)
				})

				t.Run("nosync", func(t *testing.T) {
					h := getHandler()
					var (
						petID   int = 3
						humanID int = 2
					)

					rw, req := getHttpPair(
						"POST",
						"/humans",
						h.getModelJSON(&HumanSDK{Pets: []*PetSDK{{ID: petID}}}),
					)

					humanModel := h.ModelHandlers[reflect.TypeOf(HumanSDK{})]
					require.NotNil(t, humanModel)
					humanRepo, ok := humanModel.Repository.(*MockRepository)
					require.True(t, ok)

					humanRepo.On("Create", mock.AnythingOfType("*jsonapi.Scope")).Once().
						Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						human, ok := scope.Value.(*HumanSDK)
						require.True(t, ok)

						human.ID = humanID
						if assert.NotEmpty(t, human.Pets) {
							pet := human.Pets[0]
							if assert.NotNil(t, pet) {
								assert.Equal(t, petID, pet.ID, pet)
							}
						}
					})

					assert.NotPanics(t, func() { h.Create(humanModel, humanModel.Create).ServeHTTP(rw, req) })
					assert.Equal(t, 201, rw.Code)
				})
			})
		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}
}
