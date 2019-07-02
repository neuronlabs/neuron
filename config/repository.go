package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Repository defines the repository configuration variables.
type Repository struct {
	// DriverName defines the name for the repository driver
	DriverName string `mapstructure:"driver_name" validate:"required"`

	// Host defines the access hostname or the ip address
	Host string `mapstructure:"host" validate:"isdefault|hostname|ip"`

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

	// DBName gets the database name
	DBName string `mapstructure:"dbname"`

	// SSLMode defines if the ssl is enabled
	SSLMode string `mapstructure:"sslmode"`
}

// Parse parses the repository configuration from the whitespace seperated key value string.
// I.e. 'host=172.16.1.1 port=5432 username=some password=pass drivername=pq'
func (r *Repository) Parse(s string) error {
	spaceSplit := strings.Split(s, " ")
	for _, pair := range spaceSplit {
		eqSign := strings.IndexRune(pair, '=')
		if eqSign == -1 {
			return errors.Newf(class.RepositoryConfigInvalid, "invalid repository config, key value pair: '%s' - equal sign not found", pair)
		}

		key := pair[:eqSign]
		value := pair[eqSign+1:]
		switch key {
		case "driver_name", "driver":
			r.DriverName = value
		case "host", "hostname":
			r.Host = value
		case "path":
			r.Path = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.Newf(class.RepositoryConfigInvalid, "repository port configuration is not an integer: '%s'", value)
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
				return errors.Newf(class.RepositoryConfigInvalid, "repository config max_timeout parse duration failed: '%v'", err)
			}
			r.MaxTimeout = &d
		case "dbname":
			r.DBName = value
		case "sslmode":
			r.SSLMode = value
		default:
			log.Debugf("Invalid repository configuration key: '%s'", key)
		}
	}
	return nil
}

// Validate validates the repository config.
func (r *Repository) Validate() error {
	return validate.Struct(r)
}
