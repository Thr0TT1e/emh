package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ============================================================
// Хелперы
// ============================================================

// testTime создаёт время для тестов.
func testTime(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}

// ============================================================
// Тесты Create
// ============================================================

// TestConflictUseCase_Create_SuccessWithoutParent проверяет успешное создание без родителя.
func TestConflictUseCase_Create_SuccessWithoutParent(t *testing.T) {
	var capturedParams domain.CreateConflictParams
	repo := &mockConflictRepository{
		createFn: func(ctx context.Context, p domain.CreateConflictParams) (string, error) {
			capturedParams = p
			return "conflict-123", nil
		},
	}

	uc := NewConflictUseCase(repo)

	startDate := testTime(1941, 6, 22)
	endDate := testTime(1945, 5, 9)
	params := domain.CreateConflictParams{
		Name:        "Великая Отечественная война",
		Description: "1941-1945",
		Type:        domain.ConflictTypeGlobal,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	id, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id != "conflict-123" {
		t.Errorf("Create() id = %q, want %q", id, "conflict-123")
	}
	if capturedParams.Name != "Велическая Отечественная война" && capturedParams.Name != "Великая Отечественная война" {
		t.Errorf("captured Name = %q", capturedParams.Name)
	}
}

// TestConflictUseCase_Create_SuccessWithParent проверяет успешное создание с родителем.
func TestConflictUseCase_Create_SuccessWithParent(t *testing.T) {
	parentID := "parent-123"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			if id == parentID {
				return &domain.Conflict{ID: parentID, Name: "Parent"}, nil
			}
			return nil, domain.ErrNotFound
		},
		createFn: func(ctx context.Context, p domain.CreateConflictParams) (string, error) {
			return "child-456", nil
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.CreateConflictParams{
		Name:             "Битва за Москву",
		Type:             domain.ConflictTypeGlobal,
		ParentConflictID: &parentID,
	}

	id, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id != "child-456" {
		t.Errorf("Create() id = %q, want %q", id, "child-456")
	}
}

// TestConflictUseCase_Create_EmptyName проверяет ошибку при пустом имени.
func TestConflictUseCase_Create_EmptyName(t *testing.T) {
	repo := &mockConflictRepository{}
	uc := NewConflictUseCase(repo)

	params := domain.CreateConflictParams{
		Name: "",
		Type: domain.ConflictTypeGlobal,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("Create() should return error for empty name")
	}
	if !errors.Is(err, domain.ErrConflictNameRequired) {
		t.Errorf("Create() error = %v, want ErrConflictNameRequired", err)
	}
}

// TestConflictUseCase_Create_InvalidDates проверяет ошибку при невалидных датах.
func TestConflictUseCase_Create_InvalidDates(t *testing.T) {
	repo := &mockConflictRepository{}
	uc := NewConflictUseCase(repo)

	// Дата окончания раньше даты начала
	startDate := testTime(1945, 5, 9)
	endDate := testTime(1941, 6, 22)
	params := domain.CreateConflictParams{
		Name:      "Тест",
		Type:      domain.ConflictTypeGlobal,
		StartDate: startDate,
		EndDate:   endDate,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("Create() should return error for invalid dates")
	}
	if !errors.Is(err, domain.ErrConflictDateInvalid) {
		t.Errorf("Create() error = %v, want ErrConflictDateInvalid", err)
	}
}

// TestConflictUseCase_Create_ParentNotFound проверяет ошибку при несуществующем родителе.
func TestConflictUseCase_Create_ParentNotFound(t *testing.T) {
	parentID := "nonexistent-parent"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.CreateConflictParams{
		Name:             "Child",
		Type:             domain.ConflictTypeGlobal,
		ParentConflictID: &parentID,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("Create() should return error for nonexistent parent")
	}
}

// TestConflictUseCase_Create_RepoError проверяет обработку ошибки репозитория.
func TestConflictUseCase_Create_RepoError(t *testing.T) {
	repo := &mockConflictRepository{
		createFn: func(ctx context.Context, p domain.CreateConflictParams) (string, error) {
			return "", errors.New("db connection failed")
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.CreateConflictParams{
		Name: "Тест",
		Type: domain.ConflictTypeGlobal,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("Create() should return error on repo failure")
	}
}

// ============================================================
// Тесты Update
// ============================================================

// TestConflictUseCase_Update_Success проверяет успешное обновление.
func TestConflictUseCase_Update_Success(t *testing.T) {
	existingID := "conflict-123"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return &domain.Conflict{
				ID:        existingID,
				Name:      "Old Name",
				StartDate: testTime(1941, 6, 22),
				EndDate:   testTime(1945, 5, 9),
			}, nil
		},
		updateFn: func(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error) {
			return &domain.Conflict{ID: p.ID, Name: p.Name}, nil
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.UpdateConflictParams{
		ID:        existingID,
		Name:      "New Name",
		FieldMask: []string{"name"},
	}

	result, err := uc.Update(context.Background(), params)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if result.Name != "New Name" {
		t.Errorf("Update() name = %q, want %q", result.Name, "New Name")
	}
}

// TestConflictUseCase_Update_InvalidDates проверяет ошибку при невалидных датах после обновления.
func TestConflictUseCase_Update_InvalidDates(t *testing.T) {
	existingID := "conflict-123"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return &domain.Conflict{
				ID:        existingID,
				StartDate: testTime(1941, 6, 22),
				EndDate:   testTime(1945, 5, 9),
			}, nil
		},
	}

	uc := NewConflictUseCase(repo)

	// Пытаемся обновить end_date на дату раньше start_date
	newEndDate := testTime(1940, 1, 1) // раньше 1941
	params := domain.UpdateConflictParams{
		ID:        existingID,
		EndDate:   newEndDate,
		FieldMask: []string{"end_date"},
	}

	_, err := uc.Update(context.Background(), params)
	if err == nil {
		t.Fatal("Update() should return error for invalid dates")
	}
	if !errors.Is(err, domain.ErrConflictDateInvalid) {
		t.Errorf("Update() error = %v, want ErrConflictDateInvalid", err)
	}
}

// TestConflictUseCase_Update_OwnParent проверяет защиту от самоссылки.
func TestConflictUseCase_Update_OwnParent(t *testing.T) {
	conflictID := "conflict-123"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return &domain.Conflict{ID: conflictID}, nil
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.UpdateConflictParams{
		ID:               conflictID,
		ParentConflictID: &conflictID, // самоссылка
		FieldMask:        []string{"parent_conflict_id"},
	}

	_, err := uc.Update(context.Background(), params)
	if err == nil {
		t.Fatal("Update() should return error for self-reference")
	}
	if !errors.Is(err, domain.ErrOwnParent) {
		t.Errorf("Update() error = %v, want ErrOwnParent", err)
	}
}

// TestConflictUseCase_Update_CyclicReference проверяет защиту от циклических ссылок.
func TestConflictUseCase_Update_CyclicReference(t *testing.T) {
	conflictID := "conflict-A"
	parentID := "conflict-B"
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return &domain.Conflict{ID: conflictID}, nil
		},
		hasCyclicReferenceFn: func(ctx context.Context, id, parentID string) (bool, error) {
			return true, nil // Цикл обнаружен
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.UpdateConflictParams{
		ID:               conflictID,
		ParentConflictID: &parentID,
		FieldMask:        []string{"parent_conflict_id"},
	}

	_, err := uc.Update(context.Background(), params)
	if err == nil {
		t.Fatal("Update() should return error for cyclic reference")
	}
	if !errors.Is(err, domain.ErrCyclicReference) {
		t.Errorf("Update() error = %v, want ErrCyclicReference", err)
	}
}

// TestConflictUseCase_Update_NotFound проверяет ошибку при несуществующем конфликте.
func TestConflictUseCase_Update_NotFound(t *testing.T) {
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewConflictUseCase(repo)

	params := domain.UpdateConflictParams{
		ID:        "nonexistent-id",
		Name:      "New Name",
		FieldMask: []string{"name"},
	}

	_, err := uc.Update(context.Background(), params)
	if err == nil {
		t.Fatal("Update() should return error for nonexistent conflict")
	}
}

// ============================================================
// Тесты Delete
// ============================================================

// TestConflictUseCase_Delete_Success проверяет успешное удаление.
func TestConflictUseCase_Delete_Success(t *testing.T) {
	var deletedID string
	repo := &mockConflictRepository{
		hasHeroesFn: func(ctx context.Context, id string) (bool, error) {
			return false, nil // Нет связанных героев
		},
		deleteFn: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}

	uc := NewConflictUseCase(repo)

	err := uc.Delete(context.Background(), "conflict-123")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if deletedID != "conflict-123" {
		t.Errorf("Delete() id = %q, want %q", deletedID, "conflict-123")
	}
}

// TestConflictUseCase_Delete_HasHeroes проверяет ошибку при наличии связанных героев.
func TestConflictUseCase_Delete_HasHeroes(t *testing.T) {
	repo := &mockConflictRepository{
		hasHeroesFn: func(ctx context.Context, id string) (bool, error) {
			return true, nil // Есть связанные герои
		},
	}

	uc := NewConflictUseCase(repo)

	err := uc.Delete(context.Background(), "conflict-123")
	if err == nil {
		t.Fatal("Delete() should return error when conflict has heroes")
	}
	if !errors.Is(err, domain.ErrEntityAssignedToHeroes) {
		t.Errorf("Delete() error = %v, want ErrEntityAssignedToHeroes", err)
	}
}

// TestConflictUseCase_Delete_HasHeroesError проверяет обработку ошибки HasHeroes.
func TestConflictUseCase_Delete_HasHeroesError(t *testing.T) {
	repo := &mockConflictRepository{
		hasHeroesFn: func(ctx context.Context, id string) (bool, error) {
			return false, errors.New("db error")
		},
	}

	uc := NewConflictUseCase(repo)

	err := uc.Delete(context.Background(), "conflict-123")
	if err == nil {
		t.Fatal("Delete() should return error on HasHeroes failure")
	}
}

// ============================================================
// Тесты GetByID
// ============================================================

// TestConflictUseCase_GetByID_Success проверяет успешный возврат.
func TestConflictUseCase_GetByID_Success(t *testing.T) {
	expected := &domain.Conflict{
		ID:   "conflict-123",
		Name: "Великая Отечественная война",
	}
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return expected, nil
		},
	}

	uc := NewConflictUseCase(repo)

	result, err := uc.GetByID(context.Background(), "conflict-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("GetByID() id = %q, want %q", result.ID, expected.ID)
	}
	if result.Name != expected.Name {
		t.Errorf("GetByID() name = %q, want %q", result.Name, expected.Name)
	}
}

