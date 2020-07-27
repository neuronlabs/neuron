package auth

// TwoFactorAuthenticator is the two factor authenticator
type TwoFactorAuthenticator interface {
	HasTwoFactorAuth(accountID interface{}) (bool, error)
	GenerateTwoFactorAuth(accountID interface{}) (secret string, err error)
	TwoFactorCreate(accountID interface{}) (verifyID interface{}, err error)
	TwoFactorVerify(verifyID string, code string) error
}
