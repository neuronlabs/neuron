package database

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	ErrDatabase = errors.New("database")
	// ErrRepository is an error related with the repository.
	ErrRepository = errors.Wrap(ErrDatabase, "repository")
	// ErrRepositoryNotFound is an error when repository is not found.
	ErrRepositoryNotFound = errors.Wrap(ErrRepository, "not found")
	// ErrRepositoryAlreadyRegistered class of errors when repository is already registered.
	ErrRepositoryAlreadyRegistered = errors.Wrap(ErrRepository, "already registered")
)
