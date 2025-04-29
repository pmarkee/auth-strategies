package auth

import (
	"crypto/rand"
	"golang.org/x/crypto/argon2"
)

func computeHash(s string, salt []byte) []byte {
	return argon2.IDKey([]byte(s), salt, 3, 64*1024, 2, 32)
}

func generateSalt() ([]byte, error) {
	var salt = make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}
