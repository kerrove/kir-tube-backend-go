package password

import "golang.org/x/crypto/bcrypt"

func Encode(p string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func Validate(userPassword, bodyPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(bodyPassword))

	if err != nil {
		return false
	} else {
		return true
	}
}
