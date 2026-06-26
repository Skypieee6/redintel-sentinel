package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateOpaqueToken returns a cryptographically random, URL-safe token.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the hex-encoded SHA-256 digest of a token. Opaque tokens
// (refresh, password-reset, invitation) are only ever stored as their hash.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

const apiKeyPrefixLen = 8

// GenerateAPIKey returns a new API key in the form "rin_<random>" along with a
// short non-secret prefix used to identify the key, and the SHA-256 hash to
// persist. The plaintext key is shown to the caller exactly once.
func GenerateAPIKey() (key, prefix, hash string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	secret := hex.EncodeToString(b)
	key = fmt.Sprintf("rin_%s", secret)
	prefix = secret[:apiKeyPrefixLen]
	hash = HashToken(key)
	return key, prefix, hash, nil
}
