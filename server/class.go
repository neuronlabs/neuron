package server

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	MjrServer               errors.Major
	ClassDuplicatedEndpoint errors.Class
)

func init() {
	MjrServer = errors.MustNewMajor()
	ClassDuplicatedEndpoint = errors.MustNewMajorClass(MjrServer)
}
