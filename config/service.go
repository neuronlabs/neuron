package config

import (
	"crypto/tls"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/neuronlabs/neuron/errors"
)

// Connection is the configuration for non local schemas credentials.
// The connection config can be set by providing raw_url or with host,path,protocol.
type Connection struct {
	// Host defines the access hostname or the ip address
	Host string `mapstructure:"host" validate:"hostname|ip"`
	// Port is the connection port
	Port int `mapstructure:"port"`
	// Protocol is the protocol used in the connection
	Protocol string `mapstructure:"protocol"`
	// RawURL is the raw connection url. If set it must define the protocol ('http://',
	// 'postgresql://'...)
	RawURL string `mapstructure:"raw_url" validate:"isdefault|url"`
	// Username is the username used to get connection credential
	Username string `mapstructure:"username"`
	// Password is the password used to get connection credentials
	Password string `mapstructure:"password"`
	// Options contains connection dependent specific options
	Options map[string]interface{} `mapstructure:"options"`
	// MaxTimeout defines the maximum timeout for the given repository connection
	MaxTimeout *time.Duration `mapstructure:"max_timeout"`
	// TLS defines the tls configuration for given repository.
	TLS *tls.Config
}

// Service defines the service configuration.
type Service struct {
	Connection Connection `json:"connection" mapstructure:"connection"`
	// DriverName defines the name for the service driver.
	DriverName string `json:"driver_name" mapstructure:"driver_name" validate:"required"`
	// DBName gets the database name for the repositories.
	DBName string `json:"driver_name" mapstructure:"dbname"`
	// Custom holds custom mappings for the service configuration
	Options map[string]interface{} `json:"options" mapstructure:"options"`
}

// Parse parses the repository configuration from the whitespace separated key value string.
// I.e. 'host=172.16.1.1 port=5432 username=some password=pass drivername=postgres'
func (s *Service) Parse(raw string) error {
	spaceSplit := strings.Split(raw, " ")
	for _, pair := range spaceSplit {
		eqSign := strings.IndexRune(pair, '=')
		if eqSign == -1 {
			return errors.NewDetf(ClassConfigInvalidValue, "invalid repository config, key value pair: '%s' - equal sign not found", pair)
		}

		key := pair[:eqSign]
		value := pair[eqSign+1:]
		switch key {
		case "driver_name", "driver":
			s.DriverName = value
		case "host", "hostname":
			s.Connection.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.NewDetf(ClassConfigInvalidValue, "repository port configuration is not an integer: '%s'", value)
			}
			s.Connection.Port = port
		case "protocol":
			s.Connection.Protocol = value
		case "raw_url":
			s.Connection.RawURL = value
		case "username", "user":
			s.Connection.Username = value
		case "password":
			s.Connection.Password = value
		case "max_timeout":
			d, err := time.ParseDuration(value)
			if err != nil {
				return errors.NewDetf(ClassConfigInvalidValue, "repository config max_timeout parse duration failed: '%v'", err)
			}
			s.Connection.MaxTimeout = &d
		case "dbname":
			s.DBName = value
		default:
			if s.Options == nil {
				s.Options = map[string]interface{}{}
			}
			s.Options[key] = value
		}
	}
	return nil
}

// Validate validates the repository config.
func (s *Service) Validate() error {
	if s.DriverName == "" {
		return errors.New(ClassConfigInvalidValue, "no repository driver name provided in the config")
	}
	if s.Connection.RawURL != "" {
		if _, err := url.Parse(s.Connection.RawURL); err != nil {
			return errors.Newf(ClassConfigInvalidValue, "invalid raw url for the repository config: %v", err)
		}
	} else {
		if s.Connection.Port < 0 {
			return errors.Newf(ClassConfigInvalidValue, "repository port value cannot be < 0")
		}
	}
	return nil
}
