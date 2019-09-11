package mapping

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrderedFields tests the ordered fields methods.
func TestOrderedFields(t *testing.T) {
	type Model struct {
		ID     int
		Name   string
		First  string
		Second int
		Third  string
	}
	mm := testingModelMap(t)

	err := mm.RegisterModels(Model{})
	require.NoError(t, err)

	m, err := mm.GetModelStruct(Model{})
	require.NoError(t, err)

	fields := m.Fields()

	// unsort the slice
	fields[0], fields[3], fields[2], fields[4] = fields[3], fields[0], fields[4], fields[2]

	// Remove the 'Name' field at index [1]
	fields = append(fields[:1], fields[2:]...)

	ordered := OrderedFields(fields)
	// sort the fields
	sort.Sort(ordered)

	for i, field := range ordered {
		switch i {
		case 0:
			assert.Equal(t, "ID", field.Name())
		case 1:
			assert.Equal(t, "First", field.Name())
		case 2:
			assert.Equal(t, "Second", field.Name())
		case 3:
			assert.Equal(t, "Third", field.Name())
		}
	}

	// insert back the name field
	nameField, ok := m.Attribute("Name")
	require.True(t, ok)

	orderedPtr := &ordered
	orderedPtr.Insert(nameField)

	assert.Len(t, ordered, 5)

	for i, field := range ordered {
		switch i {
		case 0:
			assert.Equal(t, "ID", field.Name())
		case 1:
			assert.Equal(t, "Name", field.Name())
		case 2:
			assert.Equal(t, "First", field.Name())
		case 3:
			assert.Equal(t, "Second", field.Name())
		case 4:
			assert.Equal(t, "Third", field.Name())
		}
	}

}
