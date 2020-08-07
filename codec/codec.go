package codec

import (
	"io"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// StructTag is a constant used as a tag that defines models codecs.
const StructTag = "codec"

type Codec interface {
	// MarshalErrors marshals given errors.
	MarshalErrors(w io.Writer, errors ...*Error) error
	// UnmarshalErrors unmarshal provided errors.
	UnmarshalErrors(r io.Reader) (MultiError, error)
	// MimeType returns the mime type that this codec is defined for.
	MimeType() string
}

// ModelMarshaler is an interface that allows to marshal provided models.
type ModelMarshaler interface {
	// MarshalModels marshal provided models into given codec encoding type. The function should
	// simply encode only provided models without any additional payload like metadata.
	MarshalModels(models []mapping.Model, options MarshalOptions) ([]byte, error)
}

// Model Unmarshaler is an interface that allows to unmarshal provided models of given model struct.
type ModelUnmarshaler interface {
	// UnmarshalModels unmarshals provided data into mapping.Model slice. The data should simply be only encoded models.
	UnmarshalModels(data []byte, options UnmarshalOptions) ([]mapping.Model, error)
}

// ParameterParser is an interface that parses parameters in given codec format.
type ParameterParser interface {
	ParseParameters(c *controller.Controller, q *query.Scope, parameters query.Parameters) error
}

// ParameterExtractor is an interface that extracts query parameters from given scope.
type ParameterExtractor interface {
	ExtractParameters(c *controller.Controller, q *query.Scope) (query.Parameters, error)
}
