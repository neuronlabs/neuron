package common

// EncodeLinks is the structure used as a key in the store that states
// to encode the links for the encoder.
type EncodeLinks struct{}

// StoreKeys
var (
	EncodeLinksCtxKey = EncodeLinks{}
)
