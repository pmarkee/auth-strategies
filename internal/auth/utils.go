package auth

import (
	"golang.org/x/crypto/argon2"
)

func ComputeHash(s string, salt []byte) []byte {
	return argon2.IDKey([]byte(s), salt, 3, 64*1024, 2, 32)
}
