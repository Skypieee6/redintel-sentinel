package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password", 4)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "s3cret-password" {
		t.Fatal("password was not hashed")
	}
	if !VerifyPassword(hash, "s3cret-password") {
		t.Fatal("valid password failed verification")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("invalid password passed verification")
	}
}

func TestJWTGenerateAndParse(t *testing.T) {
	m := NewJWTManager("test-secret", time.Minute)
	tok, exp, err := m.GenerateAccessToken("user-123", "a@b.com")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !exp.After(time.Now()) {
		t.Fatal("expiry should be in the future")
	}
	claims, err := m.ParseAccessToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != "user-123" || claims.Email != "a@b.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTRejectsTamperedToken(t *testing.T) {
	m := NewJWTManager("secret-a", time.Minute)
	tok, _, _ := m.GenerateAccessToken("u", "e")
	other := NewJWTManager("secret-b", time.Minute)
	if _, err := other.ParseAccessToken(tok); err == nil {
		t.Fatal("token signed with a different secret should be rejected")
	}
	if _, err := m.ParseAccessToken("not-a-token"); err == nil {
		t.Fatal("malformed token should be rejected")
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	m := NewJWTManager("secret", -time.Minute)
	tok, _, _ := m.GenerateAccessToken("u", "e")
	if _, err := m.ParseAccessToken(tok); err == nil {
		t.Fatal("expired token should be rejected")
	}
}

func TestOpaqueTokenAndHash(t *testing.T) {
	a, err := GenerateOpaqueToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := GenerateOpaqueToken()
	if a == b {
		t.Fatal("opaque tokens should be unique")
	}
	if HashToken(a) != HashToken(a) {
		t.Fatal("hashing should be deterministic")
	}
	if HashToken(a) == a {
		t.Fatal("hash should differ from raw token")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(prefix) != apiKeyPrefixLen {
		t.Fatalf("prefix length = %d, want %d", len(prefix), apiKeyPrefixLen)
	}
	if hash != HashToken(key) {
		t.Fatal("hash should be HashToken(key)")
	}
	if key[:4] != "rin_" {
		t.Fatalf("key should start with rin_, got %q", key[:4])
	}
}
