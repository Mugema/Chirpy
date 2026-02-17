package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPassword(password, hashString string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hashString)
	if err != nil {
		return false, err
	}
	return match, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("no auth in header")
	}

	token := strings.Split(auth, " ")
	if len(token) == 1 {
		return "", errors.New("no auth token")
	}

	return token[1], nil
}

func MakeRefreshToken() (string, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	encodedString := hex.EncodeToString(data)
	fmt.Printf("Created refreshToken for user %v \n", encodedString)
	return encodedString, nil
}
