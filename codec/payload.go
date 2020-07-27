package codec

import (
	"io"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Payload is the default structure used by codecs to marshal and unmarshal.
type Payload struct {
	// Payload defined model structure.
	ModelStruct *mapping.ModelStruct
	// Data contains models data.
	Data []mapping.Model
	// FieldSets is the index based field sets that maps it's indexes with the data.
	FieldSets []mapping.FieldSet
	// Meta is an object containing non-standard meta-information about the error.
	Meta Meta
	// IncludedRelations is the information about included relations in the payload.
	IncludedRelations []*query.IncludedRelation
	// Options are the options for the codecs how to treat the payload.
	PaginationLinks *PaginationLinks
	// MarshalLinks are the links used to marshal for specific codecs.
	MarshalLinks *LinkOptions
	// MarshalSingularFormat marshals single model in a singular format for given codec.
	MarshalSingularFormat bool
}

// PayloadMarshaler is the interface used to marshal payload into provided writer..
type PayloadMarshaler interface {
	MarshalPayload(w io.Writer, payload *Payload) error
}

// PayloadUnmarshaler is the interface used to unmarshal payload from given reader for provided codec type.
type PayloadUnmarshaler interface {
	UnmarshalPayload(r io.Reader, options UnmarshalOptions) (*Payload, error)
}
