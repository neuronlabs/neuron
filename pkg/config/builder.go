package config

type BuilderConfig struct {

	// StrictQueriesMode if true sets the strict mode for the query builder, that doesn't allow
	// unknown query keys
	StrictQueriesMode bool

	// ErrorLimits defines the upper limit of the error number while getting the query
	ErrorLimits int `validate:"min=1,max=20"`

	// IncludeNestedLimit is a maximum value for nested includes (i.e. IncludeNestedLimit = 1
	// allows ?include=posts.comments but does not allow ?include=posts.comments.author)
	IncludeNestedLimit int `validate:"min=1,max=20"`

	// FilterValueLimit is a maximum length of the filter values
	FilterValueLimit int `validate:"min=1,max=50"`
}
