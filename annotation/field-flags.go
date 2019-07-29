package annotation

// Model field's flag tags.
const (
	// Flags is the neuron model field's tag used for defining field flags.
	Flags = "flags"
	// Hidden defines that the field should be hidden from marshaling.
	Hidden = "hidden"
	// ISO8601 sets the time field format to ISO8601.
	ISO8601 = "iso8601"
	// OmitEmpty allows to omit marshaling this field if it's zero-value.
	OmitEmpty = "omitempty"
	// I18n defines that this field is internationalization ready.
	I18n = "i18n"
	// NoFilter is the neuron model field's flag that disallows to query filter for given field.
	NoFilter = "nofilter"
	// NotSortable is the neuron model field's flag that disallows to query sort on given field.
	NotSortable = "nosort"
)
