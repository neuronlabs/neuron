package mapping

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValues tests the model values functions.
func TestValues(t *testing.T) {
	type Model struct {
		ID     int
		Attr   string
		Time   *time.Time
		Age    uint
		Number int
		Flt    float64
		Bl     bool
	}

	t.Run("New", func(t *testing.T) {
		mm := testingModelMap(t)
		err := mm.RegisterModels(Model{})
		require.NoError(t, err)

		mStruct, err := mm.GetModelStruct(Model{})
		require.NoError(t, err)

		t.Run("Many", func(t *testing.T) {
			rf := NewReflectValueMany(mStruct)
			assert.Equal(t, reflect.PtrTo(reflect.SliceOf(reflect.PtrTo(mStruct.modelType))), rf.Type())
			assert.False(t, rf.IsNil())

			rv := NewValueMany(mStruct)
			rf = reflect.ValueOf(rv)

			assert.Equal(t, reflect.PtrTo(reflect.SliceOf(reflect.PtrTo(mStruct.modelType))), rf.Type())
			assert.False(t, rf.IsNil())
		})

		t.Run("Single", func(t *testing.T) {
			sf := NewReflectValueSingle(mStruct)
			assert.Equal(t, reflect.PtrTo(mStruct.modelType), sf.Type())
			assert.False(t, sf.IsNil())

			sv := NewValueSingle(mStruct)
			sf = reflect.ValueOf(sv)
			assert.Equal(t, reflect.PtrTo(mStruct.modelType), sf.Type())
			assert.False(t, sf.IsNil())
		})
	})
}
