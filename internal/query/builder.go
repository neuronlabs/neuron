package query

import (
	"github.com/kucjac/jsonapi/config"
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/i18n"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/namer"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Builder is a struct that is responsible for creating the query scopes
type Builder struct {
	// I18n defines current builder i18n support
	I18n *i18n.Support

	// Config is the configuration for the builder
	Config *config.BuilderConfig

	// schemas are the given model schemas
	schemas *models.ModelSchemas
}

var DefaultConfig *config.BuilderConfig = config.ReadDefaultControllerConfig().Builder

// DefaultBuilder returns builder with default config and no i18n support
func DefaultBuilder() *Builder {
	schemas, _ := models.NewModelSchemas(
		namer.NamingSnake,
		DefaultConfig.IncludeNestedLimit,
		map[string]*config.Schema{},
		"default",
		"default",
		flags.New(),
	)
	b, err := NewBuilder(
		schemas,
		DefaultConfig,
		nil,
	)
	if err != nil {
		panic(err)
	}
	return b
}

// NewBuilder creates new query builder
func NewBuilder(
	schemas *models.ModelSchemas,
	cfg *config.BuilderConfig,
	i18nSupport *i18n.Support,
) (*Builder, error) {
	b := &Builder{schemas: schemas, I18n: i18nSupport, Config: cfg}

	if err := b.validateConfig(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Builder) validateConfig() error {
	err := validate.Struct(b.Config)
	return errors.Wrap(err, "validateConfig failed.")
}