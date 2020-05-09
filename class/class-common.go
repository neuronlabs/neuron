package class

import (
	"github.com/neuronlabs/errors"
)

func init() {
	registerClasses()
}

func registerClasses() {
	MjrNeuron = errors.MustNewMajor()
	registerGeneratorClasses()
	registerCommonClasses()
	registerConfigClasses()
	registerEncodingClasses()
	registerInternalClasses()
	registerModelClasses()
	registerQueryClasses()
	registerRepositoryClasses()
}

// MjrNeuron is the major class for all neuron errors.
var MjrNeuron errors.Major

// MjrCommon is the common major errors classification.
var MjrCommon errors.Major

var (
	// MnrCommonParse is the 'MjrCommon' minor error classification
	// for parsing common strings.
	MnrCommonParse errors.Minor

	// CommonParseBrackets is the 'MjrCommon', 'MnrCommonParse' error classification
	// for parsing bracketed strings.
	CommonParseBrackets errors.Class

	// MnrCommonLogger is the 'MjrCommon' minor error classification
	// for logger issues.
	MnrCommonLogger errors.Minor

	// CommonLoggerNotImplement is the 'MjrCommon', 'MnrCommonLogger' error classification
	// for logger's that doesn't implement some interface.
	CommonLoggerNotImplement errors.Class

	// CommonLoggerUnknownLevel is the 'MjrCommon', 'MnrCommonLogger' error classification
	// for unknown level logger.
	CommonLoggerUnknownLevel errors.Class
)

func registerCommonClasses() {
	MjrCommon = errors.MustNewMajor()

	MnrCommonParse = errors.MustNewMinor(MjrCommon)
	CommonParseBrackets = errors.MustNewClass(MjrCommon, MnrCommonParse, errors.MustNewIndex(MjrCommon, MnrCommonParse))

	MnrCommonLogger = errors.MustNewMinor(MjrCommon)
	CommonLoggerNotImplement = errors.MustNewClass(MjrCommon, MnrCommonLogger, errors.MustNewIndex(MjrCommon, MnrCommonLogger))
	CommonLoggerUnknownLevel = errors.MustNewClass(MjrCommon, MnrCommonLogger, errors.MustNewIndex(MjrCommon, MnrCommonLogger))
}
