package common

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(h), nil
}

func PasswordHashCompare(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func PassGenerate(length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	var password string
	for i := 0; i < length; i++ {
		bintRand, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))

		if err != nil {
			panic(err)
		}

		intRand := int(bintRand.Int64())

		password += string(chars[intRand%len(chars)])
	}
	return password
}
