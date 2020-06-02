package codec

import (
	"github.com/neuronlabs/neuron/query"
)

// LinkType is the link type used for marshaling.
type LinkType int

// Link type enumerators.
const (
	NoLink LinkType = iota
	ResourceLink
	RelatedLink
	RelationshipLink
)

// PaginationLinks is the structure that contain options for the pagination links.
// https://jsonapi.org/examples/#pagination
type PaginationLinks struct {
	// Self should be just a query too
	Self  string `json:"self,omitempty"`
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`

	Total int64 `json:"total"`
}

// LinkOptions contains link options required for marshaling codec data.
type LinkOptions struct {
	// Type defines link type.
	Type LinkType

	// BaseURL should be the common base url for all the links.
	BaseURL string
	// Collection is the link root collection name.
	Collection string
	// RootID is the root collection primary value.
	RootID string

	// RelatedField is the related field name used in the relationship link type.
	RelatedField string

	// PaginationLinks contains the query values for the pagination links like:
	// 'first', 'prev', 'next', 'last'. The values should be only the query values
	// without the '?' sign.
	PaginationLinks *PaginationLinks
}

type MarshalOptions struct {
	Link LinkOptions

	// RelationQueries
	RelationQueries []*query.Scope

	// ManyResults defines if the marshaled value should be a slice.
	ManyResults    bool
	RootCollection string
	RootID         string
	RelatedField   string

	Meta             map[string]interface{}
	RelationshipMeta map[string]Meta
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}
