package v1

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
)

// --- Моки usecase ---

// mockHeroUCForAdmin мок для HeroUseCase (только нужные методы).
type mockHeroUCForAdmin struct {
	createFunc func(ctx context.Context, p domain.CreateHeroParams) (string, error)
	updateFunc func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
	deleteFunc func(ctx context.Context, id string, hard bool) error
	getFunc    func(ctx context.Context, id string) (*domain.Hero, error)
	listFunc   func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
}

func (m *mockHeroUCForAdmin) CreateHero(ctx context.Context, p domain.CreateHeroParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return newUUID(), nil
}
func (m *mockHeroUCForAdmin) UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Hero{ID: *p.ID, FirstName: "Иван", LastName: "Петров", Status: domain.StatusDraft}, nil
}
func (m *mockHeroUCForAdmin) DeleteHero(ctx context.Context, id string, hard bool) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, hard)
	}
	return nil
}
func (m *mockHeroUCForAdmin) GetByID(ctx context.Context, id string) (*domain.Hero, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &domain.Hero{ID: id}, nil
}
func (m *mockHeroUCForAdmin) ListHeroes(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Hero{}, "", 0, nil
}

// mockHeroQueryUCForAdmin мок для HeroQueryUseCase.
type mockHeroQueryUCForAdmin struct {
	getDetailFunc  func(ctx context.Context, id string) (*domain.HeroDetail, error)
	listPhotosFunc func(ctx context.Context, heroID string) ([]*domain.Photo, error)
}

func (m *mockHeroQueryUCForAdmin) GetHeroDetail(ctx context.Context, id string) (*domain.HeroDetail, error) {
	if m.getDetailFunc != nil {
		return m.getDetailFunc(ctx, id)
	}
	return &domain.HeroDetail{
		Hero: &domain.Hero{ID: id, FirstName: "Иван", LastName: "Петров", Status: domain.StatusDraft,
			BirthDate: domain.NewUnknownDate(), DeathDate: domain.NewUnknownDate()},
	}, nil
}
func (m *mockHeroQueryUCForAdmin) ListHeroPhotos(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	if m.listPhotosFunc != nil {
		return m.listPhotosFunc(ctx, heroID)
	}
	return []*domain.Photo{}, nil
}

// ListHeroPhotosPaged — заглушка для удовлетворения интерфейса.
// Admin-сервер не использует этот метод.
func (m *mockHeroQueryUCForAdmin) ListHeroPhotosPaged(ctx context.Context, heroID, cursor string, limit int) ([]*domain.Photo, string, int64, error) {
	// Возвращаем пустой результат — admin-сервер не вызывает этот метод
	return nil, "", 0, nil
}

// mockAwardUC мок для HeroAwardUseCase.
type mockAwardUC struct {
	addFunc        func(ctx context.Context, heroID, awardID string, date *string, decree string) error
	removeFunc     func(ctx context.Context, heroID, awardID string) error
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroAward, error)
}

func (m *mockAwardUC) Add(ctx context.Context, heroID, awardID string, date *string, decree string) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, heroID, awardID, date, decree)
	}
	return nil
}
func (m *mockAwardUC) Remove(ctx context.Context, heroID, awardID string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, heroID, awardID)
	}
	return nil
}
func (m *mockAwardUC) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroAward{}, nil
}

// mockConflictUC мок для HeroConflictUseCase.
type mockConflictUC struct {
	addFunc        func(ctx context.Context, heroID, conflictID, location, rank string) error
	removeFunc     func(ctx context.Context, heroID, conflictID string) error
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroConflict, error)
}

func (m *mockConflictUC) Add(ctx context.Context, heroID, conflictID, location, rank string) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, heroID, conflictID, location, rank)
	}
	return nil
}
func (m *mockConflictUC) Remove(ctx context.Context, heroID, conflictID string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, heroID, conflictID)
	}
	return nil
}
func (m *mockConflictUC) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroConflict{}, nil
}

// mockLocationUC мок для HeroLocationUseCase.
type mockLocationUC struct {
	addFunc        func(ctx context.Context, heroID, locationID string, t domain.HeroLocationType) error
	removeFunc     func(ctx context.Context, heroID, locationID string, t domain.HeroLocationType) error
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroLocation, error)
}

