package controller

import (
	"github.com/neuronlabs/neuron/errors"
)

func init() {
	MjrController = errors.MustNewMajor()
	// Repository
	MnrRepository = errors.MustNewMinor(MjrController)
	ClassRepositoryNotFound = errors.MustNewClassWIndex(MjrController, MnrRepository)
	ClassRepositoryNotMatched = errors.MustNewClassWIndex(MjrController, MnrRepository)
	ClassRepositoryAlreadyRegistered = errors.MustNewClassWIndex(MjrController, MnrRepository)

	// Config
	MnrConfig = errors.MustNewMinor(MjrController)
	ClassInvalidConfig = errors.MustNewClassWIndex(MjrController, MnrConfig)

	ClassInvalidModel = errors.MustNewMajorClass(MjrController)
}

var (
	// MjrController defines major classification for the controller.
	MjrController errors.Major
	// MnrRepository defines minor for repositories.
	MnrRepository errors.Minor
	// ClassRepositoryNotFound defines error classification.
	ClassRepositoryNotFound errors.Class
	// ClassRepositoryNotMatched defines errors classification when model doesn't have repository defined.
	ClassRepositoryNotMatched errors.Class
	// ClassRepositoryAlreadyRegistered class of errors when repository is already registered.
	ClassRepositoryAlreadyRegistered errors.Class

	// MnrConfig defines the error classification related with the config.
	MnrConfig errors.Minor
	// ClassInvalidConfig provided invalid or nil config.
	ClassInvalidConfig errors.Class
	// ClassInvalidModel is the error classification for the invalid input models.
	ClassInvalidModel errors.Class
)
