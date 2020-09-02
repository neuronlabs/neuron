package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestFind(t *testing.T) {
	m := mapping.New()
	err := m.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithDefaultRepository(repo))
	require.NoError(t, err)

	mStruct, err := m.ModelStruct(&testmodels.Blog{})
	require.NoError(t, err)

	repo.OnFind(func(ctx context.Context, s *query.Scope) error {
		assert.Equal(t, mStruct, s.ModelStruct)
		// Filter.
		if assert.Len(t, s.Filters, 1) {
			sf, ok := s.Filters[0].(filter.Simple)
			require.True(t, ok)
			assert.Equal(t, mStruct.Primary(), sf.StructField)
			assert.Equal(t, filter.OpGreaterThan, sf.Operator)
			if assert.Len(t, sf.Values, 1) {
				assert.Equal(t, 10, sf.Values[0])
			}
		}
		// Sort
		if assert.Len(t, s.SortingOrder, 1) {
			so, ok := s.SortingOrder[0].(query.SortField)
			require.True(t, ok)
			assert.Equal(t, mStruct.Primary(), so.StructField)
			assert.Equal(t, query.DescendingOrder, so.SortOrder)
		}
		// Fieldset
		if assert.Len(t, s.FieldSets, 1) {
			fieldSet := s.FieldSets[0]
			if assert.Len(t, fieldSet, 2) {
				assert.Equal(t, fieldSet[0], mStruct.Primary())
				assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
			}
		}
		s.Models = []mapping.Model{
			&testmodels.Blog{ID: 12, Title: "Simple title"},
			&testmodels.Blog{ID: 11, Title: "Second title"},
		}
		return nil
	})

	models, err := db.Query(mStruct).
		Where("ID > ?", 10).
		Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
		OrderBy(query.SortField{StructField: mStruct.Primary(), SortOrder: query.DescendingOrder}).
		Find()
	require.NoError(t, err)

	require.Len(t, models, 2)
	b1, ok := models[0].(*testmodels.Blog)
	require.True(t, ok)
	assert.Equal(t, 12, b1.ID)
	assert.Equal(t, "Simple title", b1.Title)

	b2, ok := models[1].(*testmodels.Blog)
	require.True(t, ok)
	assert.Equal(t, 11, b2.ID)
	assert.Equal(t, "Second title", b2.Title)

	t.Run("WithRelationFilter", func(t *testing.T) {
		t.Run("HasMany", func(t *testing.T) {
			// At first the function would like to find the posts that matches given query.
			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				postMStruct, err := m.ModelStruct(&testmodels.Post{})
				require.NoError(t, err)

				require.Equal(t, postMStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, postMStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)

					if assert.Len(t, sf.Values, 2) {
						assert.ElementsMatch(t, sf.Values, []interface{}{10, 12})
					}
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, postMStruct.MustFieldByName("BlogID"), fieldSet[0])
					}
				}
				s.Models = []mapping.Model{&testmodels.Post{BlogID: 20}, &testmodels.Post{BlogID: 21}}
				return nil
			})

			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				require.Equal(t, mStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{20, 21}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), mStruct.MustFieldByName("Title")}, fieldSet)
				}
				s.Models = []mapping.Model{
					&testmodels.Blog{ID: 20, Title: "twenty"},
					&testmodels.Blog{ID: 21, Title: "twenty one"},
				}
				return nil
			})

			models, err := db.Query(mStruct).
				Where("Posts.ID IN ?", 10, 12).
				Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
				Find()
			require.NoError(t, err)

			if assert.Len(t, models, 2) {
				b1, ok := models[0].(*testmodels.Blog)
				require.True(t, ok)

				assert.Equal(t, 20, b1.ID)
				assert.Equal(t, "twenty", b1.Title)

				b2, ok := models[1].(*testmodels.Blog)
				require.True(t, ok)

				assert.Equal(t, 21, b2.ID)
				assert.Equal(t, "twenty one", b2.Title)
			}

			t.Run("ForeignKeyOnly", func(t *testing.T) {
				repo.OnFind(func(ctx context.Context, s *query.Scope) error {
					require.Equal(t, mStruct, s.ModelStruct)
					if assert.Len(t, s.Filters, 1) {
						sf, ok := s.Filters[0].(filter.Simple)
						require.True(t, ok)

						assert.Equal(t, mStruct.Primary(), sf.StructField)
						assert.Equal(t, filter.OpIn, sf.Operator)
						assert.ElementsMatch(t, []interface{}{20, 21}, sf.Values)
					}
					if assert.Len(t, s.FieldSets, 1) {
						fieldSet := s.FieldSets[0]
						assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), mStruct.MustFieldByName("Title")}, fieldSet)
					}
					s.Models = []mapping.Model{
						&testmodels.Blog{ID: 20, Title: "twenty"},
						&testmodels.Blog{ID: 21, Title: "twenty one"},
					}
					return nil
				})

				models, err := db.Query(mStruct).
					Where("Posts.BlogID IN ?", 20, 21).
					Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
					Find()
				require.NoError(t, err)

				if assert.Len(t, models, 2) {
					b1, ok := models[0].(*testmodels.Blog)
					require.True(t, ok)

					assert.Equal(t, 20, b1.ID)
					assert.Equal(t, "twenty", b1.Title)

					b2, ok := models[1].(*testmodels.Blog)
					require.True(t, ok)

					assert.Equal(t, 21, b2.ID)
					assert.Equal(t, "twenty one", b2.Title)
				}
			})
		})

		t.Run("BelongsTo", func(t *testing.T) {
			// At first the function would like to find the posts that matches given query.
			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				postMStruct, err := m.ModelStruct(&testmodels.Post{})
				require.NoError(t, err)

				require.Equal(t, postMStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, postMStruct.MustFieldByName("Body"), sf.StructField)
					assert.Equal(t, filter.OpContains, sf.Operator)

					if assert.Len(t, sf.Values, 1) {
						assert.Equal(t, "relation filter", sf.Values[0])
					}
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, postMStruct.Primary(), fieldSet[0])
					}
				}
				s.Models = []mapping.Model{&testmodels.Post{ID: 10}, &testmodels.Post{ID: 11}}
				return nil
			})

			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				require.Equal(t, mStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.MustFieldByName("CurrentPostID"), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{uint64(10), uint64(11)}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), mStruct.MustFieldByName("Title")}, fieldSet)
				}
				s.Models = []mapping.Model{
					&testmodels.Blog{ID: 20, Title: "twenty"},
					&testmodels.Blog{ID: 21, Title: "twenty one"},
				}
				return nil
			})

			models, err := db.Query(mStruct).
				Where("CurrentPost.Body CONTAINS ?", "relation filter").
				Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
				Find()
			require.NoError(t, err)

			if assert.Len(t, models, 2) {
				b1, ok := models[0].(*testmodels.Blog)
				require.True(t, ok)

				assert.Equal(t, 20, b1.ID)
				assert.Equal(t, "twenty", b1.Title)

				b2, ok := models[1].(*testmodels.Blog)
				require.True(t, ok)

				assert.Equal(t, 21, b2.ID)
				assert.Equal(t, "twenty one", b2.Title)
			}

			t.Run("ForeignKeyOnly", func(t *testing.T) {
				repo.OnFind(func(ctx context.Context, s *query.Scope) error {
					require.Equal(t, mStruct, s.ModelStruct)
					if assert.Len(t, s.Filters, 1) {
						sf, ok := s.Filters[0].(filter.Simple)
						require.True(t, ok)

						assert.Equal(t, mStruct.MustFieldByName("CurrentPostID"), sf.StructField)
						assert.Equal(t, filter.OpIn, sf.Operator)
						assert.ElementsMatch(t, []interface{}{10, 11}, sf.Values)
					}
					if assert.Len(t, s.FieldSets, 1) {
						fieldSet := s.FieldSets[0]
						assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), mStruct.MustFieldByName("Title")}, fieldSet)
					}
					s.Models = []mapping.Model{
						&testmodels.Blog{ID: 20, Title: "twenty"},
						&testmodels.Blog{ID: 21, Title: "twenty one"},
					}
					return nil
				})
				models, err := db.Query(mStruct).
					Where("CurrentPost.ID IN ?", 10, 11).
					Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
					Find()
				require.NoError(t, err)

				if assert.Len(t, models, 2) {
					b1, ok := models[0].(*testmodels.Blog)
					require.True(t, ok)

					assert.Equal(t, 20, b1.ID)
					assert.Equal(t, "twenty", b1.Title)

					b2, ok := models[1].(*testmodels.Blog)
					require.True(t, ok)

					assert.Equal(t, 21, b2.ID)
					assert.Equal(t, "twenty one", b2.Title)
				}
			})
		})

		t.Run("HasOne", func(t *testing.T) {
			hom := m.MustModelStruct(&testmodels.HasOneModel{})
			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				fm := m.MustModelStruct(&testmodels.ForeignModel{})
				require.Equal(t, fm, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)
					assert.Equal(t, fm.Primary(), sf.StructField)
					assert.Equal(t, filter.OpEqual, sf.Operator)
					assert.ElementsMatch(t, []interface{}{10}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fm.MustFieldByName("ForeignKey"), fieldSet[0])
					}
				}
				s.Models = []mapping.Model{&testmodels.ForeignModel{ForeignKey: 3}}
				return nil
			})
			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				require.Equal(t, hom, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, hom.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{3}, sf.Values)
				}
				s.Models = []mapping.Model{&testmodels.HasOneModel{ID: 3}}
				return nil
			})

			models, err := db.Query(hom).
				Where("HasOne.ID = ?", 10).
				Find()
			require.NoError(t, err)

			if assert.Len(t, models, 1) {
				h1, ok := models[0].(*testmodels.HasOneModel)
				require.True(t, ok)

				assert.Equal(t, 3, h1.ID)
			}

			t.Run("ForeignKeyOnly", func(t *testing.T) {
				repo.OnFind(func(_ context.Context, s *query.Scope) error {
					require.Equal(t, hom, s.ModelStruct)

					if assert.Len(t, s.Filters, 1) {
						sf, ok := s.Filters[0].(filter.Simple)
						require.True(t, ok)

						assert.Equal(t, hom.Primary(), sf.StructField)
						assert.Equal(t, filter.OpIn, sf.Operator)
						assert.ElementsMatch(t, []interface{}{10, 11}, sf.Values)
					}
					s.Models = []mapping.Model{&testmodels.HasOneModel{ID: 20}, &testmodels.HasOneModel{ID: 25}}
					return nil
				})
				models, err := db.Query(hom).
					Where("HasOne.ForeignKey IN ?", 10, 11).
					Find()
				require.NoError(t, err)

				if assert.Len(t, models, 2) {
					h1, ok := models[0].(*testmodels.HasOneModel)
					require.True(t, ok)
					assert.Equal(t, h1.ID, 20)

					h2, ok := models[1].(*testmodels.HasOneModel)
					require.True(t, ok)
					assert.Equal(t, h2.ID, 25)
				}
			})
		})

		t.Run("ManyToMany", func(t *testing.T) {
			tm := m.MustModelStruct(&testmodels.ManyToManyModel{})

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				relModel := m.MustModelStruct(&testmodels.RelatedModel{})
				require.Equal(t, relModel, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, relModel.MustFieldByName("FloatField"), sf.StructField)
					assert.Equal(t, filter.OpGreaterEqual, sf.Operator)
					assert.ElementsMatch(t, []interface{}{10.50}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relModel.Primary(), fieldSet[0])
					}
				}
				s.Models = []mapping.Model{
					&testmodels.RelatedModel{ID: 12},
					&testmodels.RelatedModel{ID: 13},
				}
				return nil
			})

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				jm := m.MustModelStruct(&testmodels.JoinModel{})
				require.Equal(t, jm, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, jm.MustFieldByName("MtMForeignKey"), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 13}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, jm.MustFieldByName("ForeignKey"), fieldSet[0])
					}
				}
				s.Models = []mapping.Model{
					&testmodels.JoinModel{ForeignKey: 22},
					&testmodels.JoinModel{ForeignKey: 44},
				}
				return nil
			})

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				require.Equal(t, tm, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, tm.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{22, 44}, sf.Values)
				}
				s.Models = []mapping.Model{
					&testmodels.ManyToManyModel{ID: 22},
					&testmodels.ManyToManyModel{ID: 44},
				}
				return nil
			})

			models, err := db.Query(tm).
				Where("Many2Many.FloatField >= ?", 10.50).
				Find()
			require.NoError(t, err)

			if assert.Len(t, models, 2) {
				m1, ok := models[0].(*testmodels.ManyToManyModel)
				require.True(t, ok)
				assert.Equal(t, 22, m1.ID)

				m2, ok := models[1].(*testmodels.ManyToManyModel)
				require.True(t, ok)
				assert.Equal(t, 44, m2.ID)
			}

			t.Run("OnlyForeign", func(t *testing.T) {
				// This should query join model and then the model by itself.
				repo.OnFind(func(_ context.Context, s *query.Scope) error {
					jm := m.MustModelStruct(&testmodels.JoinModel{})
					require.Equal(t, jm, s.ModelStruct)

					if assert.Len(t, s.Filters, 1) {
						sf, ok := s.Filters[0].(filter.Simple)
						require.True(t, ok)

						assert.Equal(t, jm.MustFieldByName("MtMForeignKey"), sf.StructField)
						assert.Equal(t, filter.OpIn, sf.Operator)
						assert.ElementsMatch(t, []interface{}{5, 11}, sf.Values)
					}
					if assert.Len(t, s.FieldSets, 1) {
						fieldSet := s.FieldSets[0]
						if assert.Len(t, fieldSet, 1) {
							assert.ElementsMatch(t, mapping.FieldSet{jm.MustFieldByName("ForeignKey")}, fieldSet)
						}
					}
					s.Models = []mapping.Model{
						&testmodels.JoinModel{ForeignKey: 10},
						&testmodels.JoinModel{ForeignKey: 15},
					}
					return nil
				})

				repo.OnFind(func(_ context.Context, s *query.Scope) error {
					require.Equal(t, tm, s.ModelStruct)
					if assert.Len(t, s.Filters, 1) {
						sf, ok := s.Filters[0].(filter.Simple)
						require.True(t, ok)

						assert.Equal(t, tm.Primary(), sf.StructField)
						assert.Equal(t, filter.OpIn, sf.Operator)
						assert.ElementsMatch(t, []interface{}{10, 15}, sf.Values)
					}
					s.Models = []mapping.Model{
						&testmodels.ManyToManyModel{ID: 10},
						&testmodels.ManyToManyModel{ID: 15},
					}
					return nil
				})
				models, err := db.Query(tm).
					Where("Many2Many.ID IN ?", 5, 11).
					Find()
				require.NoError(t, err)

				if assert.Len(t, models, 2) {
					m1, ok := models[0].(*testmodels.ManyToManyModel)
					require.True(t, ok)
					assert.Equal(t, 10, m1.ID)

					m2, ok := models[1].(*testmodels.ManyToManyModel)
					require.True(t, ok)
					assert.Equal(t, 15, m2.ID)
				}
			})
		})
	})
}

