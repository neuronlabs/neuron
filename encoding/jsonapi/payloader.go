package jsonapi

// Payloader is used to encapsulate the One and Many payload types
type Payloader interface {
	clearIncluded()
	setIncluded(included []*node)
}

// SinglePayload is used to represent a generic JSON API payload where a single
// resource (node) was included as an {} in the "data" key
type SinglePayload struct {
	Data     *node   `json:"data"`
	Included []*node `json:"included,omitempty"`
	Links    *Links  `json:"links,omitempty"`
	Meta     *Meta   `json:"meta,omitempty"`
}

func (p *SinglePayload) clearIncluded() {
	p.Included = []*node{}
}

func (p *SinglePayload) setIncluded(included []*node) {
	p.Included = included
}

var _ Payloader = &ManyPayload{}

// ManyPayload is used to represent a generic JSON API payload where many
// resources (Nodes) were included in an [] in the "data" key
type ManyPayload struct {
	Data     []*node `json:"data"`
	Included []*node `json:"included,omitempty"`
	Links    *Links  `json:"links,omitempty"`
	Meta     *Meta   `json:"meta,omitempty"`
}

func (p *ManyPayload) clearIncluded() {
	p.Included = []*node{}
}

func (p *ManyPayload) setIncluded(included []*node) {
	p.Included = included
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

// Link is used to represent a member of the `links` object.
type Link struct {
	Href string `json:"href"`
	Meta Meta   `json:"meta,omitempty"`
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}
