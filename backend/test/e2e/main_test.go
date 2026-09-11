//go:build integration

package e2e

import (
	"os"
	"testing"
)

// TestMain выполняется один раз для всего пакета e2e.
// Используется для инициализации общей тестовой БД.
func TestMain(m *testing.M) {
	// Проверяем что TEST_DATABASE_URL задан
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		println("TEST_DATABASE_URL не задан. Запусти `just test-db-up` и повтори.")
		os.Exit(1)
	}

	// Запускаем тесты
	code := m.Run()

	// Cleanup будет выполнен в каждом тесте через defer
	os.Exit(code)
}

// cleanupTestDB очищает все таблицы между тестами.
// Вызывается в каждом тесте через defer для изоляции.
func cleanupTestDB(t *testing.T, pool interface{ Close() }) {
	t.Helper()
	// Cleanup выполняется через testutil.CleanupAllTables в setupTestServer
}