func TestGet(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(
		WithDefaultRepository(repo),
		WithModelMap(mm),
	)
	require.NoError(t, err)

	mStruct, err := mm.ModelStruct(&testmodels.Blog{})
	require.NoError(t, err)

	repo.OnFind(func(ctx context.Context, s *query.Scope) error {
		assert.Equal(t, mStruct, s.ModelStruct)
		// Filter.
		if assert.Len(t, s.Filters, 1) {
			sf, ok := s.Filters[0].(filter.Simple)
			require.True(t, ok)
			assert.Equal(t, mStruct.Primary(), sf.StructField)
			assert.Equal(t, filter.OpGreaterThan, sf.Operator)
			if assert.Len(t, sf.Values, 1) {
				assert.Equal(t, 10, sf.Values[0])
			}
		}
		// Sort
		if assert.Len(t, s.SortingOrder, 1) {
			so, ok := s.SortingOrder[0].(query.SortField)
			require.True(t, ok)
			assert.Equal(t, mStruct.Primary(), so.StructField)
			assert.Equal(t, query.DescendingOrder, so.SortOrder)
		}
		// Fieldset
		if assert.Len(t, s.FieldSets, 1) {
			fieldSet := s.FieldSets[0]
			if assert.Len(t, fieldSet, 2) {
				assert.Equal(t, fieldSet[0], mStruct.Primary())
				assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
			}
		}
		if assert.NotNil(t, s.Pagination) {
			assert.Equal(t, int64(1), s.Pagination.Limit)
			assert.Equal(t, int64(0), s.Pagination.Offset)
		}
		s.Models = []mapping.Model{
			&testmodels.Blog{ID: 12, Title: "Simple title"},
		}
		return nil
	})

	model, err := db.Query(mStruct).
		Where("ID > ?", 10).
		Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
		OrderBy(query.SortField{StructField: mStruct.Primary(), SortOrder: query.DescendingOrder}).
		Get()
	require.NoError(t, err)

	b1, ok := model.(*testmodels.Blog)
	require.True(t, ok)
	assert.Equal(t, 12, b1.ID)
	assert.Equal(t, "Simple title", b1.Title)
}
