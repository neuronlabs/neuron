package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
)

// ProcessChain is the slice (chain) of processes.
type ProcessChain []*Process

// InsertBefore adds the process before the process with the 'before' name.
func (c *ProcessChain) InsertBefore(before string, processes ...*Process) error {
	var (
		index int
		found bool
	)
	for i, p := range *c {
		if p.Name == before {
			index = i
			found = true
			break
		}
	}
	if !found {
		return errors.Newf(class.QueryProcessorNotFound, "process: '%s' not found", before)
	}

	var chain = ProcessChain{}
	if index == 0 {
		chain = append(chain, processes...)
		chain = append(chain, *c...)
	} else {
		chain = append(chain, (*c)[:index]...)
		chain = append(chain, processes...)
		chain = append(chain, (*c)[index:]...)
	}

	*c = chain
	return nil
}

// InsertAfter inserts the processes after the provided process name
func (c *ProcessChain) InsertAfter(after string, processes ...*Process) error {
	var (
		index int
		found bool
	)
	for i, p := range *c {
		if p.Name == after {
			index = i
			found = true
			break
		}
	}
	if !found {
		return errors.Newf(class.QueryProcessorNotFound, "process: '%s' not found", after)
	}

	var chain = ProcessChain{}
	if index != len(*c)-1 {
		chain = append(chain, (*c)[:index]...)
		chain = append(chain, processes...)
		chain = append(chain, (*c)[index:]...)
	} else {
		chain = append(*c, processes...)
	}
	*c = chain

	return nil
}

// Replace replaces the process within the ProcessChain
func (c *ProcessChain) Replace(toReplace string, process *Process) error {
	var (
		index int
		found bool
	)
	for i, p := range *c {
		if p.Name == toReplace {
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.Newf(class.QueryProcessorNotFound, "process: '%s' not found", toReplace)
	}

	(*c)[index] = process
	return nil
}

// DeleteProcess deletes the process from the chain
func (c *ProcessChain) DeleteProcess(processName string) error {

	var (
		index int
		found bool
	)
	for i, p := range *c {
		if p.Name == processName {
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.Newf(class.QueryProcessorNotFound, "process: '%s' not found", processName)
	}

	chain := ProcessChain{}

	if index != len(*c)-1 {
		chain = append((*c)[:index], (*c)[index+1:]...)
	} else if index == 0 {
		chain = append(chain, (*c)[index+1:]...)
	} else {
		chain = (*c)[:index]
	}

	*c = chain

	return nil

}
