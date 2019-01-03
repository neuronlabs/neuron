package config

// ControllerConfig defines the configuration for the Controller
type ControllerConfig struct {

	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lowercamel
	// - snake
	// - kebab
	NamingConvention string `validate:"oneof=camel lowercamel snake kebab"`

	// DefaultSchema is the default schema name for the models within given controller
	DefaultSchema string `validate:"alphanum"`

	// StrictUnmarshalMode is the flag that defines if the unmarshaling should be in a
	// strict mode that check if incoming values are all known
	StrictUnmarshalMode bool

	// MarshalLinks defines if the controller should add links while marshaling models
	MarshalLinks bool

	// Debug sets the debug mode for the controller.
	Debug bool

	Builder *BuilderConfig
}
