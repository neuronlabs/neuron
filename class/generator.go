package class

import (
	"github.com/neuronlabs/errors"
)

func registerGeneratorClasses() {
	MnrGenerator = errors.MustNewMinor(MjrNeuron)
	GeneratorInvalidInput = newIndexClass(MnrGenerator)
	GeneratorFieldNoZeroer = newIndexClass(MnrGenerator)
}

// Define generator classes.
var (
	MnrGenerator           errors.Minor
	GeneratorInvalidInput  errors.Class
	GeneratorFieldNoZeroer errors.Class
)
