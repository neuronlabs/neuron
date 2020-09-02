package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestAddRelations(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithModelMap(mm), WithDefaultRepository(repo))
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("HasMany", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasManyModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasMany")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()

		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasManyModel{ID: 1}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 2) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs1, r1)
					assert.Equal(t, model.ID, rs1.ForeignKey)

					rs2, ok := s.Models[1].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs2, r2)
					assert.Equal(t, model.ID, rs2.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 2, nil
			})

			err = db.AddRelations(ctx, model, relationField, r1, r2)
			require.NoError(t, err)
		})

		t.Run("NoID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasManyModel{}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.AddRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)

			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("NoRelationID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasManyModel{ID: 4}
			r1 := &testmodels.ForeignModel{}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.AddRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})
	})

	t.Run("HasOne", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasOneModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasOne")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()

		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasOneModel{ID: 1}
			r1 := &testmodels.ForeignModel{ID: 3}

			repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs1, r1)
					assert.Equal(t, model.ID, rs1.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 1, nil
			})

			err = db.AddRelations(ctx, model, relationField, r1)
			require.NoError(t, err)
		})

		t.Run("NoID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{}
			r1 := &testmodels.ForeignModel{ID: 3}

			err = db.AddRelations(ctx, model, relationField, r1)
			require.Error(t, err)

			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("NoRelationID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{ID: 4}
			r1 := &testmodels.ForeignModel{}

			err = db.AddRelations(ctx, model, relationField, r1)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("MoreThanOne", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{ID: 4}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.AddRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidInput))
		})
	})

	t.Run("Many2Many", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.ManyToManyModel{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("Many2Many")
		require.True(t, ok)

		model := &testmodels.ManyToManyModel{ID: 2}
		r1 := &testmodels.RelatedModel{ID: 3}
		r2 := &testmodels.RelatedModel{ID: 4}

		joinModel := relation.Relationship().JoinModel()

		repo.OnInsert(func(_ context.Context, s *query.Scope) error {
			require.Equal(t, joinModel, s.ModelStruct)

			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.ElementsMatch(t, mapping.FieldSet{relation.Relationship().ForeignKey(), relation.Relationship().ManyToManyForeignKey()}, fieldSet)
				}
			}
			if assert.Len(t, s.Models, 2) {
				i1, ok := s.Models[0].(*testmodels.JoinModel)
				require.True(t, ok)

				i1.ID = 5
				assert.Equal(t, i1.ForeignKey, 2)
				assert.Equal(t, i1.MtMForeignKey, 3)

				i2, ok := s.Models[1].(*testmodels.JoinModel)
				require.True(t, ok)

				i2.ID = 4
				assert.Equal(t, i2.ForeignKey, 2)
				assert.Equal(t, i2.MtMForeignKey, 4)
			}
			return nil
		})

		err = db.AddRelations(ctx, model, relation, r1, r2)
		require.NoError(t, err)
	})

	t.Run("BelongsTo", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.Post{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("LatestComment")
		require.True(t, ok)

		post := &testmodels.Post{ID: 3}
		comment := &testmodels.Comment{ID: 10}
		repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
			require.Equal(t, mStruct, s.ModelStruct)
			if assert.Len(t, s.Models, 1) {
				p1, ok := s.Models[0].(*testmodels.Post)
				require.True(t, ok)
				assert.Equal(t, post, p1)
				assert.Equal(t, comment.ID, p1.LatestCommentID)
			}
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), relation.Relationship().ForeignKey()}, fieldSet)
				}
			}
			return 1, nil
		})
		err = db.AddRelations(ctx, post, relation, comment)
		require.NoError(t, err)

		assert.Equal(t, post.LatestCommentID, comment.ID)
	})
}

