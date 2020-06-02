package auth

import (
	"context"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
)

// Authorizator is the interface used to authorize resources.
type Authorizator interface {
	// Init initialize the authorizator. This function should set all options for the authorizator.
	Init(options ...Options) error
	// Verify if the 'accountID' is allowed to access the resource. The resourceID is a unique resource identifier.
	Verify(accountID, resourceID string) error
}

// Authenticator is the interface used to authenticate the username and password.
type Authenticator interface {
	Initialize(options ...Option) error
	Authenticate(username, password string) (accountID string, err error)
}

// Accounter is the interface used for account operations.
type Accounter interface {
	CreateAccount(account Account) error
	DeleteAccount(accountID string) error
	ChangeSecret(accountID string, secret string) error
}

// Account is the interface used to define the service account.
type Account interface {
	AccountID() string
	AccountUsername() string
	AccountSecret() string
}

// Options are the authorization service options.
type Options struct {
	// Controller used for the service.
	Controller *controller.Controller
	// Secret is the authorization secret.
	Secret string
	// PublicKey is used for decoding the token public key.
	PublicKey string
	// PrivateKey is used for encoding the token private key.
	PrivateKey string
	// RepositoryName is the authorization service repository name.
	RepositoryName string
}

type Option func(o *Options)

type accountKey struct{}

// CtxWithAccountID stores account id in the context.
func CtxWithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountKey{}, accountID)
}

// CtxAccountID gets the account id from the context 'ctx'. If the context doesn't contain account id the
// function returns empty string.
func CtxAccountID(ctx context.Context) string {
	accID, ok := ctx.Value(accountKey{}).(string)
	if !ok {
		log.Debug("No account id found in the context 'ctx'.")
	}
	return accID
}
