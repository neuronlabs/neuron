package test_models

import (
	"time"

	"github.com/neuronlabs/neuron/neuron-generator/internal/test_models/wrapper"
)

type User struct {
	ID          string
	CreatedAt   time.Time
	DeletedAt   *time.Time
	Name        *string
	Age         int
	Bytes       []byte
	Wrapped     wrapper.Int
	PtrWrapped  *wrapper.Int
	FavoriteCar *Car
	Cars        []*Car
	Sons        []User
	Sister      User
}

//go:generate neuron-generator models methods  .

type Car struct {
	ID     uint
	Plates string
}
