package filestore

import (
	"context"
	"time"
)

// File is a basic interface for the file abstraction.
type File interface {
	// Name returns file name with extension.
	Name() string
	// Bucket returns file's bucket.
	Bucket() string
	// Directory returns file directory.
	Directory() string
	// Version returns file version.
	Version() string

	// Parameters obtainable after open.
	// ModifiedAt is the last modification time.
	ModifiedAt() time.Time
	// Size returns byte size of given file.
	Size() int64

	// I/O
	//
	// Read functions.
	//
	// Open opens the file in an read-only way.
	Open(ctx context.Context) error
	// Close after successful read the file needs to be closed.
	Close(ctx context.Context) error
	// Read the file content when it is opened. Implements io.Reader
	Read(data []byte) (int, error)

	// Write the file content Implements io.Writer. The write is independent of the Open -> Read -> Close cycle.
	// In order to store the changes the file needs to be put into the store.
	Write(data []byte) (int, error)
}

// Metadater is an interface used
type Metadater interface {
	Metadata() map[string]interface{}
	GetMeta(key string) (interface{}, bool)
	SetMeta(key string, value interface{})
}
