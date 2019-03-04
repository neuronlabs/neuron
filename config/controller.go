package config

// ControllerConfig defines the configuration for the Controller
type ControllerConfig struct {

	// NamingConvention is the naming convention used while preparing the models.
	// Allowed values:
	// - camel
	// - lowercamel
	// - snake
	// - kebab
	NamingConvention string `mapstructure:"naming_convention" validate:"isdefault|oneof=camel lowercamel snake kebab"`

	// DefaultSchema is the default schema name for the models within given controller
	DefaultSchema string `validate:"alphanum" mapstructure:"default_schema"`

	// ModelSchemas defines the model schemas used by api
	ModelSchemas map[string]*Schema `mapstructure:"schemas"`

	// StrictUnmarshalMode is the flag that defines if the unmarshaling should be in a
	// strict mode that checks if incoming values are all known to the controller
	// As well as the query builder doesn't allow unknown queries
	StrictUnmarshalMode bool `mapstructure:"strict_unmarshal"`

	// Debug sets the debug mode for the controller.
	Debug bool `mapstructure:"debug"`

	// Builder defines the builder config
	Builder *BuilderConfig `mapstructure:"builder"`

	// I18n defines i18n config
	I18n *I18nConfig `mapstructure:"i18n"`

	// Flags defines the controller default flags
	Flags *Flags `mapstructure:"flags"`

	// DefaultRepository
	DefaultRepository string `mapstructure:"default_repository"`

	// CreateValidatorAlias is the alias for the create validators
	CreateValidatorAlias string `mapstructure:"create_validator_alias"`

	// PatchValidatorAlias is the alis used for the Patch validator
	PatchValidatorAlias string `mapstructure:"patch_validator_alias"`

	// DefaultValidatorAlias is the alias used as a default validator alias
	DefaultValidatorAlias string `mapstructure:"default_validator_alias"`
}
