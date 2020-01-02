package config

// I18nConfig defines i18n configuration support.
type I18nConfig struct {
	// SupportedLanguages represent supported languages tags
	SupportedLanguages []string `mapstructure:"supported_languages"`
}
