package server

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/database"
)

// Params are the parameters passed to the server version.
type Params struct {
	Ctx           context.Context
	Authorizer    auth.Verifier
	Authenticator auth.Authenticator
	Tokener       auth.Tokener
	DB            database.DB
}
