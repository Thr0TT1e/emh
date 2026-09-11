//go:build integration

package pg

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

// newHeroRepo создаёт тестовый репозиторий героев с общим пулом.
func newHeroRepo(t *testing.T) (*heroRepository, func()) {
	t.Helper()
	pool, cleanupPool := testutil.NewTestPool(t)
	testutil.CleanupAllTables(t, pool)

	repo := NewHeroRepository(pool)
	cleanup := func() {
		testutil.CleanupAllTables(t, pool)
		cleanupPool()
	}

	return repo.(*heroRepository), cleanup
}

// ptrString возвращает указатель на строку.
func ptrString(s string) *string { return &s }

// ptrFlexibleDate возвращает указатель на FlexibleDate.
func ptrFlexibleDate(fd domain.FlexibleDate) *domain.FlexibleDate { return &fd }

// ptrStatus возвращает указатель на PublicationStatus.
func ptrStatus(s domain.PublicationStatus) *domain.PublicationStatus { return &s }

// exactDate создаёт точную дату.
func exactDate(y, m, d int) domain.FlexibleDate {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return domain.NewExactDate(t)
}

// newValidCreateParams возвращает валидные параметры создания героя.
func newValidCreateParams() domain.CreateHeroParams {
	return domain.CreateHeroParams{
		FirstName:        "Иван",
		LastName:         "Петров",
		Status:           domain.StatusDraft,
		BirthDate:        exactDate(1990, 5, 15),
		DeathDate:        domain.NewUnknownDate(),
		ServiceStartDate: domain.NewUnknownDate(),
	}
}

// --- Тесты Create ---

// TestHeroRepo_Create_Success проверяет базовое создание героя.
func TestHeroRepo_Create_Success(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
}

// TestHeroRepo_Create_WithAllFields проверяет создание со всеми полями.
func TestHeroRepo_Create_WithAllFields(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := domain.CreateHeroParams{
		FirstName:        "Иван",
		LastName:         "Петров",
		MiddleName:       "Сергеевич",
		ShortBio:         "Краткая биография",
		FullBio:          "Полная биография героя",
		Rank:             "Старший лейтенант",
		BirthDate:        exactDate(1990, 5, 15),
		DeathDate:        exactDate(2020, 8, 20),
		ServiceStartDate: exactDate(2010, 9, 1),
		Status:           domain.StatusPublished,
		Nickname:         "Тридцатый",
		Unit:             "в/ч 12345",
		Position:         "Командир взвода",
		ServiceBranch:    "ВДВ",
		CauseOfDeath:     "Погиб в бою",
		Memberships:      []string{"Союз десантников", "Ветераны Афгана"},
	}

	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create with all fields failed: %v", err)
	}

	// Проверяем, что все поля сохранились.
	hero, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if hero.FirstName != "Иван" {
		t.Errorf("FirstName = %q, want %q", hero.FirstName, "Иван")
	}
	if hero.MiddleName != "Сергеевич" {
		t.Errorf("MiddleName = %q, want %q", hero.MiddleName, "Сергеевич")
	}
	if hero.Rank != "Старший лейтенант" {
		t.Errorf("Rank = %q, want %q", hero.Rank, "Старший лейтенант")
	}
	if hero.Status != domain.StatusPublished {
		t.Errorf("Status = %v, want %v", hero.Status, domain.StatusPublished)
	}
	if len(hero.Memberships) != 2 {
		t.Errorf("Memberships len = %d, want 2", len(hero.Memberships))
	}
}

