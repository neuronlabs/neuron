package filestore

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrFileStore is the general error for this package.
	ErrFileStore = errors.New("file store")
	// ErrStore is the error with the store by itself.
	ErrStore = errors.Wrap(ErrFileStore, "store")
	// ErrVersionsNotAllowed is an error if the store doesn't support file versions.
	ErrVersionsNotAllowed = errors.Wrap(ErrStore, "versions not allowed")
	// ErrFileName is an error for invalid file names.
	ErrFileName = errors.Wrap(ErrFileStore, "file name")
	// ErrExists is an error if the file already exists.
	ErrExists = errors.Wrap(ErrFileStore, "exists")
	// ErrFileIsDir is an error if provided file name is a directory.
	ErrFileIsDir = errors.Wrap(ErrFileStore, "file is dir")
	// ErrNotExists is an error when the file doesn't exists.
	ErrNotExists = errors.Wrap(ErrFileStore, "not exists")
	// ErrNotOpened is an error if the file is not opened yet and wanted to be read.
	ErrNotOpened = errors.Wrap(ErrFileStore, "not opened")
	// ErrAlreadyOpened is an error thrown when the file is already opened.
	ErrAlreadyOpened = errors.Wrap(ErrFileStore, "already opened")
	// ErrClosed is an error if the file is already closed.
	ErrClosed = errors.Wrap(ErrFileStore, "closed")
	// ErrInternal is an internal error for this package.
	ErrInternal = errors.Wrap(errors.ErrInternal, "file store")
	// ErrPermission is an error with the file permissions.
	ErrPermission = errors.Wrap(ErrFileStore, "permission")
)
