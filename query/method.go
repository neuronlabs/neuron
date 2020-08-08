package query

// Method is the enum used for the query methods.
type Method int

// Enum values for the query methods.
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
