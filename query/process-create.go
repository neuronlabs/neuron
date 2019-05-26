package query

import (
	"context"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/log"

	// "github.com/kucjac/uni-db"
	"github.com/pkg/errors"
)

// ErrNoCreateRepository is thrown when the repository doesn't implement the creator interface
var ErrNoCreateRepository = errors.New("No create repository for model found.")

// CREATE process

var (
	// ProcessCreate is the process that does the Repository Create method
	ProcessCreate = &Process{
		Name: "neuron:create",
		Func: createFunc,
	}

	// ProcessBeforeCreate is the process that does the hook HBeforeCreate
	ProcessBeforeCreate = &Process{
		Name: "neuron:hook_before_create",
		Func: beforeCreateFunc,
	}

	// ProcessAfterCreate is the Process that does the hook HAfterCreate
	ProcessAfterCreate = &Process{
		Name: "neuron:hook_after_create",
		Func: afterCreateFunc,
	}
)

func createFunc(ctx context.Context, s *Scope) error {

	repo, err := repository.GetRepository(s.Controller(), s.Struct())
	if err != nil {
		log.Errorf("No repository found for the %s model. %v", s.Struct().Collection(), err)
		return ErrNoRepositoryFound
	}

	creater, ok := repo.(Creater)
	if !ok {
		log.Errorf("The repository deosn't implement Creater interface for model: %s", (*scope.Scope)(s).Struct().Collection())
		return ErrNoCreateRepository
	}

	if err := creater.Create(ctx, s); err != nil {
		return err
	}

	return nil
}

// beforeCreate is the function that is used before the create process
func beforeCreateFunc(ctx context.Context, s *Scope) error {
	beforeCreator, ok := s.Value.(BeforeCreator)
	if !ok {
		return nil
	}

	// Use the hook function before create
	err := beforeCreator.HBeforeCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

// afterCreate is the function that is used after the create process
// It uses AfterCreateR hook if the model implements it.
func afterCreateFunc(ctx context.Context, s *Scope) error {
	afterCreator, ok := s.Value.(AfterCreator)
	if !ok {
		return nil
	}

	err := afterCreator.HAfterCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}