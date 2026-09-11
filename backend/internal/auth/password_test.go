package auth

import (
	"testing"
)

// TestHashPassword_Success проверяет успешное хеширование пароля.
func TestHashPassword_Success(t *testing.T) {
	password := "my-secret-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}
	if hash == password {
		t.Error("HashPassword returned plaintext password")
	}
}

// TestCheckPassword_Correct проверяет успешную проверку правильного пароля.
func TestCheckPassword_Correct(t *testing.T) {
	password := "my-secret-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Error("CheckPassword returned false for correct password")
	}
}

// TestCheckPassword_Wrong проверяет, что неверный пароль не проходит проверку.
func TestCheckPassword_Wrong(t *testing.T) {
	password := "my-secret-password"
	wrongPassword := "wrong-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if CheckPassword(hash, wrongPassword) {
		t.Error("CheckPassword returned true for wrong password")
	}
}

// TestHashPassword_Unique проверяет, что каждый хеш уникален (из-за соли).
func TestHashPassword_Unique(t *testing.T) {
	password := "my-secret-password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword 1 failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword 2 failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword returned identical hashes for same password (salt not working)")
	}

	// Но оба хеша должны валидировать тот же пароль.
	if !CheckPassword(hash1, password) {
		t.Error("hash1 does not validate password")
	}
	if !CheckPassword(hash2, password) {
		t.Error("hash2 does not validate password")
	}
}
