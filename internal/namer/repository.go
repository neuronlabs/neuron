package namer

// RepositoryNamer is the interface used for the repositories to implement
// that defines it's name
type RepositoryNamer interface {
	RepositoryName() string
}
