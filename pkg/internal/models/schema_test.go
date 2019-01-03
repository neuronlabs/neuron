package models

import (
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func DefaultTesting() *Schema {

}

func TestRegisterModels(t *testing.T) {
	mmodels := []interface{}{&internal.Blog{}, &internal.Post{}, &internal.Comment{}}

	c := DefaultTesting()

	err := c.RegisterModels(mmodels...)
	if assert.NoError(t, err) {
		comment, err := c.GetModelStruct(&internal.Comment{})
		require.NoError(t, err)

		post, err := c.GetModelStruct(&internal.Post{})
		require.NoError(t, err)

		blog, err := c.GetModelStruct(&internal.Blog{})
		require.NoError(t, err)

		t.Run("Blog", func(t *testing.T) {

			assert.Equal(t, reflect.TypeOf(internal.Blog{}), blog.Type())

			// post is a relationship
			relField, ok := blog.RelationshipField("posts")
			if assert.True(t, ok) {

				rel := relField.Relationship()
				if assert.NotNil(t, rel) {

					// Check struct
					assert.Equal(t, post, rel.Struct())

					assert.NotNil(t, rel.Struct().Flags())

					// Posts relationship
					blogId, ok := post.ForeignKey("blog_id")
					if assert.True(t, ok) {
						assert.Equal(t, blogId, rel.ForeignKey())
					}

					assert.Equal(t, RelHasMany, rel.Kind())
				}
			}
			relField, ok = blog.RelationshipField("current_post")
			if assert.True(t, ok) {
				rel := relField.Relationship()
				if assert.NotNil(t, rel) {
					assert.Equal(t, post, rel.Struct())

					curPostId, ok := blog.ForeignKey("current_post_id")
					if assert.True(t, ok) {
						assert.Equal(t, curPostId, rel.ForeignKey())
					}

					assert.Equal(t, RelBelongsTo, rel.Kind())
				}
			}

			attrField, ok := blog.Attribute("title")
			if assert.True(t, ok) {
				assert.False(t, attrField.IsNestedStruct())
				assert.False(t, attrField.IsTime())
			}

			attrField, ok = blog.Attribute("created_at")
			if assert.True(t, ok) {
				assert.True(t, attrField.IsTime())
				assert.False(t, attrField.IsBasePtr())
			}
		})

		t.Run("Posts", func(t *testing.T) {
			relField, ok := post.RelationshipField("latest_comment")
			if assert.True(t, ok) {
				assert.True(t, relField.IsRelationship())

				rel := relField.Relationship()
				if assert.NotNil(t, rel) {
					assert.Equal(t, comment, rel.Struct())
					assert.Equal(t, RelHasOne, rel.Kind())

					postId, ok := comment.ForeignKey("post_id")
					if assert.True(t, ok) {
						assert.Equal(t, postId, rel.ForeignKey())
					}
				}
			}
		})

		t.Run("Comment", func(t *testing.T) {
			assert.Equal(t, "comments", comment.Collection())
		})
	}

}
