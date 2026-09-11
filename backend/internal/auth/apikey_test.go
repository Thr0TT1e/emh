package auth

import (
	"strings"
	"testing"
)

// TestGenerateAPIKey_Success проверяет успешную генерацию API-ключа.
func TestGenerateAPIKey_Success(t *testing.T) {
	keyID, secret, fullKey, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if keyID == "" {
		t.Error("keyID is empty")
	}
	if secret == "" {
		t.Error("secret is empty")
	}
	if fullKey == "" {
		t.Error("fullKey is empty")
	}

	// fullKey должен содержать префикс.
	if !strings.HasPrefix(fullKey, apiKeyPrefix) {
		t.Errorf("fullKey %q does not have prefix %q", fullKey, apiKeyPrefix)
	}

	// fullKey должен содержать keyID и secret.
	if !strings.Contains(fullKey, keyID) {
		t.Errorf("fullKey %q does not contain keyID %q", fullKey, keyID)
	}
}

// TestParseAPIKey_Success проверяет парсинг валидного ключа.
func TestParseAPIKey_Success(t *testing.T) {
	keyID, secret, fullKey, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	parsedKeyID, parsedSecret, ok := ParseAPIKey(fullKey)
	if !ok {
		t.Fatal("ParseAPIKey returned ok=false for valid key")
	}
	if parsedKeyID != keyID {
		t.Errorf("parsed keyID = %q, want %q", parsedKeyID, keyID)
	}
	if parsedSecret != secret {
		t.Errorf("parsed secret = %q, want %q", parsedSecret, secret)
	}
}

// TestParseAPIKey_InvalidPrefix проверяет ошибку при неверном префиксе.
func TestParseAPIKey_InvalidPrefix(t *testing.T) {
	_, _, ok := ParseAPIKey("invalid_prefix_key_secret")
	if ok {
		t.Error("ParseAPIKey should return ok=false for invalid prefix")
	}
}

// TestParseAPIKey_MalformedKey проверяет ошибку при некорректном формате.
func TestParseAPIKey_MalformedKey(t *testing.T) {
	// Ключ без разделителя.
	_, _, ok := ParseAPIKey(apiKeyPrefix + "noseparator")
	if ok {
		t.Error("ParseAPIKey should return ok=false for malformed key")
	}

	// Ключ с пустыми частями.
	_, _, ok = ParseAPIKey(apiKeyPrefix + "_")
	if ok {
		t.Error("ParseAPIKey should return ok=false for empty parts")
	}
}

// TestIsAPIKey проверяет определение API-ключа по префиксу.
func TestIsAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"valid api key", "emh_key123_secret456", true},
		{"jwt token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", false},
		{"empty string", "", false},
		{"partial prefix", "emh", false},
		{"wrong prefix", "api_key123_secret456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAPIKey(tt.token); got != tt.expected {
				t.Errorf("IsAPIKey(%q) = %v, want %v", tt.token, got, tt.expected)
			}
		})
	}
}

// TestGenerateAPIKey_Unique проверяет, что каждая генерация даёт уникальный ключ.
func TestGenerateAPIKey_Unique(t *testing.T) {
	keys := make(map[string]bool)

	for i := 0; i < 10; i++ {
		_, _, fullKey, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("GenerateAPIKey %d failed: %v", i, err)
		}
		if keys[fullKey] {
			t.Errorf("duplicate key generated: %q", fullKey)
		}
		keys[fullKey] = true
	}
}
