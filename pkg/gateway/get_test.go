package gateway

import (
	"fmt"
	"github.com/kucjac/uni-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestGatewayGet(t *testing.T) {

	getHandler := func() *Handler {
		h := prepareHandler(defaultLanguages, blogModels...)
		for _, model := range h.ModelHandlers {
			model.Get = &Endpoint{}
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
			// Case 1:
			// Getting an object correctly without accept-language header
			h := getHandler()
			model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			blogRepo, ok := model.Repository.(*MockRepository)
			require.True(t, ok)

			rw, req := getHttpPair("GET", "/blogs/1?fields[blogs]=some_attr", nil)
			blogRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
				Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)

					scope.Value = &BlogSDK{ID: 1, Lang: "pl", SomeAttr: "My Attr"}
				})

			h.Get(model, model.Get).ServeHTTP(rw, req)

			// // assert content-language is the same
			assert.Equal(t, 200, rw.Result().StatusCode)
		},
		"NonExistingObject": func(t *testing.T) {
			// Case 2:
			// Getting a non-existing object
			h := getHandler()
			model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			blogRepo, ok := model.Repository.(*MockRepository)
			require.True(t, ok)
			blogRepo.On("Get", mock.Anything).Once().Return(unidb.ErrUniqueViolation.New())
			rw, req := getHttpPair("GET", "/blogs/123", nil)
			h.Get(model, model.Get).ServeHTTP(rw, req)

			assert.Equal(t, 409, rw.Result().StatusCode)
		},
		"BadUrl": func(t *testing.T) {
			// Case 3:
			// assigning bad url - internal error
			h := getHandler()
			model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]

			rw, req := getHttpPair("GET", "/blogs", nil)
			h.Get(model, model.Get).ServeHTTP(rw, req)

			assert.Equal(t, 500, rw.Result().StatusCode)
		},
		"InvalidQuery": func(t *testing.T) {
			// Case 4:
			// User input error (i.e. invalid query)
			h := getHandler()
			model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]

			rw, req := getHttpPair("GET", "/blogs/1?include=nonexisting", nil)
			h.Get(model, model.Get).ServeHTTP(rw, req)

			assert.Equal(t, 400, rw.Result().StatusCode)
		},
		"UnsupportedLanguage": func(t *testing.T) {
			// Case 5:
			// User provided unsupported language
			h := getHandler()
			model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]

			rw, req := getHttpPair("GET", "/blogs/1?language=nonsupportedlang", nil)
			h.Get(model, model.Get).ServeHTTP(rw, req)

			assert.Equal(t, 400, rw.Result().StatusCode)
		},
		"IncludedValues": func(t *testing.T) {
			t.Run("HasOne", func(t *testing.T) {
				t.Helper()
				t.Run("NonSynced", func(t *testing.T) {

					// If a relation is non synced the relation id's are taken from the repository
					postID := 3
					h := getHandler()
					model := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					blogRepo, ok := model.Repository.(*MockRepository)
					require.True(t, ok)

					rw, req := getHttpPair("GET", "/blogs/1?include=current_post_no_sync&fields[blogs]=current_post_no_sync&fields[posts]=title", nil)

					blogRepo.On("Get", mock.Anything).Once().Return(nil).
						Run(func(args mock.Arguments) {
							arg := args.Get(0).(*Scope)
							arg.Value = &BlogSDK{
								ID:   1,
								Lang: h.SupportedLanguages[0].String(),

								CurrentPostNoSync: &PostSDK{
									ID: postID,
								}}
						})

					postModel := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
					postRepo, ok := postModel.Repository.(*MockRepository)
					require.True(t, ok)

					// Included values are taken
					postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
								}
							}
						}
						scope.Value = []*PostSDK{{ID: postID, Title: "NonSynced"}}

					})
					// postRepo.On("Get", mock.Anything).Once().Return(nil).
					// 	Run(func(args mock.Arguments) {
					// 		scope, ok := args.Get(0).(*Scope)
					// 		require.True(t, ok)

					// 		if assert.Len(t, scope.PrimaryFilters, 1) {
					// 			if assert.NotNil(t, scope.PrimaryFilters[0]) {
					// 				if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
					// 					assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
					// 				}
					// 			}
					// 		}

					// 		scope.Value = []*PostSDK{{ID: 3}}
					// 	})

					require.NotPanics(t, func() { h.Get(model, model.Get).ServeHTTP(rw, req) })
					assert.Equal(t, 200, rw.Result().StatusCode)

					payload, err := h.Controller.unmarshalPayload(rw.Body)
					if assert.NoError(t, err) {
						assert.NotNil(t, payload.Data)
						assert.NotEmpty(t, payload.Included)
					}
				})

				t.Run("Synced", func(t *testing.T) {
					h := getHandler()
					rw, req := getHttpPair("GET", "/blogs/3?include=current_post&fields[blogs]=current_post&fields[posts]=title", nil)

					var (
						blogID, postID int = 3, 12
					)

					blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

					blogRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						scope.Value = &BlogSDK{ID: blogID}
					})

					_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.Fieldset, 1) {
							assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.Fieldset["blog_id"])
						}

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.ForeignKeyFilters[0].StructField)
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID)
									}
								}
							}
						}
						scope.Value = &PostSDK{ID: postID, BlogID: blogID, Title: "Some title"}
					})

					postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.Fieldset, 2) {
							_, ok := scope.Fieldset["title"]
							assert.True(t, ok)
						}

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
									}
								}
							}
						}
					})
					require.NotPanics(t, func() { h.Get(blogModel, blogModel.Get).ServeHTTP(rw, req) })

					if assert.Equal(t, 200, rw.Code) {

						blog := &BlogSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, blog)) {
							assert.Equal(t, blogID, blog.ID)
							if assert.NotNil(t, blog.CurrentPost) {
								assert.Equal(t, postID, blog.CurrentPost.ID)
							}
						}
					}
				})

				t.Run("SyncedWithNested", func(t *testing.T) {
					h := getHandler()
					// Nested includes should be also taken for it's relationships
					rw, req := getHttpPair("GET", "/blogs/3?include=current_post.comments&fields[blogs]=current_post&fields[posts]=title", nil)

					var (
						blogID, postID int = 3, 12
					)

					blogModel, blogRepo := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

					blogRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						scope.Value = &BlogSDK{ID: blogID}
					})

					_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.Fieldset, 1) {
							assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.Fieldset["blog_id"])
						}

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.ForeignKeyFilters[0].StructField)
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID)
									}
								}
							}
						}
						scope.Value = &PostSDK{ID: postID, BlogID: blogID, Title: "Some title"}
					})

					postRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.Fieldset, 3) {
							_, ok := scope.Fieldset["title"]
							assert.True(t, ok)
						}

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
									}
								}
							}
						}
					})

					var commentID1, commentID2 int = 15, 12

					_, commentRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))

					// First List Call is about relationship values.
					commentRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						assert.Len(t, scope.ForeignKeyFilters, 1)
						assert.Contains(t, scope.Fieldset, "post_id")

						scope.Value = []*CommentSDK{{ID: commentID1, PostID: postID}, {ID: commentID2, PostID: postID}}

					})

					commentRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										fv := scope.PrimaryFilters[0].Values[0]
										assert.Contains(t, fv.Values, commentID1)
										assert.Contains(t, fv.Values, commentID2)
									}
								}
							}
						}
						scope.Value = []*CommentSDK{
							{ID: commentID1, Body: "First", PostID: postID},
							{ID: commentID2, Body: "Second", PostID: postID},
						}
					})

					require.NotPanics(t, func() { h.Get(blogModel, blogModel.Get).ServeHTTP(rw, req) })

					if assert.Equal(t, 200, rw.Code) {

						blog := &BlogSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, blog)) {
							assert.Equal(t, blogID, blog.ID)
							if assert.NotNil(t, blog.CurrentPost) {
								assert.Equal(t, postID, blog.CurrentPost.ID)
							}
						}
					}
				})
			})

			t.Run("HasMany", func(t *testing.T) {
				t.Helper()
				t.Run("Synced", func(t *testing.T) {
					h := getHandler()
					rw, req := getHttpPair("GET", "/posts/3?include=comments&fields[posts]=comments", nil)

					var (
						postID, commentID1, commentID2 int = 3, 12, 14
					)

					postModel, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))

					// Get root model from its repo
					postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						scope.Value = &PostSDK{ID: postID}
					})

					// get the relation's primaries field included within the scope
					_, commentsRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))
					commentsRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {

								assert.Equal(t, scope.Struct.foreignKeys["post_id"], scope.ForeignKeyFilters[0].StructField)
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										vals := scope.ForeignKeyFilters[0].Values[0].Values
										assert.Len(t, vals, 1)
										assert.Contains(t, vals, postID)
									}
								}
							}
						}
						scope.Value = []*CommentSDK{
							{ID: commentID1, PostID: postID},
							{ID: commentID2, PostID: postID},
						}
					})

					// Included field should get the comments with primary filter field
					commentsRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
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
						scope.Value = []*CommentSDK{
							{ID: commentID1, Body: "Body1", PostID: postID},
							{ID: commentID2, Body: "Body2", PostID: postID},
						}

					})
					// require.NotPanics(t, func() { h.Get(postModel, postModel.Get).ServeHTTP(rw, req) })
					h.Get(postModel, postModel.Get).ServeHTTP(rw, req)
					if assert.Equal(t, 200, rw.Code) {
						payload, err := h.Controller.unmarshalPayload(rw.Body)
						if assert.NoError(t, err) {
							if assert.Len(t, payload.Included, 2) {
								var founds int
								for _, inc := range payload.Included {
									switch inc.ID {
									case strconv.Itoa(commentID1), strconv.Itoa(commentID2):
										founds++
									}
								}
								assert.Equal(t, 2, founds)
							}
							if assert.NotNil(t, payload.Data) {
								assert.Equal(t, strconv.Itoa(postID), payload.Data.ID)
							}

						}
					}

				})
			})
		},
		"PrecheckValues": func(t *testing.T) {
			h := getHandler()
			blogModel := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
			blogRepo, ok := blogModel.Repository.(*MockRepository)
			require.True(t, ok)

			precheckPair := h.Controller.BuildPrecheckPair("preset=blogs.current_post&filter[blogs][id][$eq]=1", "filter[comments][post][id]")

			commentModel := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
			commentModel.AddPrecheckPair(precheckPair, Get)
			commentRepo, ok := commentModel.Repository.(*MockRepository)
			require.True(t, ok)

			postModel := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
			postRepo, ok := postModel.Repository.(*MockRepository)
			require.True(t, ok)

			rw, req := getHttpPair("GET", "/comments/1", nil)

			blogRepo.On("List", mock.Anything).Once().Return(nil).
				Run(func(args mock.Arguments) {
					arg := args.Get(0).(*Scope)
					arg.Value = []*BlogSDK{{ID: 1, CurrentPost: &PostSDK{ID: 3}}}
				})

			postRepo.On("List", mock.Anything).Once().Return(nil).Run(
				func(args mock.Arguments) {
					arg := args.Get(0).(*Scope)
					t.Logf("%v", arg.Fieldset)
					arg.Value = []*PostSDK{{ID: 3}}
				})

			commentRepo.On("Get", mock.Anything).Once().Return(nil).Run(
				func(args mock.Arguments) {
					arg := args.Get(0).(*Scope)
					arg.Value = &CommentSDK{ID: 1, Body: "Some body", Post: &PostSDK{ID: 3}}
				})

			h.Get(commentModel, commentModel.Get).ServeHTTP(rw, req)
		},
		"GetRelations": func(t *testing.T) {

			t.Run("HasOne", func(t *testing.T) {
				h := getHandler()

				var (
					blogID int = 3
					postID int = 1
				)

				rw, req := getHttpPair(
					"GET",
					"/blogs/"+strconv.Itoa(blogID)+"?fields[blogs]=current_post",
					nil,
				)
				blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
				require.True(t, ok)

				blogRepo, ok := blogModel.Repository.(*MockRepository)
				require.True(t, ok)

				blogRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
					func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

						scope.Value = &BlogSDK{ID: blogID}
					})

				postModel, ok := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
				require.True(t, ok)

				postRepo, ok := postModel.Repository.(*MockRepository)
				require.True(t, ok)

				postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID)
									}
								}
							}
						}

						scope.Value = &PostSDK{ID: postID, BlogID: blogID}
					})

				// postRepo.AssertCalled(t, "Get", mock.AnythingOfType("*jsonapi.Scope"))

				h.Get(blogModel, blogModel.Get).ServeHTTP(rw, req)

				if assert.Equal(t, 200, rw.Code) {
					blog := &BlogSDK{}
					if assert.NoError(t, h.Controller.Unmarshal(rw.Body, blog)) {
						if assert.NotNil(t, blog.CurrentPost) {
							assert.Equal(t, postID, blog.CurrentPost.ID)
						}
					}
				}
			})
			t.Run("HasMany", func(t *testing.T) {
				t.Run("InFieldset", func(t *testing.T) {
					h := getHandler()

					var (
						authorID int = 3
						blogID1  int = 1
						blogID4  int = 4
					)

					rw, req := getHttpPair(
						"GET",
						"/authors/"+strconv.Itoa(authorID),
						nil,
					)

					authorModel, ok := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
					require.True(t, ok)

					authorRepo, ok := authorModel.Repository.(*MockRepository)
					require.True(t, ok)

					authorRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
						func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)

							// set ID - authorID and blogs as blogID1 and blogID4
							scope.Value = &AuthorSDK{ID: authorID, Name: "SomeName"}
						})

					blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					require.True(t, ok)

					blogRepo, ok := blogModel.Repository.(*MockRepository)
					require.True(t, ok)

					blogRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(
						func(args mock.Arguments) {
							scope := args.Get(0).(*Scope)

							assert.Contains(t, scope.Fieldset, "author_id")

							scope.Value = []*BlogSDK{{ID: blogID1, AuthorID: authorID}, {ID: blogID4, AuthorID: authorID}}
						})

					h.Get(authorModel, authorModel.Get).ServeHTTP(rw, req)

					author := &AuthorSDK{}
					if assert.Equal(t, 200, rw.Code) {
						err := h.Controller.Unmarshal(rw.Body, author)
						if assert.NoError(t, err) {
							if assert.Len(t, author.Blogs, 2) {
								if assert.NotNil(t, author.Blogs[0]) {
									assert.Equal(t, blogID1, author.Blogs[0].ID)
								}
								if assert.NotNil(t, author.Blogs[1]) {
									assert.Equal(t, blogID4, author.Blogs[1].ID)
								}

							}
						}
					}
				})

				t.Run("NotInFieldset", func(t *testing.T) {
					var authorID int = 3
					h := getHandler()

					rw, req := getHttpPair(
						"GET",
						"/authors/"+strconv.Itoa(authorID)+"?fields[authors]=name",
						nil,
					)

					authorModel, ok := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
					require.True(t, ok)

					authorRepo, ok := authorModel.Repository.(*MockRepository)
					require.True(t, ok)

					authorRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						scope.Value = &AuthorSDK{ID: authorID}
					})

					blogModel, ok := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
					require.True(t, ok)

					blogRepo, ok := blogModel.Repository.(*MockRepository)
					require.True(t, ok)

					blogRepo.AssertNotCalled(t, "List")

					h.Get(authorModel, authorModel.Get).ServeHTTP(rw, req)

					assert.Equal(t, 200, rw.Code)
				})

			})
			t.Run("Many2Many", func(t *testing.T) {

			})
			t.Run("BelongsTo", func(t *testing.T) {
				h := getHandler()

				commentID := 4
				postID := 5

				rw, req := getHttpPair("GET", "/comments/"+strconv.Itoa(commentID), nil)

				commentModel, ok := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
				require.True(t, ok)

				commentRepo, ok := commentModel.Repository.(*MockRepository)
				require.True(t, ok)

				commentRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						scope.Value = &CommentSDK{ID: commentID, Body: "My body is great", PostID: postID}
					})

				h.Get(commentModel, commentModel.Get).ServeHTTP(rw, req)
				if assert.Equal(t, 200, rw.Code) {
					comment := &CommentSDK{}
					if assert.NoError(t, h.Controller.Unmarshal(rw.Body, comment)) {
						if assert.NotNil(t, comment.Post) {
							assert.Equal(t, postID, comment.Post.ID)
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

func TestHandlerGetRelated(t *testing.T) {
	getHandler := func() *Handler {
		h := prepareHandler(defaultLanguages, blogModels...)
		for _, model := range h.ModelHandlers {
			model.GetRelated = &Endpoint{}
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
			t.Run("RelationHasOne", func(t *testing.T) {

				var blogID int = 1
				var postID int = 15
				var commentID1, commentID2 int = 26, 61
				h := getHandler()

				rw, req := getHttpPair("GET", "/blogs/1/current_post", nil)
				blogModel, _ := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

				_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))
				postRepo.On("Get", mock.Anything).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.ForeignKeyFilters[0].StructField)
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID)
									}
								}
							}
						}

						assert.Len(t, scope.Fieldset, 1)
						scope.Value = &PostSDK{ID: postID, BlogID: blogID}
					})

				postRepo.On("Get", mock.Anything).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

						if assert.Len(t, scope.PrimaryFilters, 1) {
							if assert.NotNil(t, scope.PrimaryFilters[0]) {
								if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
									if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
										assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
									}
								}
							}
						}
						scope.Value = &PostSDK{ID: postID, Title: "This title", CreatedAt: time.Now(), BlogID: blogID}
					})

				_, commentsRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))
				commentsRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.ForeignKeyFilters, 1) {
						if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
							assert.Equal(t, scope.Struct.foreignKeys["post_id"], scope.ForeignKeyFilters[0].StructField)
							if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
								if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
									assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, postID)
								}
							}
						}
					}
					scope.Value = []*CommentSDK{{ID: commentID1, PostID: postID}, {ID: commentID2, PostID: postID}}
				})

				if assert.NotPanics(t, func() { h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req) }) {

					if assert.Equal(t, 200, rw.Result().StatusCode, rw.Body) {
						post := &PostSDK{}
						err := h.Controller.Unmarshal(rw.Body, post)
						if assert.NoError(t, err, rw.Body.String()) {
							count := 0
							for _, comment := range post.Comments {
								switch comment.ID {
								case commentID1, commentID2:
									count += 1
								}
							}

							assert.Equal(t, 2, count)
						}
					}

				}
			})

			t.Run("RelationHasMany", func(t *testing.T) {
				var postID int = 15
				var commentID1, commentID2 int = 26, 61
				h := getHandler()

				rw, req := getHttpPair("GET", "/posts/15/comments", nil)

				postModel, ok := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
				require.True(t, ok)

				_, commentRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))
				commentRepo.On("List", mock.Anything).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope, ok := args.Get(0).(*Scope)
						require.True(t, ok)

						if assert.Len(t, scope.ForeignKeyFilters, 1) {
							if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
								assert.Equal(t, scope.Struct.foreignKeys["post_id"], scope.ForeignKeyFilters[0].StructField)
								if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
									if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
										assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, postID)
									}
								}
							}
						}

						assert.Len(t, scope.Fieldset, 1)
						scope.Value = []*CommentSDK{{ID: commentID1, PostID: postID}, {ID: commentID2, PostID: postID}}
					})

				commentRepo.On("List", mock.Anything).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)

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
						scope.Value = []*CommentSDK{{ID: commentID1, Body: "Some", PostID: postID}, {ID: commentID2, Body: "Another body", PostID: postID}}
					})

				if assert.NotPanics(t, func() { h.GetRelated(postModel, postModel.GetRelated).ServeHTTP(rw, req) }) {

					if assert.Equal(t, 200, rw.Result().StatusCode, rw.Body) {
						comments := []*CommentSDK{}
						err := h.Controller.Unmarshal(rw.Body, &comments)
						if assert.NoError(t, err, rw.Body.String()) {
							count := 0
							for _, comment := range comments {
								switch comment.ID {
								case commentID1, commentID2:
									count += 1
								}
							}

							assert.Equal(t, 2, count)
						}
					}

				}
			})

			t.Run("RelationBelongsTo", func(t *testing.T) {
				h := getHandler()

				commentID := 4
				postID := 5
				blogID := 3
				commentsID2 := 5

				rw, req := getHttpPair("GET", "/comments/"+strconv.Itoa(commentID)+"/post", nil)

				commentModel, ok := h.ModelHandlers[reflect.TypeOf(CommentSDK{})]
				require.True(t, ok)

				commentRepo, ok := commentModel.Repository.(*MockRepository)
				require.True(t, ok)

				commentRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
					Run(func(args mock.Arguments) {
						scope := args.Get(0).(*Scope)
						scope.Value = &CommentSDK{ID: commentID, Body: "My body is great", PostID: postID}
					})

				postRepo, ok := h.GetRepositoryByType(reflect.TypeOf(PostSDK{})).(*MockRepository)
				require.True(t, ok)

				postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					scope.Value = &PostSDK{ID: postID, Title: "Title", BlogID: blogID, CreatedAt: time.Now(), CommentsNoSync: []*CommentSDK{{ID: 125}}}
				})

				commentRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					scope.Value = []*CommentSDK{{ID: commentID, PostID: postID}, {ID: commentsID2, PostID: postID}}
				})

				if assert.NotPanics(t, func() { h.GetRelated(commentModel, commentModel.GetRelated).ServeHTTP(rw, req) }) {
					if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
						post := &PostSDK{}
						if assert.NoError(t, h.Controller.Unmarshal(rw.Body, post)) {
							if assert.NotNil(t, post) {
								assert.Equal(t, postID, post.ID)
								ctr := 0
								for _, com := range post.Comments {
									switch com.ID {
									case commentID, commentsID2:
										ctr++
									}
								}
								assert.Equal(t, 2, ctr)
							}
						}
					}
				}
			})

			t.Run("RelationMany2Many", func(t *testing.T) {
				models := append(blogModels, &HumanSDK{}, &PetSDK{})
				h := prepareHandler(defaultLanguages, models...)
				for _, model := range h.ModelHandlers {
					model.GetRelated = &Endpoint{}
					model.Repository = &MockRepository{}
				}

				var petID int = 1
				var humanID1, humandID2 int = 5, 12
				rw, req := getHttpPair("GET", fmt.Sprintf("/pets/%d/humans", petID), nil)

				petModel, petRepo := getModelAndRepo(t, h, reflect.TypeOf(PetSDK{}))
				petRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					scope.Value = &PetSDK{ID: petID, Humans: []*HumanSDK{{ID: humanID1}, {ID: humandID2}}}
				})

				_, humanRepo := getModelAndRepo(t, h, reflect.TypeOf(HumanSDK{}))
				humanRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					// check filters
					scope.Value = []*HumanSDK{{ID: humanID1, Pets: []*PetSDK{{ID: petID}}}, {ID: humandID2, Pets: []*PetSDK{{ID: petID}}}}
				})

				petRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					scope.Value = []*PetSDK{{ID: petID, Humans: []*HumanSDK{{ID: humanID1}, {ID: humandID2}}}}
				})

				humanRepo.On("List", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					// check filters
					scope.Value = []*HumanSDK{{ID: humanID1, Name: "Mietek", Pets: []*PetSDK{{ID: petID}}}, {ID: humandID2, Name: "Maciek", Pets: []*PetSDK{{ID: petID}}}}
				})

				h.GetRelated(petModel, petModel.GetRelated).ServeHTTP(rw, req)
				if assert.Equal(t, 200, rw.Code) {
					humans := []*HumanSDK{}
					err := h.Controller.Unmarshal(rw.Body, &humans)
					if assert.NoError(t, err) {
						if assert.Len(t, humans, 2) {
							ctr := 0
							for _, human := range humans {
								switch human.ID {
								case humanID1, humandID2:
									ctr++
								}
							}
							assert.Equal(t, 2, ctr)
						}

					}

				}

			})
		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}

	// h := prepareHandler(defaultLanguages, blogModels...)
	// mockRepo := &MockRepository{}
	// h.SetDefaultRepo(mockRepo)

	// // Case 1:
	// // Correct related field

	// // Case 2:
	// // Invalid field name
	// rw, req = getHttpPair("GET", "/blogs/1/current_invalid_post", nil)
	// h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)

	// // Case 3:
	// // Invalid model's url. - internal
	// rw, req = getHttpPair("GET", "/blogs/current_invalid_post", nil)
	// h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 4:
	// // invalid language
	// // rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)
	// // req.Header.Add(headerAcceptLanguage, "invalid_language")
	// // h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// // assert.Equal(t, 400, rw.Result().StatusCode)

	// // Case 5:
	// // Root repo dberr
	// rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)

	// mockRepo.On("Get", mock.Anything).Once().Return(unidb.ErrUniqueViolation.New())
	// h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 409, rw.Result().StatusCode)

	// // Case 6:
	// // No primary filter for scope
	// rw, req = getHttpPair("GET", "/blogs/1/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &BlogSDK{ID: 1}
	// 	})
	// h.GetRelated(blogModel, blogModel.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// postsHandler := h.ModelHandlers[reflect.TypeOf(PostSDK{})]
	// postsHandler.GetRelated = &Endpoint{Type: GetRelated}

	// // Case 7:
	// // Get many relateds, correct
	// rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &PostSDK{ID: 1, Comments: []*CommentSDK{{ID: 1}, {ID: 2}}}

	// 	})
	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = []*CommentSDK{{ID: 1, Body: "First comment"}, {ID: 2, Body: "Second CommentSDK"}}
	// 	})
	// h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// // Case 8:
	// // Get many relateds, with empty map
	// rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &PostSDK{ID: 1}
	// 	})
	// h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// // Case 9:
	// // Provided nil value after getting root from repository - Internal Error
	// rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = nil
	// 	})
	// h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 9:
	// // dbErr for related
	// rw, req = getHttpPair("GET", "/posts/1/comments", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &PostSDK{ID: 1, Comments: []*CommentSDK{{ID: 1}}}
	// 	})

	// mockRepo.On("List", mock.Anything).Once().Return(unidb.ErrInternalError.New())
	// h.GetRelated(postsHandler, postsHandler.GetRelated).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 10:
	// // Related field that use language has language filter
	// rw, req = getHttpPair("GET", "/authors/1/blogs", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &AuthorSDK{ID: 1, Blogs: []*BlogSDK{{ID: 1}, {ID: 2}, {ID: 5}}}
	// 	})

	// mockRepo.On("List", mock.Anything).Once().Return(nil).Run(
	// 	func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = []*BlogSDK{{ID: 1}, {ID: 2}, {ID: 5}}
	// 	})

	// authHandler := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
	// authHandler.GetRelated = &Endpoint{Type: GetRelated}
	// h.GetRelated(authHandler, authHandler.GetRelated).ServeHTTP(rw, req)
}

