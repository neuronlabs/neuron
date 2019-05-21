package config

import (
	"time"
)

// Repository defines the repository configuration variables
type Repository struct {
	// DriverName defines the name for the repository driver
	DriverName string `mapstructure:"repository_driver"`

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
