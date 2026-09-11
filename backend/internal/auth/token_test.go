package auth

import (
	"testing"
)

// TestGenerateRefreshToken_Success проверяет успешную генерацию refresh-токена.
func TestGenerateRefreshToken_Success(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken returned empty token")
	}

	// base64url(32 bytes) = 43 символа.
	if len(token) != 43 {
		t.Errorf("token length = %d, want 43", len(token))
	}
}

// TestGenerateRefreshToken_Unique проверяет уникальность токенов.
func TestGenerateRefreshToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)

	for i := 0; i < 10; i++ {
		token, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken %d failed: %v", i, err)
		}
		if tokens[token] {
			t.Errorf("duplicate token generated: %q", token)
		}
		tokens[token] = true
	}
}

// TestHashRefreshToken_Deterministic проверяет детерминированность хеша.
func TestHashRefreshToken_Deterministic(t *testing.T) {
	token := "test-refresh-token-12345"

	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)

	if hash1 != hash2 {
		t.Errorf("HashRefreshToken is not deterministic: %q != %q", hash1, hash2)
	}
}

// TestHashRefreshToken_DifferentTokens проверяет, что разные токены дают разные хеши.
func TestHashRefreshToken_DifferentTokens(t *testing.T) {
	token1 := "token-1"
	token2 := "token-2"

	hash1 := HashRefreshToken(token1)
	hash2 := HashRefreshToken(token2)

	if hash1 == hash2 {
		t.Error("different tokens produced same hash")
	}
}

// TestHashRefreshToken_SHA256Length проверяет длину SHA-256 хеша.
func TestHashRefreshToken_SHA256Length(t *testing.T) {
	token := "test-token"
	hash := HashRefreshToken(token)

	// SHA-256 в hex = 64 символа.
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}
}

// TestGenerateAndHashRefreshToken проверяет полный цикл: генерация → хеш.
func TestGenerateAndHashRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	hash := HashRefreshToken(token)

	// Хеш не должен совпадать с токеном.
	if hash == token {
		t.Error("hash should not equal plaintext token")
	}

	// Хеш должен быть детерминированным.
	hash2 := HashRefreshToken(token)
	if hash != hash2 {
		t.Error("hash is not deterministic")
	}
}
