package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestInclude(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	db := New(c)

	t.Run("HasMany", func(t *testing.T) {
		mStruct, err := c.ModelStruct(&testmodels.HasManyModel{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("HasMany")
		require.True(t, ok)

		relatedModelStruct := relation.Relationship().RelatedModelStruct()
		m1 := &testmodels.HasManyModel{ID: 1}
		m2 := &testmodels.HasManyModel{ID: 2}

		repo.OnFind(func(_ context.Context, s *query.Scope) error {
			require.Equal(t, relatedModelStruct, s.ModelStruct)
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)

				assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				assert.ElementsMatch(t, []interface{}{1, 2}, sf.Values)
			}
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, relatedModelStruct.Primary(), fieldSet[0])
					assert.Equal(t, relation.Relationship().ForeignKey(), fieldSet[1])
				}
			}
			s.Models = []mapping.Model{
				&testmodels.ForeignModel{ID: 4, ForeignKey: 1},
				&testmodels.ForeignModel{ID: 5, ForeignKey: 2},
				&testmodels.ForeignModel{ID: 6, ForeignKey: 1},
				&testmodels.ForeignModel{ID: 7, ForeignKey: 2},
				&testmodels.ForeignModel{ID: 8, ForeignKey: 2},
			}
			return nil
		})
		err = db.IncludeRelations(context.Background(), mStruct, []mapping.Model{m1, m2}, relation)
		require.NoError(t, err)

		if assert.Len(t, m1.HasMany, 2) {
			relations, err := m1.GetRelationModels(relation)
			require.NoError(t, err)

			assert.ElementsMatch(t, []interface{}{4, 6}, mapping.Models(relations).PrimaryKeyValues())
		}
		if assert.Len(t, m2.HasMany, 3) {
			relations, err := m2.GetRelationModels(relation)
			require.NoError(t, err)

			assert.ElementsMatch(t, []interface{}{5, 7, 8}, mapping.Models(relations).PrimaryKeyValues())
		}
	})

	t.Run("HasOne", func(t *testing.T) {
		mStruct, err := c.ModelStruct(&testmodels.HasOneModel{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("HasOne")
		require.True(t, ok)

		relatedModelStruct := relation.Relationship().RelatedModelStruct()
		m1 := &testmodels.HasOneModel{ID: 1}
		m2 := &testmodels.HasOneModel{ID: 2}

		repo.OnFind(func(_ context.Context, s *query.Scope) error {
			require.Equal(t, relatedModelStruct, s.ModelStruct)
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)

				assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				assert.ElementsMatch(t, []interface{}{1, 2}, sf.Values)
			}
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, relatedModelStruct.Primary(), fieldSet[0])
					assert.Equal(t, relation.Relationship().ForeignKey(), fieldSet[1])
				}
			}
			s.Models = []mapping.Model{
				&testmodels.ForeignModel{ID: 4, ForeignKey: 1},
				&testmodels.ForeignModel{ID: 5, ForeignKey: 2},
			}
			return nil
		})
		err = db.IncludeRelations(context.Background(), mStruct, []mapping.Model{m1, m2}, relation)
		require.NoError(t, err)

		if assert.NotNil(t, m1.HasOne) {
			assert.Equal(t, 4, m1.HasOne.ID)
		}
		if assert.NotNil(t, m2.HasOne) {
			assert.Equal(t, 5, m2.HasOne.ID)
		}
	})

	t.Run("BelongsTo", func(t *testing.T) {
		mStruct, err := c.ModelStruct(&testmodels.Blog{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("CurrentPost")
		require.True(t, ok)

		relatedModelStruct := relation.Relationship().RelatedModelStruct()
		m1 := &testmodels.Blog{ID: 1, CurrentPostID: 5}
		m2 := &testmodels.Blog{ID: 2, CurrentPostID: 10}
		// The third one doesn't have current post id.
		m3 := &testmodels.Blog{ID: 3}

		repo.OnFind(func(_ context.Context, s *query.Scope) error {
			require.Equal(t, relatedModelStruct, s.ModelStruct)
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)

				assert.Equal(t, relatedModelStruct.Primary(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				assert.ElementsMatch(t, []interface{}{m1.CurrentPostID, m2.CurrentPostID}, sf.Values)
			}
			if assert.Len(t, s.FieldSets, 1) {
				assert.ElementsMatch(t, s.FieldSets[0], relatedModelStruct.Fields())
			}
			s.Models = []mapping.Model{
				&testmodels.Post{ID: 5, Title: "Title 1", Body: "Body 1", BlogID: 1},
				&testmodels.Post{ID: 10, Title: "Title 2", Body: "Body 2", BlogID: 2},
			}
			return nil
		})
		err = db.IncludeRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation)
		require.NoError(t, err)

		if assert.NotNil(t, m1.CurrentPost) {
			assert.Equal(t, uint64(5), m1.CurrentPost.ID)
		}
		if assert.NotNil(t, m2.CurrentPost) {
			assert.Equal(t, uint64(10), m2.CurrentPost.ID)
		}
		assert.Nil(t, m3.CurrentPost)
	})

	t.Run("Many2Many", func(t *testing.T) {
		t.Run("FullFieldSet", func(t *testing.T) {
			mStruct, err := c.ModelStruct(&testmodels.ManyToManyModel{})
			require.NoError(t, err)

			relation, ok := mStruct.RelationByName("Many2Many")
			require.True(t, ok)

			relatedModelStruct := relation.Relationship().RelatedModelStruct()
			joinModelStruct := relation.Relationship().JoinModel()
			m1 := &testmodels.ManyToManyModel{ID: 1}
			m2 := &testmodels.ManyToManyModel{ID: 2}
			m3 := &testmodels.ManyToManyModel{ID: 3}

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				require.Equal(t, joinModelStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{1, 2, 3}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 2) {
						assert.ElementsMatch(t, mapping.FieldSet{relation.Relationship().ForeignKey(), relation.Relationship().ManyToManyForeignKey()}, fieldSet)
					}
				}
				s.Models = []mapping.Model{
					&testmodels.JoinModel{ForeignKey: 1, MtMForeignKey: 1},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 2},
					&testmodels.JoinModel{ForeignKey: 1, MtMForeignKey: 3},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 4},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 5},
				}
				return nil
			})

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				require.Equal(t, relatedModelStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, relatedModelStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					assert.ElementsMatch(t, relatedModelStruct.Fields(), s.FieldSets[0])
				}
				s.Models = []mapping.Model{
					&testmodels.RelatedModel{ID: 1, FloatField: 12.3},
					&testmodels.RelatedModel{ID: 2, FloatField: 12.4},
					&testmodels.RelatedModel{ID: 3, FloatField: 12.5},
					&testmodels.RelatedModel{ID: 4, FloatField: 12.6},
					&testmodels.RelatedModel{ID: 5, FloatField: 12.7},
				}
				return nil
			})

			err = db.IncludeRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation)
			require.NoError(t, err)

			if assert.Len(t, m1.Many2Many, 2) {
				relations, err := m1.GetRelationModels(relation)
				require.NoError(t, err)

				assert.ElementsMatch(t, []interface{}{1, 3}, mapping.Models(relations).PrimaryKeyValues())
			}
			if assert.Len(t, m2.Many2Many, 3) {
				relations, err := m2.GetRelationModels(relation)
				require.NoError(t, err)

				assert.ElementsMatch(t, []interface{}{2, 4, 5}, mapping.Models(relations).PrimaryKeyValues())
			}
			assert.Len(t, m3.Many2Many, 0)
		})

		t.Run("OnlyPrimes", func(t *testing.T) {
			mStruct, err := c.ModelStruct(&testmodels.ManyToManyModel{})
			require.NoError(t, err)

			relation, ok := mStruct.RelationByName("Many2Many")
			require.True(t, ok)

			relatedModelStruct := relation.Relationship().RelatedModelStruct()
			joinModelStruct := relation.Relationship().JoinModel()
			m1 := &testmodels.ManyToManyModel{ID: 1}
			m2 := &testmodels.ManyToManyModel{ID: 2}
			m3 := &testmodels.ManyToManyModel{ID: 3}

			repo.OnFind(func(_ context.Context, s *query.Scope) error {
				require.Equal(t, joinModelStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{1, 2, 3}, sf.Values)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 2) {
						assert.ElementsMatch(t, mapping.FieldSet{relation.Relationship().ForeignKey(), relation.Relationship().ManyToManyForeignKey()}, fieldSet)
					}
				}
				s.Models = []mapping.Model{
					&testmodels.JoinModel{ForeignKey: 1, MtMForeignKey: 1},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 2},
					&testmodels.JoinModel{ForeignKey: 1, MtMForeignKey: 3},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 4},
					&testmodels.JoinModel{ForeignKey: 2, MtMForeignKey: 5},
				}
				return nil
			})

			err = db.IncludeRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation, relatedModelStruct.Primary())
			require.NoError(t, err)

			if assert.Len(t, m1.Many2Many, 2) {
				relations, err := m1.GetRelationModels(relation)
				require.NoError(t, err)

				assert.ElementsMatch(t, []interface{}{1, 3}, mapping.Models(relations).PrimaryKeyValues())
			}
			if assert.Len(t, m2.Many2Many, 3) {
				relations, err := m2.GetRelationModels(relation)
				require.NoError(t, err)

				assert.ElementsMatch(t, []interface{}{2, 4, 5}, mapping.Models(relations).PrimaryKeyValues())
			}
			assert.Len(t, m3.Many2Many, 0)
		})
	})
}

