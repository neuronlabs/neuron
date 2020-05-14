package tests

import (
	"time"

	"github.com/neuronlabs/neuron/neuron-generator/internal/tests/external"
)

//go:generate neuron-generator models methods --format=goimports --single-file .
//go:generate neuron-generator models collections --format=goimports --single-file .

// User is testing model.
type User struct {
	ID          [16]byte
	CreatedAt   time.Time
	DeletedAt   *time.Time
	Name        *string
	Age         int
	Bytes       []byte
	PtrBytes    *[]byte
	Wrapped     external.Int
	PtrWrapped  *external.Int
	External    *external.Model
	FavoriteCar Car
	Cars        []*Car
	Sons        []User
	Sister      *User
}

// Car is the test model for generator.
type Car struct {
	ID     *[16]byte
	Plates string
}