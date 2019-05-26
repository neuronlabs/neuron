package builder

import (
	"context"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestBuildScopeMany(t *testing.T) {

	ctx := context.Background()

	t.Run("Simple", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

		u, err := url.Parse("/api/v1/blogs")
		require.NoError(t, err)

		s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
		assert.Empty(t, errs)
		assert.Nil(t, err)

		if assert.NotNil(t, s) {

			assert.NotEmpty(t, s.Fieldset())

			assert.Empty(t, s.SortFields())

			assert.Nil(t, s.Pagination())
		}
	})

	t.Run("WithIncludes", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

		u, err := url.Parse("/api/v1/blogs?include=current_post")
		require.NoError(t, err)

		scope, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
		assert.Empty(t, errs)
		assert.Nil(t, err)
		if assert.NotNil(t, scope) {
			incl := scope.IncludedScopes()
			if assert.NotEmpty(t, incl) {
				assert.Equal(t, "posts", incl[0].Struct().Collection())

			}
		}
	})

	t.Run("WithSorts", func(t *testing.T) {
		t.Run("SingleDescending", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?sort=-title")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {

				sortFields := s.SortFields()
				if assert.Equal(t, 1, len(s.SortFields())) {
					assert.Equal(t, sorts.DescendingOrder, sortFields[0].Order())
				}
			}
		})

		t.Run("SingleAscending", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?sort=title")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {

				sortFields := s.SortFields()
				if assert.Equal(t, 1, len(s.SortFields())) {
					assert.Equal(t, sorts.AscendingOrder, sortFields[0].Order())
				}
			}
		})

		t.Run("Multiple", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?sort=-title,id")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {

				sortFields := s.SortFields()
				if assert.Equal(t, 2, len(s.SortFields())) {
					assert.Equal(t, sorts.DescendingOrder, sortFields[0].Order())
					assert.Equal(t, "title", sortFields[0].StructField().ApiName())
					assert.Equal(t, sorts.AscendingOrder, sortFields[1].Order())
					assert.Equal(t, "id", sortFields[1].StructField().ApiName())
				}
			}
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		t.Run("SizeNumber", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?page[size]=4&page[number]=5")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {

				if pagination := s.Pagination(); assert.NotNil(t, pagination) {
					assert.Equal(t, 4, pagination.PageSize)
					assert.Equal(t, 5, pagination.PageNumber)
				}
			}
		})
		t.Run("LimitOffset", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?page[limit]=10&page[offset]=5")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {

				if pagination := s.Pagination(); assert.NotNil(t, pagination) {
					assert.Equal(t, 10, pagination.Limit)
					assert.Equal(t, 5, pagination.Offset)
				}
			}
		})

		t.Run("Invalid", func(t *testing.T) {

			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?page[limit]=have&page[offset]=a&page[size]=nice&page[number]=day")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})

		t.Run("MixedTypes", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?page[limit]=2&page[number]=1")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})
	})

	t.Run("Filtered", func(t *testing.T) {

		t.Run("Simple", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?filter[blogs][id][$eq]=12,55")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {
				prims := s.PrimaryFilters()

				if assert.Len(t, prims, 1) {
					ovs := prims[0].Values()

					if assert.Len(t, ovs, 1) {
						assert.Equal(t, filters.OpEqual, ovs[0].Operator())
						assert.Len(t, ovs[0].Values, 2)
					}
				}
			}
		})

		t.Run("Invalid", func(t *testing.T) {

			invRunner := func(q string) func(t *testing.T) {
				return func(t *testing.T) {
					b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
					u, err := url.Parse(q)
					require.NoError(t, err)

					_, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
					assert.Nil(t, err)
					assert.NotEmpty(t, errs)
				}
			}

			pairs := map[string]string{
				"Bracket":     "/api/v1/blogs?filter[[blogs][id][$eq]=12,55",
				"Operator":    "/api/v1/blogs?filter[blogs][id][invalid]=125",
				"Value":       "/api/v1/blogs?filter[blogs][id]=stringval",
				"NotIncluded": "/api/v1/blogs?filter[posts][id]=12&fields[blogs]=id",
			}

			for name, p := range pairs {
				t.Run(name, invRunner(p))
			}
		})

	})
	t.Run("Fieldset", func(t *testing.T) {

		t.Run("Valid", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/v1/blogs?fields[blogs]=title,posts")
			require.NoError(t, err)

			s, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {
				fieldset := s.Fieldset()
				assert.Len(t, fieldset, 3)
			}
		})

		t.Run("Invalid", func(t *testing.T) {

			invRunner := func(q string) func(t *testing.T) {
				return func(t *testing.T) {
					b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
					u, err := url.Parse(q)
					require.NoError(t, err)

					_, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
					assert.Nil(t, err)
					assert.NotEmpty(t, errs)
				}
			}

			pairs := map[string]string{
				"Bracket":        "/api/v1/blogs?fields[[blogs]=title",
				"Nested":         "/api/v1/blogs?fields[blogs][title]=now",
				"CollectionName": "/api/v1/blogs?fields[blog]=title",
				"TooManyFields":  "/api/v1/blogs?fields[blogs]=title,id,posts,comments,this-comment,some-invalid,current_post",
			}

			for name, p := range pairs {
				t.Run(name, invRunner(p))
			}
		})
	})
	t.Run("UnsupportedParameter", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
		u, err := url.Parse("/api/v1/blogs?title=name")
		require.NoError(t, err)

		_, errs, err := b.BuildScopeMany(ctx, &internal.Blog{}, u)
		assert.Nil(t, err)
		assert.NotEmpty(t, errs)
	})
}

