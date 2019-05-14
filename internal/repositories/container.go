package repositories

import (
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/repositories"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/pkg/errors"
)

// errors used in the repository container
var (
	ErrRepoAlreadyRegistered error = errors.New("Repository already registered.")
	ErrNoRepoForModel        error = errors.New("Model doesn't have defined repository name")
	ErrRepositoryNotFound    error = errors.New("Repository not found")
	ErrNewNotRepository      error = errors.New("New method doesn't return Repository")
)

// RepositoryGetter is the interface that allows to get the repository by the model or by the name
type RepositoryGetter interface {
	RepositoryByModel(model *models.ModelStruct) (repositories.Repository, bool)
	RepositoryByName(name string) (interface{}, bool)
}

// Container is the container for the model repositories
// It contains mapping between repository name as well as the repository mapped to
// the given ModelStruct
type Container struct {
	repositories []Repository

	// models are the mapping for the model's defined
	models map[*models.ModelStruct]Repository

	defaultRepo Repository
}

// New creates the repository container
func New() *Container {
	r := &Container{
		models: map[*models.ModelStruct]Repository{},
	}

	return r
}

// MapModel maps the model with the provided repositories
func (r *Container) MapModel(model *models.ModelStruct) error {
	repoName := model.RepositoryName()

	var repo Repository
	if repoName == "" {
		repo = r.defaultRepo
		if r.defaultRepo == nil {
			return errors.New("No default repository found")
		}
	} else {
		for _, repo = range r.repositories {
			if repoName == repo.RepositoryName() {
				break
			}
		}
	}
	if repo == nil {
		log.Debugf("Repository: '%s' is not registered.", repoName)
		return ErrRepositoryNotFound
	}

	repoCopy, err := repo.New((*mapping.ModelStruct)(model))
	if err != nil {
		return err
	}

	r.models[model] = repoCopy

	log.Debugf("Model %s mapped, to repository: %s", model.Collection(), repoName)
	return nil
}

// RegisterRepository registers the repostiory by it's name
func (r *Container) RegisterRepository(repo Repository) error {
	// if names are zero
	// get the default repository name
	repoName := repo.RepositoryName()

	for _, ir := range r.repositories {
		if ir.RepositoryName() == repoName {
			log.Debugf("Repository already registered: %s", repoName)
			return ErrRepoAlreadyRegistered
		}
	}

	// set defaultRepo
	if r.defaultRepo == nil {
		r.defaultRepo = repo
	}

	r.repositories = append(r.repositories, repo)

	log.Debugf("Repository: '%s' registered succesfully.", repoName)
	return nil
}

// RepositoryByName returns repository by it's name
func (r *Container) RepositoryByName(name string) (Repository, bool) {
	for _, repo := range r.repositories {
		if repo.RepositoryName() == name {
			return repo, true
		}
	}
	return nil, false
}

// RepositoryByModel returns repository by the provided ModelsStruct
func (r *Container) RepositoryByModel(model *models.ModelStruct) (Repository, bool) {
	repo, ok := r.models[model]
	if !ok {
		return nil, false
	}
	return repo, ok
}

// SetDefault sets the default repository
func (r *Container) SetDefault(repo Repository) {
	var found bool
	// search and check if repository was already registered
	for _, registered := range r.repositories {
		if repo == registered {
			found = true
			break
		}
	}

	// if not found add the default repository
	if !found {
		r.repositories = append(r.repositories, repo)
	}

	// set the repo as default
	r.defaultRepo = repo
}