func TestClearRelations(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithDefaultRepository(repo), WithModelMap(mm))
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("HasMany", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasManyModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasMany")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()

		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasManyModel{ID: 1}

			repo.OnUpdate(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, 0, rs1.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 2, nil
			})

			res, err := db.ClearRelations(ctx, model, relationField)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})
	})

	t.Run("HasOne", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasOneModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasOne")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()
		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasOneModel{ID: 1}

			repo.OnUpdate(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, 0, rs1.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 2, nil
			})

			res, err := db.ClearRelations(ctx, model, relationField)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})
	})

	t.Run("Many2Many", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.ManyToManyModel{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("Many2Many")
		require.True(t, ok)

		model := &testmodels.ManyToManyModel{ID: 2}

		joinModel := relation.Relationship().JoinModel()

		repo.OnDelete(func(_ context.Context, s *query.Scope) (int64, error) {
			require.Equal(t, joinModel, s.ModelStruct)

			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)

				assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				assert.ElementsMatch(t, []interface{}{model.ID}, sf.Values)
			}
			return 3, nil
		})

		res, err := db.ClearRelations(ctx, model, relation)
		require.NoError(t, err)

		assert.Equal(t, int64(3), res)
	})

	t.Run("BelongsTo", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.Post{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("LatestComment")
		require.True(t, ok)

		post := &testmodels.Post{ID: 3, LatestCommentID: 5}
		repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
			require.Equal(t, mStruct, s.ModelStruct)
			if assert.Len(t, s.Models, 1) {
				p1, ok := s.Models[0].(*testmodels.Post)
				require.True(t, ok)
				assert.Equal(t, post, p1)
				assert.Equal(t, 0, p1.LatestCommentID)
			}
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 1) {
					assert.ElementsMatch(t, mapping.FieldSet{relation.Relationship().ForeignKey()}, fieldSet)
				}
			}
			return 1, nil
		})
		res, err := db.ClearRelations(ctx, post, relation)
		require.NoError(t, err)

		assert.Equal(t, int64(1), res)
	})
}

func TestSetRelations(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithModelMap(mm), WithDefaultRepository(repo))
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("HasMany", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasManyModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasMany")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()

		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasManyModel{ID: 1}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			repo.OnUpdate(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, relationField.Relationship().ForeignKey(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{model.ID}, sf.Values)
				}
				return 4, nil
			})

			repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 2) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs1, r1)
					assert.Equal(t, model.ID, rs1.ForeignKey)

					rs2, ok := s.Models[1].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs2, r2)
					assert.Equal(t, model.ID, rs2.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 2, nil
			})

			err = db.SetRelations(ctx, model, relationField, r1, r2)
			require.NoError(t, err)
		})

		t.Run("NoID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasManyModel{}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.SetRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)

			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("NoRelationID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasManyModel{ID: 4}
			r1 := &testmodels.ForeignModel{}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.SetRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})
	})

	t.Run("HasOne", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasOneModel{})
		require.NoError(t, err)

		relationField, ok := mStruct.RelationByName("HasOne")
		require.True(t, ok)

		relatedMStruct := relationField.Relationship().RelatedModelStruct()

		t.Run("Valid", func(t *testing.T) {
			model := &testmodels.HasOneModel{ID: 1}
			r1 := &testmodels.ForeignModel{ID: 3}

			repo.OnUpdate(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, 0, rs1.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 2, nil
			})

			repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, relatedMStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					rs1, ok := s.Models[0].(*testmodels.ForeignModel)
					require.True(t, ok)
					assert.Equal(t, rs1, r1)
					assert.Equal(t, model.ID, rs1.ForeignKey)
				}
				// Common fieldset update.
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, relationField.Relationship().ForeignKey(), fieldSet[0])
					}
				}
				return 1, nil
			})

			err = db.SetRelations(ctx, model, relationField, r1)
			require.NoError(t, err)
		})

		t.Run("NoID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{}
			r1 := &testmodels.ForeignModel{ID: 3}

			err = db.SetRelations(ctx, model, relationField, r1)
			require.Error(t, err)

			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("NoRelationID", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{ID: 4}
			r1 := &testmodels.ForeignModel{}

			err = db.SetRelations(ctx, model, relationField, r1)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidModels))
		})

		t.Run("MoreThanOne", func(t *testing.T) {
			// Root model have no ID.
			model := &testmodels.HasOneModel{ID: 4}
			r1 := &testmodels.ForeignModel{ID: 3}
			r2 := &testmodels.ForeignModel{ID: 5}

			err = db.SetRelations(ctx, model, relationField, r1, r2)
			require.Error(t, err)
			// one of the input relation models has invalid primary key value.
			assert.True(t, errors.Is(err, query.ErrInvalidInput))
		})
	})

	t.Run("Many2Many", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.ManyToManyModel{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("Many2Many")
		require.True(t, ok)

		model := &testmodels.ManyToManyModel{ID: 2}
		r1 := &testmodels.RelatedModel{ID: 3}
		r2 := &testmodels.RelatedModel{ID: 4}

		joinModel := relation.Relationship().JoinModel()

		repo.OnDelete(func(_ context.Context, s *query.Scope) (int64, error) {
			require.Equal(t, joinModel, s.ModelStruct)

			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)

				assert.Equal(t, relation.Relationship().ForeignKey(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				assert.ElementsMatch(t, []interface{}{model.ID}, sf.Values)
			}
			return 3, nil
		})

		repo.OnInsert(func(_ context.Context, s *query.Scope) error {
			require.Equal(t, joinModel, s.ModelStruct)

			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.ElementsMatch(t, mapping.FieldSet{relation.Relationship().ForeignKey(), relation.Relationship().ManyToManyForeignKey()}, fieldSet)
				}
			}
			if assert.Len(t, s.Models, 2) {
				i1, ok := s.Models[0].(*testmodels.JoinModel)
				require.True(t, ok)

				i1.ID = 5
				assert.Equal(t, i1.ForeignKey, 2)
				assert.Equal(t, i1.MtMForeignKey, 3)

				i2, ok := s.Models[1].(*testmodels.JoinModel)
				require.True(t, ok)

				i2.ID = 4
				assert.Equal(t, i2.ForeignKey, 2)
				assert.Equal(t, i2.MtMForeignKey, 4)
			}
			return nil
		})

		err = db.SetRelations(ctx, model, relation, r1, r2)
		require.NoError(t, err)
	})

	t.Run("BelongsTo", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.Post{})
		require.NoError(t, err)

		relation, ok := mStruct.RelationByName("LatestComment")
		require.True(t, ok)

		post := &testmodels.Post{ID: 3}
		comment := &testmodels.Comment{ID: 10}
		repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
			require.Equal(t, mStruct, s.ModelStruct)
			if assert.Len(t, s.Models, 1) {
				p1, ok := s.Models[0].(*testmodels.Post)
				require.True(t, ok)
				assert.Equal(t, post, p1)
				assert.Equal(t, comment.ID, p1.LatestCommentID)
			}
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.ElementsMatch(t, mapping.FieldSet{mStruct.Primary(), relation.Relationship().ForeignKey()}, fieldSet)
				}
			}
			return 1, nil
		})
		err = db.SetRelations(ctx, post, relation, comment)
		require.NoError(t, err)

		assert.Equal(t, post.LatestCommentID, comment.ID)
	})
}