func TestBuildScopeSingle(t *testing.T) {
	ctx := context.Background()
	t.Run("Valid", func(t *testing.T) {
		t.Run("WithId", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/55")
			require.NoError(t, err)
			s, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, 55)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {
				prims := s.PrimaryFilters()
				if assert.Len(t, prims, 1) {
					v := prims[0].Values()
					if assert.Len(t, v, 1) {
						assert.Equal(t, 55, v[0].Values[0])
					}
				}

			}
		})
		t.Run("WithoutId", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/55")
			require.NoError(t, err)
			s, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, nil)
			assert.Nil(t, err)
			assert.Empty(t, errs)

			if assert.NotNil(t, s) {
				prims := s.PrimaryFilters()
				if assert.Len(t, prims, 1) {
					v := prims[0].Values()
					if assert.Len(t, v, 1) {
						assert.Equal(t, 55, v[0].Values[0])
					}
				}
			}
		})
	})

	t.Run("WithInclude", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

		u, err := url.Parse("/api/v1/blogs/44?include=posts&fields[posts]=title")
		require.NoError(t, err)
		s, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, nil)
		assert.Nil(t, err)
		assert.Empty(t, errs)

		if assert.NotNil(t, s) {
			prims := s.PrimaryFilters()
			if assert.Len(t, prims, 1) {
				v := prims[0].Values()
				if assert.Len(t, v, 1) {
					assert.Equal(t, 44, v[0].Values[0])
				}
			}

			includes := s.IncludedScopes()
			if assert.Len(t, includes, 1) {
				pi := includes[0]
				assert.Equal(t, "posts", pi.Struct().Collection())

				fs := pi.Fieldset()
				assert.Len(t, fs, 2)
			}
		}
	})

	t.Run("WithNestedInclude", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

		u, err := url.Parse("/api/v1/blogs/44?include=posts,current_post.latest_comment&fields[posts]=title")
		require.NoError(t, err)
		s, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, nil)
		assert.Nil(t, err)
		assert.Empty(t, errs)

		if assert.NotNil(t, s) {
			prims := s.PrimaryFilters()
			if assert.Len(t, prims, 1) {
				v := prims[0].Values()
				if assert.Len(t, v, 1) {
					assert.Equal(t, 44, v[0].Values[0])
				}
			}

			includes := s.IncludedScopes()
			if assert.Len(t, includes, 2) {

			}
		}
	})

	t.Run("FilterIncluded", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

		u, err := url.Parse("/api/blogs/123?filter[posts][id]=1&include=current_post")
		require.NoError(t, err)
		_, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, nil)
		assert.Nil(t, err)
		assert.Empty(t, errs)
	})
	t.Run("Invalid", func(t *testing.T) {
		invRunner := func(q string) func(t *testing.T) {
			return func(t *testing.T) {
				b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})

				u, err := url.Parse(q)
				require.NoError(t, err)
				_, errs, err := b.BuildScopeSingle(ctx, &internal.Blog{}, u, nil)
				assert.Nil(t, err)
				assert.NotEmpty(t, errs)
			}
		}

		pairs := map[string]string{
			"BadId":                    "/api/blogs/bad-id",
			"Include":                  "/api/blogs/44?include=invalid",
			"IncludeFieldset":          "/api/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts",
			"Brackets":                 "/api/blogs/44?include=posts&fields[blogs]]=posts",
			"DuplicatedFields":         "/api/blogs/44?include=posts&fields[blogs]=title,posts&fields[blogs]=posts",
			"FieldsetCollection":       "/api/blogs/44?include=posts&fields[postis]=title",
			"QueryParameter":           "/api/blogs/123?title=some-title",
			"FilterCollection":         "/api/blogs/123?filter[postis]",
			"FilterCollectionNotMatch": "/api/blogs/123?filter[comments]",
		}

		for name, p := range pairs {
			t.Run(name, invRunner(p))
		}
	})
}

