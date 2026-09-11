//go:build integration

package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// FakeHero создаёт тестового героя в БД и возвращает его ID.
// Для тестов репозитория героев можно переопределить значения через opts.
func FakeHero(t *testing.T, pool *pgxpool.Pool, opts ...func(*domain.Hero)) string {
	t.Helper()

	h := &domain.Hero{
		FirstName: "Иван",
		LastName:  "Тестовый",
		Status:    domain.StatusDraft,
	}
	for _, opt := range opts {
		opt(h)
	}

	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO heroes (
			first_name, last_name, middle_name, short_bio, full_bio, rank,
			birth_date, birth_date_precision, birth_date_display,
			death_date, death_date_precision, death_date_display,
			service_start_date, service_start_date_precision, service_start_date_display,
			status, nickname, unit, position, service_branch, cause_of_death, memberships
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			'[]'::jsonb
		) RETURNING id`,
		h.FirstName, h.LastName, h.MiddleName, h.ShortBio, h.FullBio, h.Rank,
		h.BirthDate.Anchor, int(h.BirthDate.Precision), nullableString(h.BirthDate.DisplayText),
		h.DeathDate.Anchor, int(h.DeathDate.Precision), nullableString(h.DeathDate.DisplayText),
		h.ServiceStartDate.Anchor, int(h.ServiceStartDate.Precision), nullableString(h.ServiceStartDate.DisplayText),
		int(h.Status), h.Nickname, h.Unit, h.Position, h.ServiceBranch, h.CauseOfDeath,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert fake hero: %v", err)
	}
	return id
}

// FakeAward создаёт тестовую награду в БД и возвращает её ID.
func FakeAward(t *testing.T, pool *pgxpool.Pool, name string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO awards (name, description, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id`,
		name, "Тестовая награда "+name, 0,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert fake award: %v", err)
	}
	return id
}

// FakeConflict создаёт тестовый конфликт в БД и возвращает его ID.
func FakeConflict(t *testing.T, pool *pgxpool.Pool, name string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO conflicts (name, description, type)
		VALUES ($1, $2, 0)
		RETURNING id`,
		name, "Тестовый конфликт "+name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert fake conflict: %v", err)
	}
	return id
}

// FakeLocation создаёт тестовую локацию в БД и возвращает её ID.
func FakeLocation(t *testing.T, pool *pgxpool.Pool, name string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO locations (name, type)
		VALUES ($1, 0)
		RETURNING id`,
		name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert fake location: %v", err)
	}
	return id
}

// NewUUID генерирует новый UUIDv4 (для тестов, не для production).
// В production используется UUIDv7 из БД.
func NewUUID() string {
	return uuid.New().String()
}

// RandomEmail генерирует уникальный email для тестов.
func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.ru", NewUUID()[:8])
}

// nullableString возвращает nil для пустой строки.
// Дублируется из pg-пакета, чтобы testutil не зависел от него.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
