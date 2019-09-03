package jsonapi

// LinkType is the link type used for marshaling.
type LinkType int

// Link type enumerators.
const (
	DefaultLink LinkType = iota
	RelationshipLink
	NoLink
)

// MarshalOptions is the struct that contains marshaling options.
type MarshalOptions struct {
	Link    LinkType
	LinkURL string

	RootCollection string `validate:"required"`
	RootID         string `validate:"required"`
	RelatedField   string `validate:"required"`

	Meta             map[string]interface{}
	RelationshipMeta map[string]Meta
}
