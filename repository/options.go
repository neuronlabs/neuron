package repository

import (
	"crypto/tls"
	"time"
)

// Options is the common structure used as the options for the repositories.
type Options struct {
	// URI is the uri with the full connection credentials for the repository.
	URI string
	// Host defines the access hostname or the ip address
	Host string
	// Port is the connection port
	Port uint16
	// Database
	Database string
	// Protocol is the protocol used in the connection
	Protocol string
	// Username is the username used to get connection credential
	Username string
	// Password is the password used to get connection credentials
	Password string
	// MaxTimeout defines the maximum timeout for the given repository connection
	MaxTimeout *time.Duration
	// TLS defines the tls configuration for given repository.
	TLSConfig *tls.Config
}

// Option is a function that changes the repository options.
type Option func(o *Options)

// WithHost sets the options hostname.
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort is an option that sets the Port in the repository options.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = uint16(port)
	}
}

// WithDatabase is an option that sets the Database in the repository options.
func WithDatabase(db string) Option {
	return func(o *Options) {
		o.Database = db
	}
}

// WithProtocol is an option that sets the Protocol in the repository options.
func WithProtocol(proto string) Option {
	return func(o *Options) {
		o.Protocol = proto
	}
}

// WithUsername is an option that sets the Username in the repository options.
func WithUsername(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// WithURI is an option that sets repository connection URI.
func WithURI(uri string) Option {
	return func(o *Options) {
		o.URI = uri
	}
}

// WithPassword is an option that sets the Password in the repository options.
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithMaxTimeout is an option that sets the MaxTimeout in the repository options.
func WithMaxTimeout(maxTimeout time.Duration) Option {
	return func(o *Options) {
		o.MaxTimeout = &maxTimeout
	}
}

// WithTLSConfig is an option that sets the TLSConfig in the repository options.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = tlsConfig
	}
}
