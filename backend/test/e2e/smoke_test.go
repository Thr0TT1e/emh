//go:build integration

package e2e

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	v1 "codeberg.org/Thr0TT1e/emh/backend/internal/delivery/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// mockMediaStorage in-memory реализация MediaStorage для E2E тестов.
type mockMediaStorage struct{}

func (m *mockMediaStorage) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	return "https://mock-s3.example.com/presigned/" + key, nil
}

func (m *mockMediaStorage) PublicURL(key string) string {
	return "https://mock-s3.example.com/public/" + key
}

func (m *mockMediaStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockMediaStorage) KeyFromPublicURL(rawURL string) (string, error) {
	return "extracted-key", nil
}

func (m *mockMediaStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockMediaStorage) ListObjects(ctx context.Context) ([]domain.ObjectInfo, error) {
	return []domain.ObjectInfo{}, nil
}

func (m *mockMediaStorage) EnsureBucket(ctx context.Context) error {
	return nil
}

// Upload реализует новый метод интерфейса (защита от decompression bomb).
// В E2E тестах не используется (thumbnailEnqueuer — мок), поэтому просто возвращает nil.
func (m *mockMediaStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	return nil
}

// ListObjectsPaged — заглушка для удовлетворения интерфейса domain.MediaStorage.
// E2E тесты не используют пагинацию S3 (тестируют только Photo и Media usecase'ы).
func (m *mockMediaStorage) ListObjectsPaged(ctx context.Context, marker string, limit int) ([]domain.ObjectInfo, string, error) {
	return nil, "", nil
}

// mockThumbnailEnqueuer мок для ThumbnailWorker (не нужен в smoke test).
type mockThumbnailEnqueuer struct{}

func (m *mockThumbnailEnqueuer) Enqueue(task usecase.ThumbnailTask) {}

