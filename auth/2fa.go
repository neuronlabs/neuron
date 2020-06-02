package auth

// TwoFactorAuthenticator is the two factor authenticator
type TwoFactorAuthenticator interface {
	HasTwoFactorAuth(accountID string) (bool, error)
	GenerateTwoFactorAuth(accountID string) (secret string, err error)
	TwoFactorCreate(accountID string) (verifyID string, err error)
	TwoFactorVerify(verifyID string, code string) error
}
