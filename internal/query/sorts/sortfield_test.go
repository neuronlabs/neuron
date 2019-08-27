package sorts

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/namer"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// TestNewUniques tests the NewUniques function.
func TestNewUniques(t *testing.T) {
	ms := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	t.Run("DisallowFK", func(t *testing.T) {
		_, err := NewUniques(mStruct, true, "current_post_id")
		require.Error(t, err)
	})

	t.Run("Duplicated", func(t *testing.T) {
		_, err := NewUniques(mStruct, false, "id", "-id", "id")
		require.Error(t, err)
	})

	t.Run("TooManyPossible", func(t *testing.T) {
		sorts := []string{}
		for i := 0; i < 200; i++ {
			sorts = append(sorts, "id")
		}
		_, err := NewUniques(mStruct, false, sorts...)
		require.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		sorts, err := NewUniques(mStruct, false, "-id", "title", "posts.id", "current_post_id")
		require.NoError(t, err)

		assert.Len(t, sorts, 4)
	})
}

// TestNew tests New sort field method.
func TestNew(t *testing.T) {
	ms := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	t.Run("NoOrder", func(t *testing.T) {
		sField, err := New(mStruct, "-id", true)
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order())
	})

	t.Run("WithOrder", func(t *testing.T) {
		sField, err := New(mStruct, "id", true, DescendingOrder)
		require.NoError(t, err)

		assert.Equal(t, DescendingOrder, sField.Order())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Run("NonRelationship", func(t *testing.T) {
			_, err := New(mStruct, "invalid", false)
			require.Error(t, err)
		})

		t.Run("Relationship", func(t *testing.T) {
			_, err := New(mStruct, "some.field", false)
			require.Error(t, err)
		})

		t.Run("SubField", func(t *testing.T) {
			t.Run("DisallowedFK", func(t *testing.T) {
				_, err := New(mStruct, "posts.unknown", true)
				require.Error(t, err)
			})

			t.Run("AllowedFK", func(t *testing.T) {
				t.Run("Unknown", func(t *testing.T) {
					_, err := New(mStruct, "posts.unknwon", false)
					require.Error(t, err)
				})

				t.Run("Valid", func(t *testing.T) {
					sField, err := New(mStruct, "posts.blog_id", false)
					require.NoError(t, err)

					if assert.Len(t, sField.SubFields(), 1) {
						fk := sField.SubFields()[0]
						assert.Equal(t, models.KindForeignKey, fk.StructField().FieldKind())
					}
				})
			})
		})
	})

	t.Run("TooDeep", func(t *testing.T) {
		_, err := New(mStruct, "posts.comments.id", true)
		require.Error(t, err)
	})
}

// TestSortField tests the sortfield copy method.
func TestSortField(t *testing.T) {
	ms := models.NewModelMap(namer.NamingKebab, config.ReadDefaultControllerConfig())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	sField, err := New(mStruct, "posts.id", true)
	require.NoError(t, err)

	t.Run("Copy", func(t *testing.T) {
		copied := Copy(sField)

		require.NotEqual(t, fmt.Sprintf("%p", sField), fmt.Sprintf("%p", copied), "%p, %p", sField, copied)

		assert.Equal(t, sField.StructField(), copied.StructField())
	})

	t.Run("NewSortField", func(t *testing.T) {
		sField := NewSortField(sField.StructField(), AscendingOrder)
		require.NotNil(t, sField)
	})
}

// TestSetRelationScopeSort sets the relation scope sort field.
func TestSetRelationScopeSort(t *testing.T) {
	ms := models.NewModelMap(namer.NamingKebab, config.ReadDefaultControllerConfig())

	err := ms.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&Blog{})
	require.NoError(t, err)

	sortField := &SortField{structField: mStruct.PrimaryField()}
	err = sortField.setSubfield([]string{}, AscendingOrder, true)
	assert.Error(t, err)

	postField, ok := mStruct.RelationshipField("posts")
	require.True(t, ok)

	sortField = &SortField{structField: postField}
	err = sortField.setSubfield([]string{}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"posts", "some", "id"}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"comments", "id", "desc"}, AscendingOrder, true)
	assert.Error(t, err)

	err = sortField.setSubfield([]string{"comments", "id"}, AscendingOrder, true)
	assert.Nil(t, err)

	err = sortField.setSubfield([]string{"comments", "body"}, AscendingOrder, true)
	assert.Nil(t, err)

	err = sortField.setSubfield([]string{"comments", "id"}, AscendingOrder, true)
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
