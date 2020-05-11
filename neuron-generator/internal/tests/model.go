package tests

import (
	"time"

	"github.com/neuronlabs/neuron/neuron-generator/internal/tests/external"
)

// User is testing model.
type User struct {
	ID          string
	CreatedAt   time.Time
	DeletedAt   *time.Time
	Name        *string
	Age         int
	Bytes       []byte
	Wrapped     external.Int
	PtrWrapped  *external.Int
	FavoriteCar Car
	Cars        []*Car
	Sons        []User
	Sister      *User
}

//go:generate neuron-generator models methods .

// Car is the test model for generator.
type Car struct {
	ID     uint
	Plates string
}
