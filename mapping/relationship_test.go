package mapping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
)

// Model1WithMany2Many is the model with the many2many relationship.
type Model1WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model2WithMany2Many `neuron:"type=relation;many2many=joinModel;foreign=_,SecondForeign;on_delete=on_error=continue,order=2"`
}

// Model2WithMany2Many is the second model with the many2many relationship.
type Model2WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model1WithMany2Many `neuron:"type=relation;many2many=joinModel;foreign=SecondForeign"`
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
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
	// in belongs to relationship - foreign key must be the primary of the relation field
	BelongsTo *modelWithHasOne `neuron:"type=relation;foreign=ForeignKey"`
}

type modelWithHasOne struct {
	ID int `neuron:"type=primary"`
	// in has one relatinship - foreign key must be the field that is the same
	HasOne *modelWithBelongsTo `neuron:"type=relation;foreign=ForeignKey"`
}

// Comment defines the model for job comment.
type Comment struct {
	ID string
	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// User relation
	UserID string
	User   *User

	// Job relation
	JobID string
	Job   *Job
}

// Job is the model for a single timesheet job instance
type Job struct {
	ID string
	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// StartAt defines the start timestamp of the given job
	StartAt *time.Time
	// EndAt defines the end timestamp for given job
	EndAt *time.Time

	// Title defines shortly the job
	Title string

	// Relation with the creator
	CreatorID string `neuron:"type=foreign"`
	Creator   *User

	// HaveMany jobs relationship
	Comments []*Comment
}

// User is the model that represents user's contanct in the timesheet
type User struct {
	ID string
	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	FirstName string
	LastName  string
	Email     string

	Jobs []*Job `neuron:"foreign=CreatorID"`
}

