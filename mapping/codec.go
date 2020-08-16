package mapping

import (
	"github.com/neuronlabs/neuron/errors"
)

// Database tag is defined with 'codec':
// It's subtags are split with ';'.
// The first subtag is always codec name - if the name is empty or '_' then it's value would not be set.
// The other subtags could be composed in two ways:
// subtag_name:subtag_value1,subtag_value2 OR subtag_name.
// Subtag values are split using comma ','.
//
// Neuron defined codec subtags:
// 	'omitempty' 	- omits given field in unmarshal/marshal process.
//	'-'				- field is skipped in the codec.
//
//
// Example codec tag:
// codec:"-" 					- skip this field in marshal/unmarshal process
// codec:"codec_name;omitempty" - defined codec_name and codec field if empty is omited on marshaling process.
// codec:";omitempty" 			- codec field if empty is omited on marshaling process
// codec:"_;omitempty" 			- codec field if empty is omited on marshaling process

func (m *ModelStruct) extractCodecTags() error {
	for _, field := range m.structFields {
		tags := field.extractFieldTags("codec", AnnotationTagSeparator, AnnotationSeparator)
		if len(tags) == 0 {
			continue
		}

		// The first tag is always database name.
		name := tags[0]
		if name.Key == "-" {
			field.fieldFlags |= fCodecSkip
			continue
		}
		switch name.Key {
		case "name":
			if len(name.Values) != 1 {
				return errors.Wrapf(ErrMapping, "model's: %s field: '%s' codec tag name defined without value", m, field)
			}
			field.codecName = name.Values[0]
		case "_":
		default:
			field.codecName = name.Key
		}

		for _, tag := range tags[1:] {
			switch tag.Key {
			case "omitempty":
				field.fieldFlags |= fOmitempty
			}
		}
	}
	return nil
}
