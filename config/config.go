package config

// Config contains general configurations for the Neuron service
type Config struct {
	// Controller defines the configuration for the Controllers
	Controller *Controller `mapstructure:"controller"`

	// Gateway is the configuration for the gateway
	Gateway *Gateway `mapstructure:"gateway"`
}
