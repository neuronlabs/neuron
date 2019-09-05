package query

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

// Many2ManyModel is the model with many2many relationship.
type Many2ManyModel struct {
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