// TestConflictUseCase_GetByID_NotFound проверяет ошибку при отсутствии.
func TestConflictUseCase_GetByID_NotFound(t *testing.T) {
	repo := &mockConflictRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewConflictUseCase(repo)

	_, err := uc.GetByID(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("GetByID() should return error for nonexistent conflict")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

// ============================================================
// Тесты List
// ============================================================

// TestConflictUseCase_List_Success проверяет успешный возврат списка.
func TestConflictUseCase_List_Success(t *testing.T) {
	expected := []*domain.Conflict{
		{ID: "1", Name: "ВОВ"},
		{ID: "2", Name: "Афган"},
	}
	repo := &mockConflictRepository{
		listFn: func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
			return expected, "", 2, nil
		},
	}

	uc := NewConflictUseCase(repo)
	results, cursor, total, err := uc.List(context.Background(), domain.ConflictFilter{
		Type:  domain.ConflictTypeGlobal,
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("List() count = %d, want 2", len(results))
	}
	if cursor != "" {
		t.Errorf("cursor = %q, want empty", cursor)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
}

// TestConflictUseCase_List_Error проверяет обработку ошибки репозитория.
func TestConflictUseCase_List_Error(t *testing.T) {
	repo := &mockConflictRepository{
		listFn: func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
			return nil, "", 0, errors.New("db error")
		},
	}

	uc := NewConflictUseCase(repo)
	_, _, _, err := uc.List(context.Background(), domain.ConflictFilter{
		Type:  domain.ConflictTypeGlobal,
		Limit: 20,
	})
	if err == nil {
		t.Fatal("List() should return error on repo failure")
	}
}

// ============================================================
// Тесты вспомогательных функций
// ============================================================

// TestValidateConflictDates проверяет валидацию дат.
func TestValidateConflictDates(t *testing.T) {
	tests := []struct {
		name      string
		start     *time.Time
		end       *time.Time
		wantError bool
	}{
		{
			name:      "valid dates",
			start:     testTime(1941, 6, 22),
			end:       testTime(1945, 5, 9),
			wantError: false,
		},
		{
			name:      "same dates",
			start:     testTime(1941, 6, 22),
			end:       testTime(1941, 6, 22),
			wantError: false,
		},
		{
			name:      "end before start",
			start:     testTime(1945, 5, 9),
			end:       testTime(1941, 6, 22),
			wantError: true,
		},
		{
			name:      "nil start",
			start:     nil,
			end:       testTime(1945, 5, 9),
			wantError: false,
		},
		{
			name:      "nil end",
			start:     testTime(1941, 6, 22),
			end:       nil,
			wantError: false,
		},
		{
			name:      "both nil",
			start:     nil,
			end:       nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConflictDates(tt.start, tt.end)
			if (err != nil) != tt.wantError {
				t.Errorf("validateConflictDates() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestHasField проверяет поиск поля в маске.
func TestHasField(t *testing.T) {
	tests := []struct {
		name     string
		mask     []string
		field    string
		expected bool
	}{
		{"found", []string{"name", "description", "type"}, "description", true},
		{"not found", []string{"name", "description"}, "type", false},
		{"empty mask", []string{}, "name", false},
		{"empty field", []string{"name"}, "", false},
		{"first element", []string{"name", "description"}, "name", true},
		{"last element", []string{"name", "description"}, "description", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasField(tt.mask, tt.field)
			if result != tt.expected {
				t.Errorf("hasField(%v, %q) = %v, want %v", tt.mask, tt.field, result, tt.expected)
			}
		})
	}
}
