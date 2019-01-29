package config

// Config contains general configurations for the
type Config struct {
	// Controller defines the configuration for the Controllers
	Controller *ControllerConfig `mapstructure:"controller"`

	// Gateway is the configuration for the gateway
	Gateway *GatewayConfig `mapstructure:"gateway"`
}
