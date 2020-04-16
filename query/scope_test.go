package query

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestFormatQuery tests the format query methods
func TestFormatQuery(t *testing.T) {
	c := newController(t)
	err := c.RegisterModels(&formatter{}, &formatterRelation{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&formatter{})
	require.NoError(t, err)

	t.Run("Filters", func(t *testing.T) {
		t.Run("Primary", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			err = s.FilterField(NewFilterField(mStruct.Primary(), OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 2)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Foreign", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.ForeignKey("fk")
			require.True(t, ok)

			err = s.FilterField(NewFilterField(field, OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 2)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Attribute", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.Attribute("attr")
			require.True(t, ok)

			err = s.FilterField(NewFilterField(field, OpEqual, "some-value"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 2)

			assert.Equal(t, "some-value", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Relationship", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.RelationField("rel")
			require.True(t, ok)

			relPrim := field.Relationship().Struct().Primary()

			err = s.FilterField(newRelationshipFilter(field, NewFilterField(relPrim, OpEqual, 12)))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 2)

			assert.Equal(t, "12", q.Get(fmt.Sprintf("filter[%s][%s][%s][%s]", mStruct.Collection(), field.NeuronName(), relPrim.NeuronName(), OpEqual.URLAlias)))
		})

		t.Run("Language", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			field := mStruct.LanguageField()
			require.NotNil(t, field)

			err = s.FilterField(NewFilterField(field, OpNotEqual, "pl"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 2)

			assert.Equal(t, "pl", q.Get(ParamLanguage), fmt.Sprintf("%v", q))
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		s, err := NewC(c, &formatter{})
		require.NoError(t, err)

		err = s.Limit(12)
		require.NoError(t, err)

		q := s.FormatQuery()
		require.Len(t, q, 2)

		assert.Equal(t, "12", q.Get(ParamPageLimit))
	})

	t.Run("Sorts", func(t *testing.T) {
		s, err := NewC(c, &formatter{})
		require.NoError(t, err)

		err = s.Sort("-id")
		require.NoError(t, err)

		q := s.FormatQuery()
		require.Len(t, q, 2)

		assert.Equal(t, "-id", q.Get(ParamSort))
	})

	t.Run("Fieldset", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			s, err := NewC(c, &formatter{})
			require.NoError(t, err)

			s.setAllFields()

			q := s.FormatQuery()
			require.Len(t, q, 1)

			fieldsString := q.Get(fmt.Sprintf("fields[%s]", s.Struct().Collection()))
			fields := strings.Split(fieldsString, ",")
			assert.Len(t, fields, 5)
			assert.Contains(t, fields, "id")
			assert.Contains(t, fields, "attr")
			assert.Contains(t, fields, "rel")
			assert.Contains(t, fields, "fk")
			assert.Contains(t, fields, "lang")
		})

	})
}
