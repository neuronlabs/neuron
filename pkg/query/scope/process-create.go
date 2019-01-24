package scope

import (
	ctrl "github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"

	"github.com/kucjac/jsonapi/pkg/log"

	// "github.com/kucjac/uni-db"
	"github.com/pkg/errors"
)

type beforeCreator interface {
	BeforeCreate(s *Scope) error
}

var ErrNoCreateRepository error = errors.New("No create repository for model found.")

// CREATE process

var (
	create Process = Process{
		Name: "whiz:create",
		Func: createFunc,
	}

	beforeCreate Process = Process{
		Name: "whiz:hook_before_create",
		Func: beforeCreateFunc,
	}

	afterCreate Process = Process{
		Name: "whiz:hook_after_create",
		Func: afterCreateFunc,
	}
)

func createFunc(s *Scope) error {
	var c *ctrl.Controller = s.Controller()

	repo, ok := (*controller.Controller)(c).RepositoryByModel((*scope.Scope)(s).Struct())
	if !ok {
		log.Errorf("No repository found for the %s model.", s.Struct().Collection())
		return ErrNoRepositoryFound
	}

	creater, ok := repo.(creater)
	if !ok {
		log.Errorf("The repository deosn't implement Creater interface for model: %s", (*scope.Scope)(s).Struct().Collection())
		return ErrNoCreateRepository
	}

	if err := creater.Create(s); err != nil {
		return err
	}

	return nil
}

// beforeCreate is the function that is used before the create process
func beforeCreateFunc(s *Scope) error {
	beforeCreator, ok := s.Value.(wBeforeCreator)
	if !ok {
		return nil
	}

	// Use the hook function before create
	err := beforeCreator.WBeforeCreate(s)
	if err != nil {
		return err
	}
	return nil
}

// afterCreate is the function that is used after the create process
// It uses AfterCreateR hook if the model implements it.
func afterCreateFunc(s *Scope) error {
	afterCreator, ok := s.Value.(wAfterCreator)
	if !ok {
		return nil
	}

	err := afterCreator.WAfterCreate(s)
	if err != nil {
		return err
	}
	return nil
}