// setupTestServer создаёт HTTP-сервер со всеми сервисами для E2E тестирования.
func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	// 1. Создаём пул подключений к тестовой БД
	pool, cleanupPool := testutil.NewTestPool(t)

	// Очищаем таблицы ПЕРЕД тестом
	testutil.CleanupAllTables(t, pool)

	// 2. Создаём все репозитории
	heroRepo := pg.NewHeroRepository(pool)
	photoRepo := pg.NewPhotoRepository(pool)
	heroAwardRepo := pg.NewHeroAwardRepository(pool)
	heroConflictRepo := pg.NewHeroConflictRepository(pool)
	heroLocationRepo := pg.NewHeroLocationRepository(pool)
	heroSourceRepo := pg.NewHeroSourceRepository(pool)
	heroRelationRepo := pg.NewHeroRelationRepository(pool)
	awardRepo := pg.NewAwardRepository(pool)
	conflictRepo := pg.NewConflictRepository(pool)
	locationRepo := pg.NewLocationRepository(pool)
	submissionRepo := pg.NewSubmissionRepository(pool)
	refreshTokenRepo := pg.NewRefreshTokenRepository(pool)

	// 3. Создаём usecase'ы
	heroUC := usecase.NewHeroUseCase(heroRepo)
	heroQueryUC := usecase.NewHeroQueryUseCase(
		heroRepo, photoRepo, heroAwardRepo, heroConflictRepo,
		heroLocationRepo, heroSourceRepo, heroRelationRepo,
	)

	awardUC := usecase.NewAwardUseCase(awardRepo)
	heroAwardUC := usecase.NewHeroAwardUseCase(heroAwardRepo, heroRepo, awardRepo)

	conflictUC := usecase.NewConflictUseCase(conflictRepo)
	heroConflictUC := usecase.NewHeroConflictUseCase(heroConflictRepo, heroRepo, conflictRepo)

	locationUC := usecase.NewLocationUseCase(locationRepo)
	heroLocationUC := usecase.NewHeroLocationUseCase(heroLocationRepo, heroRepo, locationRepo)

	heroSourceUC := usecase.NewHeroSourceUseCase(heroSourceRepo, heroRepo)
	heroRelationUC := usecase.NewHeroRelationUseCase(heroRelationRepo, heroRepo)

	mediaStorage := &mockMediaStorage{}
	thumbnailEnqueuer := &mockThumbnailEnqueuer{}
	logger := discardLogger()
	photoUC := usecase.NewPhotoUseCase(photoRepo, heroRepo, mediaStorage, thumbnailEnqueuer, logger)

	mediaUC := usecase.NewMediaUseCase(mediaStorage, 15*time.Minute)
	submissionUC := usecase.NewSubmissionUseCase(submissionRepo, heroRepo)

	// 4. Создаём тестового админа
	passwordHash, err := auth.HashPassword("test-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	admins := []domain.AdminUser{
		{
			Username:     "admin",
			PasswordHash: passwordHash,
			Role:         "admin",
		},
	}

	jwtSecret := "test-jwt-secret-at-least-32-chars-long-12345"
	authUC := usecase.NewAuthUseCase(
		admins,
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
		refreshTokenRepo,
		logger,
	)

	// 5. Создаём серверы
	heroPublicServer := v1.NewHeroServer(heroUC, heroQueryUC, heroSourceUC, heroRelationUC, logger)
	heroAdminServer := v1.NewHeroAdminServer(
		heroUC, heroQueryUC, heroAwardUC, heroConflictUC,
		heroLocationUC, heroSourceUC, heroRelationUC, photoUC, logger,
	)
	awardPublicServer := v1.NewAwardServer(awardUC, logger)
	awardAdminServer := v1.NewAwardAdminServer(awardUC, logger)
	conflictPublicServer := v1.NewConflictServer(conflictUC, logger)
	conflictAdminServer := v1.NewConflictAdminServer(conflictUC, logger)
	locationPublicServer := v1.NewLocationServer(locationUC, logger)
	locationAdminServer := v1.NewLocationAdminServer(locationUC, logger)
	mediaServer := v1.NewMediaServer(mediaUC, logger)
	submissionPublicServer := v1.NewSubmissionServer(submissionUC, logger)
	authServer := v1.NewAuthServer(authUC, logger)

	// 6. Создаём interceptors
	validateInterceptor := validate.NewInterceptor()
	authInterceptor := interceptor.NewAuthInterceptor(
		jwtSecret,
		[]string{},
		nil,
		nil,
		nil,
		nil,
		"test-ip-pepper-for-e2e-at-least-16-chars",
		logger,
	)

	// 7. Регистрируем все сервисы
	mux := http.NewServeMux()

	publicPath, publicHandler := emhv1connect.NewHeroServiceHandler(
		heroPublicServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(publicPath, publicHandler)

	adminPath, adminHandler := emhv1connect.NewHeroAdminServiceHandler(
		heroAdminServer, connect.WithInterceptors(authInterceptor, validateInterceptor),
	)
	mux.Handle(adminPath, adminHandler)

	awardPublicPath, awardPublicHandler := emhv1connect.NewAwardServiceHandler(
		awardPublicServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(awardPublicPath, awardPublicHandler)

	awardAdminPath, awardAdminHandler := emhv1connect.NewAwardAdminServiceHandler(
		awardAdminServer, connect.WithInterceptors(authInterceptor, validateInterceptor),
	)
	mux.Handle(awardAdminPath, awardAdminHandler)

	conflictPublicPath, conflictPublicHandler := emhv1connect.NewConflictServiceHandler(
		conflictPublicServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(conflictPublicPath, conflictPublicHandler)

	conflictAdminPath, conflictAdminHandler := emhv1connect.NewConflictAdminServiceHandler(
		conflictAdminServer, connect.WithInterceptors(authInterceptor, validateInterceptor),
	)
	mux.Handle(conflictAdminPath, conflictAdminHandler)

	locationPublicPath, locationPublicHandler := emhv1connect.NewLocationServiceHandler(
		locationPublicServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(locationPublicPath, locationPublicHandler)

	locationAdminPath, locationAdminHandler := emhv1connect.NewLocationAdminServiceHandler(
		locationAdminServer, connect.WithInterceptors(authInterceptor, validateInterceptor),
	)
	mux.Handle(locationAdminPath, locationAdminHandler)

	mediaPath, mediaHandler := emhv1connect.NewMediaServiceHandler(
		mediaServer, connect.WithInterceptors(authInterceptor, validateInterceptor),
	)
	mux.Handle(mediaPath, mediaHandler)

	submissionPublicPath, submissionPublicHandler := emhv1connect.NewSubmissionServiceHandler(
		submissionPublicServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(submissionPublicPath, submissionPublicHandler)

	authPath, authHandler := emhv1connect.NewAuthServiceHandler(
		authServer, connect.WithInterceptors(validateInterceptor),
	)
	mux.Handle(authPath, authHandler)

	// 8. Создаём тестовый HTTP-сервер
	server := httptest.NewServer(mux)

	cleanup := func() {
		server.Close()
		cleanupPool()
	}

	return server, cleanup
}

// discardLogger возвращает логгер, который ничего не пишет.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- Тесты ---

// TestE2E_AuthFlow проверяет полный цикл аутентификации: Login → Refresh → Logout.
func TestE2E_AuthFlow(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)

	// 1. Login
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if loginResp.Msg.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if loginResp.Msg.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if loginResp.Msg.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", loginResp.Msg.TokenType)
	}

	refreshToken := loginResp.Msg.RefreshToken

	// 2. Refresh
	refreshResp, err := authClient.Refresh(ctx, connect.NewRequest(&emhv1.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}))
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if refreshResp.Msg.AccessToken == "" {
		t.Error("new AccessToken should not be empty")
	}
	if refreshResp.Msg.RefreshToken == "" {
		t.Error("new RefreshToken should not be empty")
	}

	if refreshToken == refreshResp.Msg.RefreshToken {
		t.Error("new refresh token should be different from old")
	}

	// 3. Logout
	logoutResp, err := authClient.Logout(ctx, connect.NewRequest(&emhv1.LogoutRequest{
		RefreshToken: refreshResp.Msg.RefreshToken,
	}))
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	if !logoutResp.Msg.Success {
		t.Error("Logout should return success=true")
	}
}

// TestE2E_HeroCRUD проверяет полный CRUD цикл для героя.
func TestE2E_HeroCRUD(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	heroAdminClient := emhv1connect.NewHeroAdminServiceClient(http.DefaultClient, server.URL)

	// 1. Login
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	// 2. CreateHero
	createReq := connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Иван",
		LastName:  "Петров",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED,
	})
	createReq.Header().Set("Authorization", "Bearer "+accessToken)

	createResp, err := heroAdminClient.CreateHero(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateHero failed: %v", err)
	}

	heroID := createResp.Msg.Id
	if heroID == "" {
		t.Fatal("CreateHero returned empty ID")
	}

	// 3. UpdateHero
	updateReq := connect.NewRequest(&emhv1.UpdateHeroRequest{
		Id:        heroID,
		FirstName: "Пётр",
		Rank:      "Капитан",
		FieldMask: []string{"first_name", "rank"},
	})
	updateReq.Header().Set("Authorization", "Bearer "+accessToken)

	updateResp, err := heroAdminClient.UpdateHero(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateHero failed: %v", err)
	}

	if updateResp.Msg.Hero == nil {
		t.Fatal("UpdateHero returned nil hero")
	}
	if updateResp.Msg.Hero.Summary.FirstName != "Пётр" {
		t.Errorf("UpdateHero FirstName = %q, want Пётр", updateResp.Msg.Hero.Summary.FirstName)
	}

	// 4. DeleteHero (soft delete)
	deleteReq := connect.NewRequest(&emhv1.DeleteHeroRequest{
		Id:         heroID,
		HardDelete: false,
	})
	deleteReq.Header().Set("Authorization", "Bearer "+accessToken)

	deleteResp, err := heroAdminClient.DeleteHero(ctx, deleteReq)
	if err != nil {
		t.Fatalf("DeleteHero failed: %v", err)
	}

	if !deleteResp.Msg.Success {
		t.Error("DeleteHero should return success=true")
	}
}