// TestHeroRepo_Create_FlexibleDate_AllPrecisions проверяет все типы гибких дат.
func TestHeroRepo_Create_FlexibleDate_AllPrecisions(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	tests := []struct {
		name      string
		birthDate domain.FlexibleDate
	}{
		{"exact", exactDate(1990, 5, 15)},
		{"month", domain.FlexibleDate{Precision: domain.PrecisionMonth, Anchor: ptrTime(time.Date(1990, 5, 1, 0, 0, 0, 0, time.UTC)), DisplayText: "Май 1990"}},
		{"year", domain.FlexibleDate{Precision: domain.PrecisionYear, Anchor: ptrTime(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)), DisplayText: "1990"}},
		{"season", domain.FlexibleDate{Precision: domain.PrecisionSeason, Anchor: ptrTime(time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC)), DisplayText: "Лето 1990"}},
		{"day_month", domain.FlexibleDate{Precision: domain.PrecisionDayMonth, DisplayText: "15 мая"}},
		{"unknown", domain.NewUnknownDate()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := newValidCreateParams()
			params.BirthDate = tt.birthDate

			id, err := repo.Create(context.Background(), params)
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}

			hero, err := repo.GetByID(context.Background(), id)
			if err != nil {
				t.Fatalf("GetByID failed: %v", err)
			}

			if hero.BirthDate.Precision != tt.birthDate.Precision {
				t.Errorf("precision = %v, want %v", hero.BirthDate.Precision, tt.birthDate.Precision)
			}
			if hero.BirthDate.DisplayText != tt.birthDate.DisplayText {
				t.Errorf("displayText = %q, want %q", hero.BirthDate.DisplayText, tt.birthDate.DisplayText)
			}
		})
	}
}

// ptrTime возвращает указатель на time.Time.
func ptrTime(t time.Time) *time.Time { return &t }

// --- Тесты GetByID ---

// TestHeroRepo_GetByID_Success проверяет получение существующего героя.
func TestHeroRepo_GetByID_Success(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	hero, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if hero.ID != id {
		t.Errorf("ID = %q, want %q", hero.ID, id)
	}
	if hero.FirstName != "Иван" {
		t.Errorf("FirstName = %q, want %q", hero.FirstName, "Иван")
	}
}