// TestMappedRelationships tests the mapped relationships.
func TestMappedRelationships(t *testing.T) {
	t.Run("many2many", func(t *testing.T) {
		t.Run("PredefinedFields", func(t *testing.T) {
			m := testingModelMap(t)

			err := m.RegisterModels(Model1WithMany2Many{}, Model2WithMany2Many{}, joinModel{})
			require.NoError(t, err)

			join, err := m.GetModelStruct(joinModel{})
			require.NoError(t, err)

			assert.True(t, join.isJoin)

			first, err := m.GetModelStruct(Model1WithMany2Many{})
			require.NoError(t, err)

			second, err := m.GetModelStruct(Model2WithMany2Many{})
			require.NoError(t, err)

			t.Run("First", func(t *testing.T) {
				relField, ok := first.relationshipField("synced")
				require.True(t, ok)

				rel := relField.relationship
				require.NotNil(t, rel)

				assert.True(t, rel.isMany2Many())
				require.Equal(t, RelMany2Many, relField.relationship.kind)

				assert.Equal(t, second, relField.relationship.mStruct)

				firstForeign, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstForeign, relField.relationship.foreignKey)

				secondFK, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondFK, relField.relationship.mtmRelatedForeignKey)
			})

			t.Run("Second", func(t *testing.T) {
				relField, ok := second.relationshipField("synced")
				require.True(t, ok)

				rel := relField.relationship
				require.NotNil(t, rel)
				require.Equal(t, RelMany2Many, rel.kind)

				assert.Equal(t, first, rel.mStruct)

				// check the backreference key
				secondForeign, ok := join.ForeignKey("SecondForeign")
				require.True(t, ok)

				assert.Equal(t, secondForeign, rel.foreignKey)

				// check the foreign key
				firstFK, ok := join.ForeignKey("Model1WithMany2ManyID")
				require.True(t, ok)

				assert.Equal(t, firstFK, rel.mtmRelatedForeignKey)
			})
		})

		t.Run("DefaultSettings", func(t *testing.T) {
			m := testingModelMap(t)

			err := m.RegisterModels(First{}, Second{}, FirstSeconds{})
			require.NoError(t, err)

			first, err := m.GetModelStruct(First{})
			require.NoError(t, err)

			second, err := m.GetModelStruct(Second{})
			require.NoError(t, err)

			firstSeconds, err := m.GetModelStruct(FirstSeconds{})
			require.NoError(t, err)

			firstRel, ok := first.RelationByName("Many")
			require.True(t, ok)

			fID, ok := firstSeconds.ForeignKey("FirstID")
			require.True(t, ok)

			sID, ok := firstSeconds.ForeignKey("SecondID")
			require.True(t, ok)

			relFirst := firstRel.relationship
			if assert.NotNil(t, relFirst) {
				assert.Equal(t, fID, relFirst.foreignKey)
				assert.Equal(t, sID, relFirst.mtmRelatedForeignKey)
				assert.Equal(t, RelMany2Many, relFirst.kind)
			}

			secondRel, ok := second.RelationByName("Firsts")
			require.True(t, ok)

			relSecond := secondRel.relationship
			if assert.NotNil(t, relSecond) {
				assert.Equal(t, sID, relSecond.foreignKey)
				assert.Equal(t, fID, relSecond.mtmRelatedForeignKey)
				assert.Equal(t, RelMany2Many, relSecond.kind)
			}
		})

		t.Run("WithoutJoinTable", func(t *testing.T) {
			t.Run("NotRegistered", func(t *testing.T) {
				m := testingModelMap(t)

				err := m.RegisterModels(First{}, Second{})
				require.Error(t, err)

				e, ok := err.(errors.DetailedError)
				require.True(t, ok)
				assert.Equal(t, class.ModelRelationshipJoinModel, e.Class())
			})

			t.Run("NotDefinedInTag", func(t *testing.T) {

			})
		})
	})

	t.Run("hasMany", func(t *testing.T) {
		t.Run("synced", func(t *testing.T) {
			m := testingModelMap(t)

			// get the models
			require.NoError(t, m.RegisterModels(modelWithHasMany{}, modelWithForeignKey{}))

			// get hasMany model
			hasManyModel, err := m.GetModelStruct(modelWithHasMany{})
			require.NoError(t, err)

			hasManyField, ok := hasManyModel.relationshipField("has_many")
			require.True(t, ok)

			fkModel, err := m.GetModelStruct(modelWithForeignKey{})
			require.NoError(t, err)

			fk, ok := fkModel.ForeignKey("foreign_key")
			require.True(t, ok)

			if assert.NotNil(t, hasManyField.relationship) {
				assert.Equal(t, fk, hasManyField.relationship.foreignKey)
				assert.Equal(t, RelHasMany, hasManyField.relationship.kind)
				assert.Equal(t, fkModel, hasManyField.relationship.mStruct)
			}
		})
	})

	t.Run("SingleRelations", func(t *testing.T) {
		m := testingModelMap(t)

		require.NoError(t, m.RegisterModels(modelWithBelongsTo{}, modelWithHasOne{}))

		t.Run("belongsTo", func(t *testing.T) {
			model, err := m.GetModelStruct(modelWithBelongsTo{})
			require.NoError(t, err)

			belongsToField, ok := model.relationshipField("belongs_to")
			require.True(t, ok)

			if assert.NotNil(t, belongsToField.relationship) {
				assert.Equal(t, RelBelongsTo, belongsToField.relationship.kind)
			}
			relFields := model.RelationFields()
			assert.Len(t, relFields, 1)
		})

		t.Run("hasOne", func(t *testing.T) {
			model, err := m.GetModelStruct(modelWithHasOne{})
			require.NoError(t, err)

			hasOneField, ok := model.relationshipField("has_one")
			require.True(t, ok)

			belongsToModel, err := m.GetModelStruct(modelWithBelongsTo{})
			require.NoError(t, err)

			fk, ok := belongsToModel.ForeignKey("foreign_key")
			require.True(t, ok)

			if assert.NotNil(t, hasOneField.relationship) {
				assert.Equal(t, RelHasOne, hasOneField.relationship.kind)
				assert.Equal(t, fk, hasOneField.relationship.foreignKey)
			}
			relFields := model.RelationFields()
			assert.Len(t, relFields, 1)
		})
	})

	t.Run("MultipleRelations", func(t *testing.T) {
		m := testingModelMap(t)

		err := m.RegisterModels(Comment{}, User{}, Job{})
		require.NoError(t, err)

		model, err := m.GetModelStruct(Comment{})
		require.NoError(t, err)

		relFields := model.RelationFields()
		assert.Len(t, relFields, 2)
	})
}
