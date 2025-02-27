package password_hasher

import (
	"crypto/rand"
	"fmt"
	"slices"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type passwordHasher interface {
	GenHashedPassdWithSalt(rawPassword string) (hashedPass, salt string, err error)
	isSamePassword(hashedPassword, salt, pasword string) (bool, error)
}

type passwordHashType byte

const (
	BcryptPasswordHash  passwordHashType = iota
	Argon2iPasswordHash passwordHashType = iota
)

func NewPasswordHasher(hashType passwordHashType) passwordHasher {
	switch hashType {
	case BcryptPasswordHash:
		return &bcryptPassworddHasher{cost: bcrypt.DefaultCost}

	case Argon2iPasswordHash:
		return &argon2iPassworddHasher{time: 3, memory: 32 * 1024, threads: 4, keyLen: 32}
	}

	panic(fmt.Sprintf("Unsupported password hash type: %v", hashType))
}

// ---------------------------------------------------------------------------------------------------------------
// ------------------------------------------------Bcryp---------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------

type bcryptPassworddHasher struct {
	cost int
}

func (b bcryptPassworddHasher) GenHashedPassdWithSalt(rawPassword string) (hashedPass, salt string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), b.cost)
	return string(hash), "", err
}

func (bcryptPassworddHasher) isSamePassword(hashedPassword, salt, pasword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pasword))
	return err == nil, nil
}

// ---------------------------------------------------------------------------------------------------------------
// ------------------------------------------------Argon-i---------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------

type argon2iPassworddHasher struct {
	time, memory uint32
	threads      uint8
	keyLen       uint32
}

func (argon argon2iPassworddHasher) Key(password, salt []byte) []byte {
	return argon2.Key([]byte(password), salt, argon.time, argon.memory, argon.threads, argon.keyLen)
}

func (argon argon2iPassworddHasher) GenHashedPassdWithSalt(rawPassword string) (hashedPass, salt string, err error) {
	saltBytes := generateRandomSalt(32)
	hashedPassBytes := argon.Key([]byte(rawPassword), saltBytes)
	return string(hashedPassBytes), string(saltBytes), nil
}

func (argon argon2iPassworddHasher) isSamePassword(hashedPassword, salt, pasword string) (bool, error) {
	hashedPassBytes := []byte(hashedPassword)
	generatedHashedPassBytes := argon.Key([]byte(pasword), []byte(salt))
	return slices.Equal(hashedPassBytes, generatedHashedPassBytes), nil
}

func generateRandomSalt(saltLen uint8) []byte {
	salt := make([]byte, saltLen)
	rand.Read(salt)
	return salt
}
