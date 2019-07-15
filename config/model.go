package config

import (
	"time"
)

// ModelConfig defines single model configurations.
type ModelConfig struct {
	// Collection is the model's collection name
	Collection string `mapstructure:"collection"`

	// RepositoryName is the model's repository name, with the name provided in the initialization
	// process...
	RepositoryName string `mapstructure:"repository_name"`

	// Map sets the model's Store values
	Map map[string]interface{} `mapstructure:"map"`

	// Repository defines the model's repository connection config
	Repository *Repository `mapstructure:"repository"`

	// AutoMigrate automatically migrates the model into the given repository structuring
	// I.e. sql creates or updates the table
	AutoMigrate bool `mapstructure:"automigrate"`

	// Fields contains the model fields configuration.
	Fields map[string]*Field `mapstructure:"fields"`
}

// Field is the model's field configuration.
type Field struct {
	Flags     []string          `mapstructure:"flags" validate:"isdefault|oneof=client-id nofilter hidden nosort iso8601 i18n omitempty langtag many2many"`
	Strategy  *RelationStrategy `mapstructure:"strategy"`
	Many2Many *Many2Many        `mapstructure:"many2many"`
}

// Many2Many is the many2many relation configuration.
type Many2Many struct {
	JoinTableModel    string `mapstructure:"join_table_model"`
	ForeignKey        string `mapstructure:"foreign_key"`
	RelatedForeignKey string `mapstructure:"related_foreign_key"`
}

// RelationStrategy is the configuration of the relation modification strategy.
type RelationStrategy struct {
	OnCreate Strategy         `mapstructure:"on_create"`
	OnDelete OnDeleteStrategy `mapstructure:"on_delete"`
	OnPatch  Strategy         `mapstructure:"on_patch"`
}

// Strategy is the configuration strategy for the relations.
type Strategy struct {
	QueryOrder uint   `mapstructure:"query_order"`
	OnError    string `mapstructure:"on_error" validate:"isdefault|oneof=fail continue"`
}

// OnDeleteStrategy is the configuration for the 'on delete' option, relation strategy.
// It is a strategy that defines modification query order, reaction 'on error' and the
// reaction on root deletion 'relation_change' - when the root model is being deleted, how
// should the related model change. Possible options are:
// - set null - (default) - sets the related model foreign key to 'null' or 'empty'.
// - restrict - doesn't allow to delete the root model if it contains any relation value.
// - no action - does nothing with the related foreign key.
// - cascade - related data is being deleted when the parent data is being deleted.
type OnDeleteStrategy struct {
	QueryOrder     uint   `mapstructure:"query_order"`
	OnError        string `mapstructure:"on_error" validate:"isdefault|oneof=fail continue"`
	RelationChange string `mapstructure:"relation_change" validate:"isdefault|oneof=restrict cascade noaction setnull"`
}

// Connection is the configuration for non local schemas credentials.
// The connection config can be set by providing raw_url or with host,path,protocol.
type Connection struct {
	// Host defines the access hostname or the ip address
	Host string `mapstructure:"host" validate:"hostname|ip"`

	// Path is the connection path, just after the protocol and
	Path string `mapstructure:"path" validate:"isdefault|uri"`

	// Port is the connection port
	Port interface{} `mapstructure:"port"`

	// Protocol is the protocol used in the connection
	Protocol string `mapstructure:"protocol"`

	// RawURL is the raw connection url. If set it must define the protocol ('http://',
	// 'rpc://'...)
	RawURL string `mapstructure:"raw_url" validate:"isdefault|url"`

	// Username is the username used to get connection credential
	Username string `mapstructure:"username"`

	// Password is the password used to get connection credentials
	Password string `mapstructure:"password"`

	// Options contains connection dependent specific options
	Options map[string]interface{} `mapstructure:"options"`

	// MaxTimeout defines the maximum timeout for the given repository connection
	MaxTimeout *time.Duration `mapstructure:"max_timeout"`
}
