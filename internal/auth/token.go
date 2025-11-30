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
