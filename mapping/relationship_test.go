package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate neurogonesis models methods methods .

// TestMappedRelationships tests the mapped relationships.
func TestMappedRelationships(t *testing.T) {
	t.Run("many2many", func(t *testing.T) {
		t.Run("PredefinedFields", func(t *testing.T) {
			m := testingModelMap(t)

			err := m.RegisterModels(&Model1WithMany2Many{}, &Model2WithMany2Many{}, &JoinModel{})
			require.NoError(t, err)

			join, err := m.ModelStruct(&JoinModel{})
			require.NoError(t, err)

			assert.True(t, join.isJoin)

			first, err := m.ModelStruct(&Model1WithMany2Many{})
			require.NoError(t, err)

			second, err := m.ModelStruct(&Model2WithMany2Many{})
			require.NoError(t, err)

			t.Run("First", func(t *testing.T) {
				relField, ok := first.relationshipField("synced")
				require.True(t, ok)

				rel := relField.relationship
				require.NotNil(t, rel)

				assert.True(t, rel.isMany2Many())
				require.Equal(t, RelMany2Many, relField.relationship.kind)

				assert.Equal(t, second, relField.relationship.mStruct)

				firstForeign, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstForeign, relField.relationship.foreignKey)

				secondFK, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondFK, relField.relationship.mtmRelatedForeignKey)
			})

			t.Run("Second", func(t *testing.T) {
				relField, ok := second.relationshipField("synced")
				require.True(t, ok)

				rel := relField.relationship
				require.NotNil(t, rel)
				require.Equal(t, RelMany2Many, rel.kind)

				assert.Equal(t, first, rel.mStruct)

				// check the backreference key
				secondForeign, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondForeign, rel.foreignKey)

				// check the foreign key
				firstFK, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstFK, rel.mtmRelatedForeignKey)
			})
		})

		t.Run("DefaultSettings", func(t *testing.T) {
			m := testingModelMap(t)

			err := m.RegisterModels(&First{}, &Second{}, &FirstSeconds{})
			require.NoError(t, err)

			first, err := m.ModelStruct(&First{})
			require.NoError(t, err)

			second, err := m.ModelStruct(&Second{})
			require.NoError(t, err)

			firstSeconds, err := m.ModelStruct(&FirstSeconds{})
			require.NoError(t, err)

			firstRel, ok := first.RelationByName("Many")
			require.True(t, ok)

			fID, ok := firstSeconds.ForeignKey("FirstID")
			require.True(t, ok)

			sID, ok := firstSeconds.ForeignKey("SecondID")
			require.True(t, ok)

			relFirst := firstRel.relationship
			if assert.NotNil(t, relFirst) {
				assert.Equal(t, fID, relFirst.foreignKey)
				assert.Equal(t, sID, relFirst.mtmRelatedForeignKey)
				assert.Equal(t, RelMany2Many, relFirst.kind)
			}

			secondRel, ok := second.RelationByName("Firsts")
			require.True(t, ok)

			relSecond := secondRel.relationship
			if assert.NotNil(t, relSecond) {
				assert.Equal(t, sID, relSecond.foreignKey)
				assert.Equal(t, fID, relSecond.mtmRelatedForeignKey)
				assert.Equal(t, RelMany2Many, relSecond.kind)
			}
		})
	})

	t.Run("hasMany", func(t *testing.T) {
		t.Run("synced", func(t *testing.T) {
			m := testingModelMap(t)

			// get the models
			require.NoError(t, m.RegisterModels(&ModelWithHasMany{}, &ModelWithForeignKey{}))

			// get hasMany model
			hasManyModel, err := m.ModelStruct(&ModelWithHasMany{})
			require.NoError(t, err)

			hasManyField, ok := hasManyModel.relationshipField("has_many")
			require.True(t, ok)

			fkModel, err := m.ModelStruct(&ModelWithForeignKey{})
			require.NoError(t, err)

			fk, ok := fkModel.ForeignKey("foreign_key")
			require.True(t, ok)

			if assert.NotNil(t, hasManyField.relationship) {
				assert.Equal(t, fk, hasManyField.relationship.foreignKey)
				assert.Equal(t, RelHasMany, hasManyField.relationship.kind)
				assert.Equal(t, fkModel, hasManyField.relationship.mStruct)
			}
		})
	})

	t.Run("SingleRelations", func(t *testing.T) {
		m := testingModelMap(t)

		require.NoError(t, m.RegisterModels(&ModelWithBelongsTo{}, &ModelWithHasOne{}))

		t.Run("belongsTo", func(t *testing.T) {
			model, err := m.ModelStruct(&ModelWithBelongsTo{})
			require.NoError(t, err)

			belongsToField, ok := model.relationshipField("belongs_to")
			require.True(t, ok)

			if assert.NotNil(t, belongsToField.relationship) {
				assert.Equal(t, RelBelongsTo, belongsToField.relationship.kind)
			}
			relFields := model.RelationFields()
			assert.Len(t, relFields, 1)
		})

		t.Run("hasOne", func(t *testing.T) {
			model, err := m.ModelStruct(&ModelWithHasOne{})
			require.NoError(t, err)

			hasOneField, ok := model.relationshipField("has_one")
			require.True(t, ok)

			belongsToModel, err := m.ModelStruct(&ModelWithBelongsTo{})
			require.NoError(t, err)

			fk, ok := belongsToModel.ForeignKey("foreign_key")
			require.True(t, ok)

			if assert.NotNil(t, hasOneField.relationship) {
				assert.Equal(t, RelHasOne, hasOneField.relationship.kind)
				assert.Equal(t, fk, hasOneField.relationship.foreignKey)
			}
			relFields := model.RelationFields()
			assert.Len(t, relFields, 1)
		})
	})

	t.Run("MultipleRelations", func(t *testing.T) {
		m := testingModelMap(t)

		err := m.RegisterModels(&Comment{}, &User{}, &Job{}, &Car{}, &CarBrand{})
		require.NoError(t, err)

		model, err := m.ModelStruct(&Comment{})
		require.NoError(t, err)

		relFields := model.RelationFields()
		assert.Len(t, relFields, 2)
	})
}
