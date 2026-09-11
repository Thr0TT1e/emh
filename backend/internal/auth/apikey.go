package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// apiKeyPrefix позволяет отличить API-ключ от JWT по префиксу.
const apiKeyPrefix = "emh_"

// GenerateAPIKey создаёт новый API-ключ.
// Возвращает keyID (публичный идентификатор), secret (секретная часть)
// и fullKey (полный ключ, показывается пользователю только один раз).
func GenerateAPIKey() (keyID, secret, fullKey string, err error) {
	keyID = uuid.NewString()

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", "", fmt.Errorf("generate secret: %w", err)
	}
	secret = base64.RawURLEncoding.EncodeToString(secretBytes)

	fullKey = apiKeyPrefix + keyID + "_" + secret
	return keyID, secret, fullKey, nil
}

// ParseAPIKey разбирает полный ключ на keyID и secret.
// UUID не содержит подчёркиваний, поэтому первое "_" после префикса — разделитель.
func ParseAPIKey(fullKey string) (keyID, secret string, ok bool) {
	if !strings.HasPrefix(fullKey, apiKeyPrefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(fullKey, apiKeyPrefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// IsAPIKey проверяет, является ли токен API-ключом (по префиксу).
func IsAPIKey(token string) bool {
	return strings.HasPrefix(token, apiKeyPrefix)
}
