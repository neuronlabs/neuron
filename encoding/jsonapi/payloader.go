package jsonapi

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
)

// payloader is used to encapsulate the One and Many payload types
type payloader interface {
	clearIncluded()
	setIncluded(included []*node)
}

// onePayload is used to represent a generic JSON API payload where a single
// resource (node) was included as an {} in the "data" key
type onePayload struct {
	Data     *node   `json:"data"`
	Included []*node `json:"included,omitempty"`
	Links    *Links  `json:"links,omitempty"`
	Meta     *Meta   `json:"meta,omitempty"`
}

func (p *onePayload) clearIncluded() {
	p.Included = []*node{}
}

func (p *onePayload) setIncluded(included []*node) {
	p.Included = included
}

// manyPayload is used to represent a generic JSON API payload where many
// resources (Nodes) were included in an [] in the "data" key
type manyPayload struct {
	Data     []*node `json:"data"`
	Included []*node `json:"included,omitempty"`
	Links    *Links  `json:"links,omitempty"`
	Meta     *Meta   `json:"meta,omitempty"`
}

func (p *manyPayload) clearIncluded() {
	p.Included = []*node{}
}

func (p *manyPayload) setIncluded(included []*node) {
	p.Included = included
}

type singleTypeRecognizePayload struct {
	Data *typeRecognizeNode `json:"data"`
}

type manyTypeRecognizePayload struct {
	Data []*typeRecognizeNode `json:"data"`
}

type typeRecognizeNode struct {
	Type string `json:"type"`
}

// node is used to represent a generic JSON API Resource
type node struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	// ClientID      string                 `json:"client-id,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	Relationships map[string]interface{} `json:"relationships,omitempty"`
	Links         *Links                 `json:"links,omitempty"`
	Meta          *Meta                  `json:"meta,omitempty"`
}

// relationshipOneNode is used to represent a generic has one JSON API relation
type relationshipOneNode struct {
	Data  *node  `json:"data"`
	Links *Links `json:"links,omitempty"`
	Meta  *Meta  `json:"meta,omitempty"`
}

// relationshipManyNode is used to represent a generic has many JSON API
// relation
type relationshipManyNode struct {
	Data  []*node `json:"data"`
	Links *Links  `json:"links,omitempty"`
	Meta  *Meta   `json:"meta,omitempty"`
}

// Links is used to represent a `links` object.
// http://jsonapi.org/format/#document-links
type Links map[string]interface{}

func (l *Links) validate() error {
	// Each member of a links object is a “link”. A link MUST be represented as
	// either:
	//  - a string containing the link’s URL.
	//  - an object (“link object”) which can contain the following members:
	//    - href: a string containing the link’s URL.
	//    - meta: a meta object containing non-standard meta-information about the
	//            link.
	for k, v := range *l {
		_, isString := v.(string)
		_, isLink := v.(Link)

		if !(isString || isLink) {
			return errors.Newf(class.EncodingUnmarshalInvalidType, "the %s member of the links object was not a string or link object", k)
		}
	}
	return nil
}

// Link is used to represent a member of the `links` object.
type Link struct {
	Href string `json:"href"`
	Meta Meta   `json:"meta,omitempty"`
}

// Linkable is used to include document links in response data
// e.g. {"self": "http://example.com/posts/1"}
type Linkable interface {
	JSONAPILinks() *Links
}

// Metable is used to include document meta in response data
// e.g. {"foo": "bar"}
type Metable interface {
	JSONAPIMeta() *Meta
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}

// RelationshipLinkable is used to include relationship links  in response data
// e.g. {"related": "http://example.com/posts/1/comments"}
type RelationshipLinkable interface {
	// JSONAPIRelationshipLinks will be invoked for each relationship with the corresponding relation name (e.g. `comments`)
	JSONAPIRelationshipLinks(relation string) *Links
}

// RelationshipMetable is used to include relationship meta in response data
type RelationshipMetable interface {
	// JSONRelationshipMeta will be invoked for each relationship with the corresponding relation name (e.g. `comments`)
	JSONAPIRelationshipMeta(relation string) *Meta
}
