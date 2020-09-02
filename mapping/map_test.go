package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testingModelMap(t testing.TB) *ModelMap {
	t.Helper()

	m := New(WithNamingConvention(SnakeCase))
	return m
}

// TestRegisterModel tests the register model function.
func TestRegisterModel(t *testing.T) {
	t.Run("Tags", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			m := testingModelMap(t)

			model, err := m.ModelStruct(&NotTaggedModel{})
			require.NoError(t, err)

			assert.Equal(t, "not_tagged_models", model.Collection())

			assert.Len(t, model.structFields, 7)

			assert.NotNil(t, model.Primary())
			assert.Equal(t, "id", model.primary.neuronName)

			sField, ok := model.Attribute("Name")
			assert.True(t, ok)
			assert.Equal(t, "name", sField.neuronName)

			sField, ok = model.Attribute("Age")
			assert.True(t, ok)
			assert.Equal(t, "age", sField.neuronName)

			sField, ok = model.Attribute("Created")
			assert.True(t, ok)
			assert.True(t, sField.IsTime())
			assert.Equal(t, "created", sField.neuronName)

			sField, ok = model.RelationByName("Related")
			if assert.True(t, ok) {
				fk, ok := model.ForeignKey("OtherNotTaggedModelID")
				if assert.True(t, ok) {
					assert.Equal(t, fk, sField.relationship.foreignKey)
				}
			}

			otherModel, err := m.ModelStruct(&OtherNotTaggedModel{})
			require.NoError(t, err)

			sField, ok = otherModel.RelationByName("ManyRelation")
			if assert.True(t, ok) {
				fk, ok := model.ForeignKey("ManyRelationID")
				if assert.True(t, ok) {
					assert.Equal(t, fk, sField.relationship.foreignKey)
				}
			}

			sField, ok = otherModel.RelationByName("SingleRelated")
			if assert.True(t, ok) {
				fk, ok := model.ForeignKey("OtherNotTaggedModelID")
				if assert.True(t, ok) {
					assert.Equal(t, fk, sField.relationship.foreignKey)
				}
			}
		})
	})

	t.Run("Relationship", func(t *testing.T) {
		mm := testingModelMap(t)
		err := mm.RegisterModels(&User{}, &Car{}, &CarBrand{}, &Job{}, &Comment{})
		require.NoError(t, err)

		car, err := mm.ModelStruct(&Car{})
		require.NoError(t, err)

		brandID, ok := car.ForeignKey("BrandID")
		assert.True(t, ok)

		_, ok = car.Attribute("BrandID")
		assert.False(t, ok)

		brand, ok := car.RelationByName("Brand")
		require.True(t, ok)

		assert.Equal(t, brandID, brand.Relationship().ForeignKey())

		userID, ok := car.ForeignKey("UserID")
		assert.True(t, ok)

		user, err := mm.ModelStruct(&User{})
		require.NoError(t, err)

		rel, ok := user.RelationByName("Cars")
		require.True(t, ok)

		assert.Equal(t, userID, rel.Relationship().ForeignKey())
	})
}

// TestTimeRelatedField tests the time related fields.
func TestTimeRelatedField(t *testing.T) {
	t.Run("WithDefaultNames", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			mm := testingModelMap(t)

			err := mm.RegisterModels(&Timer{})
			require.NoError(t, err)

			m, err := mm.ModelStruct(&Timer{})
			require.NoError(t, err)

			createdAtField, ok := m.Attribute("CreatedAt")
			require.True(t, ok)

			assert.True(t, createdAtField.IsTime())
			assert.True(t, createdAtField.IsCreatedAt())

			createdAt, ok := m.CreatedAt()
			if assert.True(t, ok) {
				assert.True(t, createdAt.IsTime())
			}

			updatedAt, ok := m.UpdatedAt()
			if assert.True(t, ok) {
				assert.True(t, updatedAt.IsTimePointer())
			}

			deletedAt, ok := m.DeletedAt()
			if assert.True(t, ok) {
				assert.True(t, deletedAt.IsTimePointer())
			}
		})

		t.Run("Invalid", func(t *testing.T) {
			t.Run("CreatedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidCreatedAt{})
				require.Error(t, err)
			})

			t.Run("DeletedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidDeletedAt{})
				require.Error(t, err)
			})

			t.Run("UpdatedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidUpdatedAt{})
				require.Error(t, err)
			})
		})
	})

	t.Run("WithFlags", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			mm := testingModelMap(t)

			err := mm.RegisterModels(&Timer{})
			require.NoError(t, err)

			m, err := mm.ModelStruct(&Timer{})
			require.NoError(t, err)

			createdAtField, ok := m.Attribute("CreatedAt")
			require.True(t, ok)

			assert.True(t, createdAtField.IsTime())
			assert.True(t, createdAtField.IsCreatedAt())

			createdAt, ok := m.CreatedAt()
			if assert.True(t, ok) {
				assert.True(t, createdAt.IsTime())
			}

			updatedAt, ok := m.UpdatedAt()
			if assert.True(t, ok) {
				assert.True(t, updatedAt.IsTimePointer())
				assert.True(t, updatedAt.IsUpdatedAt())
			}

			deletedAt, ok := m.DeletedAt()
			if assert.True(t, ok) {
				assert.True(t, deletedAt.IsTimePointer())
				assert.True(t, deletedAt.IsDeletedAt())
			}
		})

		t.Run("Invalid", func(t *testing.T) {
			t.Run("CreatedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidCreatedAt{})
				require.Error(t, err)
			})

			t.Run("DeletedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidDeletedAt{})
				require.Error(t, err)
			})

			t.Run("UpdatedAt", func(t *testing.T) {
				mm := testingModelMap(t)

				err := mm.RegisterModels(&InvalidUpdatedAt{})
				require.Error(t, err)
			})
		})
	})
}

func TestUnmappedModels(t *testing.T) {
	mm := testingModelMap(t)

	mStruct, err := mm.ModelStruct(&Model1WithMany2Many{})
	require.NoError(t, err)

	require.NotNil(t, mStruct)
	assert.Len(t, mm.models, 3)
}
