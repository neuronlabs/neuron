package models

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/log"
	// "github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/namer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "reflect"
	"testing"
)

const (
	defaultSchema string = "schema"
	defaultRepo   string = "repo"
)

func testingSchemas(t *testing.T) *ModelSchemas {
	t.Helper()

	cfg := config.ReadDefaultControllerConfig()
	m, err := NewModelSchemas(namer.NamingSnake, 3, cfg.ModelSchemas, defaultSchema, defaultRepo, flags.New())
	require.NoError(t, err)
	return m
}

func TestRegisterModel(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}

	s := testingSchemas(t)

	t.Run("embedded", func(t *testing.T) {
		err := s.RegisterModels(&embeddedModel{})
		assert.NoError(t, err)

		ms, err := s.GetModelStruct(&embeddedModel{})
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

}
