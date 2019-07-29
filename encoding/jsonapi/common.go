package jsonapi

import (
	"github.com/neuronlabs/neuron-core/query"
)

const (
	// MediaType is the identifier for the JSON API media type
	// see http://jsonapi.org/format/#document-structure
	MediaType = "application/vnd.api+json"

	// ISO8601TimeFormat is the time formatting for the ISO 8601.
	ISO8601TimeFormat = "2006-01-02T15:04:05Z"
)

// EncodeLinks marks provided query 's' to encode the links while marshaling to jsonapi format.
func EncodeLinks(s *query.Scope) {
	s.StoreSet(encodeLinksCtxKey, struct{}{})

}

// encodeLinks is the structure used as a key in the store that states
// to encode the links for the encoder.
type encodeLinksKeyStruct struct{}

// StoreKeys
var encodeLinksCtxKey = encodeLinksKeyStruct{}
