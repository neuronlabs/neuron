package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/namer"
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
	ID      int
	Name    string
	Age     int
	Created time.Time
}

// TestRegisterModel tests the register model function.
func TestRegisterModel(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}

	t.Run("Embedded", func(t *testing.T) {
		m := testingModelMap(t)
		err := m.RegisterModels(&embeddedModel{})
		assert.NoError(t, err)

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

		sField, ok := ms.RelationshipField("rel_field")
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
		err := m.RegisterModels(NotTaggedModel{})
		require.NoError(t, err)

		model, err := m.GetModelStruct(NotTaggedModel{})
		require.NoError(t, err)

		assert.Equal(t, "not_tagged_models", model.Collection())

		assert.Len(t, model.fields, 4)

		assert.NotNil(t, model.PrimaryField())
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
	})
}