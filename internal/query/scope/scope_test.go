package scope

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/namer"

	"github.com/neuronlabs/neuron-core/internal/models"
)

type testRelatedModel struct {
	ID int `neuron:"type=primary"`
	FK int `neuron:"type=foreign"`
}

type testModel struct {
	ID       int               `neuron:"type=primary"`
	Name     string            `neuron:"type=attribute"`
	Relation *testRelatedModel `neuron:"type=relation;foreign=FK"`
}

// TestGetPrimaryFieldValues tests the GetPrimaryFieldValues method.
func TestGetPrimaryFieldValues(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}

	schm := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

	require.NoError(t, schm.RegisterModels(&testModel{}, &testRelatedModel{}))

	t.Run("Single", func(t *testing.T) {

		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &testModel{ID: 1}

		primaries, err := s.GetPrimaryFieldValues()
		require.NoError(t, err)

		if assert.Len(t, primaries, 1) {
			assert.Equal(t, 1, primaries[0])
		}

	})

	t.Run("Slice", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &([]*testModel{{ID: 1}, {ID: 2}, {ID: 666}})

		primaries, err := s.GetPrimaryFieldValues()
		require.NoError(t, err)

		if assert.Len(t, primaries, 3) {
			assert.Equal(t, 1, primaries[0])
			assert.Equal(t, 2, primaries[1])
			assert.Equal(t, 666, primaries[2])
		}
	})

	t.Run("InvalidType", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		// non pointer to slice
		s.Value = []*testModel{{ID: 1}, {ID: 2}, {ID: 666}}
		_, err = s.GetPrimaryFieldValues()
		require.Error(t, err)
	})

	t.Run("NilValue", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		// non pointer to slice
		s.Value = nil
		_, err = s.GetPrimaryFieldValues()
		require.Error(t, err)
	})
}

func TestGetForeignKeyValues(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}

	schm := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

	require.NoError(t, schm.RegisterModels(&testModel{}, &testRelatedModel{}))

	t.Run("Single", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &testRelatedModel{ID: 1, FK: 10}
		fkField, ok := mStruct.ForeignKey("FK")
		require.True(t, ok)

		fkValues, err := s.GetForeignKeyValues(fkField)
		require.NoError(t, err)

		if assert.Len(t, fkValues, 1) {
			assert.Equal(t, 10, fkValues[0])
		}
	})

	t.Run("Slice", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &([]*testRelatedModel{{ID: 1, FK: 10}, {ID: 5, FK: 55}})
		fkField, ok := mStruct.ForeignKey("FK")
		require.True(t, ok)

		fkValues, err := s.GetForeignKeyValues(fkField)
		require.NoError(t, err)

		if assert.Len(t, fkValues, 2) {
			assert.Equal(t, 10, fkValues[0])
			assert.Equal(t, 55, fkValues[1])
		}
	})

	t.Run("InvalidType", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = []*testRelatedModel{{ID: 1, FK: 10}, {ID: 5, FK: 55}}
		fkField, ok := mStruct.ForeignKey("FK")
		require.True(t, ok)

		_, err = s.GetForeignKeyValues(fkField)
		require.Error(t, err)
	})

	t.Run("InvalidSliceType", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		testInt := 0

		s.Value = &([]interface{}{&testRelatedModel{ID: 1, FK: 10}, &testInt})
		fkField, ok := mStruct.ForeignKey("FK")
		require.True(t, ok)

		_, err = s.GetForeignKeyValues(fkField)
		require.Error(t, err)
	})

	t.Run("NonMatchedStruct", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &([]*testRelatedModel{{ID: 1, FK: 10}, {ID: 5, FK: 55}})

		mStruct2, err := schm.GetModelStruct(&testModel{})
		require.NoError(t, err)

		_, err = s.GetForeignKeyValues(mStruct2.PrimaryField())
		require.Error(t, err)
	})

	t.Run("NotAForeignKey", func(t *testing.T) {
		mStruct, err := schm.GetModelStruct(&testRelatedModel{})
		require.NoError(t, err)

		s := NewRootScope(mStruct)

		s.Value = &([]*testRelatedModel{{ID: 1, FK: 10}, {ID: 5, FK: 55}})

		_, err = s.GetForeignKeyValues(mStruct.PrimaryField())
		require.Error(t, err)
	})

}
