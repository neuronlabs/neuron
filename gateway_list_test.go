package jsonapi

import (
	"github.com/kucjac/uni-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestHandlerList(t *testing.T) {
	getHandler := func() *Handler {
		h := prepareHandler(defaultLanguages, blogModels...)
		for _, model := range h.ModelHandlers {
			model.Repository = &MockRepository{}
			model.List = &Endpoint{}
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
			t.Run("RequestWithoutLangTag", func(t *testing.T) {
				// Case 1:
				// Correct with no Accept-Language header

				h := getHandler()
				rw, req := getHttpPair("GET", "/blogs?fields[blogs]=some_attr,lang", nil)
				model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

				repo.On("List", mock.Anything).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)
						var langtag string
						if assert.NotEmpty(t, scope.LanguageFilters) {
							if assert.Len(t, scope.LanguageFilters.Values, 1) {
								filter := scope.LanguageFilters.Values[0]
								if assert.NotNil(t, filter) {
									if assert.Len(t, filter.Values, 1) {
										langtag, ok = filter.Values[0].(string)
										require.True(t, ok)
									}
								}
							}
						}

						scope.Value = []*BlogSDK{{ID: 1, Lang: langtag, SomeAttr: "some_attribute"}}
					})

				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))
				postRepo.AssertNotCalled(t, "List")

				if assert.NotPanics(t, func() { h.List(model, model.List).ServeHTTP(rw, req) }) {
					if assert.Equal(t, 200, rw.Result().StatusCode, rw.Body.String()) {
						blogs := []*BlogSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, &blogs)) {
							if assert.Len(t, blogs, 1) {
								if assert.NotNil(t, blogs[0]) {
									assert.Equal(t, 1, blogs[0].ID)
									assert.NotEmpty(t, blogs[0].Lang)
									assert.Nil(t, blogs[0].CurrentPost)
									assert.Nil(t, blogs[0].CurrentPostNoSync)
								}
							}
						}
					}
				}
			})

			t.Run("RequestWithLangTag", func(t *testing.T) {
				// Case 2:
				// Correct with Accept-Language header
				h := getHandler()

				rw, req := getHttpPair("GET", "/blogs?language=pl&fields[blogs]=some_attr,lang", nil)

				model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))
				// req.Header.Add(headerAcceptLanguage, "pl;q=0.9, en")

				repo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						scope.Value = []*BlogSDK{{ID: 1, Lang: "pl", SomeAttr: "Yes"}}
					})

				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))
				postRepo.AssertNotCalled(t, "List")

				if assert.NotPanics(t, func() { h.List(model, model.List).ServeHTTP(rw, req) }) {
					if assert.Equal(t, 200, rw.Result().StatusCode, rw.Body.String()) {
						blogs := []*BlogSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, &blogs)) {
							if assert.Len(t, blogs, 1) {
								if assert.NotNil(t, blogs[0]) {
									assert.Equal(t, 1, blogs[0].ID)
									assert.Equal(t, "pl", blogs[0].Lang)
									assert.Nil(t, blogs[0].CurrentPost)
									assert.Nil(t, blogs[0].CurrentPostNoSync)
								}

							}
						}
					}
				}
			})
		},
		"Includes": func(t *testing.T) {
			t.Run("Invalid", func(t *testing.T) {
				// Case 3:
				// User input error on query
				h := getHandler()
				rw, req := getHttpPair("GET", "/blogs?include=invalid", nil)

				model, repo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

				repo.AssertNotCalled(t, "List")
				h.List(model, model.List).ServeHTTP(rw, req)

				assert.Equal(t, 400, rw.Result().StatusCode)
			})

			t.Run("InFieldset", func(t *testing.T) {

				blogID1 := 1
				blogID2 := 3

				postID1 := 1
				postID2 := 4

				h := getHandler()

				rw, req := getHttpPair("GET", "/blogs?include=current_post&fields[blogs]=current_post&fields[posts]=created_at", nil)

				blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))
				blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						scope.Value = []*BlogSDK{{ID: blogID1}, {ID: blogID2}}
						assert.NotEmpty(t, scope.IncludedFields)
					})

				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

				/**

				Get foreign relations

				*/
				postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								fkFilter := scope.ForeignKeyFilters[0]

								blogFK := scope.Struct.foreignKeys["blog_id"]

								assert.Equal(t, fkFilter.StructField, blogFK)
								if assert.Len(t, fkFilter.Values, 1) {
									if assert.NotNil(t, fkFilter.Values[0]) {
										if assert.Len(t, fkFilter.Values[0].Values, 2) {
											assert.Contains(t, fkFilter.Values[0].Values, blogID1)
											assert.Contains(t, fkFilter.Values[0].Values, blogID2)
										}

									}
								}
							}
						}
						scope.Value = []*PostSDK{{ID: postID1, BlogID: blogID1}, {ID: postID2, BlogID: blogID2}}
					})
				/**

				Get Included relations

				*/
				postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID1)
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID2)

									}
								}
							}
						}

						scope.Value = []*PostSDK{{ID: postID1, CreatedAt: time.Now(), BlogID: blogID1}, {ID: postID2, BlogID: blogID2, CreatedAt: time.Now()}}
					})

				if assert.NotPanics(t, func() { h.List(blogModel, blogModel.List).ServeHTTP(rw, req) }) {
					// h.List(blogModel, blogModel.List).ServeHTTP(rw, req)

					if assert.Equal(t, 200, rw.Code) {
						payload, err := h.Controller.unmarshalManyPayload(rw.Body)
						if assert.NoError(t, err) {
							assert.Len(t, payload.Data, 2)
							assert.Len(t, payload.Included, 2)
						}
					}
				}
			})

		},
		"InvalidLangTag": func(t *testing.T) {
			// Case 4:
			// Getting incorrect language
			h := getHandler()

			rw, req := getHttpPair("GET", "/blogs?language=nosuchlanguage", nil)

			blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

			blogRepo.AssertNotCalled(t, "List")

			if assert.NotPanics(t, func() { h.List(blogModel, blogModel.List).ServeHTTP(rw, req) }) {

				assert.Equal(t, 400, rw.Result().StatusCode)
			}
		},
		"RepositoryError": func(t *testing.T) {

			t.Run("Internal", func(t *testing.T) {
				h := getHandler()

				rw, req := getHttpPair("GET", "/blogs", nil)

				blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

				blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(unidb.ErrInternalError.New())
				if assert.NotPanics(t, func() { h.List(blogModel, blogModel.List).ServeHTTP(rw, req) }) {

					assert.Equal(t, 500, rw.Code)
				}
			})

			t.Run("NotFound", func(t *testing.T) {
				h := getHandler()
				rw, req := getHttpPair("GET", "/blogs", nil)

				blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))
				blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(unidb.ErrNoResult.New())

				if assert.NotPanics(t, func() { h.List(blogModel, blogModel.List).ServeHTTP(rw, req) }) {

					if assert.Equal(t, 200, rw.Code) {
						blogs := []*BlogSDK{}

						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, &blogs)) {
							assert.Empty(t, blogs)
						}
					}
				}
			})
		},
		"Relations": func(t *testing.T) {
			t.Run("HasOne", func(t *testing.T) {
				h := getHandler()

				var (
					blogID1, blogID2 int = 3, 4
					postID1, postID2 int = 1, 15
				)

				rw, req := getHttpPair(
					"GET",
					"/blogs?fields[blogs]=current_post",
					nil,
				)
				blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
				require.True(t, ok)

				blogRepo, ok := blogModel.Repository.(*MockRepository)
				require.True(t, ok)

				blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
					func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

						scope.Value = []*BlogSDK{{ID: blogID1}, {ID: blogID2}}
					})

				postModel, ok := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
				require.True(t, ok)

				postRepo, ok := postModel.Repository.(*MockRepository)
				require.True(t, ok)

				postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID1)
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID2)
									}
								}
							}
						}

						scope.Value = []*PostSDK{{ID: postID1, BlogID: blogID1}, {ID: postID2, BlogID: blogID2}}
					})

				// postRepo.AssertCalled(t, "Get", mock.AnythingOfType("*jsonapi.Scope"))

				h.List(blogModel, blogModel.Get).ServeHTTP(rw, req)

				if assert.Equal(t, 200, rw.Code) {

					manyPayload, err := h.Controller.unmarshalManyPayload(rw.Body)
					if assert.NoError(t, err) {
						assert.Len(t, manyPayload.Data, 2)
					}
				}
			})

			t.Run("HasMany", func(t *testing.T) {
				t.Run("InFieldset", func(t *testing.T) {
					h := getHandler()

					var (
						authorID1, authorID2           int = 3, 12
						author1blogID1, author1blogID2 int = 1, 5
						author2blogID1, author2blogID2 int = 4, 16
					)

					rw, req := getHttpPair(
						"GET",
						"/authors",
						nil,
					)

					authorModel, ok := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
					require.True(t, ok)

					authorRepo, ok := authorModel.Repository.(*MockRepository)
					require.True(t, ok)

					authorRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
						func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)

							// set ID - authorID and blogs as blogID1 and blogID4
							scope.Value = []*AuthorSDK{{ID: authorID1, Name: "SomeName"}, {ID: authorID2, Name: "OtherName"}}
						})

					blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					require.True(t, ok)

					blogRepo, ok := blogModel.Repository.(*MockRepository)
					require.True(t, ok)

					blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
						func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)

							assert.Contains(t, scope.Fieldset, "author_id")

							assert.Len(t, scope.Fieldset, 1)

							if assert.Len(t, scope.ForeignKeyFilters, 1) {
								if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
									if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
										if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
											assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, authorID1)
											assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, authorID2)
										}
									}
								}
							}

							scope.Value = []*BlogSDK{
								{ID: author1blogID1, AuthorID: authorID1},
								{ID: author1blogID2, AuthorID: authorID1},
								{ID: author2blogID1, AuthorID: authorID2},
								{ID: author2blogID2, AuthorID: authorID2},
							}
						})

					if assert.NotPanics(t, func() { h.List(authorModel, authorModel.Get).ServeHTTP(rw, req) }) {
						// h.List(authorModel, authorModel.List).ServeHTTP(rw, req)

						if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
							many, err := h.Controller.unmarshalManyPayload(rw.Body)
							if assert.NoError(t, err) {
								if assert.Len(t, many.Data, 2) {
									for _, n := range many.Data {
										assert.Len(t, n.Relationships["blogs"], 2)
									}
								}
							}
						}

					}

				})

				t.Run("NotInFieldset", func(t *testing.T) {
					var authorID1, authorID2 int = 3, 66
					h := getHandler()

					rw, req := getHttpPair(
						"GET",
						"/authors?fields[authors]=name",
						nil,
					)

					authorModel, ok := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
					require.True(t, ok)

					authorRepo, ok := authorModel.Repository.(*MockRepository)
					require.True(t, ok)

					authorRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

						assert.Len(t, scope.Fieldset, 1)
						scope.Value = []*AuthorSDK{{ID: authorID1, Name: "First"}, {ID: authorID2, Name: "Second"}}
					})

					blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					require.True(t, ok)

					blogRepo, ok := blogModel.Repository.(*MockRepository)
					require.True(t, ok)

					blogRepo.AssertNotCalled(t, "List")

					if assert.NotPanics(t, func() { h.List(authorModel, authorModel.List).ServeHTTP(rw, req) }) {
						assert.Equal(t, 200, rw.Code)
					}
				})
			})

			t.Run("BelongsTo", func(t *testing.T) {
				h := getHandler()

				commentID1, commentID2 := 4, 12
				postID1, postID2 := 5, 30

				rw, req := getHttpPair("GET", "/comments", nil)

				commentModel, ok := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
				require.True(t, ok)

				commentRepo, ok := commentModel.Repository.(*MockRepository)
				require.True(t, ok)

				commentRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						scope.Value = []*CommentSDK{{ID: commentID1, Body: "My body is great", PostID: postID1}, {ID: commentID2, Body: "My body is great also", PostID: postID2}}
					})

				if assert.NotPanics(t, func() { h.List(commentModel, commentModel.List).ServeHTTP(rw, req) }) {
					if assert.Equal(t, 200, rw.Code) {
						comments := []*CommentSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, &comments)) {
							for _, comment := range comments {
								if assert.NotNil(t, comment.Post) {
									if comment.ID == commentID1 {
										assert.Equal(t, postID1, comment.Post.ID)
									} else {
										assert.Equal(t, postID2, comment.Post.ID)
									}
								}
							}

						}
					}
				}
			})
		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}

}
