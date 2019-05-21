package config

// Schema defines configuration for the single model schema.
// If the schema is not local the
type Schema struct {
	Name       string                  `mapstructure:"name"`
	Models     map[string]*ModelConfig `mapstructure:"models"`
	Local      bool                    `mapstructure:"local"`
	Connection *Connection             `mapstructure:"connection"`
}
