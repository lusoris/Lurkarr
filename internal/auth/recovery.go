package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const recoveryCodeCount = 10

// GenerateRecoveryCodes creates a set of random recovery codes and their bcrypt hashes.
// Returns plaintext codes (to show the user once) and hashed codes (to store in DB).
func GenerateRecoveryCodes() (plain, hashed []string, err error) {
	plain = make([]string, recoveryCodeCount)
	hashed = make([]string, recoveryCodeCount)

	for i := range recoveryCodeCount {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return nil, nil, fmt.Errorf("generate recovery code: %w", err)
		}
		code := hex.EncodeToString(b) // 8-char hex string
		plain[i] = fmt.Sprintf("%s-%s", code[:4], code[4:])

		hash, err := bcrypt.GenerateFromPassword([]byte(plain[i]), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, fmt.Errorf("hash recovery code: %w", err)
		}
		hashed[i] = string(hash)
	}
	return plain, hashed, nil
}

// ValidateRecoveryCode checks a plaintext code against the stored hashed codes.
// If valid, it returns the index of the matched code so the caller can remove it.
// Returns -1 if no code matches.
func ValidateRecoveryCode(code string, hashedCodes []string) int {
	for i, h := range hashedCodes {
		if bcrypt.CompareHashAndPassword([]byte(h), []byte(code)) == nil {
			return i
		}
	}
	return -1
}
