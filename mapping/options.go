package mapping

// MapOptions are the options for the model map.
type MapOptions struct {
	DefaultNotNull            bool
	ModelNotNull              map[Model]struct{}
	NamingConvention          NamingConvention
	DBNamingConvention        NamingConvention
	PluralCollections         bool
	DatabasePluralCollections bool
	DefaultDatabaseSchema     string
}

// MapOption is a function that sets the map options.
type MapOption func(o *MapOptions)

// WithNamingConvention sets the 'convention' as the naming convention for the model map.
func WithNamingConvention(convention NamingConvention) MapOption {
	return func(o *MapOptions) {
		o.NamingConvention = convention
	}
}

// WithDatabaseNamingConvention sets the 'convention' as the naming convention for the model map.
func WithDatabaseNamingConvention(dbNaming NamingConvention) MapOption {
	return func(o *MapOptions) {
		o.DBNamingConvention = dbNaming
	}
}

// WithPluralCollections defines if collections are named in plural way.
func WithPluralCollections(plural bool) MapOption {
	return func(o *MapOptions) {
		o.PluralCollections = plural
	}
}

// WithPluralDatabaseCollections defines if the database collections are named in plural way.
func WithPluralDatabaseCollections(plural bool) MapOption {
	return func(o *MapOptions) {
		o.DatabasePluralCollections = plural
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

// WithDefaultDatabaseSchema sets the default database schema.
func WithDefaultDatabaseSchema(defaultSchema string) MapOption {
	return func(o *MapOptions) {
		o.DefaultDatabaseSchema = defaultSchema
	}
}