func TestGetRelations(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithDefaultRepository(repo), WithModelMap(mm))
	require.NoError(t, err)

	t.Run("HasMany", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasManyModel{})
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
		relations, err := db.GetRelations(context.Background(), mStruct, []mapping.Model{m1, m2}, relation)
		require.NoError(t, err)
		assert.Len(t, relations, 5)
		assert.ElementsMatch(t, []interface{}{4, 5, 6, 7, 8}, mapping.Models(relations).PrimaryKeyValues())
	})
	t.Run("HasOne", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.HasOneModel{})
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
		relations, err := db.GetRelations(context.Background(), mStruct, []mapping.Model{m1, m2}, relation)
		require.NoError(t, err)
		assert.Len(t, relations, 2)
		assert.ElementsMatch(t, []interface{}{4, 5}, mapping.Models(relations).PrimaryKeyValues())
	})

	t.Run("Many2Many", func(t *testing.T) {
		t.Run("FullFieldset", func(t *testing.T) {
			mStruct, err := mm.ModelStruct(&testmodels.ManyToManyModel{})
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

			relations, err := db.GetRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation)
			require.NoError(t, err)
			assert.Len(t, relations, 5)
			assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5}, mapping.Models(relations).PrimaryKeyValues())
		})
		t.Run("OnlyPrimaries", func(t *testing.T) {
			mStruct, err := mm.ModelStruct(&testmodels.ManyToManyModel{})
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

			relations, err := db.GetRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation, relatedModelStruct.Primary())
			require.NoError(t, err)
			assert.Len(t, relations, 5)
			assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5}, mapping.Models(relations).PrimaryKeyValues())
		})
	})

	t.Run("BelongsTo", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&testmodels.Blog{})
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
		relations, err := db.GetRelations(context.Background(), mStruct, []mapping.Model{m1, m2, m3}, relation)
		require.NoError(t, err)
		assert.Len(t, relations, 2)
		assert.ElementsMatch(t, []interface{}{uint64(5), uint64(10)}, mapping.Models(relations).PrimaryKeyValues())
	})
}
