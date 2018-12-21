package gateway

import (
	"fmt"
	"github.com/kucjac/uni-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"reflect"
	"testing"
)

func TestGatewayDelete(t *testing.T) {
	getHandler := func(models ...interface{}) *Handler {
		h := prepareHandler(defaultLanguages, models...)
		for _, model := range h.ModelHandlers {
			model.Delete = &Endpoint{}
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
		"Success": func(t *testing.T) {
			h := getHandler(&ModelSDK{})

			modelID := 1
			rw, req := getHttpPair("DELETE", fmt.Sprintf("/models/%d", modelID), nil)

			model, repo := getModelAndRepo(t, h, reflect.TypeOf(ModelSDK{}))

			repo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).Once().
				Return(nil).Run(func(args mock.Arguments) {
				scope, ok := args.Get(0).(*Scope)
				require.True(t, ok)

				if assert.Len(t, scope.PrimaryFilters, 1) {
					if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
						fv := scope.PrimaryFilters[0].Values[0]
						assert.Equal(t, OpEqual, fv.Operator)
						assert.Contains(t, fv.Values, modelID)
					}
				}
			})

			if assert.NotPanics(t, func() { h.Delete(model, model.Delete).ServeHTTP(rw, req) }) {
				assert.Equal(t, 204, rw.Result().StatusCode)
				repo.AssertCalled(t, "Delete", mock.Anything)
			}

		},
		"Relations": func(t *testing.T) {
			t.Run("BelongsTo", func(t *testing.T) {
				// Deleting entry with belongs to relationship should
				// occur within the repository
				h := getHandler(blogModels...)

				commentID := 1
				rw, req := getHttpPair("DELETE", fmt.Sprintf("/comments/%d", commentID), nil)

				model, repo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))

				repo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).Once().
					Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
							fv := scope.PrimaryFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, commentID)
						}
					}
				})

				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

				if assert.NotPanics(t, func() { h.Delete(model, model.Delete).ServeHTTP(rw, req) }) {
					repo.AssertCalled(t, "Delete", mock.Anything)
					postRepo.AssertNotCalled(t, "Patch", mock.Anything)
					assert.Equal(t, 204, rw.Result().StatusCode)
				}
			})
			t.Run("HasOne", func(t *testing.T) {
				// Deleting entry with HasOne relationships should
				// also clear foreign keys from the relationship repositories
				// that contain given FK
				h := getHandler(blogModels...)

				var blogID int = 1

				rw, req := getHttpPair("DELETE", fmt.Sprintf("/blogs/%d", blogID), nil)

				blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

				blogRepo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).Once().
					Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
							fv := scope.PrimaryFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, blogID)
						}
					}
				})
				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))
				postRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).
					Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					post, ok := scope.Value.(*PostSDK)
					require.True(t, ok)

					blogIDField, ok := scope.Struct.FieldByName("BlogID")
					require.True(t, ok)

					if assert.Len(t, scope.ForeignKeyFilters, 1) {
						assert.Equal(t, blogIDField, scope.ForeignKeyFilters[0].StructField)
						if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
							fv := scope.ForeignKeyFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, blogID)
						}
					}
					assert.Zero(t, post.BlogID)
					assert.Contains(t, scope.SelectedFields, blogIDField)
				})

				if assert.NotPanics(t, func() { h.Delete(blogModel, blogModel.Delete).ServeHTTP(rw, req) }) {
					assert.Equal(t, 204, rw.Result().StatusCode)
					blogRepo.AssertCalled(t, "Delete", mock.Anything)
					postRepo.AssertCalled(t, "Patch", mock.Anything)
				}
			})
			t.Run("HasMany", func(t *testing.T) {
				// Deleting entry with HasMany relationships should
				// also clear all entries that contains foreign keys from the relationship
				// repositories that contain given FK
				// Deleting entry with HasOne relationships should
				// also clear foreign keys from the relationship repositories
				// that contain given FK
				h := getHandler(blogModels...)

				var postID int = 1

				rw, req := getHttpPair("DELETE", fmt.Sprintf("/posts/%d", postID), nil)

				postModel, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

				postRepo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).Once().
					Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
							fv := scope.PrimaryFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, postID)
						}
					}
				})
				_, commentRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))
				commentRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).
					Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					comment, ok := scope.Value.(*CommentSDK)
					require.True(t, ok)

					postIDField, ok := scope.Struct.FieldByName("PostID")
					require.True(t, ok)

					if assert.Len(t, scope.ForeignKeyFilters, 1) {
						assert.Equal(t, postIDField, scope.ForeignKeyFilters[0].StructField)
						if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
							fv := scope.ForeignKeyFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, postID)
						}
					}
					assert.Contains(t, scope.SelectedFields, postIDField)
					assert.Zero(t, comment.PostID)
				})

				if assert.NotPanics(t, func() { h.Delete(postModel, postModel.Delete).ServeHTTP(rw, req) }) {
					assert.Equal(t, 204, rw.Result().StatusCode)
					postRepo.AssertCalled(t, "Delete", mock.Anything)
					commentRepo.AssertCalled(t, "Patch", mock.Anything)
				}
			})
			t.Run("Many2Many", func(t *testing.T) {
				// Deleting entry with Many2Many relationships the relationship should
				// Should affect Synced relationships
				h := getHandler(&HumanSDK{}, &PetSDK{})

				var petID int = 1
				rw, req := getHttpPair("DELETE", fmt.Sprintf("/pets/%d", petID), nil)

				petModel, petRepo := getModelAndRepo(t, h, reflect.TypeOf(PetSDK{}))

				petRepo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).
					Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
							fv := scope.PrimaryFilters[0].Values[0]
							assert.Equal(t, OpEqual, fv.Operator)
							assert.Contains(t, fv.Values, petID)
						}
					}
				})

				_, humanRepo := getModelAndRepo(t, h, reflect.TypeOf(HumanSDK{}))
				humanRepo.On("Patch", mock.AnythingOfType("*jsonapi.Scope")).
					Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					petsField, ok := scope.Struct.FieldByName("Pets")
					require.True(t, ok)

					assert.Contains(t, scope.SelectedFields, petsField)
					if assert.Len(t, scope.RelationshipFilters, 1) {
						if assert.Equal(t, petsField, scope.RelationshipFilters[0].StructField) {
							if assert.Len(t, scope.RelationshipFilters[0].Nested, 1) {
								relFilter := scope.RelationshipFilters[0].Nested[0]
								assert.Equal(t, petsField.relatedStruct.primary, relFilter.StructField)

								if assert.Len(t, relFilter.Values, 1) {
									assert.Equal(t, OpEqual, relFilter.Values[0].Operator)
									assert.Contains(t, relFilter.Values[0].Values, petID)
								}
							}
						}
					}
					human, ok := scope.Value.(*HumanSDK)
					require.True(t, ok)

					assert.Zero(t, human.Pets)
				})

				// if assert.NotPanics(t, func() { h.Delete(petModel, petModel.Delete).ServeHTTP(rw, req) }) {
				h.Delete(petModel, petModel.Delete).ServeHTTP(rw, req)
				humanRepo.AssertCalled(t, "Patch", mock.Anything)
				assert.Equal(t, http.StatusNoContent, rw.Code)
				// }
			})
		},
		"NotFound": func(t *testing.T) {
			h := getHandler(&ModelSDK{})

			modelID := 1
			rw, req := getHttpPair("DELETE", fmt.Sprintf("/models/%d", modelID), nil)

			model, repo := getModelAndRepo(t, h, reflect.TypeOf(ModelSDK{}))

			repo.On("Delete", mock.AnythingOfType("*jsonapi.Scope")).Once().
				Return(unidb.ErrNoResult.New()).Run(func(args mock.Arguments) {
				scope, ok := args.Get(0).(*Scope)
				require.True(t, ok)

				if assert.Len(t, scope.PrimaryFilters, 1) {
					if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
						fv := scope.PrimaryFilters[0].Values[0]
						assert.Equal(t, OpEqual, fv.Operator)
						assert.Contains(t, fv.Values, modelID)
					}
				}
			})

			if assert.NotPanics(t, func() { h.Delete(model, model.Delete).ServeHTTP(rw, req) }) {
				assert.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
			}
		},
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}
}
