package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/mapping"
)

// TestNewUniques tests the NewUniques function.
func TestNewUniques(t *testing.T) {
	ms := mapping.NewModelMap(mapping.NamingSnake, config.DefaultController())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	t.Run("DisallowFK", func(t *testing.T) {
		_, err := NewSortFields(mStruct, true, "current_post_id")
		require.Error(t, err)
	})

	t.Run("Duplicated", func(t *testing.T) {
		_, err := NewSortFields(mStruct, false, "TransactionID", "-TransactionID", "TransactionID")
		require.Error(t, err)
	})

	t.Run("TooManyPossible", func(t *testing.T) {
		sorts := []string{}
		for i := 0; i < 200; i++ {
			sorts = append(sorts, "TransactionID")
		}

		_, err := NewSortFields(mStruct, false, sorts...)
		require.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		sorts, err := NewSortFields(mStruct, false, "-TransactionID", "title", "posts.TransactionID", "current_post_id")
		require.NoError(t, err)

		assert.Len(t, sorts, 4)
	})
}

// TestNew tests New sort field method.
func TestNew(t *testing.T) {
	ms := mapping.NewModelMap(mapping.NamingSnake, config.DefaultController())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	t.Run("NoOrder", func(t *testing.T) {
		sField, err := NewSort(mStruct, "-TransactionID", true)
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order)
	})

	t.Run("WithOrder", func(t *testing.T) {
		sField, err := NewSort(mStruct, "TransactionID", true, DescendingOrder)
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Run("NonRelationship", func(t *testing.T) {
			_, err := NewSort(mStruct, "invalid", false)
			require.Error(t, err)
		})

		t.Run("Relationship", func(t *testing.T) {
			_, err := NewSort(mStruct, "some.field", false)
			require.Error(t, err)
		})

		t.Run("SubField", func(t *testing.T) {
			t.Run("DisallowedFK", func(t *testing.T) {
				_, err := NewSort(mStruct, "posts.unknown", true)
				require.Error(t, err)
			})

			t.Run("AllowedFK", func(t *testing.T) {
				t.Run("Unknown", func(t *testing.T) {
					_, err := NewSort(mStruct, "posts.unknwon", false)
					require.Error(t, err)
				})

				t.Run("Valid", func(t *testing.T) {
					sField, err := NewSort(mStruct, "posts.blog_id", false)
					require.NoError(t, err)

					if assert.Len(t, sField.SubFields, 1) {
						fk := sField.SubFields[0]
						assert.Equal(t, mapping.KindForeignKey, fk.StructField.Kind())
					}
				})
			})
		})
	})
}

// TestSortField tests the sortfield copy method.
func TestSortField(t *testing.T) {
	ms := mapping.NewModelMap(mapping.NamingKebab, config.DefaultController())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	sField, err := NewSort(mStruct, "posts.TransactionID", true)
	require.NoError(t, err)

	t.Run("Copy", func(t *testing.T) {
		copied := sField.Copy()

		require.NotEqual(t, fmt.Sprintf("%p", sField), fmt.Sprintf("%p", copied), "%p, %p", sField, copied)

		assert.Equal(t, sField.StructField, copied.StructField)
	})

	t.Run("NewSortField", func(t *testing.T) {
		sField := newSortField(sField.StructField, AscendingOrder)
		require.NotNil(t, sField)
	})
}

// TestSetRelationScopeSort sets the relation scope sort field.
func TestSetRelationScopeSort(t *testing.T) {
	ms := mapping.NewModelMap(mapping.NamingKebab, config.DefaultController())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	sortField := &SortField{StructField: mStruct.Primary()}
	err = sortField.setSubfield([]string{}, AscendingOrder, true)
	assert.Error(t, err)

	postField, ok := mStruct.RelationByName("posts")
	require.True(t, ok)

	sortField = &SortField{StructField: postField}
	err = sortField.setSubfield([]string{}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"posts", "some", "TransactionID"}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"comments", "TransactionID", "desc"}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"comments", "TransactionID"}, AscendingOrder, true)
	assert.Nil(t, err)

	err = sortField.setSubfield([]string{"comments", "body"}, AscendingOrder, true)
	assert.Nil(t, err)

	err = sortField.setSubfield([]string{"comments", "TransactionID"}, AscendingOrder, true)
	assert.Nil(t, err)
}

type Blog struct {
	ID            int       `neuron:"type=primary"`
	Title         string    `neuron:"type=attr;name=title"`
	Posts         []*Post   `neuron:"type=relation;name=posts;foreign=BlogID"`
	CurrentPost   *Post     `neuron:"type=relation;name=current_post"`
	CurrentPostID uint64    `neuron:"type=foreign;name=current_post_id"`
	CreatedAt     time.Time `neuron:"type=attr;name=created_at;flags=iso8601"`
	ViewCount     int       `neuron:"type=attr;name=view_count;flags=omitempty"`
}

type Post struct {
	ID            uint64     `neuron:"type=primary"`
	BlogID        int        `neuron:"type=foreign"`
	Title         string     `neuron:"type=attr;name=title"`
	Body          string     `neuron:"type=attr;name=body"`
	Comments      []*Comment `neuron:"type=relation;name=comments;foreign=PostID"`
	LatestComment *Comment   `neuron:"type=relation;name=latest_comment;foreign=PostID"`
}

type Comment struct {
	ID     int    `neuron:"type=primary"`
	PostID uint64 `neuron:"type=foreign"`
	Body   string `neuron:"type=attr;name=body"`
}
