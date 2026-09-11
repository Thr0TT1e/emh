//go:build integration

// Package testutil предоставляет хелперы для интеграционных тестов репозиториев.
// Используется только при сборке с тегом `-tags=integration`.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер для database/sql
	"github.com/pressly/goose/v3"
)

// TestDBURL возвращает URL тестовой БД из env или падает с понятной ошибкой.
func TestDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("TEST_DATABASE_URL не задан. Запусти `just test-db-up` и повтори.")
	}
	return url
}

// NewTestPool создаёт пул подключений к тестовой БД и применяет все миграции.
// Возвращает пул и функцию очистки (закрытие пула).
// Использование:
//
//	pool, cleanup := testutil.NewTestPool(t)
//	defer cleanup()
func NewTestPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()
	url := TestDBURL(t)

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse test db config: %v", err)
	}
	cfg.MaxConns = 5
	cfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping test db: %v", err)
	}

	// Применяем миграции через database/sql (goose не работает с pgxpool).
	if err := runMigrations(); err != nil {
		pool.Close()
		t.Fatalf("apply migrations: %v", err)
	}

	cleanup := func() { pool.Close() }
	return pool, cleanup
}

// resolveMigrationsDir находит директорию с миграциями, обходя три сценария:
//  1. Env TEST_MIGRATIONS_DIR задан и существует → используем.
//  2. Авто-поиск от положения setup.go (работает из IDE и `go test`):
//     setup.go лежит в backend/internal/repository/pg/testutil/,
//     миграции — в backend/migrations/.
//  3. Fallback: ./migrations относительно CWD.
func resolveMigrationsDir() (string, error) {
	// 1. Env override (только если валиден).
	if envDir := os.Getenv("TEST_MIGRATIONS_DIR"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			return envDir, nil
		}
		// Env задан, но невалиден — продолжаем искать (не падаем).
		// Это позволяет работать из IDE без переопределения env.
	}

	// 2. Авто-поиск от файла setup.go через runtime.Caller.
	// runtime.Caller(0) даёт setup.go — файл, где объявлена функция.
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// setup.go лежит в backend/internal/repository/pg/testutil/setup.go
		// Поднимаемся на 4 уровня: testutil → pg → repository → internal → backend
		candidate := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations")
		candidate = filepath.Clean(candidate)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	// 3. Fallback: ./migrations относительно CWD.
	cwd, _ := os.Getwd()
	fallback := filepath.Join(cwd, "migrations")
	if info, err := os.Stat(fallback); err == nil && info.IsDir() {
		return fallback, nil
	}

	return "", fmt.Errorf("cannot locate migrations directory; set TEST_MIGRATIONS_DIR explicitly")
}

// runMigrations применяет все SQL-миграции к тестовой БД.
// goose требует *sql.DB, поэтому открываем отдельное соединение через pgx/v5/stdlib.
func runMigrations() error {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("TEST_DATABASE_URL is required")
	}

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return goose.UpContext(ctx, db, migrationsDir)
}

// CleanupTable удаляет все записи из таблицы. Используется между тестами.
// ВНИМАНИЕ: не сбрасывает sequence'ы — UUIDv7 генерируются в приложении.
func CleanupTable(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate %s: %v", table, err)
	}
}

// CleanupAllTables очищает все таблицы с пользовательскими данными.
// Используется в TestMain или setup/teardown тестового пакета.
func CleanupAllTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"contact_messages",
		"auth_audit_log",
		"refresh_tokens",
		"api_keys",
		"submissions",
		"hero_relations",
		"hero_sources",
		"hero_locations",
		"hero_conflicts",
		"hero_awards",
		"photos",
		"heroes",
		"awards",
		"conflicts",
		"locations",
	}
	for _, tbl := range tables {
		CleanupTable(t, pool, tbl)
	}
}
