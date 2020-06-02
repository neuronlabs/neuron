package codec

import (
	"io"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

var codecs = map[string]Codec{}

// RegisterCodec registers provided codec for given mime type.
func RegisterCodec(mime string, codec Codec) {
	_, ok := codecs[mime]
	if ok {
		log.Panicf("Codec for mime type '%s' is already registered", mime)
	}
	codecs[mime] = codec
}

// GetCodec gets the codec by it's mime type.
func GetCodec(mime string) (Codec, bool) {
	c, ok := codecs[mime]
	return c, ok
}

type Codec interface {
	// MarshalModels of type 'modelStruct' into provided writer 'w'. Returns number of bytes written or an error.
	MarshalModels(w io.Writer, modelStruct *mapping.ModelStruct, models []mapping.Model, options *MarshalOptions) error
	// UnmarshalModels from given reader 'r' of type 'modelStruct'.
	UnmarshalModels(modelStruct *mapping.ModelStruct, r io.Reader, strict bool) ([]mapping.Model, error)
	// MarshalQuery 'scope' into given writer 'w'.
	MarshalQuery(w io.Writer, s *query.Scope, options *MarshalOptions) error
	// UnmarshalQuery 'scope' for 'modelStruct' from the 'r' reader.
	UnmarshalQuery(r io.Reader, c *controller.Controller, modelStruct *mapping.ModelStruct, strict bool) (*query.Scope, error)
	// MarshalErrors marshals given errors.
	MarshalErrors(w io.Writer, errors MultiError) error
	// UnmarshalErrors unmarshal provided errors.
	UnmarshalErrors(r io.Reader) (MultiError, error)
}