// TestE2E_HeroWithRelations проверяет создание героя с привязками.
func TestE2E_HeroWithRelations(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	heroAdminClient := emhv1connect.NewHeroAdminServiceClient(http.DefaultClient, server.URL)
	awardAdminClient := emhv1connect.NewAwardAdminServiceClient(http.DefaultClient, server.URL)
	conflictAdminClient := emhv1connect.NewConflictAdminServiceClient(http.DefaultClient, server.URL)
	locationAdminClient := emhv1connect.NewLocationAdminServiceClient(http.DefaultClient, server.URL)
	heroPublicClient := emhv1connect.NewHeroServiceClient(http.DefaultClient, server.URL)

	// 1. Login
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	// 2. Создаём справочники
	awardReq := connect.NewRequest(&emhv1.CreateAwardRequest{
		Name:        "Орден Мужества",
		Description: "Государственная награда РФ",
		ImageUrl:    "https://example.com/awards/order-of-courage.jpg",
	})
	awardReq.Header().Set("Authorization", "Bearer "+accessToken)
	awardResp, err := awardAdminClient.CreateAward(ctx, awardReq)
	if err != nil {
		t.Fatalf("CreateAward failed: %v", err)
	}
	awardID := awardResp.Msg.Id

	conflictReq := connect.NewRequest(&emhv1.CreateConflictRequest{
		Name:        "Специальная военная операция",
		Description: "СВО 2022-2026",
		Type:        emhv1.ConflictType_CONFLICT_TYPE_SPECIAL_OPERATION,
		StartDate:   "2022-02-24",
	})
	conflictReq.Header().Set("Authorization", "Bearer "+accessToken)
	conflictResp, err := conflictAdminClient.CreateConflict(ctx, conflictReq)
	if err != nil {
		t.Fatalf("CreateConflict failed: %v", err)
	}
	conflictID := conflictResp.Msg.Id

	locationReq := connect.NewRequest(&emhv1.CreateLocationRequest{
		Name: "Москва",
		Type: emhv1.LocationType_LOCATION_TYPE_CITY,
	})
	locationReq.Header().Set("Authorization", "Bearer "+accessToken)
	locationResp, err := locationAdminClient.CreateLocation(ctx, locationReq)
	if err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	locationID := locationResp.Msg.Id

	// 3. Создаём героя
	createHeroReq := connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Александр",
		LastName:  "Смирнов",
		Rank:      "Майор",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED,
	})
	createHeroReq.Header().Set("Authorization", "Bearer "+accessToken)
	heroResp, err := heroAdminClient.CreateHero(ctx, createHeroReq)
	if err != nil {
		t.Fatalf("CreateHero failed: %v", err)
	}
	heroID := heroResp.Msg.Id

	// 4. Привязываем награду
	addAwardReq := connect.NewRequest(&emhv1.AddHeroAwardRequest{
		HeroId:       heroID,
		AwardId:      awardID,
		AwardDate:    "2023-06-15",
		DecreeNumber: "123",
	})
	addAwardReq.Header().Set("Authorization", "Bearer "+accessToken)
	_, err = heroAdminClient.AddHeroAward(ctx, addAwardReq)
	if err != nil {
		t.Fatalf("AddHeroAward failed: %v", err)
	}

	// 5. Привязываем конфликт
	addConflictReq := connect.NewRequest(&emhv1.AddHeroConflictRequest{
		HeroId:           heroID,
		ConflictId:       conflictID,
		SpecificLocation: "Запорожское направление",
		RankAtConflict:   "Майор",
	})
	addConflictReq.Header().Set("Authorization", "Bearer "+accessToken)
	_, err = heroAdminClient.AddHeroConflict(ctx, addConflictReq)
	if err != nil {
		t.Fatalf("AddHeroConflict failed: %v", err)
	}

	// 6. Привязываем локацию
	addLocationReq := connect.NewRequest(&emhv1.AddHeroLocationRequest{
		HeroId:     heroID,
		LocationId: locationID,
		Type:       emhv1.HeroLocationType_HERO_LOCATION_TYPE_BIRTH,
	})
	addLocationReq.Header().Set("Authorization", "Bearer "+accessToken)
	_, err = heroAdminClient.AddHeroLocation(ctx, addLocationReq)
	if err != nil {
		t.Fatalf("AddHeroLocation failed: %v", err)
	}

	// 7. Получаем героя и проверяем связи
	getResp, err := heroPublicClient.GetHero(ctx, connect.NewRequest(&emhv1.GetHeroRequest{
		Id: heroID,
	}))
	if err != nil {
		t.Fatalf("GetHero failed: %v", err)
	}

	if getResp.Msg.Hero == nil {
		t.Fatal("GetHero returned nil hero")
	}

	if len(getResp.Msg.Hero.Awards) != 1 {
		t.Errorf("Awards count = %d, want 1", len(getResp.Msg.Hero.Awards))
	}

	if len(getResp.Msg.Hero.Conflicts) != 1 {
		t.Errorf("Conflicts count = %d, want 1", len(getResp.Msg.Hero.Conflicts))
	}

	if len(getResp.Msg.Hero.Locations) != 1 {
		t.Errorf("Locations count = %d, want 1", len(getResp.Msg.Hero.Locations))
	}
}

