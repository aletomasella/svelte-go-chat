package common

import "golang.org/x/crypto/bcrypt"

func AdressHash(adress string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(adress), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(h), nil
}

func AdressHashCompare(adress string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(adress))
}
