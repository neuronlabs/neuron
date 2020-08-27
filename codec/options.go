package codec

import (
	"github.com/neuronlabs/neuron/mapping"
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
	// Self should define this page query url.
	Self string `json:"self,omitempty"`
	// First should specify the first page query url.
	First string `json:"first,omitempty"`
	// Prev specify previous page query url.
	Prev string `json:"prev,omitempty"`
	// Next specify next page query url.
	Next string `json:"next,omitempty"`
	// Last specify the last page query url.
	Last string `json:"last,omitempty"`
	// Total defines the total number of the pages.
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

// MarshalOption is the option function that sets up the marshal options.
type MarshalOption func(o *MarshalOptions)

// MarshalWithLinks marshals the output with provided links
func MarshalWithLinks(links LinkOptions) MarshalOption {
	return func(o *MarshalOptions) {
		o.Link = links
	}
}

// MarshalSingleModel marshals the output with single model encoding type.
func MarshalSingleModel() MarshalOption {
	return func(o *MarshalOptions) {
		o.SingleResult = true
	}
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}

// UnmarshalOptions is the structure that contains unmarshal options.
// UnmarshalOptions requires model or model struct to be defined.
type UnmarshalOptions struct {
	StrictUnmarshal, ExpectSingle bool
	ModelStruct                   *mapping.ModelStruct
	Model                         mapping.Model
}

// UnmarshalOption is function that changes unmarshal options.
type UnmarshalOption func(o *UnmarshalOptions)

// UnmarshalStrictly is an unmarshal option that sets up option to strictly check the fields when unmarshaled the fields.
func UnmarshalStrictly() UnmarshalOption {
	return func(o *UnmarshalOptions) {
		o.StrictUnmarshal = true
	}
}

// UnmarshalWithModel is an unmarshal option that sets up model that should be unmarshaled.
func UnmarshalWithModel(model mapping.Model) UnmarshalOption {
	return func(o *UnmarshalOptions) {
		o.Model = model
	}
}

// UnmarshalWithModelStruct is an unmarshal option that sets up model struct that should be unmarshaled.
func UnmarshalWithModelStruct(modelStruct *mapping.ModelStruct) UnmarshalOption {
	return func(o *UnmarshalOptions) {
		o.ModelStruct = modelStruct
	}
}

// UnmarshalWithSingleExpectation is an unmarshal option that sets up the single model unmarshal expectation.
func UnmarshalWithSingleExpectation() UnmarshalOption {
	return func(o *UnmarshalOptions) {
		o.ExpectSingle = true
	}
}

// FieldAnnotations is structure that is extracted from the given sField codec specific annotation.
type FieldAnnotations struct {
	Name        string
	IsHidden    bool
	IsOmitEmpty bool
	Custom      []*mapping.FieldTag
}

// ExtractFieldAnnotations extracts codec specific tags.
func ExtractFieldAnnotations(sField *mapping.StructField, tag, tagSeparator, valuesSeparator string) FieldAnnotations {
	tags := sField.ExtractCustomFieldTags(tag, tagSeparator, valuesSeparator)
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
