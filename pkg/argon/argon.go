package argon

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
const (
	Memory      = 7 * 1024
	Iterations  = 5
	Parallelism = 1
	SaltLength  = 23
	KeyLength   = 32
)

var (
	ErrInvalidEncodedHash   = errors.New("invalid format for encoded hash")
	ErrUnhandledEncodedHash = errors.New("unhandled encoded hash")
	ErrUnhandledVersion     = errors.New("unhandled version")
	ErrHashDontMatch        = errors.New("hashes don't match")
)

func GenerateFromPassword(password string) (string, error) {
	salt, err := salt(SaltLength)
	if err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, Iterations, Memory, Parallelism, KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, Memory, Iterations, Parallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func salt(length uint) ([]byte, error) {
	payload := make([]byte, length)

	if _, err := rand.Read(payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func CompareHashAndPassword(encoded, password string) error {
	memory, iterations, parallelism, salt, hash, err := parseHash(encoded)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	testedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	if subtle.ConstantTimeCompare(testedHash, hash) == 1 {
		return nil
	}

	return ErrHashDontMatch
}

func parseHash(encoded string) (uint32, uint32, uint8, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return 0, 0, 0, nil, nil, ErrInvalidEncodedHash
	}

	if parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, ErrUnhandledEncodedHash
	}

	var version int

	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("decode version: %w", err)
	}

	if version != argon2.Version {
		return 0, 0, 0, nil, nil, ErrUnhandledVersion
	}

	var memory, iterations uint32
	var parallelism uint8

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("decode params: %w", err)
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	return memory, iterations, parallelism, salt, hash, nil
}
