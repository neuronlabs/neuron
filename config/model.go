package config

import (
	"time"
)

// Connection is the configuration for non local schemas credentials
// The connection config can be set by providing raw_url or with host,path,protocol/
type Connection struct {
	// Host defines the access hostname or the ip address
	Host string `mapstructure:"host" validate:"hostname|ip"`

	// Path is the connection path, just after the protocol and
	Path string `mapstructure:"path" validate:"isdefault|uri"`

	// Port is the connection port
	Port interface{} `mapstructure:"port"`

	// Protocol is the protocol used in the connection
	Protocol string `mapstructure:"protocol"`

	// RawURL is the raw connection url. If set it must define the protocol ('http://',
	// 'rpc://'...)
	RawURL string `mapstructure:"raw_url" validate:"isdefault|url"`

	// Username is the username used to get connection credential
	Username string `mapstructure:"username"`

	// Password is the password used to get connection credentials
	Password string `mapstructure:"password"`

	// Options contains connection dependent specific options
	Options map[string]interface{} `mapstructure:"options"`

	// MaxTimeout defines the maximum timeout for the given repository connection
	MaxTimeout *time.Duration `mapstructure:"max_timeout"`
}

// ModelConfig defines single model configurations
type ModelConfig struct {
	// Collection is the model's collection name
	Collection string `mapstructure:"collection"`

	// RepositoryName is the model's repository name, with the name provided in the initialization
	// process...
	RepositoryName string `mapstructure:"repository_name"`

	// Endpoints defines model's api endpoints configuration
	Endpoints ModelEndpoints `mapstructure:"endpoints"`

	// Map sets the model's Store values
	Map map[string]interface{} `mapstructure:"map"`

	// Repository defines the model's repository connection config
	Repository *Repository `mapstructure:"repository"`

	// AutoMigrate automatically migrates the model into the given repository structuring
	// I.e. sql creates or updates the table
	AutoMigrate bool `mapstructure:"automigrate"`
}

// ModelEndpoints is the api endpoint's configuration for the given model
type ModelEndpoints struct {
	Create          Endpoint `mapstructure:"create"`
	Get             Endpoint `mapstructure:"get"`
	GetRelated      Endpoint `mapstructure:"get_related"`
	GetRelationship Endpoint `mapstructure:"get_relationship"`
	List            Endpoint `mapstructure:"list"`
	Patch           Endpoint `mapstructure:"patch"`
	Delete          Endpoint `mapstructure:"delete"`
}

// Endpoint is the configuration struct for the single endpoint for model
type Endpoint struct {

	// Forbidden defines if the endpoint should not be used
	Forbidden bool `mapstructure:"forbidden"`

	// PresetFilters defines preset filters definitions for an endpoint
	PresetFilters []string `mapstructure:"preset_filters"`

	// PresetScope defines preset scopes for an endpoint
	PresetScope []string `mapstructure:"preset_scope"`

	// PrecheckFilters are the endpoints filters used to check the consistency of the query
	PrecheckFilters []string `mapstructure:"precheck_filters"`

	// PrecheckScope are the endpoint scopes used to check the consistency of the query
	PrecheckScope []string `mapstructure:"precheck_scope"`

	// PresetSorts are the sort fields used by default on given endpoint
	PresetSorts []string `mapstructure:"preset_sorts"`

	// PresetPagination are the default pagination for provided endpoint
	PresetPagination *Pagination `mapstructure:"preset_pagination"`

	// Flags contains boolean flags used by default on given endpoint
	Flags Flags `mapstructure:"flags"`

	// RelatedField contains constraints for the relationship or related (endpoint) field
	RelatedField EndpointConstraints `mapstructure:"related"`

	// CustomMiddlewares are the middlewares used for the provided endpoint
	CustomMiddlewares []string `mapstructure:"custom_middlewares"`
}

// EndpointConstraints contains constraints for provided endpoints
type EndpointConstraints struct {
	// PrecheckFilters are the endpoints filters used to check the consistency of the query
	PresetFilters []string `mapstructure:"preset_filters"`

	// PresetScope defines preset scopes for an endpoint
	PresetScope []string `mapstructure:"preset_scope"`

	// PrecheckFilters are the endpoints filters used to check the consistency of the query
	PrecheckFilters []string `mapstructure:"precheck_filters"`

	// PrecheckScope are the endpoint scopes used to check the consistency of the query
	PrecheckScope []string `mapstructure:"precheck_scope"`

	// PresetSorts are the sort fields used by default on given endpoint
	PresetSorts []string `mapstructure:"preset_sorts"`

	// PresetPagination are the default pagination for provided endpoint
	PresetPagination *Pagination `mapstructure:"preset_pagination"`
}

// Pagination defines the pagination configuration
type Pagination struct {
	Limit      int `mapstructure:"limit"`
	Offset     int `mapstructure:"offset"`
	PageSize   int `mapstructure:"page_size"`
	PageNumber int `mapstructure:"page_number"`
}

// IsZero defines if the pagination configis not set
func (p Pagination) IsZero() bool {
	if p.Limit == 0 && p.Offset == 0 && p.PageSize == 0 && p.PageNumber == 0 {
		return true
	}
	return false
}

// PresetQuery is the query that is used to set up some values initialy from the other collections
// or by applying the filters
type PresetQuery struct {
	// Field defines the field to that the values should be set with
	Field string

	// ScopeRelations gets the values by the collection relations
	// i.e. if the collection 'blog' has a relationship of kind to many 'posts'
	//
	// Trying to filter the 'posts' by the specific blogs values can be made by providing
	// the ScopeRelations value i.e.: 'blogs.posts' with the Filter: '[blogs][id][$eq]=1'
	// This result with the possible
	ScopeRelations string

	Filters []string
}

// PresetScope is the preset scope values
type PresetScope struct {
	// Scope Query defines the preset scope
	ScopeQuery string
	Filter     string
}
