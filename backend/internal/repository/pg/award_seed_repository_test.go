//go:build integration

package pg

import (
	"context"
	"errors"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

func newAwardSeedRepo(t *testing.T) (*awardRepository, func()) {
	t.Helper()
	pool, cleanupPool := testutil.NewTestPool(t)
	testutil.CleanupTable(t, pool, "awards")

	repo := NewAwardRepository(pool).(*awardRepository)
	return repo, func() {
		testutil.CleanupTable(t, pool, "awards")
		cleanupPool()
	}
}

func seedParams(name string, tp domain.AwardType, j domain.AwardJurisdiction) domain.CreateAwardParams {
	return domain.CreateAwardParams{
		Name:           name,
		Description:    "тест",
		SortOrder:      7,
		ImageURL:       "https://cdn/awards/images/x.webp",
		RibbonImageURL: "https://cdn/awards/ribbons/x.webp",
		Type:           tp,
		Jurisdiction:   j,
		WornWithoutBar: true,
		IsJubilee:      false,
	}
}

func TestAwardSeedKey_ExistsByKey(t *testing.T) {
	repo, cleanup := newAwardSeedRepo(t)
	defer cleanup()
	ctx := context.Background()

	name := "Орден «Тест»"
	exists, err := repo.ExistsByKey(ctx, name, domain.AwardTypeOrder, domain.AwardJurisdictionUSSR)
	if err != nil || exists {
		t.Fatalf("fresh key should not exist: exists=%v err=%v", exists, err)
	}

	if _, err := repo.Create(ctx, seedParams(name, domain.AwardTypeOrder, domain.AwardJurisdictionUSSR)); err != nil {
		t.Fatal(err)
	}
	// Совпавший ключ.
	if exists, err = repo.ExistsByKey(ctx, name, domain.AwardTypeOrder, domain.AwardJurisdictionUSSR); err != nil || !exists {
		t.Fatalf("key should exist: exists=%v err=%v", exists, err)
	}
	// Тот же name, другой type — не ключ.
	if exists, err = repo.ExistsByKey(ctx, name, domain.AwardTypeMedal, domain.AwardJurisdictionUSSR); err != nil || exists {
		t.Fatalf("type is part of key: exists=%v err=%v", exists, err)
	}
	// Тот же name/type, другая юрисдикция — не ключ.
	if exists, err = repo.ExistsByKey(ctx, name, domain.AwardTypeOrder, domain.AwardJurisdictionRussianFederation); err != nil || exists {
		t.Fatalf("jurisdiction is part of key: exists=%v err=%v", exists, err)
	}
}

func TestAwardSeedKey_InsertIfAbsent(t *testing.T) {
	repo, cleanup := newAwardSeedRepo(t)
	defer cleanup()
	ctx := context.Background()

	p := seedParams("Медаль «Идемпотентность»", domain.AwardTypeMedal, domain.AwardJurisdictionRussianFederation)

	inserted, err := repo.InsertIfAbsent(ctx, p)
	if err != nil || !inserted {
		t.Fatalf("first insert: inserted=%v err=%v", inserted, err)
	}
	// Повтор — пропуск без ошибки (идемпотентность сида, ON CONFLICT DO NOTHING).
	inserted, err = repo.InsertIfAbsent(ctx, p)
	if err != nil || inserted {
		t.Fatalf("second insert must be skipped: inserted=%v err=%v", inserted, err)
	}
	// Изменение ручки админки не перезаписывается повторным сидом.
	awards, _, _, err := repo.List(ctx, domain.AwardFilter{Limit: 10, SearchQuery: p.Name})
	if err != nil || len(awards) != 1 {
		t.Fatalf("list after seed: n=%d err=%v", len(awards), err)
	}
	got := awards[0]
	if got.SortOrder != p.SortOrder || got.ImageURL != p.ImageURL {
		t.Errorf("record mismatch: %+v", got)
	}
	if !got.WornWithoutBar {
		t.Error("worn_without_bar not stored")
	}
}

func TestAwardSeedKey_CreateMapsDuplicate(t *testing.T) {
	repo, cleanup := newAwardSeedRepo(t)
	defer cleanup()
	ctx := context.Background()

	p := seedParams("Дубль-щит", domain.AwardTypeBadge, domain.AwardJurisdictionDepartmental)
	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	_, err := repo.Create(ctx, p)
	if !errors.Is(err, domain.ErrAwardDuplicate) {
		t.Fatalf("expected ErrAwardDuplicate, got %v", err)
	}
}

func TestAwardSeedKey_UpdateMapsDuplicate(t *testing.T) {
	repo, cleanup := newAwardSeedRepo(t)
	defer cleanup()
	ctx := context.Background()

	a, err := repo.Create(ctx, seedParams("Первая", domain.AwardTypeOrder, domain.AwardJurisdictionUSSR))
	if err != nil {
		t.Fatal(err)
	}
	b, err := repo.Create(ctx, seedParams("Вторая", domain.AwardTypeOrder, domain.AwardJurisdictionUSSR))
	if err != nil {
		t.Fatal(err)
	}

	// Переименование «Вторая» → «Первая» при том же type/jurisdiction — конфликт ключа.
	_, err = repo.Update(ctx, domain.UpdateAwardParams{
		ID: b, Name: "Первая", FieldMask: []string{"name"},
	})
	if !errors.Is(err, domain.ErrAwardDuplicate) {
		t.Fatalf("expected ErrAwardDuplicate on update, got %v", err)
	}
	_ = a
}
