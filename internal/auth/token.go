package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"aidanwoods.dev/go-paseto"
)

const tokenTTL = 24 * time.Hour

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
	hexKey := os.Getenv("PASETO_KEY")
	if hexKey == "" {
		return paseto.V4SymmetricKey{}, errors.New("PASETO_KEY not set")
	}
	return paseto.V4SymmetricKeyFromHex(hexKey)
}
