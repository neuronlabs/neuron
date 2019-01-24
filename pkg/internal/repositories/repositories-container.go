package repositories

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/pkg/errors"
)

var (
	ErrRepoAlreadyRegistered error = errors.New("Repository already registered.")
	ErrNoRepoForModel        error = errors.New("Model have no repository name")
	ErrRepositoryNotFound    error = errors.New("Repository not found")
	ErrNewNotRepository      error = errors.New("New method doesn't return Repository")
)

// RepositoryGetter is the interface that allows to get the repository by the model or by the nameg
type RepositoryGetter interface {
	RepositoryByModel(model *models.ModelStruct) (interface{}, bool)
	RepositoryByName(name string) (interface{}, bool)
}

// RepositoryContainer is the container for the model repositories
// It contains mapping between repository name as well as the repository mapped to
// the given ModelStruct
type RepositoryContainer struct {
	reposiotries []Repository

	// models are the mapping for the model's defined
	models map[*models.ModelStruct]Repository
}

// NewRepoContainer creates the repository container
func NewRepoContainer() *RepositoryContainer {
	r := &RepositoryContainer{
		models: map[*models.ModelStruct]Repository{},
	}

	return r
}

// MapModel maps the model with the provided repositories
func (r *RepositoryContainer) MapModel(model *models.ModelStruct) error {
	repoName := model.RepositoryName()

	if repoName == "" {
		return ErrNoRepoForModel
	}

	var repo Repository
	for _, repo = range r.reposiotries {
		if repoName == repo.RepositoryName() {
			break
		}
	}
	if repo == nil {
		log.Errorf("Repository: '%s' not registered.", repo.RepositoryName())
		return ErrRepositoryNotFound
	}

	repoCopy := repo.New()
	r.models[model] = repoCopy.(Repository)

	log.Debugf("Model %s mapped, to repository: %s", model.Collection(), repoName)
	return nil
}

// RegisterRepository registers the repostiory by it's name
func (r *RepositoryContainer) RegisterRepository(repo Repository) error {
	// if names are zero
	// get the default repository name
	repoName := repo.RepositoryName()

	// check if New function creates a repository
	_, isNewRepo := repo.New().(Repository)
	if !isNewRepo {
		return ErrNewNotRepository
	}

	for _, repo := range r.reposiotries {
		if repo.RepositoryName() == repoName {
			return ErrRepoAlreadyRegistered
		}
	}

	r.reposiotries = append(r.reposiotries, repo)

	log.Debugf("Repository: '%s' registered succesfully.", repoName)
	return nil
}

// RepositoryByName returns repository by it's name
func (r *RepositoryContainer) RepositoryByName(name string) (Repository, bool) {
	for _, repo := range r.reposiotries {
		if repo.RepositoryName() == name {
			return repo, true
		}
	}
	return nil, false
}

// RepositoryByModel returns repository by the provided ModelsStruct
func (r *RepositoryContainer) RepositoryByModel(model *models.ModelStruct) (Repository, bool) {
	repo, ok := r.models[model]
	if !ok {
		return nil, false
	}
	return repo, ok
}
