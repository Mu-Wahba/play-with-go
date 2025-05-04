package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func ComparePassHash(hashed string, pass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pass))
	if err != nil {
		return errors.New("user not found")
	}
	return nil
}
