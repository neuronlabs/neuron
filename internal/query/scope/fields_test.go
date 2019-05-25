package scope

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/namer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAutoSelectFields(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}

	t.Run("NonZeros", func(t *testing.T) {
		schm, err := models.NewModelSchemas(namer.NamingSnake, config.ReadDefaultControllerConfig(), flags.New())
		require.NoError(t, err)

		require.NoError(t, schm.RegisterModels(&testModel{}, &testRelatedModel{}))

		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &testModel{ID: 1, Name: "Some", Relation: &testRelatedModel{}}

		require.Nil(t, s.selectedFields)

		require.NoError(t, s.AutoSelectFields())

		require.NotNil(t, s.selectedFields)

		assert.Len(t, s.selectedFields, 3)
	})

	t.Run("WithZeros", func(t *testing.T) {
		schm, err := models.NewModelSchemas(namer.NamingSnake, config.ReadDefaultControllerConfig(), flags.New())
		require.NoError(t, err)

		require.NoError(t, schm.RegisterModels(&testModel{}, &testRelatedModel{}))

		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &testModel{ID: 1}

		require.Nil(t, s.selectedFields)

		require.NoError(t, s.AutoSelectFields())

		require.NotNil(t, s.selectedFields)

		assert.Len(t, s.selectedFields, 1)
	})
}
