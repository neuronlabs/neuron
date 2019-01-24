package repositories

type Repository interface {
	RepositoryName() string
	New() interface{}
}
