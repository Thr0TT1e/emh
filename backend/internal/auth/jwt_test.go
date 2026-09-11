package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

// TestGenerateToken_Success проверяет успешную генерацию JWT.
func TestGenerateToken_Success(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	token, err := GenerateToken(secret, "admin", "admin", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken returned empty token")
	}
}

// TestValidateToken_Success проверяет валидацию валидного токена.
func TestValidateToken_Success(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	username := "admin"
	role := "admin"

	token, err := GenerateToken(secret, username, role, 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Username != username {
		t.Errorf("Username = %q, want %q", claims.Username, username)
	}
	if claims.Role != role {
		t.Errorf("Role = %q, want %q", claims.Role, role)
	}
}

// TestValidateToken_Expired проверяет ошибку при истёкшем токене.
func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"

	// Генерируем токен с отрицательным expiry (уже истёк).
	token, err := GenerateToken(secret, "admin", "admin", -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Fatal("ValidateToken should fail for expired token")
	}
}

// TestValidateToken_WrongSecret проверяет ошибку при неверном секрете.
func TestValidateToken_WrongSecret(t *testing.T) {
	secret1 := "test-secret-at-least-32-characters-long-1"
	secret2 := "test-secret-at-least-32-characters-long-2"

	token, err := GenerateToken(secret1, "admin", "admin", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, secret2)
	if err == nil {
		t.Fatal("ValidateToken should fail for wrong secret")
	}
}

// TestValidateToken_InvalidFormat проверяет ошибку при невалидном формате.
func TestValidateToken_InvalidFormat(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"

	_, err := ValidateToken("not-a-valid-jwt", secret)
	if err == nil {
		t.Fatal("ValidateToken should fail for invalid format")
	}
}

// TestValidateToken_TamperedToken проверяет ошибку при подделке токена.
func TestValidateToken_TamperedToken(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"

	token, err := GenerateToken(secret, "admin", "admin", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Подделываем токен: изменяем payload (username в claims).
	// JWT = header.payload.signature, разделённые точками.
	// Изменяем payload — подпись не будет совпадать.
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	// Декодируем payload из base64url
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}

	// Изменяем username в payload
	tamperedPayload := strings.Replace(string(payload), "admin", "attacker", 1)

	// Кодируем обратно в base64url
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(tamperedPayload))

	// Собираем поддельный токен (подпись остаётся оригинальной — не совпадёт)
	tampered := strings.Join(parts, ".")

	_, err = ValidateToken(tampered, secret)
	if err == nil {
		t.Fatal("ValidateToken should fail for tampered token")
	}
}
