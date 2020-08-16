package mapping

// MapOptions are the options for the model map.
type MapOptions struct {
	DefaultNotNull   bool
	ModelNotNull     map[Model]struct{}
	NamingConvention NamingConvention
}

// MapOption is a function that sets the map options.
type MapOption func(o *MapOptions)

// WithNamingConvention sets the 'convention' as the naming convention for the model map.
func WithNamingConvention(convention NamingConvention) MapOption {
	return func(o *MapOptions) {
		o.NamingConvention = convention
	}
}

// WithDefaultNotNullModel sets the not null as the default option for all the non pointer fields in 'model'.
func WithDefaultNotNullModel(model Model) MapOption {
	return func(o *MapOptions) {
		o.ModelNotNull[model] = struct{}{}
	}
}

// WithDefaultNotNull sets the default not null option for all non-pointer fields in all models.
func WithDefaultNotNull(o *MapOptions) {
	o.DefaultNotNull = true
}
