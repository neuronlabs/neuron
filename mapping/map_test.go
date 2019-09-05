package mapping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/namer"
)

const defaultRepo = "repo"

func testingModelMap(t *testing.T) *ModelMap {
	t.Helper()

	cfg := config.ReadDefaultControllerConfig()
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
			assert.Equal(t, []int{0, 1}, sa.fieldIndex)
		}

		// get 'this' attribute
		ia, ok := ms.Attribute("int_attr")
		if assert.True(t, ok) {
			assert.Equal(t, []int{1}, ia.fieldIndex)
		}

		sField, ok := ms.RelationField("rel_field")
		if assert.True(t, ok) {
			assert.Equal(t, []int{0, 2}, sField.getFieldIndex())
		}

		sField, ok = ms.ForeignKey("rel_field_id")
		if assert.True(t, ok) {
			assert.Equal(t, []int{0, 3}, sField.getFieldIndex())
		}
	})

	t.Run("NoTags", func(t *testing.T) {
		m := testingModelMap(t)
		err := m.RegisterModels(NotTaggedModel{}, OtherNotTaggedModel{})
		require.NoError(t, err)

		model, err := m.GetModelStruct(NotTaggedModel{})
		require.NoError(t, err)

		assert.Equal(t, "not_tagged_models", model.Collection())

		assert.Len(t, model.fields, 7)

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

		sField, ok = model.RelationField("Related")
		if assert.True(t, ok) {
			fk, ok := model.ForeignKey("OtherNotTaggedModelID")
			if assert.True(t, ok) {
				assert.Equal(t, fk, sField.relationship.foreignKey)
			}
		}

		otherModel, err := m.GetModelStruct(OtherNotTaggedModel{})
		require.NoError(t, err)

		sField, ok = otherModel.RelationField("ManyRelation")
		if assert.True(t, ok) {
			fk, ok := model.ForeignKey("ManyRelationID")
			if assert.True(t, ok) {
				assert.Equal(t, fk, sField.relationship.foreignKey)
			}
		}

		sField, ok = otherModel.RelationField("SingleRelated")
		if assert.True(t, ok) {
			fk, ok := model.ForeignKey("OtherNotTaggedModelID")
			if assert.True(t, ok) {
				assert.Equal(t, fk, sField.relationship.foreignKey)
			}
		}
	})
}
