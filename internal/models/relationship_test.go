package models

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// definining many2many requires join model, backref field, and foreign key
// join model is the join table model used for getting the many 2 many relationships values
// backref field is a field in the join model

// foreign key is the related model's backref field in the join model

type Model1WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model2WithMany2Many `neuron:"type=relation;many2many=joinModel;foreign=SecondForeign"`
}

type Model2WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model1WithMany2Many `neuron:"type=relation;many2many=joinModel,SecondForeign;foreign=model1WithMany2ManyID"`
}

type joinModel struct {
	ID int `neuron:"type=primary"`

	// First model
	First                 *Model1WithMany2Many `neuron:"type=relation;foreign=Model1WithMany2ManyID"`
	Model1WithMany2ManyID int                  `neuron:"type=foreign"`

	// Second
	Second        *Model2WithMany2Many `neuron:"type=foreign;foreign=SecondForeign"`
	SecondForeign int                  `neuron:"type=foreign"`
}

// First is the many2many model
type First struct {
	ID   int       `neuron:"type=primary"`
	Many []*Second `neuron:"type=relation;many2many"`
}

// Second is the many2many model
type Second struct {
	ID     int      `neuron:"type=primary"`
	Firsts []*First `neuron:"type=relation;many2many"`
}

// FirstSeconds is the join table
type FirstSeconds struct {
	ID       int `neuron:"type=primary"`
	FirstID  int `neuron:"type=foreign"`
	SecondID int `neuron:"type=foreign"`
}

type modelWithHasMany struct {
	ID      int                    `neuron:"type=primary"`
	HasMany []*modelWithForeignKey `neuron:"type=relation;foreign=ForeignKey"`
}

type modelWithForeignKey struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
}

type modelWithBelongsTo struct {
	ID         int              `neuron:"type=primary"`
	ForeignKey int              `neuron:"type=foreign"`
	BelongsTo  *modelWithHasOne `neuron:"type=relation;foreign=ForeignKey"`
}

type modelWithHasOne struct {
	ID     int                 `neuron:"type=primary"`
	HasOne *modelWithBelongsTo `neuron:"type=relation;foreign=ForeignKey"`
}

func TestMappedRelationships(t *testing.T) {

	t.Run("many2many", func(t *testing.T) {

		t.Run("PredefinedFields", func(t *testing.T) {
			s := testingSchemas(t)

			err := s.RegisterModels(Model1WithMany2Many{}, Model2WithMany2Many{}, joinModel{})
			require.NoError(t, err)

			join, err := s.GetModelStruct(joinModel{})
			require.NoError(t, err)

			assert.True(t, join.isJoin)

			first, err := s.GetModelStruct(Model1WithMany2Many{})
			require.NoError(t, err)

			second, err := s.GetModelStruct(Model2WithMany2Many{})
			require.NoError(t, err)
			t.Run("First", func(t *testing.T) {

				relField, ok := first.relationships["synced"]
				require.True(t, ok)

				require.NotNil(t, relField.relationship)
				require.Equal(t, RelMany2Many, relField.relationship.kind)

				assert.Equal(t, second, relField.relationship.mStruct)

				firstBackref, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstBackref, relField.relationship.backReferenceForeignKey)

				secondFK, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondFK, relField.relationship.foreignKey)
			})

			t.Run("Second", func(t *testing.T) {
				relField, ok := second.relationships["synced"]
				require.True(t, ok)

				rel := relField.relationship
				require.NotNil(t, rel)
				require.Equal(t, RelMany2Many, rel.kind)

				assert.Equal(t, first, rel.mStruct)

				// check the backreference key
				secondBackRef, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondBackRef, rel.backReferenceForeignKey)

				// check the foreign key
				firstFK, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstFK, rel.foreignKey)
			})
		})

		t.Run("DefaultSettings", func(t *testing.T) {
			s := testingSchemas(t)

			err := s.RegisterModels(First{}, Second{}, FirstSeconds{})
			require.NoError(t, err)

			first, err := s.GetModelStruct(First{})
			require.NoError(t, err)

			second, err := s.GetModelStruct(Second{})
			require.NoError(t, err)

			firstSeconds, err := s.GetModelStruct(FirstSeconds{})
			require.NoError(t, err)

			firstRel, ok := first.RelationshipField("Many")
			require.True(t, ok)

			fID, ok := firstSeconds.ForeignKey("FirstID")
			require.True(t, ok)

			sID, ok := firstSeconds.ForeignKey("SecondID")
			require.True(t, ok)

			relFirst := firstRel.relationship
			if assert.NotNil(t, relFirst) {
				assert.Equal(t, fID, relFirst.backReferenceForeignKey)
				assert.Equal(t, sID, relFirst.foreignKey)
				assert.Equal(t, RelMany2Many, relFirst.kind)
			}

			secondRel, ok := second.RelationshipField("Firsts")
			require.True(t, ok)

			relSecond := secondRel.relationship
			if assert.NotNil(t, relSecond) {
				assert.Equal(t, sID, relSecond.backReferenceForeignKey)
				assert.Equal(t, fID, relSecond.foreignKey)
				assert.Equal(t, RelMany2Many, relSecond.kind)
			}
		})

	})

	t.Run("hasMany", func(t *testing.T) {
		t.Run("synced", func(t *testing.T) {
			s := testingSchemas(t)

			// get the models
			require.NoError(t, s.RegisterModels(modelWithHasMany{}, modelWithForeignKey{}))

			// get hasMany model
			hasManyModel, err := s.getModelStruct(modelWithHasMany{})
			require.NoError(t, err)

			hasManyField, ok := hasManyModel.relationships["has_many"]
			require.True(t, ok)

			fkModel, err := s.getModelStruct(modelWithForeignKey{})
			require.NoError(t, err)

			fk, ok := fkModel.foreignKeys["foreign_key"]
			require.True(t, ok)

			if assert.NotNil(t, hasManyField.relationship) {
				assert.Equal(t, fk, hasManyField.relationship.foreignKey)
				assert.Equal(t, RelHasMany, hasManyField.relationship.kind)
				assert.Equal(t, fkModel, hasManyField.relationship.mStruct)

			}
		})
	})

	s := testingSchemas(t)

	require.NoError(t, s.RegisterModels(modelWithBelongsTo{}, modelWithHasOne{}))

	t.Run("belongsTo", func(t *testing.T) {
		model, err := s.getModelStruct(modelWithBelongsTo{})
		require.NoError(t, err)

		belongsToField, ok := model.relationships["belongs_to"]
		require.True(t, ok)

		if assert.NotNil(t, belongsToField.relationship) {
			assert.Equal(t, RelBelongsTo, belongsToField.relationship.kind)
		}
	})

	t.Run("hasOne", func(t *testing.T) {

		model, err := s.getModelStruct(modelWithHasOne{})
		require.NoError(t, err)

		hasOneField, ok := model.relationships["has_one"]
		require.True(t, ok)

		belongsToModel, err := s.getModelStruct(modelWithBelongsTo{})
		require.NoError(t, err)

		fk, ok := belongsToModel.foreignKeys["foreign_key"]
		require.True(t, ok)

		if assert.NotNil(t, hasOneField.relationship) {
			assert.Equal(t, RelHasOne, hasOneField.relationship.kind)
			assert.Equal(t, fk, hasOneField.relationship.foreignKey)
		}
	})
}
