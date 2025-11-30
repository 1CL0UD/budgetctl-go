package auth

import (
	"os"
	"strconv"
	"sync"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	once      sync.Once
	cachedKey paseto.V4SymmetricKey
	cachedErr error
	tokenTTL  = 24 * time.Hour
)

// GenerateToken creates a signed PASETO v4 token for a user ID.
func GenerateToken(userID int64) (string, error) {
	key, err := loadKey()
	if err != nil {
		return "", err
	}

	token := paseto.NewToken()
	token.SetString("sub", strconv.FormatInt(userID, 10))
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(tokenTTL))

	return token.V4Encrypt(key, nil), nil
}

// ParseToken validates a token and returns the subject user ID.
func ParseToken(tokenStr string) (int64, error) {
	key, err := loadKey()
	if err != nil {
		return 0, err
	}

	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(key, tokenStr, nil)
	if err != nil {
		return 0, err
	}

	subject, err := token.GetString("sub")
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func loadKey() (paseto.V4SymmetricKey, error) {
	once.Do(func() {
		if hexKey := os.Getenv("PASETO_KEY"); hexKey != "" {
			cachedKey, cachedErr = paseto.V4SymmetricKeyFromHex(hexKey)
			return
		}
		// Auto-generate a key if none is provided; replace with env in production.
		cachedKey = paseto.NewV4SymmetricKey()
	})

	return cachedKey, cachedErr
}
