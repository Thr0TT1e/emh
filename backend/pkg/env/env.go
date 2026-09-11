// Package env содержит утилиты для работы с переменными окружения.
package env

import (
	"os"
	"strconv"
	"strings"
)

// Get возвращает значение переменной окружения key.
// Если переменная не задана или пуста, возвращается значение по умолчанию def.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// GetRequired возвращает значение переменной окружения key.
// Вторым аргументом возвращает false, если переменная не задана.
func GetRequired(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	return v, ok && v != ""
}

// GetBool возвращает bool из env или значение по умолчанию.
// Истина: "true", "1" (без учёта регистра).
func GetBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return strings.ToLower(v) == "true" || v == "1"
}

// GetInt возвращает int из env или значение по умолчанию.
func GetInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
