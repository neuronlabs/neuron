package sorts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/namer"

	"github.com/neuronlabs/neuron-core/internal/models"
)

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
	CurrentPostID uint64    `neuron:"type=foreign"`
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
