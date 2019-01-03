package i18n

import (
	"fmt"
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Support defines the i18n coverage
type Support struct {
	Matcher language.Matcher
	Locale  language.Coverage
}

// New creates new I18n Support
func New(cfg *config.I18nConfig) (*Support, error) {
	var tags []*language.Tag

	for _, langTag := range cfg.SupportedLanguages {
		tag, err := language.Parse(langTag)
		if err != nil {
			return nil, errors.Wrapf(err, "language.Parse langtag: '%s' failed.", langTag)
		}
		tags = append(tags, tag)
	}

	s := &Support{
		Locale: language.NewCoverage(tags...),
	}
	s.Matcher = language.NewMatcher(s.Locale.Tags()...)
	return s, nil
}

// PrettyLangauges return prettiefied supported languages strings
func (s *Support) PrettyLanguages() []string {
	namer := display.Tags(language.English)
	var names []string = make([]string, len(s.Locale.Tags()))
	for i, lang := range s.Locale.Tags() {
		names[i] = fmt.Sprintf("%s - '%s'", namer.Name(lang), lang.String())
	}

	return names
}
