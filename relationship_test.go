package jsonapi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMappedRelationships(t *testing.T) {
	clearMap()

	require.NoError(t, c.PrecomputeModels(&User{}, &Pet{}, &Driver{}, &Car{}))

	t.Run("many2many", func(t *testing.T) {
		t.Run("sync", func(t *testing.T) {
			model, err := c.GetModelStruct(&User{})
			require.NoError(t, err)

			pets, ok := model.relationships[c.NamerFunc("Pets")]
			require.True(t, ok)

			require.NotNil(t, pets.relationship)
			assert.Equal(t, RelMany2Many, pets.relationship.Kind)
			if assert.NotNil(t, pets.relationship.Sync) {
				assert.True(t, *pets.relationship.Sync)
			}

			petModel, err := c.GetModelStruct(&Pet{})
			require.NoError(t, err)

			owners, ok := petModel.relationships[c.NamerFunc("Owners")]
			require.True(t, ok)

			assert.Equal(t, owners, pets.relationship.BackReferenceField)
			assert.Equal(t, pets, owners.relationship.BackReferenceField)
		})
		t.Run("nosync", func(t *testing.T) {

		})
	})

	t.Run("hasMany", func(t *testing.T) {
		t.Run("synced", func(t *testing.T) {
			driverModel, err := c.getModelStruct(&Driver{})
			require.NoError(t, err)

			cars, ok := driverModel.relationships["cars"]
			require.True(t, ok)

			carModel, err := c.getModelStruct(&Car{})
			require.NoError(t, err)

			fk, ok := carModel.foreignKeys["driver_id"]
			require.True(t, ok)
			if assert.NotNil(t, cars.relationship) {
				assert.Equal(t, fk, cars.relationship.ForeignKey)
				assert.Equal(t, RelHasMany, cars.relationship.Kind)
				if assert.NotNil(t, cars.relationship.Sync) {
					assert.True(t, *cars.relationship.Sync)
				}
			}
		})
	})

	t.Run("belongsTo", func(t *testing.T) {
		driverModel, err := c.getModelStruct(&Driver{})
		require.NoError(t, err)

		favoriteCar, ok := driverModel.relationships["favorite-car"]
		require.True(t, ok)

		if assert.NotNil(t, favoriteCar.relationship) {
			assert.Equal(t, RelBelongsTo, favoriteCar.relationship.Kind)
			assert.Nil(t, favoriteCar.relationship.Sync)
		}

	})

	t.Run("hasOne", func(t *testing.T) {
		clearMap()

		require.NoError(t, c.PrecomputeModels(&Post{}, &Comment{}))

		postModel, err := c.getModelStruct(&Post{})
		require.NoError(t, err)

		lc, ok := postModel.relationships["latest_comment"]
		require.True(t, ok)

		commentModel, err := c.getModelStruct(&Comment{})
		require.NoError(t, err)

		postID, ok := commentModel.foreignKeys["post_id"]
		require.True(t, ok)

		if assert.NotNil(t, lc.relationship) {
			assert.Equal(t, RelHasOne, lc.relationship.Kind)
			assert.Equal(t, postID, lc.relationship.ForeignKey)
		}

	})
}
