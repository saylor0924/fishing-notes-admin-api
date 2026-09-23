package auth

import "golang.org/x/crypto/bcrypt"

func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
