package config

import (
	"time"
)

// ModelConfig defines single model configurations.
type ModelConfig struct {
	// Collection is the model's collection name
	Collection string `mapstructure:"collection"`

	// RepositoryName is the model's repository name, with the name provided in the initialization
	// process...
	RepositoryName string `mapstructure:"repository_name"`

	// Map sets the model's Store values
	Map map[string]interface{} `mapstructure:"map"`

	// Repository defines the model's repository connection config
	Repository *Repository `mapstructure:"repository"`

	// AutoMigrate automatically migrates the model into the given repository structuring
	// I.e. sql creates or updates the table
	AutoMigrate bool `mapstructure:"automigrate"`
}

// Connection is the configuration for non local schemas credentials.
// The connection config can be set by providing raw_url or with host,path,protocol.
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