func TestBuildScopeRelated(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
		u, err := url.Parse("/api/blogs/1/posts")
		require.NoError(t, err)

		s, errs, err := b.BuildScopeRelated(ctx, &internal.Blog{}, u)
		assert.Nil(t, err)
		assert.Empty(t, errs)

		if assert.NotNil(t, s) {
			assert.True(t, s.IsRoot())
			incFields := s.IncludedFields()
			if assert.Len(t, incFields, 1) {
				assert.Equal(t, "posts", incFields[0].Relationship().Struct().Collection())
				assert.NotEqual(t, 1, len(incFields[0].Scope.Fieldset()))
			}
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NonExistingField", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/1/invalid_field")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeRelated(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})
		t.Run("Attribute", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/1/title")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeRelated(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})
	})
}

func TestBuildScopeRelationship(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
		u, err := url.Parse("/api/blogs/1/relationships/posts")
		require.NoError(t, err)
		s, errs, err := b.BuildScopeRelationship(ctx, &internal.Blog{}, u)
		assert.Nil(t, err)
		assert.Empty(t, errs)

		if assert.NotNil(t, s) {
			assert.True(t, s.IsRoot())
			incFields := s.IncludedFields()
			if assert.Len(t, incFields, 1) {
				assert.Equal(t, "posts", incFields[0].Relationship().Struct().Collection())
				assert.Len(t, incFields[0].Scope.Fieldset(), 1)
			}
		}
	})

	// req = httptest.NewRequest("GET", "/api/v1/blogs/1/relationships/invalid_field", nil)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NonExistingField", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/1/relationships/invalid_field")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeRelationship(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})
		t.Run("Attribute", func(t *testing.T) {
			b := DefaultJSONAPIWithModels(t, &internal.Blog{}, &internal.Comment{}, &internal.Post{})
			u, err := url.Parse("/api/blogs/1/relationships/title")
			require.NoError(t, err)

			_, errs, err := b.BuildScopeRelationship(ctx, &internal.Blog{}, u)
			assert.Nil(t, err)
			assert.NotEmpty(t, errs)
		})
	})
}
