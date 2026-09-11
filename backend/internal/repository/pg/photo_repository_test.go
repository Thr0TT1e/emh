//go:build integration

package pg

import (
	"context"
	"errors"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

// newPhotoRepo создаёт тестовый репозиторий фотографий с общим пулом.
// Cleanup выполняется в t.Cleanup() автоматически.
func newPhotoRepo(t *testing.T) (*photoRepository, func()) {
	t.Helper()
	pool, cleanupPool := testutil.NewTestPool(t)

	// Очищаем таблицы перед каждым тестом (порядок важен из-за FK).
	testutil.CleanupTable(t, pool, "photos")
	testutil.CleanupTable(t, pool, "heroes")

	repo := NewPhotoRepository(pool)
	cleanup := func() {
		testutil.CleanupTable(t, pool, "photos")
		testutil.CleanupTable(t, pool, "heroes")
		cleanupPool()
	}

	return repo.(*photoRepository), cleanup
}

// createTestHero создаёт тестового героя и возвращает его ID.
func createTestHero(t *testing.T, repo *photoRepository) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO heroes (first_name, last_name, status, birth_date, birth_date_precision, death_date_precision, service_start_date_precision)
		VALUES ('Иван', 'Тестовый', 1, '1990-01-01', 1, 7, 7)
		RETURNING id
	`).Scan(&id)
	if err != nil {
		t.Fatalf("create test hero: %v", err)
	}
	return id
}

// --- Тесты Add ---

// TestPhotoRepo_Add_Success проверяет базовое создание фотографии.
func TestPhotoRepo_Add_Success(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	params := domain.AddPhotoParams{
		HeroID:      heroID,
		URL:         "https://s3.example.com/photos/test.jpg",
		Description: "Тестовое фото",
		SortOrder:   0,
		IsMain:      false,
	}

	id, err := repo.Add(context.Background(), params)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == "" {
		t.Error("Add returned empty id")
	}
}

// TestPhotoRepo_Add_WithFaceBox проверяет сериализацию face_box в JSONB.
func TestPhotoRepo_Add_WithFaceBox(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	faceBox := &domain.FaceBox{
		X:      100,
		Y:      50,
		Width:  200,
		Height: 300,
	}
	params := domain.AddPhotoParams{
		HeroID:    heroID,
		URL:       "https://s3.example.com/photos/face.jpg",
		FaceBox:   faceBox,
		SortOrder: 0,
	}

	id, err := repo.Add(context.Background(), params)
	if err != nil {
		t.Fatalf("Add with face_box failed: %v", err)
	}

	// Проверяем, что face_box сохранился корректно.
	photos, err := repo.ListByHero(context.Background(), heroID)
	if err != nil {
		t.Fatalf("ListByHero failed: %v", err)
	}
	if len(photos) != 1 {
		t.Fatalf("expected 1 photo, got %d", len(photos))
	}
	if photos[0].ID != id {
		t.Errorf("expected id %s, got %s", id, photos[0].ID)
	}
	if photos[0].FaceBox == nil {
		t.Fatal("face_box is nil")
	}
	if photos[0].FaceBox.X != 100 || photos[0].FaceBox.Y != 50 {
		t.Errorf("face_box coordinates mismatch: %+v", photos[0].FaceBox)
	}
}

// TestPhotoRepo_ListByHero_Order проверяет сортировку по sort_order.
func TestPhotoRepo_ListByHero_Order(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Добавляем 3 фото в произвольном порядке sort_order.
	for i, so := range []int{2, 0, 1} {
		_, err := repo.Add(ctx, domain.AddPhotoParams{
			HeroID:    heroID,
			URL:       "https://s3.example.com/photos/" + string(rune('a'+i)) + ".jpg",
			SortOrder: so,
		})
		if err != nil {
			t.Fatalf("Add %d failed: %v", i, err)
		}
	}

	photos, err := repo.ListByHero(ctx, heroID)
	if err != nil {
		t.Fatalf("ListByHero failed: %v", err)
	}
	if len(photos) != 3 {
		t.Fatalf("expected 3 photos, got %d", len(photos))
	}

	// Проверяем порядок: sort_order должен быть 0, 1, 2.
	expectedOrder := []int{0, 1, 2}
	for i, p := range photos {
		if p.SortOrder != expectedOrder[i] {
			t.Errorf("photo[%d].SortOrder = %d, want %d", i, p.SortOrder, expectedOrder[i])
		}
	}
}

// --- Тесты BatchAdd ---

// TestPhotoRepo_BatchAdd_Transaction проверяет атомарность пакетного добавления.
func TestPhotoRepo_BatchAdd_Transaction(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	photos := []domain.AddPhotoParams{
		{URL: "https://s3.example.com/photos/batch1.jpg", SortOrder: 0},
		{URL: "https://s3.example.com/photos/batch2.jpg", SortOrder: 1},
		{URL: "https://s3.example.com/photos/batch3.jpg", SortOrder: 2},
	}

	added, err := repo.BatchAdd(context.Background(), heroID, photos)
	if err != nil {
		t.Fatalf("BatchAdd failed: %v", err)
	}
	if len(added) != 3 {
		t.Errorf("expected 3 added photos, got %d", len(added))
	}

	// Проверяем, что все фото сохранились.
	list, err := repo.ListByHero(context.Background(), heroID)
	if err != nil {
		t.Fatalf("ListByHero failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 photos in DB, got %d", len(list))
	}
}

// TestPhotoRepo_BatchAdd_AutoMain проверяет автоназначение главного фото.
func TestPhotoRepo_BatchAdd_AutoMain(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	photos := []domain.AddPhotoParams{
		{URL: "https://s3.example.com/photos/1.jpg", SortOrder: 0, IsMain: false},
		{URL: "https://s3.example.com/photos/2.jpg", SortOrder: 1, IsMain: false},
	}

	added, err := repo.BatchAdd(context.Background(), heroID, photos)
	if err != nil {
		t.Fatalf("BatchAdd failed: %v", err)
	}

	// Первое фото должно стать главным.
	if !added[0].IsMain {
		t.Error("first photo should be main")
	}
	if added[1].IsMain {
		t.Error("second photo should not be main")
	}
}

// TestPhotoRepo_BatchAdd_ExistingMain проверяет, что при наличии главного фото
// ни одно фото из пакета не становится главным.
func TestPhotoRepo_BatchAdd_ExistingMain(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём первое фото (станет главным).
	_, err := repo.Add(ctx, domain.AddPhotoParams{
		HeroID:    heroID,
		URL:       "https://s3.example.com/photos/existing.jpg",
		SortOrder: 0,
		IsMain:    true,
	})
	if err != nil {
		t.Fatalf("Add existing failed: %v", err)
	}

	// Добавляем пакет новых фото.
	photos := []domain.AddPhotoParams{
		{URL: "https://s3.example.com/photos/new1.jpg", SortOrder: 1, IsMain: true}, // явно true
		{URL: "https://s3.example.com/photos/new2.jpg", SortOrder: 2, IsMain: false},
	}

	added, err := repo.BatchAdd(ctx, heroID, photos)
	if err != nil {
		t.Fatalf("BatchAdd failed: %v", err)
	}

	// Ни одно новое фото не должно стать главным (уже есть главное).
	for i, p := range added {
		if p.IsMain {
			t.Errorf("added[%d] should not be main (existing main exists)", i)
		}
	}
}

// --- Тесты триггеров ---

// TestPhotoRepo_Trigger_SingleMainPhoto проверяет триггер trg_single_main_photo.
// При INSERT с is_main=true все остальные фото героя должны сбросить is_main.
func TestPhotoRepo_Trigger_SingleMainPhoto(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём 3 фото, первое — главное.
	for i := 0; i < 3; i++ {
		_, err := repo.Add(ctx, domain.AddPhotoParams{
			HeroID:    heroID,
			URL:       "https://s3.example.com/photos/" + string(rune('a'+i)) + ".jpg",
			SortOrder: i,
			IsMain:    i == 0,
		})
		if err != nil {
			t.Fatalf("Add %d failed: %v", i, err)
		}
	}

	// Добавляем новое фото с is_main=true.
	_, err := repo.Add(ctx, domain.AddPhotoParams{
		HeroID:    heroID,
		URL:       "https://s3.example.com/photos/new-main.jpg",
		SortOrder: 3,
		IsMain:    true,
	})
	if err != nil {
		t.Fatalf("Add new main failed: %v", err)
	}

	// Проверяем, что только новое фото — главное.
	photos, err := repo.ListByHero(ctx, heroID)
	if err != nil {
		t.Fatalf("ListByHero failed: %v", err)
	}

	mainCount := 0
	for _, p := range photos {
		if p.IsMain {
			mainCount++
			if p.URL != "https://s3.example.com/photos/new-main.jpg" {
				t.Errorf("wrong photo is main: %s", p.URL)
			}
		}
	}
	if mainCount != 1 {
		t.Errorf("expected 1 main photo, got %d", mainCount)
	}
}

// TestPhotoRepo_Trigger_ReassignMainPhoto проверяет триггер trg_reassign_main_photo.
// При DELETE главного фото следующее по sort_order должно стать главным.
func TestPhotoRepo_Trigger_ReassignMainPhoto(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём 3 фото, первое — главное.
	var mainPhotoID string
	for i := 0; i < 3; i++ {
		id, err := repo.Add(ctx, domain.AddPhotoParams{
			HeroID:    heroID,
			URL:       "https://s3.example.com/photos/" + string(rune('a'+i)) + ".jpg",
			SortOrder: i,
			IsMain:    i == 0,
		})
		if err != nil {
			t.Fatalf("Add %d failed: %v", i, err)
		}
		if i == 0 {
			mainPhotoID = id
		}
	}

	// Удаляем главное фото.
	_, err := repo.Delete(ctx, heroID, mainPhotoID)
	if err != nil {
		t.Fatalf("Delete main failed: %v", err)
	}

	// Проверяем, что второе фото (sort_order=1) стало главным.
	photos, err := repo.ListByHero(ctx, heroID)
	if err != nil {
		t.Fatalf("ListByHero failed: %v", err)
	}

	if len(photos) != 2 {
		t.Fatalf("expected 2 photos after delete, got %d", len(photos))
	}

	mainCount := 0
	for _, p := range photos {
		if p.IsMain {
			mainCount++
			if p.SortOrder != 1 {
				t.Errorf("wrong photo became main: sort_order=%d", p.SortOrder)
			}
		}
	}
	if mainCount != 1 {
		t.Errorf("expected 1 main photo after reassign, got %d", mainCount)
	}
}

// --- Тесты IDOR protection ---

// TestPhotoRepo_DeleteByIDs_IDOR проверяет, что чужие фото не удаляются.
func TestPhotoRepo_DeleteByIDs_IDOR(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroA := createTestHero(t, repo)
	heroB := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём фото для героя A.
	photoA, err := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroA,
		URL:    "https://s3.example.com/photos/heroA.jpg",
	})
	if err != nil {
		t.Fatalf("Add photoA failed: %v", err)
	}

	// Создаём фото для героя B.
	photoB, err := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroB,
		URL:    "https://s3.example.com/photos/heroB.jpg",
	})
	if err != nil {
		t.Fatalf("Add photoB failed: %v", err)
	}

	// Пытаемся удалить фото B от имени героя A.
	_, err = repo.DeleteByIDs(ctx, heroA, []string{photoA, photoB})
	if err == nil {
		t.Fatal("DeleteByIDs should fail for IDOR attempt")
	}
	if !errors.Is(err, domain.ErrPhotosNotBelongToHero) {
		t.Errorf("expected ErrPhotosNotBelongToHero, got %v", err)
	}

	// Проверяем, что фото B не удалено.
	photosB, err := repo.ListByHero(ctx, heroB)
	if err != nil {
		t.Fatalf("ListByHero heroB failed: %v", err)
	}
	if len(photosB) != 1 {
		t.Errorf("expected 1 photo for heroB, got %d", len(photosB))
	}
}

// TestPhotoRepo_Reorder_IDOR проверяет, что чужие фото не переупорядочиваются.
func TestPhotoRepo_Reorder_IDOR(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroA := createTestHero(t, repo)
	heroB := createTestHero(t, repo)
	ctx := context.Background()

	photoA, _ := repo.Add(ctx, domain.AddPhotoParams{HeroID: heroA, URL: "https://s3.example.com/a.jpg"})
	photoB, _ := repo.Add(ctx, domain.AddPhotoParams{HeroID: heroB, URL: "https://s3.example.com/b.jpg"})

	// Пытаемся переупорядочить фото B от имени героя A.
	err := repo.Reorder(ctx, heroA, []string{photoA, photoB})
	if err == nil {
		t.Fatal("Reorder should fail for IDOR attempt")
	}
	if !errors.Is(err, domain.ErrPhotosNotBelongToHero) {
		t.Errorf("expected ErrPhotosNotBelongToHero, got %v", err)
	}
}

// --- Тесты SetMain ---

// TestPhotoRepo_SetMain_Success проверяет назначение главного фото.
func TestPhotoRepo_SetMain_Success(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём 2 фото, первое — главное.
	id1, _ := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroID, URL: "https://s3.example.com/1.jpg", SortOrder: 0, IsMain: true,
	})
	id2, _ := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroID, URL: "https://s3.example.com/2.jpg", SortOrder: 1, IsMain: false,
	})

	// Назначаем второе фото главным.
	err := repo.SetMain(ctx, heroID, id2)
	if err != nil {
		t.Fatalf("SetMain failed: %v", err)
	}

	// Проверяем, что только второе фото — главное.
	mainURL, err := repo.GetMainPhotoURL(ctx, heroID)
	if err != nil {
		t.Fatalf("GetMainPhotoURL failed: %v", err)
	}
	if mainURL != "https://s3.example.com/2.jpg" {
		t.Errorf("expected main URL https://s3.example.com/2.jpg, got %s", mainURL)
	}

	// Первое фото больше не главное.
	photos, _ := repo.ListByHero(ctx, heroID)
	for _, p := range photos {
		if p.ID == id1 && p.IsMain {
			t.Error("first photo should not be main after SetMain")
		}
	}
}

// --- Тесты Update ---

// TestPhotoRepo_Update_FieldMask проверяет partial update с field_mask.
func TestPhotoRepo_Update_FieldMask(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	id, _ := repo.Add(ctx, domain.AddPhotoParams{
		HeroID:      heroID,
		URL:         "https://s3.example.com/original.jpg",
		Description: "Оригинальное описание",
	})

	// Обновляем только description.
	updated, err := repo.Update(ctx, domain.UpdatePhotoParams{
		PhotoID:     id,
		HeroID:      heroID,
		Description: "Новое описание",
		FieldMask:   []string{"description"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Description != "Новое описание" {
		t.Errorf("description not updated: %s", updated.Description)
	}
}

// TestPhotoRepo_UpdateAfterProcessing проверяет обновление URL после thumbnail worker.
func TestPhotoRepo_UpdateAfterProcessing(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	id, _ := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroID,
		URL:    "https://s3.example.com/original.jpg",
	})

	newURL := "https://s3.example.com/photos/" + id + ".webp"
	newThumbnail := "https://s3.example.com/thumbnails/" + id + ".webp"

	err := repo.UpdateAfterProcessing(ctx, id, newURL, newThumbnail)
	if err != nil {
		t.Fatalf("UpdateAfterProcessing failed: %v", err)
	}

	// Проверяем, что URL обновились.
	photos, _ := repo.ListByHero(ctx, heroID)
	if len(photos) != 1 {
		t.Fatalf("expected 1 photo, got %d", len(photos))
	}
	if photos[0].URL != newURL {
		t.Errorf("URL not updated: %s", photos[0].URL)
	}
	if photos[0].ThumbnailURL != newThumbnail {
		t.Errorf("ThumbnailURL not updated: %s", photos[0].ThumbnailURL)
	}
}

// --- Тесты служебных методов ---

// TestPhotoRepo_ListWithoutThumbnails проверяет поиск фото без превью.
func TestPhotoRepo_ListWithoutThumbnails(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	// Создаём 2 фото: одно с thumbnail, одно без.
	_, _ = repo.Add(ctx, domain.AddPhotoParams{
		HeroID:       heroID,
		URL:          "https://s3.example.com/with-thumb.jpg",
		ThumbnailURL: "https://s3.example.com/thumb.jpg",
	})
	idWithout, _ := repo.Add(ctx, domain.AddPhotoParams{
		HeroID: heroID,
		URL:    "https://s3.example.com/without-thumb.jpg",
	})

	photos, err := repo.ListWithoutThumbnails(ctx)
	if err != nil {
		t.Fatalf("ListWithoutThumbnails failed: %v", err)
	}

	if len(photos) != 1 {
		t.Errorf("expected 1 photo without thumbnail, got %d", len(photos))
	}
	if photos[0].ID != idWithout {
		t.Errorf("wrong photo returned: %s", photos[0].ID)
	}
}

// TestPhotoRepo_ListAllMediaURLs проверяет сбор всех URL для orphan cleanup.
func TestPhotoRepo_ListAllMediaURLs(t *testing.T) {
	repo, cleanup := newPhotoRepo(t)
	defer cleanup()

	heroID := createTestHero(t, repo)
	ctx := context.Background()

	_, _ = repo.Add(ctx, domain.AddPhotoParams{
		HeroID:       heroID,
		URL:          "https://s3.example.com/photo1.jpg",
		ThumbnailURL: "https://s3.example.com/thumb1.jpg",
	})
	_, _ = repo.Add(ctx, domain.AddPhotoParams{
		HeroID:       heroID,
		URL:          "https://s3.example.com/photo2.jpg",
		ThumbnailURL: "https://s3.example.com/thumb2.jpg",
	})

	urls, err := repo.ListAllMediaURLs(ctx)
	if err != nil {
		t.Fatalf("ListAllMediaURLs failed: %v", err)
	}

	// Должно вернуть 4 URL (2 оригинала + 2 превью).
	if len(urls) != 4 {
		t.Errorf("expected 4 URLs, got %d", len(urls))
	}
}
