package config

// BuilderConfig is the config used for building the queries on incoming requests
type BuilderConfig struct {

	// StrictQueriesMode if true sets the strict mode for the query builder, that doesn't allow
	// unknown query keys
	StrictQueriesMode bool `mapstructure:"strict_queries"`

	// ErrorLimits defines the upper limit of the error number while getting the query
	ErrorLimits int `validate:"min=1,max=20" mapstructure:"error_limit"`

	// IncludeNestedLimit is a maximum value for nested includes (i.e. IncludeNestedLimit = 1
	// allows ?include=posts.comments but does not allow ?include=posts.comments.author)
	IncludeNestedLimit int `validate:"min=1,max=20" mapstructure:"include_nested_limit"`

	// FilterValueLimit is a maximum length of the filter values
	FilterValueLimit int `validate:"min=1,max=50" mapstructure:"filter_value_limit"`
}
