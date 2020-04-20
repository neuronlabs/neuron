package config

import (
	"crypto/tls"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

// Repository defines the repository configuration variables.
type Repository struct {
	// DriverName defines the name for the repository driver
	DriverName string `mapstructure:"driver_name" validate:"required"`
	// Host defines the access hostname or the ip address
	Host string `mapstructure:"host" validate:"isdefault|hostname|ip"`
	// Port is the connection port
	Port int `mapstructure:"port"`
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
	// DBName gets the database name
	DBName string `mapstructure:"dbname"`
	// TLS defines the tls configuration for given repository.
	TLS *tls.Config
}

// Parse parses the repository configuration from the whitespace separated key value string.
// I.e. 'host=172.16.1.1 port=5432 username=some password=pass drivername=postgres'
func (r *Repository) Parse(s string) error {
	spaceSplit := strings.Split(s, " ")
	for _, pair := range spaceSplit {
		eqSign := strings.IndexRune(pair, '=')
		if eqSign == -1 {
			return errors.NewDetf(class.RepositoryConfigInvalid, "invalid repository config, key value pair: '%s' - equal sign not found", pair)
		}

		key := pair[:eqSign]
		value := pair[eqSign+1:]
		switch key {
		case "driver_name", "driver":
			r.DriverName = value
		case "host", "hostname":
			r.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.NewDetf(class.RepositoryConfigInvalid, "repository port configuration is not an integer: '%s'", value)
			}
			r.Port = port
		case "protocol":
			r.Protocol = value
		case "raw_url":
			r.RawURL = value
		case "username", "user":
			r.Username = value
		case "password":
			r.Password = value
		case "max_timeout":
			d, err := time.ParseDuration(value)
			if err != nil {
				return errors.NewDetf(class.RepositoryConfigInvalid, "repository config max_timeout parse duration failed: '%v'", err)
			}
			r.MaxTimeout = &d
		case "dbname":
			r.DBName = value
		default:
			if r.Options == nil {
				r.Options = map[string]interface{}{}
			}
			r.Options[key] = value
		}
	}
	return nil
}

// Validate validates the repository config.
func (r *Repository) Validate() error {
	if r.DriverName == "" {
		return errors.New(class.ConfigValueNil, "no repository driver name provided in the config")
	}
	if r.RawURL != "" {
		if _, err := url.Parse(r.RawURL); err != nil {
			return errors.Newf(class.ConfigValueInvalid, "invalid raw url for the repository config: %v", err)
		}
	} else {
		if r.Port < 0 {
			return errors.Newf(class.ConfigValueNil, "repository port value cannot be < 0")
		}
	}
	return nil
}
