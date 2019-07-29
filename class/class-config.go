package class

import (
	"github.com/neuronlabs/errors"
)

// MjrConfig - major that classifies errors related with the config.
var MjrConfig errors.Major

func registerConfigClasses() {
	MjrConfig = errors.NewMajor()

	registerConfigRead()
	registerConfigValue()
}

var (
	// MnrConfigRead is the 'MjrConfig' minor error classification
	// for the config read issues.
	MnrConfigRead errors.Minor

	// ConfigReadNotFound is the 'MjrConfig', 'MnrConfigRead' error classification
	// for the read config not found issue.
	ConfigReadNotFound errors.Class
)

func registerConfigRead() {
	MnrConfigRead = errors.MustNewMinor(MjrConfig)

	ConfigReadNotFound = errors.MustNewClass(MjrConfig, MnrConfigRead, errors.MustNewIndex(MjrConfig, MnrConfigRead))
}

var (
	// MnrConfigValue is the 'MjrConfig' minor error classification
	// for the config value issues.
	MnrConfigValue errors.Minor

	// ConfigValueGateway is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of gateway.
	ConfigValueGateway errors.Class

	// ConfigValueProcessor is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of processor.
	ConfigValueProcessor errors.Class

	// ConfigValueFlag is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of flag.
	ConfigValueFlag errors.Class

	// ConfigValueNil is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues related to the nil config value.
	ConfigValueNil errors.Class

	// ConfigValueInvalid is the 'MjrConfig', 'MnrConfigValue' error classification
	// for config validation failures.
	ConfigValueInvalid errors.Class
)

func registerConfigValue() {
	MnrConfigValue = errors.MustNewMinor(MjrConfig)

	ConfigValueGateway = errors.MustNewClass(MjrConfig, MnrConfigValue, errors.MustNewIndex(MjrConfig, MnrConfigValue))
	ConfigValueProcessor = errors.MustNewClass(MjrConfig, MnrConfigValue, errors.MustNewIndex(MjrConfig, MnrConfigValue))
	ConfigValueFlag = errors.MustNewClass(MjrConfig, MnrConfigValue, errors.MustNewIndex(MjrConfig, MnrConfigValue))
	ConfigValueNil = errors.MustNewClass(MjrConfig, MnrConfigValue, errors.MustNewIndex(MjrConfig, MnrConfigValue))
	ConfigValueInvalid = errors.MustNewClass(MjrConfig, MnrConfigValue, errors.MustNewIndex(MjrConfig, MnrConfigValue))
}
