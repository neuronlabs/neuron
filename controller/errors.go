package controller

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrController defines major classification for the controller.
	ErrController = errors.New("controller")
	// ErrRepository is an error related with the repository.
	ErrRepository = errors.Wrap(ErrController, "repository")
	// ErrRepositoryNotFound is an error when repository is not found.
	ErrRepositoryNotFound = errors.Wrap(ErrRepository, "not found")
	// ErrRepositoryAlreadyRegistered class of errors when repository is already registered.
	ErrRepositoryAlreadyRegistered = errors.Wrap(ErrRepository, "already registered")
)
