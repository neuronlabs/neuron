package config

import (
	"errors"
	"strings"
	"time"

	"github.com/neuronlabs/neuron-core/internal"
)

// Processor is the config used for the scope processor.
type Processor struct {
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`

	// CreateProcesses are the default processes used in the create method.
	CreateProcesses ProcessList `mapstructure:"create_processes"`

	// DeleteProcesses are the default processes used in the delete method.
	DeleteProcesses ProcessList `mapstructure:"delete_processes"`

	// GetProcesses are the default processes used in the get method.
	GetProcesses ProcessList `mapstructure:"get_processes"`

	// ListProcesses are the default processes used in the list method.
	ListProcesses ProcessList `mapstructure:"list_processes"`

	// PatchProcesses are the default processes used in the patch method.
	PatchProcesses ProcessList `mapstructure:"patch_processes"`
}

// Validate validates the processor values.
func (p *Processor) Validate() error {
	err := &multiProcessError{}

	if len(p.CreateProcesses) == 0 {
		return errors.New("No create processes in configuration")
	}

	if len(p.DeleteProcesses) == 0 {
		return errors.New("No create processes in configuration")
	}

	if len(p.GetProcesses) == 0 {
		return errors.New("No create processes in configuration")
	}

	if len(p.ListProcesses) == 0 {
		return errors.New("No create processes in configuration")
	}

	if len(p.PatchProcesses) == 0 {
		return errors.New("No create processes in configuration")
	}

	p.CreateProcesses.validate(err)
	p.DeleteProcesses.validate(err)
	p.GetProcesses.validate(err)
	p.ListProcesses.validate(err)
	p.PatchProcesses.validate(err)

	if len(err.processes) == 0 {
		return nil
	}
	return err
}

// DefaultProcessorConfig creates default config for the Processor.
func DefaultProcessorConfig() map[string]interface{} {
	return map[string]interface{}{
		"default_timeout": time.Second * 30,
		"create_processes": []string{
			"neuron:begin_transaction",
			"neuron:hook_before_create",
			"neuron:set_belongs_to_relationships",
			"neuron:create",
			"neuron:store_scope_primaries",
			"neuron:patch_foreign_relationships_safe",
			"neuron:hook_after_create",
			"neuron:commit_or_rollback_transaction",
		},
		"get_processes": []string{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters_safe",
			"neuron:hook_before_get",
			"neuron:convert_relationship_filters_safe",
			"neuron:get",
			"neuron:get_foreign_relationships",
		},
		"list_processes": []string{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters_safe",
			"neuron:hook_before_list",
			"neuron:convert_relationship_filters_safe",
			"neuron:list",
			"neuron:get_foreign_relationships_safe",
			"neuron:hook_after_list",
			"neuron:get_included_safe",
		},
		"patch_processes": []string{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_patch",
			"neuron:reduce_primary_filters",
			"neuron:patch_belongs_to_relationships",
			"neuron:patch",
			"neuron:patch_foreign_relationships_safe",
			"neuron:hook_after_patch",
			"neuron:commit_or_rollback_transaction",
		},
		"delete_processes": []string{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_delete",
			"neuron:reduce_primary_filters",
			"neuron:delete",
			"neuron:delete_foreign_relationships_safe",
			"neuron:hook_after_delete",
			"neuron:commit_or_rollback_transaction",
		},
	}
}

// ThreadSafeProcessor creates the goroutine safe query processor configuration.
func ThreadSafeProcessor() *Processor {
	return &Processor{
		DefaultTimeout: time.Second * 30,
		CreateProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:hook_before_create",
			"neuron:set_belongs_to_relationships",
			"neuron:create",
			"neuron:store_scope_primaries",
			"neuron:patch_foreign_relationships_safe",
			"neuron:hook_after_create",
			"neuron:commit_or_rollback_transaction",
		},
		GetProcesses: ProcessList{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters_safe",
			"neuron:hook_before_get",
			"neuron:convert_relationship_filters_safe",
			"neuron:get",
			"neuron:get_foreign_relationships",
		},
		ListProcesses: ProcessList{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters_safe",
			"neuron:hook_before_list",
			"neuron:convert_relationship_filters_safe",
			"neuron:list",
			"neuron:get_foreign_relationships_safe",
			"neuron:hook_after_list",
			"neuron:get_included",
		},
		PatchProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_patch",
			"neuron:reduce_primary_filters",
			"neuron:patch_belongs_to_relationships",
			"neuron:patch",
			"neuron:patch_foreign_relationships_safe",
			"neuron:hook_after_patch",
			"neuron:commit_or_rollback_transaction",
		},
		DeleteProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_delete",
			"neuron:reduce_primary_filters",
			"neuron:delete",
			"neuron:delete_foreign_relationships_safe",
			"neuron:hook_after_delete",
			"neuron:commit_or_rollback_transaction",
		},
	}
}

// ConcurrentProcessor creates the concurrent processor confuration.
func ConcurrentProcessor() *Processor {
	return &Processor{
		DefaultTimeout: time.Second * 30,
		CreateProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:hook_before_create",
			"neuron:set_belongs_to_relationships",
			"neuron:create",
			"neuron:store_scope_primaries",
			"neuron:patch_foreign_relationships",
			"neuron:hook_after_create",
			"neuron:commit_or_rollback_transaction",
		},
		GetProcesses: ProcessList{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters",
			"neuron:hook_before_get",
			"neuron:convert_relationship_filters",
			"neuron:get",
			"neuron:get_foreign_relationships",
		},
		ListProcesses: ProcessList{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters",
			"neuron:hook_before_list",
			"neuron:convert_relationship_filters",
			"neuron:list",
			"neuron:get_foreign_relationships",
			"neuron:hook_after_list",
			"neuron:get_included",
		},
		PatchProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_patch",
			"neuron:reduce_primary_filters",
			"neuron:patch_belongs_to_relationships",
			"neuron:patch",
			"neuron:patch_foreign_relationships",
			"neuron:hook_after_patch",
			"neuron:commit_or_rollback_transaction",
		},
		DeleteProcesses: ProcessList{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_delete",
			"neuron:reduce_primary_filters",
			"neuron:delete",
			"neuron:delete_foreign_relationships",
			"neuron:hook_after_delete",
			"neuron:commit_or_rollback_transaction",
		},
	}
}

// DefaultConcurrentProcessorConfig creates default concurrent config for the Processor.
func DefaultConcurrentProcessorConfig() map[string]interface{} {
	return map[string]interface{}{
		"default_timeout": time.Second * 30,
		"create_processes": []string{
			"neuron:begin_transaction",
			"neuron:hook_before_create",
			"neuron:set_belongs_to_relationships",
			"neuron:create",
			"neuron:store_scope_primaries",
			"neuron:patch_foreign_relationships",
			"neuron:hook_after_create",
			"neuron:commit_or_rollback_transaction",
		},
		"get_processes": []string{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters",
			"neuron:hook_before_get",
			"neuron:convert_relationship_filters",
			"neuron:get",
			"neuron:get_foreign_relationships",
		},
		"list_processes": []string{
			"neuron:fill_empty_fieldset",
			"neuron:convert_relationship_filters",
			"neuron:hook_before_list",
			"neuron:convert_relationship_filters",
			"neuron:list",
			"neuron:get_foreign_relationships",
			"neuron:hook_after_list",
			"neuron:get_included",
		},
		"patch_processes": []string{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_patch",
			"neuron:reduce_primary_filters",
			"neuron:patch_belongs_to_relationships",
			"neuron:patch",
			"neuron:patch_foreign_relationships",
			"neuron:hook_after_patch",
			"neuron:commit_or_rollback_transaction",
		},
		"delete_processes": []string{
			"neuron:begin_transaction",
			"neuron:reduce_primary_filters",
			"neuron:hook_before_delete",
			"neuron:reduce_primary_filters",
			"neuron:delete",
			"neuron:delete_foreign_relationships",
			"neuron:hook_after_delete",
			"neuron:commit_or_rollback_transaction",
		},
	}
}

// ProcessList is a list of the processes.
type ProcessList []string

func (p ProcessList) validate(err *multiProcessError) {
	for _, process := range p {
		if _, ok := internal.Processes[process]; !ok {
			err.add(process)
		}
	}
}

type multiProcessError struct {
	processes []string
}

func (m *multiProcessError) Error() string {
	sb := &strings.Builder{}

	sb.WriteString(strings.Join(m.processes, ","))

	if len(m.processes) > 1 {
		sb.WriteString(" query processes are")
	} else {
		sb.WriteString(" query process is")
	}

	sb.WriteString(" not registered")

	return sb.String()
}

func (m *multiProcessError) add(p string) {
	m.processes = append(m.processes, p)
}