func (m *mockLocationUC) Add(ctx context.Context, heroID, locationID string, t domain.HeroLocationType) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, heroID, locationID, t)
	}
	return nil
}
func (m *mockLocationUC) Remove(ctx context.Context, heroID, locationID string, t domain.HeroLocationType) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, heroID, locationID, t)
	}
	return nil
}
func (m *mockLocationUC) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroLocation{}, nil
}

// mockSourceUCForAdmin мок для HeroSourceUseCase.
type mockSourceUCForAdmin struct {
	addFunc    func(ctx context.Context, p domain.AddHeroSourceParams) (string, error)
	removeFunc func(ctx context.Context, heroID, sourceID string) error
}

func (m *mockSourceUCForAdmin) Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error) {
	if m.addFunc != nil {
		return m.addFunc(ctx, p)
	}
	return newUUID(), nil
}
func (m *mockSourceUCForAdmin) Remove(ctx context.Context, heroID, sourceID string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, heroID, sourceID)
	}
	return nil
}
func (m *mockSourceUCForAdmin) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
	return []*domain.HeroSource{}, nil
}

// mockRelationUCForAdmin мок для HeroRelationUseCase.
type mockRelationUCForAdmin struct {
	addFunc    func(ctx context.Context, p domain.AddHeroRelationParams) (string, error)
	removeFunc func(ctx context.Context, id string) error
}

func (m *mockRelationUCForAdmin) Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
	if m.addFunc != nil {
		return m.addFunc(ctx, p)
	}
	return newUUID(), nil
}
func (m *mockRelationUCForAdmin) Remove(ctx context.Context, id string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, id)
	}
	return nil
}
func (m *mockRelationUCForAdmin) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
	return []*domain.HeroRelation{}, nil
}

// mockPhotoUCForAdmin мок для PhotoUseCase.
type mockPhotoUCForAdmin struct {
	addFunc         func(ctx context.Context, p domain.AddPhotoParams) (string, error)
	batchAddFunc    func(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error)
	deleteFunc      func(ctx context.Context, heroID, photoID string) error
	deleteBatchFunc func(ctx context.Context, heroID string, ids []string) (int, error)
	reorderFunc     func(ctx context.Context, heroID string, ids []string) error
	setMainFunc     func(ctx context.Context, heroID, photoID string) error
	updateFunc      func(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error)
	listFunc        func(ctx context.Context, heroID string) ([]*domain.Photo, error)
}

func (m *mockPhotoUCForAdmin) Add(ctx context.Context, p domain.AddPhotoParams) (string, error) {
	if m.addFunc != nil {
		return m.addFunc(ctx, p)
	}
	return newUUID(), nil
}
func (m *mockPhotoUCForAdmin) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
	if m.batchAddFunc != nil {
		return m.batchAddFunc(ctx, heroID, photos)
	}
	result := make([]*domain.Photo, len(photos))
	for i, p := range photos {
		result[i] = &domain.Photo{ID: newUUID(), HeroID: heroID, URL: p.URL, IsMain: i == 0}
	}
	return result, nil
}
func (m *mockPhotoUCForAdmin) Delete(ctx context.Context, heroID, photoID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, heroID, photoID)
	}
	return nil
}
func (m *mockPhotoUCForAdmin) DeleteBatch(ctx context.Context, heroID string, ids []string) (int, error) {
	if m.deleteBatchFunc != nil {
		return m.deleteBatchFunc(ctx, heroID, ids)
	}
	return len(ids), nil
}
func (m *mockPhotoUCForAdmin) Reorder(ctx context.Context, heroID string, ids []string) error {
	if m.reorderFunc != nil {
		return m.reorderFunc(ctx, heroID, ids)
	}
	return nil
}
func (m *mockPhotoUCForAdmin) SetMain(ctx context.Context, heroID, photoID string) error {
	if m.setMainFunc != nil {
		return m.setMainFunc(ctx, heroID, photoID)
	}
	return nil
}
func (m *mockPhotoUCForAdmin) Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Photo{ID: p.PhotoID, HeroID: p.HeroID}, nil
}
func (m *mockPhotoUCForAdmin) ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, heroID)
	}
	return []*domain.Photo{}, nil
}

// --- Setup ---

func discardAdminLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// setupHeroAdminServer создаёт тестовый сервер HeroAdminService с моками.
func setupHeroAdminServer(
	t *testing.T,
	heroUC *mockHeroUCForAdmin,
	queryUC *mockHeroQueryUCForAdmin,
	awardUC *mockAwardUC,
	conflictUC *mockConflictUC,
	locationUC *mockLocationUC,
	sourceUC *mockSourceUCForAdmin,
	relationUC *mockRelationUCForAdmin,
	photoUC *mockPhotoUCForAdmin,
) (emhv1connect.HeroAdminServiceClient, func()) {
	t.Helper()

	if heroUC == nil {
		heroUC = &mockHeroUCForAdmin{}
	}
	if queryUC == nil {
		queryUC = &mockHeroQueryUCForAdmin{}
	}
	if awardUC == nil {
		awardUC = &mockAwardUC{}
	}
	if conflictUC == nil {
		conflictUC = &mockConflictUC{}
	}
	if locationUC == nil {
		locationUC = &mockLocationUC{}
	}
	if sourceUC == nil {
		sourceUC = &mockSourceUCForAdmin{}
	}
	if relationUC == nil {
		relationUC = &mockRelationUCForAdmin{}
	}
	if photoUC == nil {
		photoUC = &mockPhotoUCForAdmin{}
	}

	server := NewHeroAdminServer(
		heroUC, queryUC, awardUC, conflictUC, locationUC, sourceUC, relationUC, photoUC,
		discardAdminLogger(),
	)

	validateInterceptor := validate.NewInterceptor()
	path, handler := emhv1connect.NewHeroAdminServiceHandler(
		server,
		connect.WithInterceptors(validateInterceptor),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	httpServer := httptest.NewServer(mux)

	client := emhv1connect.NewHeroAdminServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() { httpServer.Close() }
	return client, cleanup
}

// --- CreateHero ---

// TestHeroAdminServer_CreateHero_Success проверяет успешное создание с гибкими датами.
func TestHeroAdminServer_CreateHero_Success(t *testing.T) {
	var capturedParams domain.CreateHeroParams
	heroUC := &mockHeroUCForAdmin{
		createFunc: func(ctx context.Context, p domain.CreateHeroParams) (string, error) {
			capturedParams = p
			return newUUID(), nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	resp, err := client.CreateHero(context.Background(), connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Иван",
		LastName:  "Петров",
		Rank:      "Старший лейтенант",
		BirthDate: "1990-05-15",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_DRAFT,
	}))
	if err != nil {
		t.Fatalf("CreateHero failed: %v", err)
	}
	if resp.Msg.Id == "" {
		t.Error("CreateHero returned empty id")
	}
	if capturedParams.FirstName != "Иван" {
		t.Errorf("FirstName = %q", capturedParams.FirstName)
	}
	if capturedParams.BirthDate.Precision != domain.PrecisionExact {
		t.Errorf("BirthDate.Precision = %v, want Exact", capturedParams.BirthDate.Precision)
	}
}

// TestHeroAdminServer_CreateHero_WithFlexibleDates проверяет приоритет info над legacy.
func TestHeroAdminServer_CreateHero_WithFlexibleDates(t *testing.T) {
	var capturedParams domain.CreateHeroParams
	heroUC := &mockHeroUCForAdmin{
		createFunc: func(ctx context.Context, p domain.CreateHeroParams) (string, error) {
			capturedParams = p
			return newUUID(), nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	resp, err := client.CreateHero(context.Background(), connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Иван",
		LastName:  "Петров",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_DRAFT,
		BirthDateInfo: &emhv1.FlexibleDate{
			Precision:   emhv1.DatePrecision_DATE_PRECISION_MONTH,
			DisplayText: "Февраль 1994",
		},
		DeathDateInfo: &emhv1.FlexibleDate{
			Precision:   emhv1.DatePrecision_DATE_PRECISION_DAY_MONTH,
			DisplayText: "28 июля",
		},
	}))
	if err != nil {
		t.Fatalf("CreateHero with flexible dates failed: %v", err)
	}
	if resp.Msg.Id == "" {
		t.Error("empty id")
	}
	if capturedParams.BirthDate.Precision != domain.PrecisionMonth {
		t.Errorf("BirthDate precision = %v, want Month", capturedParams.BirthDate.Precision)
	}
	if capturedParams.BirthDate.DisplayText != "Февраль 1994" {
		t.Errorf("BirthDate display = %q", capturedParams.BirthDate.DisplayText)
	}
	if capturedParams.DeathDate.Precision != domain.PrecisionDayMonth {
		t.Errorf("DeathDate precision = %v, want DayMonth", capturedParams.DeathDate.Precision)
	}
}

// TestHeroAdminServer_CreateHero_DeathBeforeBirth проверяет маппинг бизнес-ошибки.
func TestHeroAdminServer_CreateHero_DeathBeforeBirth(t *testing.T) {
	heroUC := &mockHeroUCForAdmin{
		createFunc: func(ctx context.Context, p domain.CreateHeroParams) (string, error) {
			return "", domain.ErrDeathBeforeBirth
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.CreateHero(context.Background(), connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Иван",
		LastName:  "Петров",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_DRAFT,
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connectErr.Code())
	}
}

// TestHeroAdminServer_CreateHero_ValidationFailure проверяет proto-валидацию (min_len).
func TestHeroAdminServer_CreateHero_ValidationFailure(t *testing.T) {
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.CreateHero(context.Background(), connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "", // min_len = 1
		LastName:  "Петров",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_DRAFT,
	}))
	if err == nil {
		t.Fatal("expected validation error for empty first_name")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connectErr.Code())
	}
}

