package auth

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"strings"
	"unicode"

	"github.com/neuronlabs/neuron/errors"
)

const passwordSpecials = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// Password is a structure that defines the password and its properties.
type Password struct {
	// Password is the string value of the provided password.
	Password string
	// Uppers is a count of the uppercase letters.
	Uppers int
	// Lowers is a count of the lowercase letters.
	Lowers int
	// Specials is a count of special symbols.
	Specials int
	// Numbers is a count of numbers in the password.
	Numbers int
	// UniqueRunes is the number of unique runes int
	UniqueRunes int
	// Score is a password strength score.
	Score int
}

// UniqueRunesRatio gets the ratio of unique runes to the total password length.
func (p *Password) UniqueRunesRatio() float64 {
	return float64(p.UniqueRunes) / float64(len(p.Password))
}

// GenerateSalt creates a crypto random byte slice salt.
func GenerateSalt(saltLength int) ([]byte, error) {
	return getRandomBytes(saltLength)
}

// MD5 creates salted hash password using MD5 function.
func (p *Password) MD5(salt []byte) (password []byte, err error) {
	h := md5.New()
	return hashPassword(h, p.Password, salt)
}

// SHA256 creates salted hash password using SHA256 function.
func (p *Password) SHA256(salt []byte) ([]byte, error) {
	h := sha256.New()
	return hashPassword(h, p.Password, salt)
}

// SHA512 creates salted hash password using SHA512 function.
func (p *Password) SHA512(salt []byte) ([]byte, error) {
	h := sha512.New()
	return hashPassword(h, p.Password, salt)
}

// Hash gets the password hash with salted with 'salt'.
func (p *Password) Hash(h hash.Hash, salt []byte) ([]byte, error) {
	return hashPassword(h, p.Password, salt)
}

// CompareMD5Password compares if provided password matches the md5 hashed password with given salt.
func CompareMD5Password(password string, hashedPassword, salt []byte) (bool, error) {
	h := md5.New()
	return equalPassword(h, password, hashedPassword, salt)
}

// CompareSHA256Password compares if provided password matches the sha256 hashed password with given salt.
func CompareSHA256Password(password string, hashedPassword, salt []byte) (bool, error) {
	h := sha256.New()
	return equalPassword(h, password, hashedPassword, salt)
}

// CompareSHA512Password compares if provided password matches the sha512 hashed password with given salt.
func CompareSHA512Password(password string, hashedPassword, salt []byte) (bool, error) {
	h := sha512.New()
	return equalPassword(h, password, hashedPassword, salt)
}

// CompareHashPassword compares if provided password matches the sha512 hashed password with given salt.
func CompareHashPassword(h hash.Hash, password string, hashedPassword, salt []byte) (bool, error) {
	return equalPassword(h, password, hashedPassword, salt)
}

// NewPassword creates and analyze the 'password' using provided (optional) scorer function. If no 'scorer' is provided
// than the 'DefaultPasswordScorer' would be used.
func NewPassword(password string, scorer ...PasswordScorer) *Password {
	if len(password) == 0 {
		return nil
	}

	score := DefaultPasswordScorer
	if len(scorer) == 1 {
		score = scorer[0]
	}
	pw := &Password{
		Password: password,
	}
	uniqueRunes := map[rune]struct{}{}
	for _, rn := range password {
		_, ok := uniqueRunes[rn]
		if !ok {
			pw.UniqueRunes++
			uniqueRunes[rn] = struct{}{}
		}
		switch {
		case unicode.IsNumber(rn):
			pw.Numbers++
		case unicode.IsLower(rn):
			pw.Lowers++
		case unicode.IsUpper(rn):
			pw.Uppers++
		default:
			if isSpecial(rn) {
				pw.Specials++
			}
		}
	}
	score(pw)
	return pw
}

// PasswordScorer is a function that sets the score for given password.
type PasswordScorer func(pw *Password)

// DefaultPasswordScore
func DefaultPasswordScorer(pw *Password) {
	// If all of the runes are the same the the password has no score at all, no matter what length it is.
	if pw.UniqueRunes == 1 && len(pw.Password) != 1 {
		return
	}

	if len(pw.Password) > 8 {
		pw.Score++
	}
	if len(pw.Password) > 12 {
		pw.Score++
	}

	// Check if at least 80% of the runes are unique.
	if pw.UniqueRunesRatio() > 0.8 {
		pw.Score++
	}

	// If all runes are unique give another point.
	if pw.UniqueRunes == len(pw.Password) {
		pw.Score++
	}

	if pw.Uppers > 0 {
		pw.Score++
	}
	if pw.Numbers > 0 {
		pw.Score++
	}
	if pw.Specials == 1 {
		pw.Score++
	} else if pw.Specials > 1 {
		pw.Score += 2
	}
}

// PasswordValidator is a function that validates the password.
type PasswordValidator func(*Password) error

// DefaultPasswordValidator is the default password validator function.
func DefaultPasswordValidator(p *Password) error {
	if p.Score == 0 {
		return errors.Wrap(ErrInvalidSecret, "provided password is too weak")
	}

	if p.UniqueRunesRatio() < 0.3 {
		return errors.WrapDet(ErrInvalidSecret, "password is too weak").
			WithDetail("Password has too many similar characters.")
	}
	sb := strings.Builder{}
	var failed bool
	if len(p.Password) < 6 {
		failed = true
		sb.WriteString("Password needs to be at least 6 characters long.")
	}
	if p.Numbers == 0 {
		if failed {
			sb.WriteString(" ")
		}
		sb.WriteString("Password requires at least one number.")
		failed = true
	}

	if p.Specials == 0 {
		if failed {
			sb.WriteString(" ")
		}
		sb.WriteString("Password requires at least one special character.")
		failed = true
	}

	if p.Uppers == 0 {
		if failed {
			sb.WriteString(" ")
		}
		sb.WriteString("Password requires at least one capital letter.")
		failed = true
	}

	if failed {
		return errors.WrapDet(ErrInvalidSecret, "provided invalid password").WithDetail(sb.String())
	}
	return nil
}

func equalPassword(h hash.Hash, password string, hashedPassword, salt []byte) (bool, error) {
	thisHashed, err := hashPassword(h, password, salt)
	if err != nil {
		return false, err
	}
	return bytes.Equal(thisHashed, hashedPassword), nil
}

// getRandomBytes will generate random bytes.  This is for internal
// use in the library itself.
func getRandomBytes(length int) ([]byte, error) {
	randomData := make([]byte, length)
	if _, err := rand.Read(randomData); err != nil {
		return nil, errors.Wrap(ErrInternalError, "getting random bytes failed")
	}
	return randomData, nil
}

func hashPassword(h hash.Hash, password string, salt []byte) ([]byte, error) {
	_, err := h.Write([]byte(password))
	if err != nil {
		return nil, errors.Wrap(ErrInternalError, "writing to hash failed")
	}
	_, err = h.Write(salt)
	if err != nil {
		return nil, errors.Wrap(ErrInternalError, "writing to hash failed")
	}
	return h.Sum(nil), nil
}

func isSpecial(rn int32) bool {
	for _, special := range passwordSpecials {
		if rn == special {
			return true
		}
	}
	return false
}