// TestHeroRepo_GetByID_NotFound проверяет ошибку при отсутствии героя.
func TestHeroRepo_GetByID_NotFound(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	_, err := repo.GetByID(context.Background(), testutil.NewUUID())
	if err == nil {
		t.Fatal("GetByID should fail for non-existent id")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Тесты Update ---

// TestHeroRepo_Update_PartialFieldMask проверяет partial update.
func TestHeroRepo_Update_PartialFieldMask(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Обновляем только first_name и rank.
	newFirstName := "Пётр"
	newRank := "Капитан"
	updateParams := domain.UpdateHeroParams{
		ID:        &id,
		FirstName: &newFirstName,
		Rank:      &newRank,
		FieldMask: []string{"first_name", "rank"},
	}

	updated, err := repo.Update(context.Background(), updateParams)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.FirstName != "Пётр" {
		t.Errorf("FirstName = %q, want %q", updated.FirstName, "Пётр")
	}
	if updated.Rank != "Капитан" {
		t.Errorf("Rank = %q, want %q", updated.Rank, "Капитан")
	}
	// LastName не должен измениться.
	if updated.LastName != "Петров" {
		t.Errorf("LastName changed unexpectedly: %q", updated.LastName)
	}
}

// TestHeroRepo_Update_FlexibleDates проверяет обновление гибких дат.
func TestHeroRepo_Update_FlexibleDates(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Обновляем birth_date на month precision.
	newBirth := domain.FlexibleDate{
		Precision:   domain.PrecisionMonth,
		Anchor:      ptrTime(time.Date(1995, 3, 1, 0, 0, 0, 0, time.UTC)),
		DisplayText: "Март 1995",
	}
	updateParams := domain.UpdateHeroParams{
		ID:        &id,
		BirthDate: &newBirth,
		FieldMask: []string{"birth_date"},
	}

	updated, err := repo.Update(context.Background(), updateParams)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.BirthDate.Precision != domain.PrecisionMonth {
		t.Errorf("precision = %v, want %v", updated.BirthDate.Precision, domain.PrecisionMonth)
	}
	if updated.BirthDate.DisplayText != "Март 1995" {
		t.Errorf("displayText = %q, want %q", updated.BirthDate.DisplayText, "Март 1995")
	}
}

// TestHeroRepo_Update_NotFound проверяет ошибку при обновлении несуществующего героя.
func TestHeroRepo_Update_NotFound(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	nonExistentID := testutil.NewUUID()
	updateParams := domain.UpdateHeroParams{
		ID:        &nonExistentID,
		FirstName: ptrString("Новое имя"),
		FieldMask: []string{"first_name"},
	}

	_, err := repo.Update(context.Background(), updateParams)
	if err == nil {
		t.Fatal("Update should fail for non-existent id")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Тесты Delete ---

// TestHeroRepo_Delete_Soft проверяет soft delete (архивация).
func TestHeroRepo_Delete_Soft(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	params.Status = domain.StatusPublished
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Soft delete.
	err = repo.Delete(context.Background(), id, false)
	if err != nil {
		t.Fatalf("Delete (soft) failed: %v", err)
	}

	// Проверяем, что статус изменился на Archived.
	hero, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID after soft delete failed: %v", err)
	}
	if hero.Status != domain.StatusArchived {
		t.Errorf("Status = %v, want %v", hero.Status, domain.StatusArchived)
	}
}

// TestHeroRepo_Delete_Hard проверяет hard delete (физическое удаление).
func TestHeroRepo_Delete_Hard(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Hard delete.
	err = repo.Delete(context.Background(), id, true)
	if err != nil {
		t.Fatalf("Delete (hard) failed: %v", err)
	}

	// Проверяем, что герой удалён.
	_, err = repo.GetByID(context.Background(), id)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound after hard delete, got %v", err)
	}
}

// TestHeroRepo_Delete_NotFound проверяет ошибку при удалении несуществующего героя.
func TestHeroRepo_Delete_NotFound(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	err := repo.Delete(context.Background(), testutil.NewUUID(), false)
	if err == nil {
		t.Fatal("Delete should fail for non-existent id")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Тесты List с cursor-based pagination ---

// TestHeroRepo_List_FirstPage_WithTotalCount проверяет первую страницу с total_count.
func TestHeroRepo_List_FirstPage_WithTotalCount(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	// Создаём 5 героев.
	for i := 0; i < 5; i++ {
		params := newValidCreateParams()
		params.FirstName = "Герой" + string(rune('A'+i))
		_, err := repo.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond) // UUIDv7 монотонность
	}

	filter := domain.HeroFilter{Limit: 3, Cursor: ""}
	heroes, nextCursor, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 3 {
		t.Errorf("expected 3 heroes, got %d", len(heroes))
	}
	if nextCursor == "" {
		t.Error("expected non-empty nextCursor")
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
}

// TestHeroRepo_List_SecondPage_NoTotalCount проверяет, что total_count=0 на последующих страницах.
func TestHeroRepo_List_SecondPage_NoTotalCount(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	// Создаём 5 героев.
	for i := 0; i < 5; i++ {
		params := newValidCreateParams()
		_, err := repo.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	// Первая страница.
	filter := domain.HeroFilter{Limit: 2, Cursor: ""}
	_, nextCursor, _, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List (page 1) failed: %v", err)
	}

	// Вторая страница.
	filter.Cursor = nextCursor
	heroes, _, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List (page 2) failed: %v", err)
	}

	if len(heroes) != 2 {
		t.Errorf("expected 2 heroes on page 2, got %d", len(heroes))
	}
	if total != 0 {
		t.Errorf("total on page 2 = %d, want 0 (optimization)", total)
	}
}

// TestHeroRepo_List_CursorPagination_Order проверяет порядок UUIDv7 DESC.
func TestHeroRepo_List_CursorPagination_Order(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	// Создаём 3 героев с задержкой для монотонности UUIDv7.
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		params := newValidCreateParams()
		id, err := repo.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		ids[i] = id
		time.Sleep(2 * time.Millisecond)
	}

	// List должен вернуть героев в обратном порядке (DESC).
	filter := domain.HeroFilter{Limit: 10}
	heroes, _, _, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 3 {
		t.Fatalf("expected 3 heroes, got %d", len(heroes))
	}

	// Последний созданный должен быть первым в списке.
	if heroes[0].ID != ids[2] {
		t.Errorf("heroes[0].ID = %q, want %q (last created)", heroes[0].ID, ids[2])
	}
	if heroes[2].ID != ids[0] {
		t.Errorf("heroes[2].ID = %q, want %q (first created)", heroes[2].ID, ids[0])
	}
}

// TestHeroRepo_List_EmptyResult проверяет пустой результат.
func TestHeroRepo_List_EmptyResult(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	filter := domain.HeroFilter{Limit: 10}
	heroes, nextCursor, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 0 {
		t.Errorf("expected 0 heroes, got %d", len(heroes))
	}
	if nextCursor != "" {
		t.Errorf("expected empty nextCursor, got %q", nextCursor)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

// --- Тесты фильтров ---

// TestHeroRepo_List_SearchQuery_Trigram проверяет trigram ILIKE поиск.
func TestHeroRepo_List_SearchQuery_Trigram(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	// Создаём героев с разными именами.
	names := []string{"Иван Петров", "Пётр Иванов", "Сергей Сидоров"}
	for _, name := range names {
		parts := splitName(name)
		params := newValidCreateParams()
		params.FirstName = parts[0]
		params.LastName = parts[1]
		_, err := repo.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("Create %q failed: %v", name, err)
		}
	}

	// Ищем "Иван" — должно найти 2 записи (Иван Петров и Пётр Иванов).
	filter := domain.HeroFilter{SearchQuery: "Иван", Limit: 10}
	heroes, _, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 2 {
		t.Errorf("total = %d, want 2 (trigram search)", total)
	}
	if len(heroes) != 2 {
		t.Errorf("expected 2 heroes, got %d", len(heroes))
	}
}

// splitName разделяет "Имя Фамилия" на части.
func splitName(full string) [2]string {
	parts := [2]string{}
	for i, r := range full {
		if r == ' ' {
			parts[0] = full[:i]
			parts[1] = full[i+1:]
			return parts
		}
	}
	parts[0] = full
	return parts
}

// TestHeroRepo_List_FilterByDateRange проверяет фильтрацию по диапазону death_date.
func TestHeroRepo_List_FilterByDateRange(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	// Создаём героев с разными датами смерти.
	deathDates := []time.Time{
		time.Date(2019, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 8, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 5, 10, 0, 0, 0, 0, time.UTC),
	}

	for _, d := range deathDates {
		params := newValidCreateParams()
		params.DeathDate = domain.NewExactDate(d)
		_, err := repo.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Фильтр: 2020-01-01 до 2020-12-31.
	dateFrom := "2020-01-01"
	dateTo := "2020-12-31"
	filter := domain.HeroFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    10,
	}

	heroes, _, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 1 {
		t.Errorf("total = %d, want 1 (only 2020 death)", total)
	}
	if len(heroes) != 1 {
		t.Errorf("expected 1 hero, got %d", len(heroes))
	}
}

// --- Тесты денормализации ---

// TestHeroRepo_List_DenormalizeMainPhotoURL проверяет COALESCE для главного фото.
func TestHeroRepo_List_DenormalizeMainPhotoURL(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Создаём героя.
	params := newValidCreateParams()
	heroID, err := repo.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Добавляем 2 фото: первое не главное, второе главное.
	_, err = repo.pool.Exec(ctx, `
		INSERT INTO photos (hero_id, url, sort_order, is_main)
		VALUES ($1, 'https://s3.example.com/first.jpg', 0, false)
	`, heroID)
	if err != nil {
		t.Fatalf("Insert first photo failed: %v", err)
	}

	_, err = repo.pool.Exec(ctx, `
		INSERT INTO photos (hero_id, url, sort_order, is_main)
		VALUES ($1, 'https://s3.example.com/main.jpg', 1, true)
	`, heroID)
	if err != nil {
		t.Fatalf("Insert main photo failed: %v", err)
	}

	// List должен вернуть main_photo_url = 'https://s3.example.com/main.jpg'.
	filter := domain.HeroFilter{Limit: 10}
	heroes, _, _, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 1 {
		t.Fatalf("expected 1 hero, got %d", len(heroes))
	}

	if heroes[0].MainPhotoURL == nil {
		t.Fatal("MainPhotoURL is nil, want non-nil")
	}
	if *heroes[0].MainPhotoURL != "https://s3.example.com/main.jpg" {
		t.Errorf("MainPhotoURL = %q, want %q", *heroes[0].MainPhotoURL, "https://s3.example.com/main.jpg")
	}
}

// TestHeroRepo_List_DenormalizeAwardNames проверяет ARRAY_AGG для наград.
func TestHeroRepo_List_DenormalizeAwardNames(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Создаём героя.
	params := newValidCreateParams()
	heroID, err := repo.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Создаём 2 награды.
	var award1ID, award2ID string
	err = repo.pool.QueryRow(ctx, `
		INSERT INTO awards (name, description, sort_order)
		VALUES ('Герой России', 'Высшая награда', 1)
		RETURNING id
	`).Scan(&award1ID)
	if err != nil {
		t.Fatalf("Insert award 1 failed: %v", err)
	}

	err = repo.pool.QueryRow(ctx, `
		INSERT INTO awards (name, description, sort_order)
		VALUES ('Орден Мужества', 'За отвагу', 2)
		RETURNING id
	`).Scan(&award2ID)
	if err != nil {
		t.Fatalf("Insert award 2 failed: %v", err)
	}

	// Привязываем награды к герою.
	_, err = repo.pool.Exec(ctx, `
		INSERT INTO hero_awards (hero_id, award_id) VALUES ($1, $2)
	`, heroID, award1ID)
	if err != nil {
		t.Fatalf("Insert hero_award 1 failed: %v", err)
	}

	_, err = repo.pool.Exec(ctx, `
		INSERT INTO hero_awards (hero_id, award_id) VALUES ($1, $2)
	`, heroID, award2ID)
	if err != nil {
		t.Fatalf("Insert hero_award 2 failed: %v", err)
	}

	// List должен вернуть award_names = ['Герой России', 'Орден Мужества'].
	filter := domain.HeroFilter{Limit: 10}
	heroes, _, _, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 1 {
		t.Fatalf("expected 1 hero, got %d", len(heroes))
	}

	if len(heroes[0].AwardNames) != 2 {
		t.Fatalf("AwardNames len = %d, want 2", len(heroes[0].AwardNames))
	}

	// Проверяем порядок (ORDER BY a.sort_order).
	if heroes[0].AwardNames[0] != "Герой России" {
		t.Errorf("AwardNames[0] = %q, want %q", heroes[0].AwardNames[0], "Герой России")
	}
	if heroes[0].AwardNames[1] != "Орден Мужества" {
		t.Errorf("AwardNames[1] = %q, want %q", heroes[0].AwardNames[1], "Орден Мужества")
	}
}

// TestHeroRepo_List_NoPhotos_NoAwards проверяет денормализацию при отсутствии связей.
func TestHeroRepo_List_NoPhotos_NoAwards(t *testing.T) {
	repo, cleanup := newHeroRepo(t)
	defer cleanup()

	params := newValidCreateParams()
	_, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	filter := domain.HeroFilter{Limit: 10}
	heroes, _, _, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(heroes) != 1 {
		t.Fatalf("expected 1 hero, got %d", len(heroes))
	}

	// MainPhotoURL должен быть nil (нет фото).
	if heroes[0].MainPhotoURL != nil {
		t.Errorf("MainPhotoURL = %q, want nil", *heroes[0].MainPhotoURL)
	}

	// AwardNames должен быть nil или пустой (нет наград).
	if heroes[0].AwardNames != nil && len(heroes[0].AwardNames) > 0 {
		t.Errorf("AwardNames = %v, want nil or empty", heroes[0].AwardNames)
	}
}