// TestE2E_MediaUpload проверяет получение presigned URL.
func TestE2E_MediaUpload(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	mediaClient := emhv1connect.NewMediaServiceClient(http.DefaultClient, server.URL)

	// 1. Login — MediaService теперь за auth
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	// 2. GetUploadUrl (с JWT)
	getReq := connect.NewRequest(&emhv1.GetUploadUrlRequest{
		Type:        emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Filename:    "hero-photo.jpg",
		ContentType: "image/jpeg",
	})
	getReq.Header().Set("Authorization", "Bearer "+accessToken)

	singleResp, err := mediaClient.GetUploadUrl(ctx, getReq)
	if err != nil {
		t.Fatalf("GetUploadUrl failed: %v", err)
	}
	if singleResp.Msg.UploadUrl == "" {
		t.Error("UploadUrl should not be empty")
	}
	if singleResp.Msg.PublicUrl == "" {
		t.Error("PublicUrl should not be empty")
	}

	// 3. BatchGetUploadUrls (с JWT)
	batchReq := connect.NewRequest(&emhv1.BatchGetUploadUrlsRequest{
		Type: emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Files: []*emhv1.FileToUpload{
			{Filename: "photo1.jpg", ContentType: "image/jpeg"},
			{Filename: "photo2.png", ContentType: "image/png"},
		},
	})
	batchReq.Header().Set("Authorization", "Bearer "+accessToken)

	batchResp, err := mediaClient.BatchGetUploadUrls(ctx, batchReq)
	if err != nil {
		t.Fatalf("BatchGetUploadUrls failed: %v", err)
	}
	if batchResp.Msg.Count != 2 {
		t.Errorf("Batch count = %d, want 2", batchResp.Msg.Count)
	}
}

