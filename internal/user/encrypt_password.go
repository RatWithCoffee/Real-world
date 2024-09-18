package user

import (
	"crypto/rand"
	"golang.org/x/crypto/argon2"
	"log"
	"reflect"
)

const saltSize = 16

func GetPasswordSalt(password []byte) (encrPass []byte, salt []byte) {
	salt = GenerateRandomSalt(saltSize)
	return PasswordArgon2WithSalt(password, salt), salt
}

func PasswordArgon2WithSalt(plainPassword []byte, salt []byte) []byte {
	return argon2.IDKey(plainPassword, salt, 1, 64*1024, 4, 32)
}

func IsValidPassword(plainPassword []byte, salt []byte, encryptedPassword []byte) bool {
	return reflect.DeepEqual(PasswordArgon2WithSalt(plainPassword, salt), encryptedPassword)
}

func GenerateRandomSalt(saltSize int) []byte {
	var salt = make([]byte, saltSize)
	_, err := rand.Read(salt[:])
	if err != nil {
		log.Println(err)
	}

	return salt
}
