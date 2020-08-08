package codec

import (
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// ISO8601TimeFormat is the time formatting for the ISO 8601.
const ISO8601TimeFormat = "2006-01-02T15:04:05Z"

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
	RelationField string
}

// MarshalOptions is a structure that contains marshaling information.
type MarshalOptions struct {
	Link         LinkOptions
	SingleResult bool
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}

// UnmarshalOptions is the structure that contains unmarshal options.
type UnmarshalOptions struct {
	StrictUnmarshal   bool
	IncludedRelations []*query.IncludedRelation
	ModelStruct       *mapping.ModelStruct
}

// FieldAnnotations is structure that is extracted from the given sField codec specific annotation.
type FieldAnnotations struct {
	Name        string
	IsHidden    bool
	IsOmitEmpty bool
	Custom      []*mapping.FieldTag
}

// ExtractFieldAnnotations extracts codec specific tags.
func ExtractFieldAnnotations(sField *mapping.StructField, tag string) FieldAnnotations {
	tags := sField.ExtractCustomFieldTags(tag, ",", " ")
	if len(tags) == 0 {
		return FieldAnnotations{}
	}
	annotations := FieldAnnotations{}
	// The first tag would always be a
	annotations.Name = tags[0].Key
	if len(tags) > 1 {
		for _, tag := range tags[1:] {
			switch tag.Key {
			case "-":
				annotations.IsHidden = true
			case "omitempty":
				annotations.IsOmitEmpty = true
			default:
				annotations.Custom = append(annotations.Custom, tag)
			}
		}
	}
	return annotations
}
