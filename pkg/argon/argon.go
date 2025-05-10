package argon

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
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

	strictBase64Decoder = base64.RawStdEncoding.Strict()
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
	var version int
	var memory, iterations uint32
	var parallelism uint8
	var salt, hash []byte

	var start int
	var err error
	var partCount int

	for index := 0; index < len(encoded); index++ {
		if encoded[index] != '$' {
			continue
		}

		part := encoded[start:index]

		switch partCount {
		case 1:
			if part != "argon2id" {
				return 0, 0, 0, nil, nil, ErrUnhandledEncodedHash
			}

		case 2:
			if version, err = strconv.Atoi(strings.TrimPrefix(part, "v=")); err != nil {
				return 0, 0, 0, nil, nil, fmt.Errorf("decode version `%s` : %w", part, err)
			}

			if version != argon2.Version {
				return 0, 0, 0, nil, nil, ErrUnhandledVersion
			}

		case 3:
			var seen, lastEqual, lastStart int

			for position := 0; position < len(part); position++ {
				if part[position] == '=' {
					lastEqual = position
					continue
				}

				if part[position] == ',' {
					switch part[lastStart:lastEqual] {
					case "m":
						bigMemory, err := strconv.ParseUint(part[lastEqual+1:position], 10, 32)
						if err != nil {
							return 0, 0, 0, nil, nil, fmt.Errorf("decode memory `%s`: %w", part, err)
						}

						seen |= 1 << 0
						memory = uint32(bigMemory)

					case "t":
						bigIterations, err := strconv.ParseUint(part[lastEqual+1:position], 10, 32)
						if err != nil {
							return 0, 0, 0, nil, nil, fmt.Errorf("decode iteration `%s`: %w", part, err)
						}

						seen |= 1 << 1
						iterations = uint32(bigIterations)
					}

					lastStart = position + 1
					continue
				}
			}

			if part[lastStart:lastEqual] == "p" {
				bigParallelism, err := strconv.ParseUint(part[lastEqual+1:], 10, 8)
				if err != nil {
					return 0, 0, 0, nil, nil, fmt.Errorf("decode parallelism `%s`: %w", part, err)
				}

				seen |= 1 << 2
				parallelism = uint8(bigParallelism)
			}

			if seen != 7 {
				return 0, 0, 0, nil, nil, fmt.Errorf("decode params `%s`: %w", part, err)
			}

		case 4:
			salt, err = strictBase64Decoder.DecodeString(part)
			if err != nil {
				return 0, 0, 0, nil, nil, fmt.Errorf("decode salt: %w", err)
			}
		}

		start = index + 1
		partCount += 1
	}

	if partCount != 5 {
		return 0, 0, 0, nil, nil, ErrInvalidEncodedHash
	}

	hash, err = strictBase64Decoder.DecodeString(encoded[start:])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	return memory, iterations, parallelism, salt, hash, nil
}
