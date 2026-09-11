//go:build integration

package pg

import (
	"context"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

// TestLocationRepository_HasCyclicReference_Integration проверяет WITH RECURSIVE
// на реальной БД с цепочкой A → B → C.
func TestLocationRepository_HasCyclicReference_Integration(t *testing.T) {
	pool, cleanup := testutil.NewTestPool(t)
	defer cleanup()
	testutil.CleanupTable(t, pool, "hero_locations")
	testutil.CleanupTable(t, pool, "locations")

	repo := NewLocationRepository(pool)
	ctx := context.Background()

	// 1. Создаём иерархию A (корень) → B → C
	aID, err := repo.Create(ctx, domain.CreateLocationParams{
		Name: "Корневая",
		Type: domain.LocationTypeCountry,
	})
	if err != nil {
		t.Fatalf("create A: %v", err)
	}
	bID, err := repo.Create(ctx, domain.CreateLocationParams{
		Name:     "Средняя",
		Type:     domain.LocationTypeRegion,
		ParentID: &aID,
	})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	cID, err := repo.Create(ctx, domain.CreateLocationParams{
		Name:     "Лист",
		Type:     domain.LocationTypeCity,
		ParentID: &bID,
	})
	if err != nil {
		t.Fatalf("create C: %v", err)
	}

	// 2. Проверка: обновление C.parent = A — валидно (A предок C, не потомок)
	hasCycle, err := repo.HasCyclicReference(ctx, cID, aID)
	if err != nil {
		t.Fatalf("check C.parent=A: %v", err)
	}
	if hasCycle {
		t.Error("expected no cycle when setting C.parent = A (A is ancestor of C)")
	}

	// 3. Проверка: обновление A.parent = C — ЦИКЛ (C потомок A)
	hasCycle, err = repo.HasCyclicReference(ctx, aID, cID)
	if err != nil {
		t.Fatalf("check A.parent=C: %v", err)
	}
	if !hasCycle {
		t.Error("expected cycle when setting A.parent = C (C is descendant of A)")
	}

	// 4. Проверка: обновление A.parent = B — ЦИКЛ (B потомок A)
	hasCycle, err = repo.HasCyclicReference(ctx, aID, bID)
	if err != nil {
		t.Fatalf("check A.parent=B: %v", err)
	}
	if !hasCycle {
		t.Error("expected cycle when setting A.parent = B")
	}

	// 5. Проверка: обновление B.parent = C — ЦИКЛ (C потомок B)
	hasCycle, err = repo.HasCyclicReference(ctx, bID, cID)
	if err != nil {
		t.Fatalf("check B.parent=C: %v", err)
	}
	if !hasCycle {
		t.Error("expected cycle when setting B.parent = C")
	}
}

// TestConflictRepository_HasCyclicReference_Integration аналогичен location,
// но для конфликтов.
func TestConflictRepository_HasCyclicReference_Integration(t *testing.T) {
	pool, cleanup := testutil.NewTestPool(t)
	defer cleanup()
	testutil.CleanupTable(t, pool, "hero_conflicts")
	testutil.CleanupTable(t, pool, "conflicts")

	repo := NewConflictRepository(pool)
	ctx := context.Background()

	// Создаём иерархию: ВОВ → Битва под Москвой → Операция "Тайфун"
	ww2ID, err := repo.Create(ctx, domain.CreateConflictParams{
		Name: "Великая Отечественная война",
		Type: domain.ConflictTypeGlobal,
	})
	if err != nil {
		t.Fatalf("create WW2: %v", err)
	}
	battleID, err := repo.Create(ctx, domain.CreateConflictParams{
		Name:             "Битва под Москвой",
		Type:             domain.ConflictTypeGlobal,
		ParentConflictID: &ww2ID,
	})
	if err != nil {
		t.Fatalf("create battle: %v", err)
	}
	operationID, err := repo.Create(ctx, domain.CreateConflictParams{
		Name:             "Операция Тайфун",
		Type:             domain.ConflictTypeGlobal,
		ParentConflictID: &battleID,
	})
	if err != nil {
		t.Fatalf("create operation: %v", err)
	}

	// Валидно: операция.parent = ВОВ (ВОВ предок операции)
	hasCycle, err := repo.HasCyclicReference(ctx, operationID, ww2ID)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if hasCycle {
		t.Error("expected no cycle when operation.parent = WW2")
	}

	// Цикл: ВОВ.parent = операция (операция потомок ВОВ)
	hasCycle, err = repo.HasCyclicReference(ctx, ww2ID, operationID)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !hasCycle {
		t.Error("expected cycle when WW2.parent = operation")
	}
}
