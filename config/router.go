package config

// Router contains information about the router used in the gateway
type Router struct {
	// Name gets the router by it's registered name
	Name string `mapstructure:"name"`

	// DefaultMiddlewares are the middlewares used as default for each endpoint
	// without middlewares set from the
	DefaultMiddlewares []string `mapstructure:"default_middlewares"`

	// Prefix is the url prefix for the API
	Prefix string `mapstructure:"prefix"`

	// DefaultPagination defines default ListPagination for the gateway
	DefaultPagination *Pagination `mapstructure:"default_pagination"`

	// CompressionLevel defines the compression level for the handler function writers
	CompressionLevel int `mapstructure:"compression_level" validate:"max=9,min=-2"`
}
