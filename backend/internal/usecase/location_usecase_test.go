package usecase

import (
	"context"
	"errors"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// --- Хелперы ---

// newLocationUseCase создаёт LocationUseCase с моком.
func newLocationUseCase(repo *mockLocationRepository) LocationUseCase {
	return NewLocationUseCase(repo)
}

// ptrFloat64 возвращает указатель на float64.
func ptrFloat64(f float64) *float64 {
	return &f
}

// --- Тесты Create ---

// TestLocationUseCase_Create_Success проверяет успешное создание локации.
func TestLocationUseCase_Create_Success(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	id, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name: "Москва",
		Type: domain.LocationTypeCity,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
}

// TestLocationUseCase_Create_WithCoordinates проверяет создание с координатами.
func TestLocationUseCase_Create_WithCoordinates(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	id, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name:      "Москва",
		Type:      domain.LocationTypeCity,
		Latitude:  ptrFloat64(55.7558),
		Longitude: ptrFloat64(37.6173),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
}

// TestLocationUseCase_Create_EmptyName проверяет ошибку при пустом имени.
func TestLocationUseCase_Create_EmptyName(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	_, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name: "",
		Type: domain.LocationTypeCity,
	})
	if !errors.Is(err, domain.ErrLocationNameRequired) {
		t.Errorf("expected ErrLocationNameRequired, got %v", err)
	}
}

// TestLocationUseCase_Create_InvalidLatitude проверяет ошибку при невалидной широте.
func TestLocationUseCase_Create_InvalidLatitude(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
	}{
		{"below -90", -90.1},
		{"above 90", 90.1},
		{"way below", -1000},
		{"way above", 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLocationRepository{}
			uc := newLocationUseCase(repo)

			_, err := uc.Create(context.Background(), domain.CreateLocationParams{
				Name:     "Тест",
				Type:     domain.LocationTypeCity,
				Latitude: ptrFloat64(tt.lat),
			})
			if !errors.Is(err, domain.ErrInvalidLatitude) {
				t.Errorf("expected ErrInvalidLatitude, got %v", err)
			}
		})
	}
}

// TestLocationUseCase_Create_InvalidLongitude проверяет ошибку при невалидной долготе.
func TestLocationUseCase_Create_InvalidLongitude(t *testing.T) {
	tests := []struct {
		name string
		lon  float64
	}{
		{"below -180", -180.1},
		{"above 180", 180.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLocationRepository{}
			uc := newLocationUseCase(repo)

			_, err := uc.Create(context.Background(), domain.CreateLocationParams{
				Name:      "Тест",
				Type:      domain.LocationTypeCity,
				Longitude: ptrFloat64(tt.lon),
			})
			if !errors.Is(err, domain.ErrInvalidLongitude) {
				t.Errorf("expected ErrInvalidLongitude, got %v", err)
			}
		})
	}
}

// TestLocationUseCase_Create_ZeroCoordinates проверяет, что координаты 0,0 валидны.
// Это защита от бага `!= 0` вместо `!= nil` (техдолг #22).
func TestLocationUseCase_Create_ZeroCoordinates(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	// Нулевой меридиан и экватор — валидные координаты
	_, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name:      "Нулевая точка",
		Type:      domain.LocationTypeCity,
		Latitude:  ptrFloat64(0),
		Longitude: ptrFloat64(0),
	})
	if err != nil {
		t.Errorf("zero coordinates should be valid, got error: %v", err)
	}
}

// TestLocationUseCase_Create_ParentNotFound проверяет ошибку при несуществующем родителе.
func TestLocationUseCase_Create_ParentNotFound(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "nonexistent-parent"
	_, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name:     "Дочерняя",
		Type:     domain.LocationTypeCity,
		ParentID: &parentID,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent parent")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected wrapped ErrNotFound, got %v", err)
	}
}

