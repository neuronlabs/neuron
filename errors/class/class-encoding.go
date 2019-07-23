package class

// MjrEncoding - major that classifies errors related with the 'jsonapi' encoding
var MjrEncoding Major

func registerEncodingClasses() {
	MjrEncoding = MustRegisterMajor("Encoding", "encoding related issues")

	registerEncodingMarshal()
	registerEncodingUnmarshal()
}

/**

Encoding Marshal

*/
var (
	// MnrEncodingMarshal is the minor error classification for the 'MjrEncoding' major.
	// It is related with the encoding marshal errors.
	MnrEncodingMarshal Minor

	// EncodingMarshalOutputValue is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the setting the output value.
	EncodingMarshalOutputValue Class

	// EncodingMarshalModelNotMapped is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when marshaling unmapped model.
	EncodingMarshalModelNotMapped Class

	// EncodingMarshalNilValue is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is nil.
	EncodingMarshalNilValue Class

	// EncodingMarshalNonAddressable is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is not addressable.
	EncodingMarshalNonAddressable Class

	// EncodingMarshalInput is the 'MjrEncoding', 'MnrEncodingMarshal' error classification
	// when the provided marshaling value is invalid. I.e. provided value is not a model structure.
	EncodingMarshalInput Class
)

func registerEncodingMarshal() {
	MnrEncodingMarshal = MjrEncoding.MustRegisterMinor("Marshal", "marshaling to given encoding")

	EncodingMarshalOutputValue = MnrEncodingMarshal.MustRegisterIndex("Output Value", "marshaling process failed on setting the output value").Class()
	EncodingMarshalModelNotMapped = MnrEncodingMarshal.MustRegisterIndex("Model Not Mapped", "trying to marshal unmapped model").Class()
	EncodingMarshalNilValue = MnrEncodingMarshal.MustRegisterIndex("Nil value", "marshaling nil value when it should exists").Class()
	EncodingMarshalNonAddressable = MnrEncodingMarshal.MustRegisterIndex("Non Addressable", "marshaling unaddressable value").Class()
	EncodingMarshalInput = MnrEncodingMarshal.MustRegisterIndex("Input", "marshaling invalid input value / type").Class()
}

/**

Encoding Unmarshal

*/

var (
	// MnrEncodingUnmarshal is the 'MjrEncoding' minor classification
	// for the unmarshall process.
	MnrEncodingUnmarshal Minor

	// EncodingUnmarshal is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// for general purpose unmarshal errors.
	EncodingUnmarshal Class

	// EncodingUnmarshalCollection is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when unmarshaling the input data that has undefined model collection.
	EncodingUnmarshalCollection Class

	// EncodingUnmarshalFieldValue is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// When unmarshaling field with invalid value.
	EncodingUnmarshalFieldValue Class

	// EncodingUnmarshalInvalidFormat is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the input data has invalid format. I.e. invalid json formatting.
	EncodingUnmarshalInvalidFormat Class

	// EncodingUnmarshalInvalidID is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the document with invalid primary field format or value.
	EncodingUnmarshalInvalidID Class

	// EncodingUnmarshalInvalidOutput is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the document with invalid output model or slice.
	EncodingUnmarshalInvalidOutput Class

	// EncodingUnmarshalInvalidTime is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when unmarshaling time defined field with invalid type or value.
	EncodingUnmarshalInvalidTime Class

	// EncodingUnmarshalInvalidType is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling field with invalid value type.
	EncodingUnmarshalInvalidType Class

	// EncodingUnmarshalNoData is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// while unmarshaling the empty data, when it is not allowed.
	EncodingUnmarshalNoData Class

	// EncodingUnmarshalUnknownField is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when trying to unmarshal undefined or unknown field. Used when 'strict' mode is set.
	EncodingUnmarshalUnknownField Class

	// EncodingUnmarshalValueOutOfRange is a 'MjrEncoding', 'MnrEncodingUnmarshal' error classification
	// when trying to unmarshal value with some range that exceeds defined maximum.
	EncodingUnmarshalValueOutOfRange Class
)

func registerEncodingUnmarshal() {
	MnrEncodingUnmarshal = MjrEncoding.MustRegisterMinor("Unmarshal", "unmarshaling from some encoding failed")

	EncodingUnmarshal = MustNewMinorClass(MnrEncodingUnmarshal)
	EncodingUnmarshalCollection = MnrEncodingUnmarshal.MustRegisterIndex("No Collection", "unmarshaling the input with no or invalid model collection").Class()
	EncodingUnmarshalFieldValue = MnrEncodingUnmarshal.MustRegisterIndex("Field Value", "unmarshaling the field with invalid value").Class()
	EncodingUnmarshalInvalidFormat = MnrEncodingUnmarshal.MustRegisterIndex("Invalid Format", "unmarshaling the input with invalid format").Class()
	EncodingUnmarshalInvalidID = MnrEncodingUnmarshal.MustRegisterIndex("Invalid ID", "unmarshaling primary field with invalid type, value or formatting").Class()
	EncodingUnmarshalInvalidOutput = MnrEncodingUnmarshal.MustRegisterIndex("Invalid Output", "invalid output value, type or model type").Class()
	EncodingUnmarshalInvalidTime = MnrEncodingUnmarshal.MustRegisterIndex("Invalid Time", "unmarshaling time defined field with invalid value or formatting").Class()
	EncodingUnmarshalInvalidType = MnrEncodingUnmarshal.MustRegisterIndex("Invalid Type", "unmarshaling the field or other data with invalid type").Class()
	EncodingUnmarshalNoData = MnrEncodingUnmarshal.MustRegisterIndex("No Data", "unmarshaling the input without or with nil data").Class()
	EncodingUnmarshalUnknownField = MnrEncodingUnmarshal.MustRegisterIndex("Unknown Field", "unmarshaling undefined field - in strict mode").Class()
	EncodingUnmarshalValueOutOfRange = MnrEncodingUnmarshal.MustRegisterIndex("Out Of Range", "unmarshaling some data with out of possible range").Class()
}
