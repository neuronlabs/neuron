package query

// Method is the enum used for the query methods.
type Method int

const (
	InvalidMethod Method = iota
	Insert
	InsertMany
	InsertRelationship
	Get
	GetRelationship
	GetRelated
	List
	Update
	UpdateMany
	UpdateRelationship
	Delete
	DeleteMany
	DeleteRelationship
)
