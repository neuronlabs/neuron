package mapping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/namer"
)

const defaultRepo = "repo"

func testingModelMap(t testing.TB) *ModelMap {
	t.Helper()

	cfg := config.DefaultController()
	m := NewModelMap(namer.NamingSnake, cfg)
	return m
}

// NotTaggedModel is the model used for testing that has no tagged fields.
type NotTaggedModel struct {
	ID                    int
	Name                  string
	Age                   int
	Created               time.Time
	OtherNotTaggedModelID int
	Related               *OtherNotTaggedModel
	ManyRelationID        int `neuron:"type=foreign"`
}

// OtherNotTaggedModel is the testing model with no neuron tags, related to the NotTaggedModel.
type OtherNotTaggedModel struct {
	ID            int
	SingleRelated *NotTaggedModel
	ManyRelation  []*NotTaggedModel
}

// TestRegisterModel tests the register model function.
func TestRegisterModel(t *testing.T) {
	t.Run("Embedded", func(t *testing.T) {
		m := testingModelMap(t)
		err := m.RegisterModels(&embeddedModel{})
		require.NoError(t, err)

		ms, err := m.GetModelStruct(&embeddedModel{})
		require.NoError(t, err)

		// get embedded attribute
		sa, ok := ms.Attribute("string_attr")
		if assert.True(t, ok) {
			assert.Equal(t, []int{0, 1}, sa.Index)
		}

		// get 'this' attribute
		ia, ok := ms.Attribute("int_attr")
		if assert.True(t, ok) {
			assert.Equal(t, []int{1}, ia.Index)
		}

		sField, ok := ms.RelationByName("rel_field")
		if assert.True(t, ok) {
			assert.Equal(t, []int{0, 2}, sField.getFieldIndex())
		}

		sField, ok = ms.ForeignKey("rel_field_id")
		if assert.True(t, ok) {
			assert.Equal(t, []int{0, 3}, sField.getFieldIndex())
		}
	})

	t.Run("Tags", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			// ModelComma is the model where the type hase more than,' instead of ';'.
			type ModelComma struct {
				ID string `neuron:"type=primary,flags=client-id"`
			}

			m := testingModelMap(t)
			err := m.RegisterModels(&ModelComma{})
			assert.Error(t, err)
		})

		t.Run("Empty", func(t *testing.T) {
			m := testingModelMap(t)
			err := m.RegisterModels(NotTaggedModel{}, OtherNotTaggedModel{})
			require.NoError(t, err)

			model, err := m.GetModelStruct(NotTaggedModel{})
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

			otherModel, err := m.GetModelStruct(OtherNotTaggedModel{})
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

		t.Run("NoType", func(t *testing.T) {
			// ModelNoFlags is the model where the flags are not written due to lack of type:
			type ModelNoFlags struct {
				ID string `neuron:"flags=client-id"`
			}

			mm := testingModelMap(t)
			err := mm.RegisterModels(&ModelNoFlags{})
			require.NoError(t, err)

			m, err := mm.GetModelStruct(&ModelNoFlags{})
			require.NoError(t, err)

			primary := m.Primary()
			require.NotNil(t, primary)

			assert.True(t, primary.allowClientID())
		})
	})

	t.Run("Relationship", func(t *testing.T) {
		// CarBrand is the included testing model.
		type CarBrand struct {
			ID int
		}

		// Car is the included testing model.
		type Car struct {
			ID     int
			UserID int

			Brand   *CarBrand
			BrandID int
		}

		// User is the include testing model.
		type User struct {
			ID   int
			Cars []*Car
		}

		mm := testingModelMap(t)
		err := mm.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		car, err := mm.GetModelStruct(Car{})
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

		user, err := mm.GetModelStruct(User{})
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
			type timer struct {
				ID        int
				CreatedAt time.Time
				UpdatedAt *time.Time
				DeletedAt *time.Time
			}

			mm := testingModelMap(t)

			err := mm.RegisterModels(timer{})
			require.NoError(t, err)

			m, err := mm.GetModelStruct(timer{})
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
				type timer struct {
					ID        int
					CreatedAt string
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})

			t.Run("DeletedAt", func(t *testing.T) {
				type timer struct {
					ID        int
					DeletedAt time.Time
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})

			t.Run("UpdatedAt", func(t *testing.T) {
				type timer struct {
					ID        int
					UpdatedAt int
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})
		})
	})

	t.Run("WithFlags", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			type timer struct {
				ID          int
				CreatedTime time.Time  `neuron:"flags=created_at"`
				UpdatedTime *time.Time `neuron:"flags=updated_at"`
				DeletedTime *time.Time `neuron:"flags=deleted_at"`
			}

			mm := testingModelMap(t)

			err := mm.RegisterModels(timer{})
			require.NoError(t, err)

			m, err := mm.GetModelStruct(timer{})
			require.NoError(t, err)

			createdAtField, ok := m.Attribute("CreatedTime")
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
				type timer struct {
					ID        int
					CreatedAt string `neuron:"flags=created_at"`
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})

			t.Run("DeletedAt", func(t *testing.T) {
				type timer struct {
					ID        int
					DeletedAt time.Time `neuron:"flags=deleted_at"`
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})

			t.Run("UpdatedAt", func(t *testing.T) {
				type timer struct {
					ID        int
					UpdatedAt int `neuron:"flags=updated_at"`
				}

				mm := testingModelMap(t)

				err := mm.RegisterModels(timer{})
				require.Error(t, err)
			})
		})
	})
}
