package query

//go:generate neuron-generator models methods --single-file --exclude=Transaction,Operator .

// Formatter is the test model for the query tests.
type Formatter struct {
	ID   int                `neuron:"type=primary"`
	Attr string             `neuron:"type=attr"`
	Rel  *FormatterRelation `neuron:"type=relation;foreign=FK"`
	FK   int                `neuron:"type=foreign"`
	Lang string             `neuron:"type=attr;flags=lang"`
}

// FormatterRelation is the relation model for the query tests.
type FormatterRelation struct {
	ID int `neuron:"type=primary"`
}

// TestingModel is one of testing models for the query package.
type TestingModel struct {
	ID         int                  `neuron:"type=primary"`
	Attr       string               `neuron:"type=attr"`
	Relation   *FilterRelationModel `neuron:"type=relation;foreign=ForeignKey"`
	ForeignKey int                  `neuron:"type=foreign"`
	Nested     *FilterNestedModel   `neuron:"type=attr"`
}

// FilterRelationModel is a relation for the filter tests.
type FilterRelationModel struct {
	ID int `neuron:"type=primary"`
}

// FilterNestedModel is nested field for the model.
type FilterNestedModel struct {
	Field string
}

// HasOneModel is the model that have has-one relationship.
type HasOneModel struct {
	ID     int           `neuron:"type=primary"`
	HasOne *ForeignModel `neuron:"type=relation;foreign=ForeignKey"`
}

// HasManyModel is the model with the has-many relationship.
type HasManyModel struct {
	ID      int             `neuron:"type=primary"`
	HasMany []*ForeignModel `neuron:"type=relation;foreign=ForeignKey"`
}

// ForeignModel is the model that have foreign key.
type ForeignModel struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
}

// ManyToManyModel is the model with many2many relationship.
type ManyToManyModel struct {
	ID        int             `neuron:"type=primary"`
	Many2Many []*RelatedModel `neuron:"type=relation;many2many=JoinModel;foreign=ForeignKey,MtMForeignKey"`
}

// JoinModel is the model used as a join model for the many2many relationships.
type JoinModel struct {
	ID            int `neuron:"type=primary"`
	ForeignKey    int `neuron:"type=foreign"`
	MtMForeignKey int `neuron:"type=foreign"`
}

// RelatedModel is the related model in the many2many relationship.
type RelatedModel struct {
	ID         int     `neuron:"type=primary"`
	FloatField float64 `neuron:"type=attr"`
}

type ForeignWithRelation struct {
	ID         int
	Relation   *HasManyWithRelation `neuron:"foreign=ForeignKey"`
	ForeignKey int                  `neuron:"type=foreign"`
}

type HasManyWithRelation struct {
	ID       int
	Relation []*ForeignWithRelation `neuron:"foreign=ForeignKey"`
}
