package mapping

// OptionsSetter is the interface used to set the options from the field's StructField.
// Used in models to prepare custom structures for the defined options.
type OptionsSetter interface {
	SetOptions(field *StructField)
}
