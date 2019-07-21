package class

// MjrLanguage is the major language error classification
var MjrLanguage Major

var (
	// MnrLanguageUnsupported is the 'MjrLanguage' minor error classification
	// for unsupported languages.
	MnrLanguageUnsupported Minor

	// LanguageUnsupportedTag is the 'MjrLanguage', 'MnrLanguageUnsupported' error
	// classification for unsupported language tags.
	LanguageUnsupportedTag Class

	// MnrLanguageParsing is the 'MjrLanguage' minor error classification for
	// parsing the languages issues.
	MnrLanguageParsing Minor

	// LanguageParsingFailed is the error classification when parsing language failed.
	LanguageParsingFailed Class
)

func registerLanguageClasses() {
	MjrLanguage = MustRegisterMajor("Language", "language - i18n issues")

	MnrLanguageUnsupported = MjrLanguage.MustRegisterMinor("Unsupported", "unsupported languages issues")
	LanguageUnsupportedTag = MnrLanguageUnsupported.MustRegisterIndex("Tag", "unsupported language tag").Class()

	MnrLanguageParsing = MjrLanguage.MustRegisterMinor("Parsing", "parsing languages issues")
	LanguageParsingFailed = MnrLanguageParsing.MustRegisterIndex("Failed", "parsing language failed").Class()
}
