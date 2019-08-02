package class

import (
	"github.com/neuronlabs/errors"
)

// MjrEncoding - major that classifies errors related with the 'jsonapi' encoding
var MjrEncoding errors.Major

func registerEncodingClasses() {
	MjrEncoding = errors.MustNewMajor()

	registerEncodingMarshal()
	registerEncodingUnmarshal()
}

/**

Encoding Marshal

*/
var (
	// MnrEncodingMarshal is the minor error classification for the 'MjrEncoding' major.
	// It is related with the encoding marshal errors.
	MnrEncodingMarshal errors.Minor

	// EncodingMarshalOutputValue is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the setting the output value.
	EncodingMarshalOutputValue errors.Class

	// EncodingMarshalModelNotMapped is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when marshaling unmapped model.
	EncodingMarshalModelNotMapped errors.Class

	// EncodingMarshalNilValue is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is nil.
	EncodingMarshalNilValue errors.Class

	// EncodingMarshalNonAddressable is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is not addressable.
	EncodingMarshalNonAddressable errors.Class

	// EncodingMarshalInput is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is invalid. I.e. provided value is not a model structure.
	EncodingMarshalInput errors.Class
)

func registerEncodingMarshal() {
	MnrEncodingMarshal = errors.MustNewMinor(MjrEncoding)

	mjr, mnr := MjrEncoding, MnrEncodingMarshal
	EncodingMarshalOutputValue = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingMarshalModelNotMapped = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingMarshalNilValue = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingMarshalNonAddressable = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingMarshalInput = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
}

/**

Encoding Unmarshal

*/

var (
	// MnrEncodingUnmarshal is the 'MjrEncoding' minor classification
	// for the unmarshall process.
	MnrEncodingUnmarshal errors.Minor

	// EncodingUnmarshal is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// for general purpose unmarshal errors.
	EncodingUnmarshal errors.Class

	// EncodingUnmarshalCollection is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when unmarshaling the input data that has undefined model collection.
	EncodingUnmarshalCollection errors.Class

	// EncodingUnmarshalFieldValue is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// When unmarshaling field with invalid value.
	EncodingUnmarshalFieldValue errors.Class

	// EncodingUnmarshalInvalidFormat is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the input data has invalid format. I.e. invalid json formatting.
	EncodingUnmarshalInvalidFormat errors.Class

	// EncodingUnmarshalInvalidID is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the document with invalid primary field format or value.
	EncodingUnmarshalInvalidID errors.Class

	// EncodingUnmarshalInvalidOutput is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the document with invalid output model or slice.
	EncodingUnmarshalInvalidOutput errors.Class

	// EncodingUnmarshalInvalidTime is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when unmarshaling time defined field with invalid type or value.
	EncodingUnmarshalInvalidTime errors.Class

	// EncodingUnmarshalInvalidType is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling field with invalid value type.
	EncodingUnmarshalInvalidType errors.Class

	// EncodingUnmarshalNoData is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the empty data, when it is not allowed.
	EncodingUnmarshalNoData errors.Class

	// EncodingUnmarshalUnknownField is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when trying to unmarshal undefined or unknown field. Used when 'strict' mode is set.
	EncodingUnmarshalUnknownField errors.Class

	// EncodingUnmarshalValueOutOfRange is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when trying to unmarshal value with some range that exceeds defined maximum.
	EncodingUnmarshalValueOutOfRange errors.Class
)

func registerEncodingUnmarshal() {
	MnrEncodingUnmarshal = errors.MustNewMinor(MjrEncoding)

	mjr, mnr := MjrEncoding, MnrEncodingUnmarshal
	EncodingUnmarshal = errors.MustNewMinorClass(mjr, mnr)
	EncodingUnmarshalCollection = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalFieldValue = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalInvalidFormat = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalInvalidID = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalInvalidOutput = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalInvalidTime = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalInvalidType = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalNoData = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalUnknownField = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	EncodingUnmarshalValueOutOfRange = errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
}
