package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractTags tests the ExtractTags function.
func TestExtractTags(t *testing.T) {
	m := testingModelMap(t)

	err := m.RegisterModels(Model1WithMany2Many{}, Model2WithMany2Many{}, joinModel{})
	require.NoError(t, err)

	first, err := m.GetModelStruct(Model1WithMany2Many{})
	require.NoError(t, err)

	synced, ok := first.RelationByName("Synced")
	require.True(t, ok)

	extracted := synced.ExtractFieldTags("neuron")
	assert.Len(t, extracted, 4)
}