// TestLocationUseCase_Create_WithParent проверяет успешное создание с родителем.
func TestLocationUseCase_Create_WithParent(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	parentID := "parent-id"
	id, err := uc.Create(context.Background(), domain.CreateLocationParams{
		Name:     "Дочерняя",
		Type:     domain.LocationTypeCity,
		ParentID: &parentID,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
}

// --- Тесты Update ---

// TestLocationUseCase_Update_Success проверяет успешное обновление.
func TestLocationUseCase_Update_Success(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	result, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		Name:      "Новое имя",
		FieldMask: []string{"name"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result == nil {
		t.Fatal("Update returned nil")
	}
}

// TestLocationUseCase_Update_NotFound проверяет ошибку при отсутствии локации.
func TestLocationUseCase_Update_NotFound(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	uc := newLocationUseCase(repo)

	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "nonexistent",
		Name:      "Имя",
		FieldMask: []string{"name"},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestLocationUseCase_Update_InvalidCoordinates проверяет валидацию координат при обновлении.
func TestLocationUseCase_Update_InvalidCoordinates(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{ID: id}, nil
		},
	}
	uc := newLocationUseCase(repo)

	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		Latitude:  ptrFloat64(91),
		FieldMask: []string{"latitude"},
	})
	if !errors.Is(err, domain.ErrInvalidLatitude) {
		t.Errorf("expected ErrInvalidLatitude, got %v", err)
	}
}

// TestLocationUseCase_Update_CoordinatesValidationWithMask проверяет, что
// валидация координат учитывает field_mask (итоговые значения).
func TestLocationUseCase_Update_CoordinatesValidationWithMask(t *testing.T) {
	// Существующая локация имеет валидные координаты
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{
				ID:        id,
				Latitude:  ptrFloat64(55.7558),
				Longitude: ptrFloat64(37.6173),
			}, nil
		},
	}
	uc := newLocationUseCase(repo)

	// Обновляем только name — координаты не должны влиять
	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		Name:      "Новое имя",
		FieldMask: []string{"name"},
	})
	if err != nil {
		t.Errorf("expected no error when coordinates not in mask, got %v", err)
	}
}

// TestLocationUseCase_Update_OwnParent проверяет ошибку при установке себя родителем.
func TestLocationUseCase_Update_OwnParent(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{ID: id}, nil
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "location-1" // тот же ID
	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		ParentID:  &parentID,
		FieldMask: []string{"parent_id"},
	})
	if !errors.Is(err, domain.ErrOwnParent) {
		t.Errorf("expected ErrOwnParent, got %v", err)
	}
}

// TestLocationUseCase_Update_CyclicReference проверяет обнаружение транзитивного цикла.
// Иерархия: A → B → C. Попытка сделать A.parent = C создаёт цикл.
func TestLocationUseCase_Update_CyclicReference(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{ID: id}, nil
		},
		hasCyclicReferenceFunc: func(ctx context.Context, selfID, candidateParentID string) (bool, error) {
			// Имитация: candidateParentID является потомком selfID
			return true, nil
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "location-C"
	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-A",
		ParentID:  &parentID,
		FieldMask: []string{"parent_id"},
	})
	if !errors.Is(err, domain.ErrCyclicReference) {
		t.Errorf("expected ErrCyclicReference, got %v", err)
	}
}

// TestLocationUseCase_Update_NoCyclicReference пропускает валидную иерархию.
func TestLocationUseCase_Update_NoCyclicReference(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{ID: id}, nil
		},
		hasCyclicReferenceFunc: func(ctx context.Context, selfID, candidateParentID string) (bool, error) {
			return false, nil // цикла нет
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "valid-parent"
	result, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		ParentID:  &parentID,
		FieldMask: []string{"parent_id"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result == nil {
		t.Fatal("Update returned nil")
	}
}

// TestLocationUseCase_Update_ParentNotFound проверяет ошибку при несуществующем новом родителе.
func TestLocationUseCase_Update_ParentNotFound(t *testing.T) {
	callCount := 0
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			callCount++
			// Первый вызов — сама локация, второй — родитель
			if callCount == 1 {
				return &domain.Location{ID: id}, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "nonexistent-parent"
	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		ParentID:  &parentID,
		FieldMask: []string{"parent_id"},
	})
	if err == nil {
		t.Fatal("expected error for nonexistent parent")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected wrapped ErrNotFound, got %v", err)
	}
}

// TestLocationUseCase_Update_CyclicReferenceCheckError проверяет обработку ошибки проверки цикла.
func TestLocationUseCase_Update_CyclicReferenceCheckError(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{ID: id}, nil
		},
		hasCyclicReferenceFunc: func(ctx context.Context, selfID, candidateParentID string) (bool, error) {
			return false, errors.New("db connection lost")
		},
	}
	uc := newLocationUseCase(repo)

	parentID := "some-parent"
	_, err := uc.Update(context.Background(), domain.UpdateLocationParams{
		ID:        "location-1",
		ParentID:  &parentID,
		FieldMask: []string{"parent_id"},
	})
	if err == nil {
		t.Fatal("expected error from HasCyclicReference")
	}
}

