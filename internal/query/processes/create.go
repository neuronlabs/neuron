package processes

import (
	"github.com/kucjac/jsonapi/internal/query/scope"
	irepos "github.com/kucjac/jsonapi/internal/repositories"
	"github.com/kucjac/jsonapi/log"
	qScope "github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/jsonapi/repositories"

	// "github.com/kucjac/uni-db"
	"github.com/pkg/errors"
)

var ErrNoCreateRepository error = errors.New("No create repository for model found.")

func Create(c irepos.RepositoryGetter, s *scope.Scope) error {
	repo, ok := c.RepositoryByModel(s.Struct())
	if !ok {
		log.Errorf("No repository found for the %s model.", s.Struct().Collection())
		return ErrNoRepositoryFound
	}

	creater, ok := repo.(repositories.Creater)
	if !ok {
		log.Errorf("The repository deosn't implement Creater interface for model: %s", s.Struct().Collection())
		return ErrNoCreateRepository
	}

	if err := creater.Create((*qScope.Scope)(s)); err != nil {
		return err
	}

	return nil
}

// BeforeCreate is the function that is used before the create process
func BeforeCreate(c irepos.RepositoryGetter, s *scope.Scope) error {
	beforeCreator, ok := s.Value.(repositories.BeforeCreatorR)
	if !ok {
		return nil
	}

	// Use the hook function before create
	err := beforeCreator.BeforeCreateR((*qScope.Scope)(s))
	if err != nil {
		return err
	}
	return nil
}

// AfterCreate is the function that is used after the create process
// It uses AfterCreateR hook if the model implements it.
func AfterCreate(c irepos.RepositoryGetter, s *scope.Scope) error {
	afterCreator, ok := s.Value.(repositories.AfterCreatorR)
	if !ok {
		return nil
	}

	err := afterCreator.AfterCreateR((*qScope.Scope)(s))
	if err != nil {
		return err
	}
	return nil
}
