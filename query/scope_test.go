package query

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

type formatter struct {
	ID   int                `neuron:"type=primary"`
	Attr string             `neuron:"type=attr"`
	Rel  *formatterRelation `neuron:"type=relation;foreign=FK"`
	FK   int                `neuron:"type=foreign"`
	Lang string             `neuron:"type=attr;flags=lang"`
}

type formatterRelation struct {
	ID int `neuron:"type=primary"`
}

func newController(t *testing.T) *controller.Controller {
	c := controller.NewDefault()
	err := c.RegisterRepository(repoName, &config.Repository{
		DriverName: repoName,
	})
	require.NoError(t, err)
	return c
}

// TestFormatQuery tests the format query methods
func TestFormatQuery(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&formatter{}, &formatterRelation{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&formatter{})
	require.NoError(t, err)

	db := NewComposer(c)

	t.Run("Filters", func(t *testing.T) {
		t.Run("Primary", func(t *testing.T) {
			s := newScope(db, mStruct)

			err = s.Filter(NewFilterField(mStruct.Primary(), OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Foreign", func(t *testing.T) {
			s := newScope(db, mStruct)

			field, ok := mStruct.ForeignKey("fk")
			require.True(t, ok)

			err = s.Filter(NewFilterField(field, OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Attribute", func(t *testing.T) {
			s := newScope(db, mStruct)
			require.NoError(t, err)

			field, ok := mStruct.Attribute("attr")
			require.True(t, ok)

			err = s.Filter(NewFilterField(field, OpEqual, "some-value"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-value", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Relationship", func(t *testing.T) {
			s := newScope(db, mStruct)
			require.NoError(t, err)

			field, ok := mStruct.RelationByName("rel")
			require.True(t, ok)

			relPrim := field.Relationship().Struct().Primary()

			err = s.Filter(newRelationshipFilter(field, NewFilterField(relPrim, OpEqual, 12)))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "12", q.Get(fmt.Sprintf("filter[%s][%s][%s][%s]", mStruct.Collection(), field.NeuronName(), relPrim.NeuronName(), OpEqual.URLAlias)))
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		s := newScope(db, mStruct)
		require.NoError(t, err)

		s.Limit(12)
		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "12", q.Get(ParamPageLimit))
	})

	t.Run("Sorts", func(t *testing.T) {
		s := newScope(db, mStruct)
		require.NoError(t, err)

		err = s.OrderBy("-id")
		require.NoError(t, err)

		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "-id", q.Get(ParamSort))
	})

	t.Run("Fieldset", func(t *testing.T) {
		t.Run("DefaultController", func(t *testing.T) {
			s := newScope(db, mStruct)
			require.NoError(t, err)

			s.Fieldset = append(mapping.FieldSet{}, mStruct.Fields()...)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			fieldsString := q.Get(fmt.Sprintf("fields[%s]", s.Struct().Collection()))
			fields := strings.Split(fieldsString, ",")
			assert.Len(t, fields, 4)

			assert.Contains(t, fields, "id")
			assert.Contains(t, fields, "attr")
			assert.Contains(t, fields, "fk")
			assert.Contains(t, fields, "lang")
		})
	})
}
