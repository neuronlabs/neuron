package repository

// Repository is the interface that defines the basic neuron Repository
// it may be extended by the interfaces from the scope package
type Repository interface {
	RepositoryName() string
}
