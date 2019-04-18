package models

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type model1WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*model2WithMany2Many `neuron:"type=relation;relation=many2many,sync,Synced"`
}

type model2WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*model1WithMany2Many `neuron:"type=relation;relation=many2many,sync,Synced"`
}

type modelWithHasMany struct {
	ID      int                    `neuron:"type=primary"`
	HasMany []*modelWithForeignKey `neuron:"type=relation;foreign=ForeignKey"`
}

type modelWithForeignKey struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
}

type modelWithBelongsTo struct {
	ID         int              `neuron:"type=primary"`
	ForeignKey int              `neuron:"type=foreign"`
	BelongsTo  *modelWithHasOne `neuron:"type=relation;foreign=ForeignKey"`
}

type modelWithHasOne struct {
	ID     int                 `neuron:"type=primary"`
	HasOne *modelWithBelongsTo `neuron:"type=relation;foreign=ForeignKey"`
}

func TestMappedRelationships(t *testing.T) {

	t.Run("many2many", func(t *testing.T) {
		t.Run("sync", func(t *testing.T) {
			s := testingSchemas(t)
			err := s.RegisterModels(model1WithMany2Many{}, model2WithMany2Many{})
			require.NoError(t, err)

			model, err := s.GetModelStruct(model1WithMany2Many{})
			require.NoError(t, err)

			m2m2, ok := model.relationships["synced"]
			require.True(t, ok)

			require.NotNil(t, m2m2.relationship)
			assert.Equal(t, RelMany2Many, m2m2.relationship.kind)
			if assert.NotNil(t, m2m2.relationship.sync) {
				assert.True(t, *m2m2.relationship.sync)
			}

			// check the other direction
			model2, err := s.GetModelStruct(model2WithMany2Many{})
			require.NoError(t, err)

			m2m1, ok := model2.relationships["synced"]
			require.True(t, ok)

			assert.Equal(t, m2m2, m2m1.relationship.backReferenceField)
			assert.Equal(t, m2m1, m2m2.relationship.backReferenceField)
		})

	})

	t.Run("hasMany", func(t *testing.T) {
		t.Run("synced", func(t *testing.T) {
			s := testingSchemas(t)

			// get the models
			require.NoError(t, s.RegisterModels(modelWithHasMany{}, modelWithForeignKey{}))

			// get hasMany model
			hasManyModel, err := s.getModelStruct(modelWithHasMany{})
			require.NoError(t, err)

			hasManyField, ok := hasManyModel.relationships["has_many"]
			require.True(t, ok)

			fkModel, err := s.getModelStruct(modelWithForeignKey{})
			require.NoError(t, err)

			fk, ok := fkModel.foreignKeys["foreign_key"]
			require.True(t, ok)

			if assert.NotNil(t, hasManyField.relationship) {
				assert.Equal(t, fk, hasManyField.relationship.foreignKey)
				assert.Equal(t, RelHasMany, hasManyField.relationship.kind)
				if assert.NotNil(t, hasManyField.relationship.sync) {
					assert.True(t, *hasManyField.relationship.sync)
				}
			}
		})
	})

	s := testingSchemas(t)

	require.NoError(t, s.RegisterModels(modelWithBelongsTo{}, modelWithHasOne{}))

	t.Run("belongsTo", func(t *testing.T) {
		model, err := s.getModelStruct(modelWithBelongsTo{})
		require.NoError(t, err)

		belongsToField, ok := model.relationships["belongs_to"]
		require.True(t, ok)

		if assert.NotNil(t, belongsToField.relationship) {
			assert.Equal(t, RelBelongsTo, belongsToField.relationship.kind)
			assert.Nil(t, belongsToField.relationship.sync)
		}
	})

	t.Run("hasOne", func(t *testing.T) {

		model, err := s.getModelStruct(modelWithHasOne{})
		require.NoError(t, err)

		hasOneField, ok := model.relationships["has_one"]
		require.True(t, ok)

		belongsToModel, err := s.getModelStruct(modelWithBelongsTo{})
		require.NoError(t, err)

		fk, ok := belongsToModel.foreignKeys["foreign_key"]
		require.True(t, ok)

		if assert.NotNil(t, hasOneField.relationship) {
			assert.Equal(t, RelHasOne, hasOneField.relationship.kind)
			assert.Equal(t, fk, hasOneField.relationship.foreignKey)
		}
	})
}
