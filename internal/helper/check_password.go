package helper

import "golang.org/x/crypto/bcrypt"

// CheckPasswordHash checks if the input password matches the hashed password
func CheckPasswordHash(hashedPassword, inputPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
	return err == nil
}