// --- UpdateHero ---

// TestHeroAdminServer_UpdateHero_Success проверяет частичное обновление.
func TestHeroAdminServer_UpdateHero_Success(t *testing.T) {
	heroID := newUUID()
	var capturedMask []string
	heroUC := &mockHeroUCForAdmin{
		updateFunc: func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
			capturedMask = p.FieldMask
			return &domain.Hero{ID: heroID, FirstName: "Пётр", LastName: "Петров", Status: domain.StatusDraft}, nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	resp, err := client.UpdateHero(context.Background(), connect.NewRequest(&emhv1.UpdateHeroRequest{
		Id:        heroID,
		FirstName: "Пётр",
		FieldMask: []string{"first_name"},
	}))
	if err != nil {
		t.Fatalf("UpdateHero failed: %v", err)
	}
	if resp.Msg.Hero == nil {
		t.Fatal("response hero is nil")
	}
	if len(capturedMask) != 1 || capturedMask[0] != "first_name" {
		t.Errorf("field_mask = %v", capturedMask)
	}
}

// TestHeroAdminServer_UpdateHero_RequiredFieldEmpty проверяет защиту от очистки first_name.
func TestHeroAdminServer_UpdateHero_RequiredFieldEmpty(t *testing.T) {
	heroID := newUUID()
	heroUC := &mockHeroUCForAdmin{
		updateFunc: func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
			return nil, domain.ErrRequiredFieldEmpty
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.UpdateHero(context.Background(), connect.NewRequest(&emhv1.UpdateHeroRequest{
		Id:        heroID,
		FirstName: "", // попытка очистить
		FieldMask: []string{"first_name"},
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connectErr.Code())
	}
}

// TestHeroAdminServer_UpdateHero_InvalidUUID проверяет proto-валидацию UUID.
func TestHeroAdminServer_UpdateHero_InvalidUUID(t *testing.T) {
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.UpdateHero(context.Background(), connect.NewRequest(&emhv1.UpdateHeroRequest{
		Id: "not-a-uuid",
	}))
	if err == nil {
		t.Fatal("expected validation error")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connectErr.Code())
	}
}

// --- DeleteHero ---

// TestHeroAdminServer_DeleteHero_Soft проверяет soft delete (архивация).
func TestHeroAdminServer_DeleteHero_Soft(t *testing.T) {
	heroID := newUUID()
	var capturedHard bool
	heroUC := &mockHeroUCForAdmin{
		deleteFunc: func(ctx context.Context, id string, hard bool) error {
			capturedHard = hard
			return nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	resp, err := client.DeleteHero(context.Background(), connect.NewRequest(&emhv1.DeleteHeroRequest{
		Id:         heroID,
		HardDelete: false,
	}))
	if err != nil {
		t.Fatalf("DeleteHero failed: %v", err)
	}
	if !resp.Msg.Success {
		t.Error("expected success=true")
	}
	if capturedHard {
		t.Error("expected soft delete, got hard")
	}
}

// TestHeroAdminServer_DeleteHero_Hard проверяет hard delete.
func TestHeroAdminServer_DeleteHero_Hard(t *testing.T) {
	heroID := newUUID()
	var capturedHard bool
	heroUC := &mockHeroUCForAdmin{
		deleteFunc: func(ctx context.Context, id string, hard bool) error {
			capturedHard = hard
			return nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.DeleteHero(context.Background(), connect.NewRequest(&emhv1.DeleteHeroRequest{
		Id:         heroID,
		HardDelete: true,
	}))
	if err != nil {
		t.Fatalf("DeleteHero hard failed: %v", err)
	}
	if !capturedHard {
		t.Error("expected hard delete")
	}
}

// TestHeroAdminServer_DeleteHero_NotFound проверяет маппинг ErrNotFound на CodeNotFound.
func TestHeroAdminServer_DeleteHero_NotFound(t *testing.T) {
	heroID := newUUID()
	heroUC := &mockHeroUCForAdmin{
		deleteFunc: func(ctx context.Context, id string, hard bool) error {
			return domain.ErrNotFound
		},
	}
	client, cleanup := setupHeroAdminServer(t, heroUC, nil, nil, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.DeleteHero(context.Background(), connect.NewRequest(&emhv1.DeleteHeroRequest{
		Id: heroID,
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want NotFound", connectErr.Code())
	}
}

// --- Photos ---

// TestHeroAdminServer_AddHeroPhoto_Success проверяет добавление фото.
func TestHeroAdminServer_AddHeroPhoto_Success(t *testing.T) {
	heroID := newUUID()
	var captured domain.AddPhotoParams
	photoUC := &mockPhotoUCForAdmin{
		addFunc: func(ctx context.Context, p domain.AddPhotoParams) (string, error) {
			captured = p
			return newUUID(), nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	_, err := client.AddHeroPhoto(context.Background(), connect.NewRequest(&emhv1.AddHeroPhotoRequest{
		HeroId:      heroID,
		Url:         "https://s3.example.com/photo.jpg",
		Description: "Тестовое фото",
		IsMain:      true,
	}))
	if err != nil {
		t.Fatalf("AddHeroPhoto failed: %v", err)
	}
	if captured.HeroID != heroID {
		t.Errorf("HeroID = %q", captured.HeroID)
	}
	if !captured.IsMain {
		t.Error("IsMain should be true")
	}
}

// TestHeroAdminServer_BatchAddHeroPhotos_Success проверяет пакетное добавление.
func TestHeroAdminServer_BatchAddHeroPhotos_Success(t *testing.T) {
	heroID := newUUID()
	photoUC := &mockPhotoUCForAdmin{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	resp, err := client.BatchAddHeroPhotos(context.Background(), connect.NewRequest(&emhv1.BatchAddHeroPhotosRequest{
		HeroId: heroID,
		Photos: []*emhv1.NewPhoto{
			{Url: "https://s3.example.com/1.jpg"},
			{Url: "https://s3.example.com/2.jpg"},
		},
	}))
	if err != nil {
		t.Fatalf("BatchAddHeroPhotos failed: %v", err)
	}
	if resp.Msg.AddedCount != 2 {
		t.Errorf("added_count = %d, want 2", resp.Msg.AddedCount)
	}
	if len(resp.Msg.Added) != 2 {
		t.Errorf("added len = %d", len(resp.Msg.Added))
	}
}

// TestHeroAdminServer_DeleteHeroPhotos_Success проверяет пакетное удаление.
func TestHeroAdminServer_DeleteHeroPhotos_Success(t *testing.T) {
	heroID := newUUID()
	photoIDs := []string{newUUID(), newUUID()}
	photoUC := &mockPhotoUCForAdmin{
		deleteBatchFunc: func(ctx context.Context, hid string, ids []string) (int, error) {
			if hid != heroID {
				t.Errorf("heroID = %q, want %q", hid, heroID)
			}
			return len(ids), nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	resp, err := client.DeleteHeroPhotos(context.Background(), connect.NewRequest(&emhv1.DeleteHeroPhotosRequest{
		HeroId:   heroID,
		PhotoIds: photoIDs,
	}))
	if err != nil {
		t.Fatalf("DeleteHeroPhotos failed: %v", err)
	}
	if resp.Msg.DeletedCount != 2 {
		t.Errorf("deleted_count = %d, want 2", resp.Msg.DeletedCount)
	}
	if !resp.Msg.Success {
		t.Error("success should be true")
	}
}

// TestHeroAdminServer_SetMainHeroPhoto_Success проверяет назначение главного фото.
func TestHeroAdminServer_SetMainHeroPhoto_Success(t *testing.T) {
	heroID := newUUID()
	photoID := newUUID()
	var capturedHero, capturedPhoto string
	photoUC := &mockPhotoUCForAdmin{
		setMainFunc: func(ctx context.Context, hid, pid string) error {
			capturedHero = hid
			capturedPhoto = pid
			return nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	_, err := client.SetMainHeroPhoto(context.Background(), connect.NewRequest(&emhv1.SetMainHeroPhotoRequest{
		HeroId:  heroID,
		PhotoId: photoID,
	}))
	if err != nil {
		t.Fatalf("SetMainHeroPhoto failed: %v", err)
	}
	if capturedHero != heroID || capturedPhoto != photoID {
		t.Errorf("captured: hero=%q photo=%q", capturedHero, capturedPhoto)
	}
}

// TestHeroAdminServer_ReorderHeroPhotos_Success проверяет переупорядочивание.
func TestHeroAdminServer_ReorderHeroPhotos_Success(t *testing.T) {
	heroID := newUUID()
	photoIDs := []string{newUUID(), newUUID(), newUUID()}
	photoUC := &mockPhotoUCForAdmin{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	_, err := client.ReorderHeroPhotos(context.Background(), connect.NewRequest(&emhv1.ReorderHeroPhotosRequest{
		HeroId:   heroID,
		PhotoIds: photoIDs,
	}))
	if err != nil {
		t.Fatalf("ReorderHeroPhotos failed: %v", err)
	}
}

// TestHeroAdminServer_UpdateHeroPhoto_Success проверяет обновление метаданных фото.
func TestHeroAdminServer_UpdateHeroPhoto_Success(t *testing.T) {
	heroID := newUUID()
	photoID := newUUID()
	photoUC := &mockPhotoUCForAdmin{
		updateFunc: func(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
			return &domain.Photo{ID: photoID, HeroID: heroID, Description: p.Description}, nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, nil, photoUC)
	defer cleanup()

	resp, err := client.UpdateHeroPhoto(context.Background(), connect.NewRequest(&emhv1.UpdateHeroPhotoRequest{
		PhotoId:     photoID,
		HeroId:      heroID,
		Description: "Новое описание",
		FieldMask:   []string{"description"},
	}))
	if err != nil {
		t.Fatalf("UpdateHeroPhoto failed: %v", err)
	}
	if resp.Msg.Photo.Description != "Новое описание" {
		t.Errorf("description = %q", resp.Msg.Photo.Description)
	}
}

// --- Awards ---

// TestHeroAdminServer_AddHeroAward_Success проверяет привязку награды.
func TestHeroAdminServer_AddHeroAward_Success(t *testing.T) {
	heroID := newUUID()
	awardID := newUUID()
	var capturedHero, capturedAward string
	awardUC := &mockAwardUC{
		addFunc: func(ctx context.Context, hid, aid string, date *string, decree string) error {
			capturedHero = hid
			capturedAward = aid
			return nil
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, awardUC, nil, nil, nil, nil, nil)
	defer cleanup()

	resp, err := client.AddHeroAward(context.Background(), connect.NewRequest(&emhv1.AddHeroAwardRequest{
		HeroId:       heroID,
		AwardId:      awardID,
		AwardDate:    "2020-05-09",
		DecreeNumber: "№123",
	}))
	if err != nil {
		t.Fatalf("AddHeroAward failed: %v", err)
	}
	if capturedHero != heroID || capturedAward != awardID {
		t.Errorf("captured: hero=%q award=%q", capturedHero, capturedAward)
	}
	// Echo-паттерн: возвращается запрос.
	if resp.Msg.HeroId != heroID {
		t.Errorf("response hero_id = %q", resp.Msg.HeroId)
	}
}

// TestHeroAdminServer_RemoveHeroAward_Success проверяет отвязку награды.
func TestHeroAdminServer_RemoveHeroAward_Success(t *testing.T) {
	heroID := newUUID()
	awardID := newUUID()
	awardUC := &mockAwardUC{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, awardUC, nil, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.RemoveHeroAward(context.Background(), connect.NewRequest(&emhv1.RemoveHeroAwardRequest{
		HeroId:  heroID,
		AwardId: awardID,
	}))
	if err != nil {
		t.Fatalf("RemoveHeroAward failed: %v", err)
	}
}

// --- Relations ---

// TestHeroAdminServer_AddHeroRelation_SelfRelation проверяет маппинг ErrSelfRelation.
func TestHeroAdminServer_AddHeroRelation_SelfRelation(t *testing.T) {
	heroID := newUUID()
	relationUC := &mockRelationUCForAdmin{
		addFunc: func(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
			return "", domain.ErrSelfRelation
		},
	}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, relationUC, nil)
	defer cleanup()

	_, err := client.AddHeroRelation(context.Background(), connect.NewRequest(&emhv1.AddHeroRelationRequest{
		FromHeroId:   heroID,
		ToHeroId:     heroID,
		RelationType: "comrade",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connectErr.Code())
	}
}

// TestHeroAdminServer_AddHeroRelation_Success проверяет создание связи.
func TestHeroAdminServer_AddHeroRelation_Success(t *testing.T) {
	fromID := newUUID()
	toID := newUUID()
	relationUC := &mockRelationUCForAdmin{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, nil, relationUC, nil)
	defer cleanup()

	_, err := client.AddHeroRelation(context.Background(), connect.NewRequest(&emhv1.AddHeroRelationRequest{
		FromHeroId:   fromID,
		ToHeroId:     toID,
		RelationType: "comrade",
		Description:  "Сослуживцы",
	}))
	if err != nil {
		t.Fatalf("AddHeroRelation failed: %v", err)
	}
}

// --- Sources ---

// TestHeroAdminServer_AddHeroSource_Success проверяет добавление источника.
func TestHeroAdminServer_AddHeroSource_Success(t *testing.T) {
	heroID := newUUID()
	sourceUC := &mockSourceUCForAdmin{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, nil, sourceUC, nil, nil)
	defer cleanup()

	_, err := client.AddHeroSource(context.Background(), connect.NewRequest(&emhv1.AddHeroSourceRequest{
		HeroId:     heroID,
		Url:        "https://example.com/article",
		Title:      "Статья о герое",
		SourceType: "website",
	}))
	if err != nil {
		t.Fatalf("AddHeroSource failed: %v", err)
	}
}

// --- Conflicts ---

// TestHeroAdminServer_AddHeroConflict_Success проверяет привязку конфликта.
func TestHeroAdminServer_AddHeroConflict_Success(t *testing.T) {
	heroID := newUUID()
	conflictID := newUUID()
	conflictUC := &mockConflictUC{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, conflictUC, nil, nil, nil, nil)
	defer cleanup()

	_, err := client.AddHeroConflict(context.Background(), connect.NewRequest(&emhv1.AddHeroConflictRequest{
		HeroId:           heroID,
		ConflictId:       conflictID,
		SpecificLocation: "Сталинград",
		RankAtConflict:   "Младший сержант",
	}))
	if err != nil {
		t.Fatalf("AddHeroConflict failed: %v", err)
	}
}

// --- Locations ---

// TestHeroAdminServer_AddHeroLocation_Success проверяет привязку локации.
func TestHeroAdminServer_AddHeroLocation_Success(t *testing.T) {
	heroID := newUUID()
	locationID := newUUID()
	locationUC := &mockLocationUC{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, locationUC, nil, nil, nil)
	defer cleanup()

	_, err := client.AddHeroLocation(context.Background(), connect.NewRequest(&emhv1.AddHeroLocationRequest{
		HeroId:     heroID,
		LocationId: locationID,
		Type:       emhv1.HeroLocationType_HERO_LOCATION_TYPE_BIRTH,
	}))
	if err != nil {
		t.Fatalf("AddHeroLocation failed: %v", err)
	}
}

// TestHeroAdminServer_RemoveHeroLocation_Success проверяет удаление связи с локацией.
func TestHeroAdminServer_RemoveHeroLocation_Success(t *testing.T) {
	heroID := newUUID()
	locationID := newUUID()
	locationUC := &mockLocationUC{}
	client, cleanup := setupHeroAdminServer(t, nil, nil, nil, nil, locationUC, nil, nil, nil)
	defer cleanup()

	_, err := client.RemoveHeroLocation(context.Background(), connect.NewRequest(&emhv1.RemoveHeroLocationRequest{
		HeroId:     heroID,
		LocationId: locationID,
		Type:       emhv1.HeroLocationType_HERO_LOCATION_TYPE_BIRTH,
	}))
	if err != nil {
		t.Fatalf("RemoveHeroLocation failed: %v", err)
	}
}