func TestIncludeOnFind(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	db := New(c)

	mStruct, err := c.ModelStruct(&testmodels.HasManyModel{})
	require.NoError(t, err)

	relation, ok := mStruct.RelationByName("HasMany")
	require.True(t, ok)

	relatedModelStruct := relation.Relationship().RelatedModelStruct()

	repo.OnFind(func(_ context.Context, s *query.Scope) error {
		require.Equal(t, mStruct, s.ModelStruct)
		s.Models = mapping.Models{
			&testmodels.HasManyModel{ID: 1},
			&testmodels.HasManyModel{ID: 2},
		}
		return nil
	})

	repo.OnFind(func(_ context.Context, s *query.Scope) error {
		require.Equal(t, relatedModelStruct, s.ModelStruct)
		if assert.Len(t, s.Filters, 1) {
			sf, ok := s.Filters[0].(filter.Simple)
			require.True(t, ok)

			assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
			assert.Equal(t, filter.OpIn, sf.Operator)
			assert.ElementsMatch(t, []interface{}{1, 2}, sf.Values)
		}
		if assert.Len(t, s.FieldSets, 1) {
			fieldSet := s.FieldSets[0]
			if assert.Len(t, fieldSet, 2) {
				assert.Equal(t, relatedModelStruct.Primary(), fieldSet[0])
				assert.Equal(t, relation.Relationship().ForeignKey(), fieldSet[1])
			}
		}
		s.Models = []mapping.Model{
			&testmodels.ForeignModel{ID: 4, ForeignKey: 1},
			&testmodels.ForeignModel{ID: 5, ForeignKey: 2},
			&testmodels.ForeignModel{ID: 6, ForeignKey: 1},
			&testmodels.ForeignModel{ID: 7, ForeignKey: 2},
			&testmodels.ForeignModel{ID: 8, ForeignKey: 2},
		}
		return nil
	})

	models, err := db.Query(mStruct).Include(relation).Find()
	require.NoError(t, err)

	if assert.Len(t, models, 2) {
		m1, ok := models[0].(*testmodels.HasManyModel)
		require.True(t, ok)
		if assert.Len(t, m1.HasMany, 2) {
			relations, err := m1.GetRelationModels(relation)
			require.NoError(t, err)

			assert.ElementsMatch(t, []interface{}{4, 6}, mapping.Models(relations).PrimaryKeyValues())
		}
		m2, ok := models[1].(*testmodels.HasManyModel)
		require.True(t, ok)
		if assert.Len(t, m2.HasMany, 3) {
			relations, err := m2.GetRelationModels(relation)
			require.NoError(t, err)

			assert.ElementsMatch(t, []interface{}{5, 7, 8}, mapping.Models(relations).PrimaryKeyValues())
		}
	}
}