func TestGetRelationship(t *testing.T) {
	getHandler := func() *Handler {
		h := prepareHandler(defaultLanguages, blogModels...)
		for _, model := range h.ModelHandlers {
			model.GetRelated = &Endpoint{}
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
		"RelationBelongsTo": func(t *testing.T) {
			h := getHandler()

			commentID := 4
			postID := 5

			rw, req := getHttpPair("GET", "/comments/"+strconv.Itoa(commentID)+"/relationships/post", nil)

			commentModel, commentRepo := getModelAndRepo(t, h, reflect.TypeOf(CommentSDK{}))
			commentRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).
				Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)
					scope.Value = &CommentSDK{ID: commentID, Body: "My body is great", PostID: postID}
				})

			postRepo, ok := h.GetRepositoryByType(reflect.TypeOf(PostSDK{})).(*MockRepository)
			require.True(t, ok)

			postRepo.On("Get", mock.AnythingOfType("*jsonapi.Scope")).Once().Return(nil).Run(func(args mock.Arguments) {
				scope, ok := args.Get(0).(*Scope)
				require.True(t, ok)

				scope.Value = &PostSDK{ID: postID}
			})

			if assert.NotPanics(t, func() { h.GetRelationship(commentModel, commentModel.GetRelated).ServeHTTP(rw, req) }) {
				if assert.Equal(t, 200, rw.Code, rw.Body.String()) {
					post := &PostSDK{}
					if assert.NoError(t, h.Controller.Unmarshal(rw.Body, post)) {
						if assert.NotNil(t, post) {
							assert.Equal(t, postID, post.ID)

						}
					}
				}
			}
		},
		"RelationHasOne": func(t *testing.T) {
			var blogID int = 1
			var postID int = 15

			h := getHandler()

			rw, req := getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
			blogModel, _ := getModelAndRepo(t, h, reflect.TypeOf(BlogSDK{}))

			_, postRepo := getModelAndRepo(t, h, reflect.TypeOf(PostSDK{}))
			postRepo.On("Get", mock.Anything).Once().Return(nil).
				Run(func(args mock.Arguments) {
					scope, ok := args.Get(0).(*Scope)
					require.True(t, ok)

					if assert.Len(t, scope.ForeignKeyFilters, 1) {
						if assert.NotNil(t, scope.ForeignKeyFilters[0]) {
							assert.Equal(t, scope.Struct.foreignKeys["blog_id"], scope.ForeignKeyFilters[0].StructField)
							if assert.Len(t, scope.ForeignKeyFilters[0].Values, 1) {
								if assert.NotNil(t, scope.ForeignKeyFilters[0].Values[0]) {
									assert.Contains(t, scope.ForeignKeyFilters[0].Values[0].Values, blogID)
								}
							}
						}
					}

					assert.Len(t, scope.Fieldset, 1)
					scope.Value = &PostSDK{ID: postID, BlogID: blogID}
				})

			postRepo.On("Get", mock.Anything).Once().Return(nil).
				Run(func(args mock.Arguments) {
					scope := args.Get(0).(*Scope)

					if assert.Len(t, scope.PrimaryFilters, 1) {
						if assert.NotNil(t, scope.PrimaryFilters[0]) {
							if assert.Len(t, scope.PrimaryFilters[0].Values, 1) {
								if assert.NotNil(t, scope.PrimaryFilters[0].Values[0]) {
									assert.Contains(t, scope.PrimaryFilters[0].Values[0].Values, postID)
								}
							}
						}
					}
					scope.Value = &PostSDK{ID: postID, BlogID: blogID}
				})

			if assert.NotPanics(t, func() { h.GetRelationship(blogModel, blogModel.GetRelationship).ServeHTTP(rw, req) }) {

				if assert.Equal(t, 200, rw.Result().StatusCode, rw.Body.String()) {
					post := &PostSDK{}

					err := h.Controller.Unmarshal(rw.Body, post)
					if assert.NoError(t, err) {
						assert.Equal(t, postID, post.ID)
					}
				}

			}
		},
		"RelationHasMany": func(t *testing.T) {

		},
		"RelationsMany2Many": func(t *testing.T) {

		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}

	// h := prepareHandler(defaultLanguages, blogModels...)
	// mockRepo := &MockRepository{}
	// h.SetDefaultRepo(mockRepo)

	// blogHandler := h.ModelHandlers[reflect.TypeOf(BlogSDK{})]
	// blogHandler.GetRelationship = &Endpoint{Type: GetRelationship}

	// // Case 1:
	// // Correct Relationship
	// rw, req := getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).
	// 	Run(func(args mock.Arguments) {
	// 		arg := args.Get(0).(*Scope)
	// 		arg.Value = &BlogSDK{ID: 1, CurrentPost: &PostSDK{ID: 1}}
	// 	})

	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// // Case 2:
	// // Invalid url - Internal Error
	// rw, req = getHttpPair("GET", "/blogs/relationships/current_post", nil)
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 3:
	// // Invalid field name
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/different_post", nil)
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 400, rw.Result().StatusCode)
	// t.Log(rw.Body)

	// // Case 4:
	// // Bad languge provided
	// // rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// // req.Header.Add(headerAcceptLanguage, "invalid language name")
	// // h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// // assert.Equal(t, 400, rw.Result().StatusCode)

	// // Case 5:
	// // Error while getting from root repo
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().
	// 	Return(unidb.ErrNoResult.New())
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 404, rw.Result().StatusCode)

	// // Case 6:
	// // Error while getting relationship scope. I.e. assigned bad value type - Internal error
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
	// 	arg := args[0].(*Scope)
	// 	arg.Value = &AuthorSDK{ID: 1}
	// })
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 7:
	// // Provided empty value for the scope.Value
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
	// 	arg := args[0].(*Scope)
	// 	arg.Value = nil
	// })
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 500, rw.Result().StatusCode)

	// // Case 8:
	// // Provided nil relationship value within scope
	// rw, req = getHttpPair("GET", "/blogs/1/relationships/current_post", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
	// 	arg := args[0].(*Scope)
	// 	arg.Value = &BlogSDK{ID: 1}
	// })
	// h.GetRelationship(blogHandler, blogHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

	// authorHandler := h.ModelHandlers[reflect.TypeOf(AuthorSDK{})]
	// authorHandler.GetRelationship = &Endpoint{Type: GetRelationship}

	// // Case 9:
	// // Provided empty slice in hasMany relationship within scope
	// rw, req = getHttpPair("GET", "/authors/1/relationships/blogs", nil)
	// mockRepo.On("Get", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
	// 	arg := args[0].(*Scope)
	// 	arg.Value = &AuthorSDK{ID: 1}
	// })

	// h.GetRelationship(authorHandler, authorHandler.GetRelationship).ServeHTTP(rw, req)
	// assert.Equal(t, 200, rw.Result().StatusCode)

}
