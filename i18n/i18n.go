package i18n

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/config"
)

// Support defines the internationalization coverage.
type Support struct {
	Matcher language.Matcher
	Locale  language.Coverage
}

// New creates new I18n Support.
func New(cfg *config.I18nConfig) (*Support, error) {
	var tags []interface{}

	for _, langTag := range cfg.SupportedLanguages {
		tag, err := language.Parse(langTag)
		if err != nil {
			return nil, errors.NewDetf(class.LanguageParsingFailed, "parsing language: '%s' failed. %s'", langTag, err.Error())
		}
		tags = append(tags, tag)
	}

	s := &Support{
		Locale: language.NewCoverage(tags...),
	}
	s.Matcher = language.NewMatcher(s.Locale.Tags())
	return s, nil
}

// PrettyLanguages return prettified supported languages strings.
func (s *Support) PrettyLanguages() []string {
	namer := display.Tags(language.English)
	names := make([]string, len(s.Locale.Tags()))
	for i, lang := range s.Locale.Tags() {
		names[i] = fmt.Sprintf("%s - '%s'", namer.Name(lang), lang.String())
	}
	return names
}
