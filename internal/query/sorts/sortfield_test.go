package sorts

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/namer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type blog struct {
	ID            int       `neuron:"type=primary"`
	Title         string    `neuron:"type=attr;name=title"`
	Posts         []*post   `neuron:"type=relation;name=posts;foreign=BlogID"`
	CurrentPost   *post     `neuron:"type=relation;name=current_post"`
	CurrentPostID uint64    `neuron:"type=foreign"`
	CreatedAt     time.Time `neuron:"type=attr;name=created_at;flags=iso8601"`
	ViewCount     int       `neuron:"type=attr;name=view_count;flags=omitempty"`
}

type post struct {
	ID            uint64     `neuron:"type=primary"`
	BlogID        int        `neuron:"type=foreign"`
	Title         string     `neuron:"type=attr;name=title"`
	Body          string     `neuron:"type=attr;name=body"`
	Comments      []*comment `neuron:"type=relation;name=comments;foreign=PostID"`
	LatestComment *comment   `neuron:"type=relation;name=latest_comment;foreign=PostID"`
}

type comment struct {
	ID     int    `neuron:"type=primary"`
	PostID uint64 `neuron:"type=foreign"`
	Body   string `neuron:"type=attr;name=body"`
}

func TestSetRelationScopeSort(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}
	ms, err := models.NewModelSchemas(namer.NamingKebab, config.ReadDefaultControllerConfig(), flags.New())
	require.NoError(t, err)

	err = ms.RegisterModels(&blog{}, &post{}, &comment{})
	require.NoError(t, err)

	mStruct, err := ms.GetModelStruct(&blog{})
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
	assert.NoError(t, err)

	err = sortField.setSubfield([]string{"comments", "body"}, AscendingOrder, true)
	assert.NoError(t, err)

	err = sortField.setSubfield([]string{"comments", "id"}, AscendingOrder, true)
	assert.NoError(t, err)

}
