package mapping

import (
	"time"

	"github.com/google/uuid"
)

//go:generate neurogns models methods methods --single-file --format=goimports .

// Model1WithMany2Many is the model with the many2many relationship.
type Model1WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model2WithMany2Many `neuron:"type=relation;many2many=JoinModel;foreign=_,SecondForeign"`
}

// Model2WithMany2Many is the second model with the many2many relationship.
type Model2WithMany2Many struct {
	ID     int                    `neuron:"type=primary"`
	Synced []*Model1WithMany2Many `neuron:"type=relation;many2many=JoinModel;foreign=SecondForeign"`
}

// JoinModel is the relation join model.
type JoinModel struct {
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

// ModelWithHasMany is a model with has many relation.
type ModelWithHasMany struct {
	ID      int                    `neuron:"type=primary"`
	HasMany []*ModelWithForeignKey `neuron:"type=relation;foreign=ForeignKey"`
}

// ModelWithForeignKey is a model with foreign key.
type ModelWithForeignKey struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
}

// ModelWithBelongsTo is a model with belongs to relationship.
type ModelWithBelongsTo struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
	// in belongs to relationship - foreign key must be the primary of the relation field
	BelongsTo *ModelWithHasOne `neuron:"type=relation;foreign=ForeignKey"`
}

// ModelWithHasOne is a model with HasOne relationship.
type ModelWithHasOne struct {
	ID int `neuron:"type=primary"`
	// in has one relationship - foreign key must be the field that is the same
	HasOne *ModelWithBelongsTo `neuron:"type=relation;foreign=ForeignKey"`
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

// Job is the model for a single time sheet job instance
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

// User is the model that represents user's contact in the time sheet
type User struct {
	ID string

	Cars []*Car
	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	FirstName string
	LastName  string
	Email     string

	Jobs []*Job `neuron:"foreign=CreatorID"`
}

// NotTaggedModel is the model used for testing that has no tagged fields.
type NotTaggedModel struct {
	ID                    int
	Name                  string
	Age                   int
	Created               time.Time
	OtherNotTaggedModelID int
	Related               *OtherNotTaggedModel
	ManyRelationID        int `neuron:"type=foreign"`
}

// OtherNotTaggedModel is the testing model with no neuron tags, related to the NotTaggedModel.
type OtherNotTaggedModel struct {
	ID            int
	SingleRelated *NotTaggedModel
	ManyRelation  []*NotTaggedModel
}

// SubNested is a model for nested fields tests.
type SubNested struct {
	InceptionFirst  int
	InceptionSecond float64 `neuron:"name=second;flags=omitempty"`
}

// NestedAttribute is a nested attribute field for the ModelWithNested.
type NestedAttribute struct {
	Float     float64
	Int       int
	String    string
	Slice     []int
	Inception SubNested

	FloatTagged float64 `neuron:"name=float-tag"`
	IntTagged   int     `neuron:"name=int-tag"`
}

// ModelWithNested is a model with nested field.
type ModelWithNested struct {
	ID          int              `neuron:"type=primary"`
	PtrComposed *NestedAttribute `neuron:"type=attr;name=ptr-composed"`
}

// OrderModel is a model used to test order of fields.
type OrderModel struct {
	ID     uuid.UUID
	Name   string
	First  string
	Second int
	Third  string
}

// CarBrand is the included testing model.
type CarBrand struct {
	ID int
}

// Car is the included testing model.
type Car struct {
	ID     int
	UserID string

	Brand   *CarBrand
	BrandID int
}

// TModel is test model for mapping.
type TModel struct {
	ID          int
	CreatedTime time.Time  `neuron:"flags=created_at"`
	UpdatedTime *time.Time `neuron:"flags=updated_at"`
	DeletedTime *time.Time `neuron:"flags=deleted_at"`
	Number      int
	String      string
}

// BenchModel is testing model for mapping benchmarks.
type BenchModel struct {
	ID     int
	Name   string
	First  string
	Second int
	Third  string
}

// Timer is a model with time related fields.
type Timer struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

// InvalidCreatedAt is a model with invalid created at field.
type InvalidCreatedAt struct {
	ID        int
	CreatedAt string
}

// InvalidDeletedAt is an invalid deleted at model.
type InvalidDeletedAt struct {
	ID        int
	DeletedAt time.Time `neuron:"flags=deleted_at"`
}

// InvalidUpdatedAt is an invalid updated at model.
type InvalidUpdatedAt struct {
	ID        int
	UpdatedAt int `neuron:"flags=updated_at"`
}
