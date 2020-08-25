package filestore

import (
	"context"
)

// Store is an interface that allows to execute operation on files.
type Store interface {
	// Type returns the name of the file store.
	Type() string
	// NewFile creates new file with given name and with provided options.
	NewFile(ctx context.Context, name string, options ...FileOption) (File, error)
	// PutFile stores the file created previously.
	PutFile(ctx context.Context, file File, options ...PutOption) error
	// GetFile gets the file header with provided name (only file name with extension) and possible options.
	GetFile(ctx context.Context, name string, options ...GetOption) (File, error)
	// ListFiles lists the file headers for provided directory with optional filters.
	ListFiles(ctx context.Context, dir string, options ...ListOption) ([]File, error)
	// DeleteFile deletes provided file.
	DeleteFile(ctx context.Context, file File) error
}

// FileOptions are the options file creating new file.
type FileOptions struct {
	Directory string
	Bucket    string
}

// FileOption is function that changes file options.
type FileOption func(o *FileOptions)

// FileWithDirectory create the file with given 'directory'.
func FileWithDirectory(directory string) FileOption {
	return func(o *FileOptions) {
		o.Directory = directory
	}
}

// FileWithBucket create file with provided bucket.
func FileWithBucket(bucket string) FileOption {
	return func(o *FileOptions) {
		o.Bucket = bucket
	}
}

// PutOptions are the options used on putting the file.
type PutOptions struct {
	// Overwrite defines if the file with given version should be overwritten by the input.
	Overwrite bool
}

// PutOption is the function that changes the put options.
type PutOption func(o *PutOptions)

// PutWithOverwrite is the put option that forces the put to overwrite given file.
func PutWithOverwrite() PutOption {
	return func(o *PutOptions) {
		o.Overwrite = true
	}
}

// GetOptions are the options while getting the file.
type GetOptions struct {
	Directory string
	Bucket    string
	Version   string
}

// GetOption is an option function that set get options.
type GetOption func(o *GetOptions)

// GetWithDirectory sets the get file options with 'directory'.
func GetWithDirectory(directory string) GetOption {
	return func(o *GetOptions) {
		o.Directory = directory
	}
}

// GetWithBucket get file with provided bucket.
func GetWithBucket(bucket string) GetOption {
	return func(o *GetOptions) {
		o.Bucket = bucket
	}
}

// GetWithVersion gets the file with given version. Used only for get method.
func GetWithVersion(version string) GetOption {
	return func(o *GetOptions) {
		o.Version = version
	}
}

// ListOptions are the settings while listing the files.
type ListOptions struct {
	// Limit, Offset sets the limit and offset while listing the files.
	Limit, Offset int
	// Extension is a filter type of the file extension.
	Extension string
	// Bucket could be used for S3 implementations.
	Bucket string
}

// ListOption is a function that sets up list options.
type ListOption func(o *ListOptions)

// ListWithLimit sets the list limit filter.
func ListWithLimit(limit int) ListOption {
	return func(o *ListOptions) {
		o.Limit = limit
	}
}

// ListWithOffset sets the list offset filter.
func ListWithOffset(offset int) ListOption {
	return func(o *ListOptions) {
		o.Offset = offset
	}
}

// ListWithExtension sets the list extension filter.
func ListWithExtension(ext string) ListOption {
	return func(o *ListOptions) {
		o.Extension = ext
	}
}

// ListWithBucket sets the list bucket filter.
func ListWithBucket(bucket string) ListOption {
	return func(o *ListOptions) {
		o.Bucket = bucket
	}
}
