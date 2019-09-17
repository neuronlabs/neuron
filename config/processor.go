package config

import (
	"errors"
	"github.com/spf13/viper"
	"strings"
	"time"

	"github.com/neuronlabs/neuron-core/internal"
)

// Processor is the config used for the scope processor.
type Processor struct {
	// DefaultTimeout is the default timeout used for given processor.
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
	// CountProcesses are the processes used for count method.=
	CountProcesses ProcessList `mapstructure:"count_processes"`
}

// Validate validates the processor values.
func (p *Processor) Validate() error {
	err := &multiProcessError{}

	if len(p.CreateProcesses) == 0 {
		return errors.New("no create processes in configuration")
	}
	if len(p.DeleteProcesses) == 0 {
		return errors.New("no delete processes in configuration")
	}
	if len(p.GetProcesses) == 0 {
		return errors.New("no get processes in configuration")
	}
	if len(p.ListProcesses) == 0 {
		return errors.New("no list processes in configuration")
	}
	if len(p.PatchProcesses) == 0 {
		return errors.New("no patch processes in configuration")
	}
	if len(p.CountProcesses) == 0 {
		return errors.New("no count processes in configuration")
	}

	p.CreateProcesses.validate(err)
	p.DeleteProcesses.validate(err)
	p.GetProcesses.validate(err)
	p.ListProcesses.validate(err)
	p.PatchProcesses.validate(err)
	p.CountProcesses.validate(err)

	if len(err.processes) == 0 {
		return nil
	}
	return err
}

// DefaultThreadsafeProcessorConfig creates default config for the Processor.
func DefaultThreadsafeProcessorConfig() map[string]interface{} {
	return defaultThreadsafeProcessorConfig()
}

func defaultThreadsafeProcessorConfig() map[string]interface{} {
	return map[string]interface{}{
		"default_timeout": time.Second * 30,
		"create_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessHookBeforeCreate,
			internal.ProcessSetBelongsToRelations,
			internal.ProcessCreate,
			internal.ProcessStoreScopePrimaries,
			internal.ProcessPatchForeignRelationsSafe,
			internal.ProcessHookAfterCreate,
			internal.ProcessTxCommitOrRollback,
		},
		"get_processes": []string{
			internal.ProcessCheckPagination,
			internal.ProcessFillEmptyFieldset,
			internal.ProcessConvertRelationFiltersSafe,
			internal.ProcessHookBeforeGet,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCheckPagination,
			internal.ProcessConvertRelationFiltersSafe,
			internal.ProcessGet,
			internal.ProcessGetForeignRelations,
			internal.ProcessHookAfterGet,
			internal.ProcessGetIncludedSafe,
		},
		"list_processes": []string{
			internal.ProcessCheckPagination,
			internal.ProcessFillEmptyFieldset,
			internal.ProcessConvertRelationFiltersSafe,
			internal.ProcessHookBeforeList,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCheckPagination,
			internal.ProcessConvertRelationFiltersSafe,
			internal.ProcessList,
			internal.ProcessGetForeignRelationsSafe,
			internal.ProcessHookAfterList,
			internal.ProcessGetIncludedSafe,
		},
		"patch_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessHookBeforePatch,
			internal.ProcessDeletedAtFilter,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessPatchBelongsToRelations,
			internal.ProcessPatch,
			internal.ProcessPatchForeignRelationsSafe,
			internal.ProcessHookAfterPatch,
			internal.ProcessTxCommitOrRollback,
		},
		"delete_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessHookBeforeDelete,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessDelete,
			internal.ProcessDeleteForeignRelationsSafe,
			internal.ProcessHookAfterDelete,
			internal.ProcessTxCommitOrRollback,
		},
		"count_processes": []string{
			internal.ProcessConvertRelationFiltersSafe,
			internal.ProcessHookBeforeCount,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCount,
			internal.ProcessHookAfterCount,
		},
	}
}

// ThreadSafeProcessor creates the goroutine safe query processor configuration.
func ThreadSafeProcessor() *Processor {
	p := &Processor{}
	v := viper.New()

	k := "ts_processor"
	v.Set(k, defaultThreadsafeProcessorConfig())
	err := v.UnmarshalKey(k, p)
	if err != nil {
		panic(err)
	}
	return p
}

// ConcurrentProcessor creates the concurrent processor confuration.
func ConcurrentProcessor() *Processor {
	p := &Processor{}
	v := viper.New()

	k := "ts_processor"
	v.Set(k, defaultConcurrentProcessorConfig())
	err := v.UnmarshalKey(k, p)
	if err != nil {
		panic(err)
	}
	return p
}

// DefaultConcurrentProcessorConfig creates default concurrent config for the Processor.
func DefaultConcurrentProcessorConfig() map[string]interface{} {
	return defaultConcurrentProcessorConfig()
}

func defaultConcurrentProcessorConfig() map[string]interface{} {
	return map[string]interface{}{
		"default_timeout": time.Second * 30,
		"create_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessHookBeforeCreate,
			internal.ProcessSetBelongsToRelations,
			internal.ProcessCreate,
			internal.ProcessStoreScopePrimaries,
			internal.ProcessPatchForeignRelations,
			internal.ProcessHookAfterCreate,
			internal.ProcessTxCommitOrRollback,
		},
		"get_processes": []string{
			internal.ProcessCheckPagination,
			internal.ProcessFillEmptyFieldset,
			internal.ProcessConvertRelationFilters,
			internal.ProcessHookBeforeGet,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCheckPagination,
			internal.ProcessConvertRelationFilters,
			internal.ProcessGet,
			internal.ProcessGetForeignRelations,
		},
		"list_processes": []string{
			internal.ProcessCheckPagination,
			internal.ProcessFillEmptyFieldset,
			internal.ProcessConvertRelationFilters,
			internal.ProcessHookBeforeList,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCheckPagination,
			internal.ProcessConvertRelationFilters,
			internal.ProcessList,
			internal.ProcessGetForeignRelations,
			internal.ProcessHookAfterList,
			internal.ProcessGetIncluded,
		},
		"patch_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessHookBeforePatch,
			internal.ProcessDeletedAtFilter,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessPatchBelongsToRelations,
			internal.ProcessPatch,
			internal.ProcessPatchForeignRelations,
			internal.ProcessHookAfterPatch,
			internal.ProcessTxCommitOrRollback,
		},
		"delete_processes": []string{
			internal.ProcessTxBegin,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessHookBeforeDelete,
			internal.ProcessReducePrimaryFilters,
			internal.ProcessDelete,
			internal.ProcessDeleteForeignRelations,
			internal.ProcessHookAfterDelete,
			internal.ProcessTxCommitOrRollback,
		},
		"count_processes": []string{
			internal.ProcessConvertRelationFilters,
			internal.ProcessHookBeforeCount,
			internal.ProcessDeletedAtFilter,
			internal.ProcessCount,
			internal.ProcessHookAfterCount,
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
