package class

// MjrConfig - major that classifies errors related with the config.
var MjrConfig Major

func registerConfigClasses() {
	MjrConfig = MustRegisterMajor("Config", "config related issues")

	registerConfigRead()
	registerConfigValue()
}

var (
	// MnrConfigRead is the 'MjrConfig' minor error classification
	// for the config read issues.
	MnrConfigRead Minor

	// ConfigReadNotFound is the 'MjrConfig', 'MnrConfigRead' error classification
	// for the read config not found issue.
	ConfigReadNotFound Class
)

func registerConfigRead() {
	MnrConfigRead = MjrConfig.MustRegisterMinor("Read", "config read issues")

	ConfigReadNotFound = MnrConfigRead.MustRegisterIndex("Not Found", "config not found while reading").Class()
}

var (
	// MnrConfigValue is the 'MjrConfig' minor error classification
	// for the config value issues.
	MnrConfigValue Minor

	// ConfigValueGateway is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of gateway.
	ConfigValueGateway Class

	// ConfigValueProcessor is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of processor.
	ConfigValueProcessor Class

	// ConfigValueFlag is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues with the value of flag.
	ConfigValueFlag Class

	// ConfigValueNil is the 'MjrConfig', 'MnrConfigValue' error classification
	// for the issues related to the nil config value.
	ConfigValueNil Class

	// ConfigValueInvalid is the 'MjrConfig', 'MnrConfigValue' error classification
	// for config validation failures.
	ConfigValueInvalid Class
)

func registerConfigValue() {
	MnrConfigValue = MjrConfig.MustRegisterMinor("Value", "config value issues")

	ConfigValueGateway = MnrConfigValue.MustRegisterIndex("Gateway", "config gateway value issues").Class()
	ConfigValueProcessor = MnrConfigValue.MustRegisterIndex("Processor", "config processor value issues").Class()
	ConfigValueFlag = MnrConfigValue.MustRegisterIndex("Flag", "config flag value issues").Class()
	ConfigValueNil = MnrConfigValue.MustRegisterIndex("Nil", "provided nil config value").Class()
	ConfigValueInvalid = MnrConfigValue.MustRegisterIndex("Invalid", "validating config failed").Class()
}