// TestE2E_Submission проверяет создание пользовательской заявки.
func TestE2E_Submission(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	submissionClient := emhv1connect.NewSubmissionServiceClient(http.DefaultClient, server.URL)
	heroAdminClient := emhv1connect.NewHeroAdminServiceClient(http.DefaultClient, server.URL)

	// 1. Создаём героя
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	createHeroReq := connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Тестовый",
		LastName:  "Герой",
		Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED,
	})
	createHeroReq.Header().Set("Authorization", "Bearer "+accessToken)
	heroResp, err := heroAdminClient.CreateHero(ctx, createHeroReq)
	if err != nil {
		t.Fatalf("CreateHero failed: %v", err)
	}
	heroID := heroResp.Msg.Id

	// 2. Создаём заявку
	submissionResp, err := submissionClient.CreateSubmission(ctx, connect.NewRequest(&emhv1.CreateSubmissionRequest{
		SubmitterName:  "Иван Иванов",
		SubmitterEmail: "ivan@example.com",
		TargetHeroId:   heroID,
		PayloadJson:    `{"correction":"Исправление даты"}`,
	}))
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}

	if submissionResp.Msg.SubmissionId == "" {
		t.Error("SubmissionId should not be empty")
	}
}

// TestE2E_FlexibleDate проверяет работу с гибкими датами.
func TestE2E_FlexibleDate(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	heroAdminClient := emhv1connect.NewHeroAdminServiceClient(http.DefaultClient, server.URL)
	heroPublicClient := emhv1connect.NewHeroServiceClient(http.DefaultClient, server.URL)

	// 1. Login
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	// 2. Создаём героя с гибкой датой
	createReq := connect.NewRequest(&emhv1.CreateHeroRequest{
		FirstName: "Василий",
		LastName:  "Кузнецов",
		BirthDateInfo: &emhv1.FlexibleDate{
			Precision:   emhv1.DatePrecision_DATE_PRECISION_YEAR,
			DisplayText: "1980",
		},
		Status: emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED,
	})
	createReq.Header().Set("Authorization", "Bearer "+accessToken)

	createResp, err := heroAdminClient.CreateHero(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateHero with flexible date failed: %v", err)
	}
	heroID := createResp.Msg.Id

	// 3. Получаем героя
	getResp, err := heroPublicClient.GetHero(ctx, connect.NewRequest(&emhv1.GetHeroRequest{
		Id: heroID,
	}))
	if err != nil {
		t.Fatalf("GetHero failed: %v", err)
	}

	if getResp.Msg.Hero.Summary.BirthDateInfo == nil {
		t.Fatal("BirthDateInfo is nil")
	}
	if getResp.Msg.Hero.Summary.BirthDateInfo.Precision != emhv1.DatePrecision_DATE_PRECISION_YEAR {
		t.Errorf("BirthDateInfo.Precision = %v, want YEAR", getResp.Msg.Hero.Summary.BirthDateInfo.Precision)
	}
}

