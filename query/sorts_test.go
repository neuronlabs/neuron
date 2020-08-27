package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
)

// TestNewUniques tests the NewUniques function.
func TestNewUniques(t *testing.T) {
	ms := mapping.NewModelMap(mapping.WithNamingConvention(mapping.SnakeCase))

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, ok := ms.GetModelStruct(&Blog{})
	require.True(t, ok)

	t.Run("Duplicated", func(t *testing.T) {
		_, err := NewSortFields(mStruct, "id", "-id", "id")
		require.Error(t, err)
	})

	t.Run("TooManyPossible", func(t *testing.T) {
		var sorts []string
		for i := 0; i < 200; i++ {
			sorts = append(sorts, "id")
		}

		_, err := NewSortFields(mStruct, sorts...)
		require.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		sorts, err := NewSortFields(mStruct, "-id", "title", "posts.id", "current_post_id")
		require.NoError(t, err)

		assert.Len(t, sorts, 4)
	})
}

// TestNew tests New sort field method.
func TestNew(t *testing.T) {
	ms := mapping.NewModelMap(mapping.WithNamingConvention(mapping.SnakeCase))

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, ok := ms.GetModelStruct(&Blog{})
	require.True(t, ok)

	t.Run("NoOrder", func(t *testing.T) {
		sField, err := NewSort(mStruct, "-id")
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order())
	})

	t.Run("WithOrder", func(t *testing.T) {
		sField, err := NewSort(mStruct, "id", DescendingOrder)
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Run("NonRelationship", func(t *testing.T) {
			_, err := NewSort(mStruct, "invalid")
			require.Error(t, err)
		})

		t.Run("Relationship", func(t *testing.T) {
			_, err := NewSort(mStruct, "some.field")
			require.Error(t, err)
		})

		t.Run("SubField", func(t *testing.T) {
			t.Run("AllowedFK", func(t *testing.T) {
				t.Run("Unknown", func(t *testing.T) {
					_, err := NewSort(mStruct, "posts.unknown")
					require.Error(t, err)
				})

				t.Run("Valid", func(t *testing.T) {
					sField, err := NewSort(mStruct, "posts.blog_id")
					require.NoError(t, err)

					relationSort, ok := sField.(RelationSort)
					require.True(t, ok)
					if assert.Len(t, relationSort.RelationFields, 1) {
						fk := relationSort.RelationFields[0]
						assert.Equal(t, mapping.KindForeignKey, fk.Kind())
					}
				})
			})
		})
	})
}

// TestSortField tests the sort field copy method.
func TestSortField(t *testing.T) {
	ms := mapping.NewModelMap(mapping.WithNamingConvention(mapping.KebabCase))

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, ok := ms.GetModelStruct(&Blog{})
	require.True(t, ok)

	sField, err := NewSort(mStruct, "posts.id")
	require.NoError(t, err)

	t.Run("Copy", func(t *testing.T) {
		copied := sField.Copy()

		assert.Equal(t, sField.Field(), copied.Field())
	})
}
