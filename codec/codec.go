package codec

import (
	"io"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

var codecs = map[string]Codec{}

// StructTag is a constant used as a tag that defines models codecs.
const StructTag = "codec"

// RegisterCodec registers provided codec for given mime type.
func RegisterCodec(mime string, codec Codec) {
	_, ok := codecs[mime]
	if ok {
		log.Panicf("Codec for mime type '%s' is already registered", mime)
	}
	codecs[mime] = codec
}

// ListCodecs lists all registered codecs.
func ListCodecs() (codecList []Codec) {
	for _, c := range codecs {
		codecList = append(codecList, c)
	}
	return codecList
}

// GetCodec gets the codec by it's mime type.
func GetCodec(mime string) (Codec, bool) {
	c, ok := codecs[mime]
	return c, ok
}

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
	UnmarshalModels(data []byte, options *UnmarshalOptions) ([]mapping.Model, error)
}

// ParameterParser is an interface that parses parameters in given codec format.
type ParameterParser interface {
	ParseParameters(c *controller.Controller, q *query.Scope, parameters query.Parameters) error
}

// ParameterExtractor is an interface that extracts query parameters from given scope.
type ParameterExtractor interface {
	ExtractParameters(c *controller.Controller, q *query.Scope) (query.Parameters, error)
}
