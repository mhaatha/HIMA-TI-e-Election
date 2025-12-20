package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

var DefaultPasswordLength = 12

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func HashNIM(nim string) string {
	hash := sha256.Sum256([]byte(nim))
	return hex.EncodeToString(hash[:])
}

func GeneratePassword(length int) (string, error) {
	const charset = "abcdefghjkmnopqrstuvwxyz" +
		"ABCDEFGHJKMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

	bytePassword := make([]byte, length)
	charsetLength := byte(len(charset))

	for i := range bytePassword {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(charsetLength)))
		if err != nil {
			return "", err
		}
		bytePassword[i] = charset[num.Int64()]
	}

	return string(bytePassword), nil
}