// TestE2E_ListHeroes проверяет список героев с пагинацией.
func TestE2E_ListHeroes(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	authClient := emhv1connect.NewAuthServiceClient(http.DefaultClient, server.URL)
	heroAdminClient := emhv1connect.NewHeroAdminServiceClient(http.DefaultClient, server.URL)
	heroPublicClient := emhv1connect.NewHeroServiceClient(http.DefaultClient, server.URL)

	// 1. Login
	loginResp, err := authClient.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "test-password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	accessToken := loginResp.Msg.AccessToken

	// 2. Создаём несколько героев
	for i := 0; i < 5; i++ {
		createReq := connect.NewRequest(&emhv1.CreateHeroRequest{
			FirstName: "Герой" + string(rune('0'+i)),
			LastName:  "Тестовый",
			Status:    emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED,
		})
		createReq.Header().Set("Authorization", "Bearer "+accessToken)
		_, err := heroAdminClient.CreateHero(ctx, createReq)
		if err != nil {
			t.Fatalf("CreateHero %d failed: %v", i, err)
		}
	}

	// 3. Получаем список
	listResp, err := heroPublicClient.ListHeroes(ctx, connect.NewRequest(&emhv1.ListHeroesRequest{
		Pagination: &emhv1.PaginationRequest{
			PageSize: 3,
		},
	}))
	if err != nil {
		t.Fatalf("ListHeroes failed: %v", err)
	}

	if len(listResp.Msg.Heroes) != 3 {
		t.Errorf("Heroes count = %d, want 3", len(listResp.Msg.Heroes))
	}
	if listResp.Msg.Pagination.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", listResp.Msg.Pagination.TotalCount)
	}
}

// TestE2E_MediaUpload_AnonymousRejected проверяет, что без JWT
// MediaService возвращает Unauthenticated.
func TestE2E_MediaUpload_AnonymousRejected(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mediaClient := emhv1connect.NewMediaServiceClient(http.DefaultClient, server.URL)

	// Без Authorization header
	_, err := mediaClient.GetUploadUrl(ctx, connect.NewRequest(&emhv1.GetUploadUrlRequest{
		Type:        emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Filename:    "hero-photo.jpg",
		ContentType: "image/jpeg",
	}))
	if err == nil {
		t.Fatal("expected Unauthenticated for anonymous MediaService access")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}
}
