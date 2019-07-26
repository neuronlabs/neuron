package class

import (
	"github.com/neuronlabs/errors"
)

// MjrLanguage is the major language error classification
var MjrLanguage errors.Major

var (
	// MnrLanguageUnsupported is the 'MjrLanguage' minor error classification
	// for unsupported languages.
	MnrLanguageUnsupported errors.Minor

	// LanguageUnsupportedTag is the 'MjrLanguage', 'MnrLanguageUnsupported' error
	// classification for unsupported language tags.
	LanguageUnsupportedTag errors.Class

	// MnrLanguageParsing is the 'MjrLanguage' minor error classification for
	// parsing the languages issues.
	MnrLanguageParsing errors.Minor

	// LanguageParsingFailed is the error classification when parsing language failed.
	LanguageParsingFailed errors.Class
)

func registerLanguageClasses() {
	MjrLanguage = errors.NewMajor()

	MnrLanguageUnsupported = errors.MustNewMinor(MjrLanguage)
	LanguageUnsupportedTag = errors.MustNewClass(MjrLanguage, MnrLanguageUnsupported, errors.MustNewIndex(MjrLanguage, MnrLanguageUnsupported))

	MnrLanguageParsing = errors.MustNewMinor(MjrLanguage)
	LanguageParsingFailed = errors.MustNewClass(MjrLanguage, MnrLanguageParsing, errors.MustNewIndex(MjrLanguage, MnrLanguageParsing))
}
