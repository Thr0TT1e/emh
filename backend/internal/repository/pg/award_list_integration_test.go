//go:build integration

package pg

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

// Инвариант keyset-пагинации (regression на старый курсор id > cursor):
// обход всех страниц при pageSize < N возвращает РОВНО N уникальных записей
// в порядке sort_order, name, id — без потерь и повторов, даже когда id
// упорядочены обратно sort_order.
func TestAwardList_KeysetFullTraversal(t *testing.T) {
	pool, cleanupPool := testutil.NewTestPool(t)
	defer cleanupPool()
	testutil.CleanupTable(t, pool, "awards")

	ctx := context.Background()
	repo := NewAwardRepository(pool)

	// 7 наград: sort_order убывает вместе с возрастанием id (uuidv7 монотонный),
	// т.е. порядок сортировки — строго обратный порядку id. Старый курсор
	// "id > X" на такой выборке терял бы почти всё со 2-й страницы.
	const total = 7
	for i := 0; i < total; i++ {
		p := domain.CreateAwardParams{
			Name:         fmt.Sprintf("Тест Keyset %d", i),
			Type:         domain.AwardTypeMedal,
			Jurisdiction: domain.AwardJurisdictionUSSR,
			SortOrder:    total - i,
		}
		if _, err := repo.Create(ctx, p); err != nil {
			t.Fatal(err)
		}
	}

	seen := make(map[string]int, total)
	order := make([]string, 0, total)
	cursor := ""
	pages := 0
	for {
		awards, next, _, err := repo.List(ctx, domain.AwardFilter{Limit: 2, Cursor: cursor})
		if err != nil {
			t.Fatalf("page %d: %v", pages, err)
		}
		pages++
		for _, a := range awards {
			seen[a.ID]++
			order = append(order, a.Name)
		}
		if next == "" {
			break
		}
		if pages > 10 {
			t.Fatal("cursor pagination does not terminate")
		}
		cursor = next
	}

	if len(order) != total {
		t.Fatalf("traversal saw %d records, want %d (lost: %v)", len(order), total, order)
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("record %s repeated %d times", id, n)
		}
	}
	// Порядок: sort_order ASC → «Тест Keyset 6» (sort_order=1) первым.
	if order[0] != "Тест Keyset 6" || order[total-1] != "Тест Keyset 0" {
		t.Fatalf("wrong order: %v", order)
	}
}

func TestAwardList_InvalidCursorMapsError(t *testing.T) {
	pool, cleanupPool := testutil.NewTestPool(t)
	defer cleanupPool()
	repo := NewAwardRepository(pool)

	_, _, _, err := repo.List(context.Background(), domain.AwardFilter{Limit: 10, Cursor: "bogus//cursor"})
	if !errors.Is(err, domain.ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

// Regression: строки с NULL image_url/description (колонки nullable с миграции
// 00002; сырой SQL/импорт без полей) не должны ронять весь List в 500
// («cannot scan NULL into *string» до COALESCE в awardColumns).
func TestAwardList_ScansNullColumns(t *testing.T) {
	pool, cleanupPool := testutil.NewTestPool(t)
	defer cleanupPool()
	testutil.CleanupTable(t, pool, "awards")
	ctx := context.Background()

	if _, err := pool.Exec(ctx,
		`INSERT INTO awards (name, description, image_url, sort_order) VALUES ('NULL-тест', NULL, NULL, 5)`); err != nil {
		t.Fatal(err)
	}

	awards, _, _, err := NewAwardRepository(pool).List(ctx, domain.AwardFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List with NULL columns: %v", err)
	}
	if len(awards) != 1 || awards[0].ImageURL != "" || awards[0].Description != "" {
		t.Fatalf("unexpected scan result: %+v", awards)
	}
}
