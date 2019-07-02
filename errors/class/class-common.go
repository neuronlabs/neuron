package class

// MjrCommon is the common major errors classification.
var MjrCommon Major

var (
	// MnrCommonParse is the 'MjrCommon' minor error classification
	// for parsing common strings.
	MnrCommonParse Minor

	// CommonParseBrackets is the 'MjrCommon', 'MnrCommonParse' error classification
	// for parsing bracketed strings.
	CommonParseBrackets Class

	// MnrCommonLogger is the 'MjrCommon' minor error classification
	// for logger issues.
	MnrCommonLogger Minor

	// CommonLoggerNotImplement is the 'MjrCommon', 'MnrCommonLogger' error classification
	// for logger's that doesn't implement some interface.
	CommonLoggerNotImplement Class

	// CommonLoggerUnknownLevel is the 'MjrCommon', 'MnrCommonLogger' error classification
	// for unknown level logger.
	CommonLoggerUnknownLevel Class
)

func registerCommonClasses() {
	MjrCommon = MustRegisterMajor("Common", "common error classification")

	MnrCommonParse = MjrCommon.MustRegisterMinor("Parse", "common parsing issues")
	CommonParseBrackets = MnrCommonParse.MustRegisterIndex("Brackets", "parsing string brackets failed").Class()

	MnrCommonLogger = MjrCommon.MustRegisterMinor("Logger", "common logger issues")
	CommonLoggerNotImplement = MnrCommonLogger.MustRegisterIndex("Not Implement", "logger issues that doesn't implement some interface").Class()
	CommonLoggerUnknownLevel = MnrCommonLogger.MustRegisterIndex("Unknown Level", "unknown level issue").Class()
}