// --- Тесты Delete ---

// TestLocationUseCase_Delete_Success проверяет успешное удаление.
func TestLocationUseCase_Delete_Success(t *testing.T) {
	repo := &mockLocationRepository{
		hasHeroesFunc: func(ctx context.Context, locationID string) (bool, error) {
			return false, nil
		},
	}
	uc := newLocationUseCase(repo)

	err := uc.Delete(context.Background(), "location-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// TestLocationUseCase_Delete_AssignedToHeroes проверяет ошибку при привязке к героям.
func TestLocationUseCase_Delete_AssignedToHeroes(t *testing.T) {
	repo := &mockLocationRepository{
		hasHeroesFunc: func(ctx context.Context, locationID string) (bool, error) {
			return true, nil
		},
	}
	uc := newLocationUseCase(repo)

	err := uc.Delete(context.Background(), "location-1")
	if !errors.Is(err, domain.ErrEntityAssignedToHeroes) {
		t.Errorf("expected ErrEntityAssignedToHeroes, got %v", err)
	}
}

// TestLocationUseCase_Delete_HasHeroesError проверяет обработку ошибки проверки.
func TestLocationUseCase_Delete_HasHeroesError(t *testing.T) {
	repo := &mockLocationRepository{
		hasHeroesFunc: func(ctx context.Context, locationID string) (bool, error) {
			return false, errors.New("db error")
		},
	}
	uc := newLocationUseCase(repo)

	err := uc.Delete(context.Background(), "location-1")
	if err == nil {
		t.Fatal("expected error from HasHeroes")
	}
}

// --- Тесты Reparent ---

// TestLocationUseCase_Reparent_Success проверяет успешный перенос детей.
func TestLocationUseCase_Reparent_Success(t *testing.T) {
	repo := &mockLocationRepository{
		reparentFunc: func(ctx context.Context, oldParentID, newParentID string) (int, error) {
			return 5, nil
		},
	}
	uc := newLocationUseCase(repo)

	affected, err := uc.Reparent(context.Background(), "old-parent", "new-parent")
	if err != nil {
		t.Fatalf("Reparent failed: %v", err)
	}
	if affected != 5 {
		t.Errorf("affected = %d, want 5", affected)
	}
}

// TestLocationUseCase_Reparent_ToRoot проверяет перенос детей в корень (пустой newParent).
func TestLocationUseCase_Reparent_ToRoot(t *testing.T) {
	repo := &mockLocationRepository{
		reparentFunc: func(ctx context.Context, oldParentID, newParentID string) (int, error) {
			if newParentID != "" {
				t.Errorf("newParentID should be empty for root, got %q", newParentID)
			}
			return 3, nil
		},
	}
	uc := newLocationUseCase(repo)

	affected, err := uc.Reparent(context.Background(), "old-parent", "")
	if err != nil {
		t.Fatalf("Reparent failed: %v", err)
	}
	if affected != 3 {
		t.Errorf("affected = %d, want 3", affected)
	}
}

// TestLocationUseCase_Reparent_NewParentNotFound проверяет ошибку при несуществующем новом родителе.
func TestLocationUseCase_Reparent_NewParentNotFound(t *testing.T) {
	repo := &mockLocationRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	uc := newLocationUseCase(repo)

	_, err := uc.Reparent(context.Background(), "old-parent", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent new parent")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected wrapped ErrNotFound, got %v", err)
	}
}

// --- Тесты делегирования ---

// TestLocationUseCase_GetByID проверяет делегирование в репозиторий.
func TestLocationUseCase_GetByID(t *testing.T) {
	repo := &mockLocationRepository{}
	uc := newLocationUseCase(repo)

	result, err := uc.GetByID(context.Background(), "location-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if result.ID != "location-1" {
		t.Errorf("ID = %q, want location-1", result.ID)
	}
}

// TestLocationUseCase_List проверяет делегирование в репозиторий.
func TestLocationUseCase_List(t *testing.T) {
	repo := &mockLocationRepository{
		listFunc: func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
			return []*domain.Location{
				{ID: "loc-1", Name: "Москва"},
				{ID: "loc-2", Name: "Санкт-Петербург"},
			}, "", 2, nil
		},
	}
	uc := newLocationUseCase(repo)
	result, _, _, err := uc.List(context.Background(), domain.LocationFilter{
		ParentID:    "",
		SearchQuery: "",
		Type:        domain.LocationTypeUnspecified,
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}
