package mapping

import (
	"reflect"
	"strconv"
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

	t.Run("FieldValue", func(t *testing.T) {
		mm := testingModelMap(t)
		err := mm.RegisterModels(Model{})
		require.NoError(t, err)

		mStruct, err := mm.GetModelStruct(Model{})
		require.NoError(t, err)

		attr, ok := mStruct.Field("Attr")
		require.True(t, ok)

		v, err := attr.ValueFromString("some")
		require.NoError(t, err)

		assert.Equal(t, "some", v)

		attr, ok = mStruct.Field("Number")
		require.True(t, ok)

		v, err = attr.ValueFromString("32")
		require.NoError(t, err)

		assert.Equal(t, 32, v)

		attr, ok = mStruct.Field("Age")
		require.True(t, ok)

		v, err = attr.ValueFromString("65")
		require.NoError(t, err)

		assert.Equal(t, uint(65), v)

		_, err = attr.ValueFromString("-12")
		assert.Error(t, err)

		attr, ok = mStruct.Field("Time")
		require.True(t, ok)

		tm := time.Now()

		_, err = attr.ValueFromString(strconv.FormatInt(tm.UnixNano(), 10))
		assert.Error(t, err)

		attr, ok = mStruct.Field("Flt")
		require.True(t, ok)

		v, err = attr.ValueFromString("-12.321")
		if assert.NoError(t, err) {
			assert.InDelta(t, -12.321, v, 0.001)
		}

		_, err = attr.ValueFromString("invalid")
		assert.Error(t, err)

		attr, ok = mStruct.Field("Bl")
		require.True(t, ok)

		v, err = attr.ValueFromString("false")
		if assert.NoError(t, err) {
			assert.Equal(t, false, v)
		}
		_, err = attr.ValueFromString("invalid")
		assert.Error(t, err)

		v, err = mStruct.primary.ValueFromString("-12431")
		if assert.NoError(t, err) {
			assert.Equal(t, -12431, v)
		}
	})

	t.Run("Primary", func(t *testing.T) {
		mm := testingModelMap(t)
		err := mm.RegisterModels(Model{})
		require.NoError(t, err)

		mStruct, err := mm.GetModelStruct(Model{})
		require.NoError(t, err)

		t.Run("Single", func(t *testing.T) {
			model := &Model{ID: 1}
			pm, err := PrimaryValues(mStruct, reflect.ValueOf(model))
			require.NoError(t, err)

			if assert.Len(t, pm, 1) {
				assert.Equal(t, model.ID, pm[0])
			}
		})

		t.Run("Many", func(t *testing.T) {
			t.Run("Ptr", func(t *testing.T) {
				models := []*Model{{ID: 1}, {ID: 12}}
				mv := reflect.ValueOf(&models)
				pm, err := PrimaryValues(mStruct, mv)
				require.NoError(t, err)

				if assert.Len(t, pm, 2) {
					assert.Contains(t, pm, 1)
					assert.Contains(t, pm, 12)
				}
			})

			t.Run("NonPtr", func(t *testing.T) {
				models := []*Model{{ID: 1}, {ID: 12}}
				mv := reflect.ValueOf(models)
				pm, err := PrimaryValues(mStruct, mv)
				require.NoError(t, err)

				if assert.Len(t, pm, 2) {
					assert.Contains(t, pm, 1)
					assert.Contains(t, pm, 12)
				}
			})
		})
	})

	t.Run("StringValue", func(t *testing.T) {
		svalues := StringValues(1, nil)
		assert.Contains(t, svalues, "1")

		svalues = []string{}
		svalues = StringValues("invalid", &svalues)

		assert.Contains(t, svalues, "invalid")
	})
}
