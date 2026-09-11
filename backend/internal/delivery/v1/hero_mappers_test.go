package v1

import (
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
)

// --- Хелперы ---

// ptrStr возвращает указатель на строку.
func ptrStr(s string) *string { return &s }

// --- mapFaceBoxToProto ---

// TestMapFaceBoxToProto_Nil возвращает nil для nil-входа.
func TestMapFaceBoxToProto_Nil(t *testing.T) {
	if got := mapFaceBoxToProto(nil); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// TestMapFaceBoxToProto_Values проверяет корректность передачи координат.
func TestMapFaceBoxToProto_Values(t *testing.T) {
	fb := &domain.FaceBox{X: 100, Y: 50, Width: 200, Height: 300}
	got := mapFaceBoxToProto(fb)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.X != 100 || got.Y != 50 || got.Width != 200 || got.Height != 300 {
		t.Errorf("face_box values mismatch: %+v", got)
	}
}

// --- mapPhotoToProto ---

// TestMapPhotoToProto_Full проверяет маппинг фото со всеми полями.
func TestMapPhotoToProto_Full(t *testing.T) {
	p := &domain.Photo{
		ID:           "photo-1",
		URL:          "https://s3.example.com/photos/1.webp",
		ThumbnailURL: "https://s3.example.com/thumbnails/1.webp",
		Description:  "Герой на параде",
		SortOrder:    2,
		IsMain:       true,
		FaceBox:      &domain.FaceBox{X: 10, Y: 20, Width: 100, Height: 150},
	}

	got := mapPhotoToProto(p)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Id != "photo-1" {
		t.Errorf("Id = %q, want %q", got.Id, "photo-1")
	}
	if got.Url != p.URL {
		t.Errorf("Url = %q, want %q", got.Url, p.URL)
	}
	if got.ThumbnailUrl != p.ThumbnailURL {
		t.Errorf("ThumbnailUrl = %q, want %q", got.ThumbnailUrl, p.ThumbnailURL)
	}
	if got.Description != "Герой на параде" {
		t.Errorf("Description = %q", got.Description)
	}
	if got.SortOrder != 2 {
		t.Errorf("SortOrder = %d, want 2", got.SortOrder)
	}
	if !got.IsMain {
		t.Error("IsMain should be true")
	}
	if got.FaceBox == nil {
		t.Fatal("FaceBox is nil")
	}
	if got.FaceBox.Width != 100 {
		t.Errorf("FaceBox.Width = %d, want 100", got.FaceBox.Width)
	}
}

// TestMapPhotoToProto_NoFaceBox проверяет, что nil face_box даёт nil в proto.
func TestMapPhotoToProto_NoFaceBox(t *testing.T) {
	p := &domain.Photo{
		ID:        "photo-2",
		URL:       "https://s3.example.com/2.webp",
		SortOrder: 0,
		IsMain:    false,
	}

	got := mapPhotoToProto(p)
	if got.FaceBox != nil {
		t.Errorf("expected nil FaceBox, got %+v", got.FaceBox)
	}
}

// --- mapHeroAwardToProto ---

// TestMapHeroAwardToProto_WithDate проверяет маппинг с датой награждения.
func TestMapHeroAwardToProto_WithDate(t *testing.T) {
	date := utcDate(2020, 5, 9)
	a := &domain.HeroAward{
		AwardID:      "award-1",
		AwardName:    "Герой России",
		AwardDate:    &date,
		DecreeNumber: "Указ №123",
	}

	got := mapHeroAwardToProto(a)
	if got.AwardId != "award-1" {
		t.Errorf("AwardId = %q", got.AwardId)
	}
	if got.AwardName != "Герой России" {
		t.Errorf("AwardName = %q", got.AwardName)
	}
	if got.DecreeNumber != "Указ №123" {
		t.Errorf("DecreeNumber = %q", got.DecreeNumber)
	}
	if got.AwardDate == nil {
		t.Fatal("AwardDate is nil")
	}
}

// TestMapHeroAwardToProto_WithoutDate проверяет маппинг без даты.
func TestMapHeroAwardToProto_WithoutDate(t *testing.T) {
	a := &domain.HeroAward{
		AwardID:   "award-2",
		AwardName: "Орден Мужества",
	}

	got := mapHeroAwardToProto(a)
	if got.AwardDate != nil {
		t.Errorf("expected nil AwardDate, got %v", got.AwardDate)
	}
}

// --- mapHeroConflictToProto ---

// TestMapHeroConflictToProto проверяет маппинг конфликта героя.
func TestMapHeroConflictToProto(t *testing.T) {
	c := &domain.HeroConflict{
		ConflictID:       "conflict-1",
		ConflictName:     "Великая Отечественная война",
		SpecificLocation: "Сталинград",
		RankAtConflict:   "Младший сержант",
	}

	got := mapHeroConflictToProto(c)
	if got.ConflictId != "conflict-1" {
		t.Errorf("ConflictId = %q", got.ConflictId)
	}
	if got.ConflictName != "Великая Отечественная война" {
		t.Errorf("ConflictName = %q", got.ConflictName)
	}
	if got.SpecificLocation != "Сталинград" {
		t.Errorf("SpecificLocation = %q", got.SpecificLocation)
	}
	if got.RankAtConflict != "Младший сержант" {
		t.Errorf("RankAtConflict = %q", got.RankAtConflict)
	}
}

// --- mapHeroLocationToProto ---

// TestMapHeroLocationToProto проверяет маппинг локации героя с вложенной Location.
func TestMapHeroLocationToProto(t *testing.T) {
	hl := &domain.HeroLocation{
		LocationID: "loc-1",
		Type:       1, // BIRTH_PLACE или аналог
		Location: &domain.Location{
			ID:        "loc-1",
			Name:      "Москва",
			Type:      1,
			Latitude:  ptrFloat64(55.7558),
			Longitude: ptrFloat64(37.6173),
		},
	}

	got := mapHeroLocationToProto(hl)
	if got.LocationId != "loc-1" {
		t.Errorf("LocationId = %q", got.LocationId)
	}
	if got.Location == nil {
		t.Fatal("Location is nil")
	}
	if got.Location.Name != "Москва" {
		t.Errorf("Location.Name = %q", got.Location.Name)
	}
}

// --- mapHeroSourceToProto ---

// TestMapHeroSourceToProto проверяет маппинг источника.
func TestMapHeroSourceToProto(t *testing.T) {
	s := &domain.HeroSource{
		ID:         "source-1",
		URL:        "https://example.com/article",
		Title:      "Статья о герое",
		SourceType: "website",
		Excerpt:    "Фрагмент текста...",
	}

	got := mapHeroSourceToProto(s)
	if got.Id != "source-1" {
		t.Errorf("Id = %q", got.Id)
	}
	if got.Url != "https://example.com/article" {
		t.Errorf("Url = %q", got.Url)
	}
	if got.Title != "Статья о герое" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.SourceType != "website" {
		t.Errorf("SourceType = %q", got.SourceType)
	}
	if got.Excerpt != "Фрагмент текста..." {
		t.Errorf("Excerpt = %q", got.Excerpt)
	}
}

// --- mapHeroRelationToProto ---

// TestMapHeroRelationToProto проверяет маппинг связи героев.
func TestMapHeroRelationToProto(t *testing.T) {
	r := &domain.HeroRelation{
		ID:              "rel-1",
		FromHeroID:      "hero-1",
		ToHeroID:        "hero-2",
		RelationType:    "comrade",
		Description:     "Служили вместе",
		RelatedHeroName: "Пётр Иванов",
	}

	got := mapHeroRelationToProto(r)
	if got.Id != "rel-1" {
		t.Errorf("Id = %q", got.Id)
	}
	if got.FromHeroId != "hero-1" {
		t.Errorf("FromHeroId = %q", got.FromHeroId)
	}
	if got.ToHeroId != "hero-2" {
		t.Errorf("ToHeroId = %q", got.ToHeroId)
	}
	if got.RelationType != "comrade" {
		t.Errorf("RelationType = %q", got.RelationType)
	}
	if got.RelatedHeroName != "Пётр Иванов" {
		t.Errorf("RelatedHeroName = %q", got.RelatedHeroName)
	}
}

// --- mapHeroToSummary ---

// TestMapHeroToSummary_AllFields проверяет полный маппинг с денормализацией.
func TestMapHeroToSummary_AllFields(t *testing.T) {
	birthAnchor := utcDate(1990, 5, 15)
	deathAnchor := utcDate(2020, 8, 20) // теперь используется

	h := &domain.Hero{
		ID:            "hero-1",
		FirstName:     "Иван",
		LastName:      "Петров",
		MiddleName:    "Сергеевич",
		Rank:          "Старший лейтенант",
		ShortBio:      "Краткая биография",
		Nickname:      "Тридцатый",
		Unit:          "в/ч 12345",
		ServiceBranch: "ВДВ",
		BirthDate:     domain.FlexibleDate{Precision: domain.PrecisionExact, Anchor: &birthAnchor},
		DeathDate: domain.FlexibleDate{
			Precision:   domain.PrecisionMonth,
			Anchor:      &deathAnchor,
			DisplayText: "Август 2020",
		},
		MainPhotoURL:     ptrStr("https://s3.example.com/main.jpg"),
		MainThumbnailURL: ptrStr("https://s3.example.com/thumb.jpg"),
		AwardNames:       []string{"Герой России", "Орден Мужества"},
	}

	got := mapHeroToSummary(h)

	// Базовые поля.
	if got.Id != "hero-1" {
		t.Errorf("Id = %q", got.Id)
	}
	if got.FirstName != "Иван" {
		t.Errorf("FirstName = %q", got.FirstName)
	}
	if got.LastName != "Петров" {
		t.Errorf("LastName = %q", got.LastName)
	}
	if got.MiddleName != "Сергеевич" {
		t.Errorf("MiddleName = %q", got.MiddleName)
	}
	if got.Rank != "Старший лейтенант" {
		t.Errorf("Rank = %q", got.Rank)
	}
	if got.Nickname != "Тридцатый" {
		t.Errorf("Nickname = %q", got.Nickname)
	}

	// Денормализация.
	if got.MainPhotoUrl != "https://s3.example.com/main.jpg" {
		t.Errorf("MainPhotoUrl = %q", got.MainPhotoUrl)
	}
	if got.MainThumbnailUrl != "https://s3.example.com/thumb.jpg" {
		t.Errorf("MainThumbnailUrl = %q", got.MainThumbnailUrl)
	}
	if len(got.AwardNames) != 2 {
		t.Fatalf("AwardNames len = %d, want 2", len(got.AwardNames))
	}

	// Обратная совместимость: старые Timestamp-поля.
	if got.BirthDate == nil {
		t.Error("BirthDate (legacy) is nil")
	}
	if got.DeathDate == nil {
		t.Error("DeathDate (legacy) is nil — должен быть заполнен из anchor")
	}

	// Новые гибкие даты.
	if got.BirthDateInfo == nil {
		t.Fatal("BirthDateInfo is nil")
	}
	if got.BirthDateInfo.Precision != emhv1.DatePrecision_DATE_PRECISION_EXACT {
		t.Errorf("BirthDateInfo.Precision = %v", got.BirthDateInfo.Precision)
	}
	if got.DeathDateInfo == nil {
		t.Fatal("DeathDateInfo is nil")
	}
	if got.DeathDateInfo.Precision != emhv1.DatePrecision_DATE_PRECISION_MONTH {
		t.Errorf("DeathDateInfo.Precision = %v", got.DeathDateInfo.Precision)
	}
	if got.DeathDateInfo.DisplayText != "Август 2020" {
		t.Errorf("DeathDateInfo.DisplayText = %q", got.DeathDateInfo.DisplayText)
	}
}

// TestMapHeroToSummary_NoPhotosNoAwards проверяет денормализацию при отсутствии связей.
func TestMapHeroToSummary_NoPhotosNoAwards(t *testing.T) {
	h := &domain.Hero{
		ID:               "hero-2",
		FirstName:        "Пётр",
		LastName:         "Иванов",
		MainPhotoURL:     nil,
		MainThumbnailURL: nil,
		AwardNames:       nil,
	}

	got := mapHeroToSummary(h)
	if got.MainPhotoUrl != "" {
		t.Errorf("MainPhotoUrl = %q, want empty", got.MainPhotoUrl)
	}
	if got.MainThumbnailUrl != "" {
		t.Errorf("MainThumbnailUrl = %q, want empty", got.MainThumbnailUrl)
	}
	if len(got.AwardNames) != 0 {
		t.Errorf("AwardNames should be empty, got %v", got.AwardNames)
	}
}

// TestMapHeroToSummary_UnknownDates проверяет маппинг unknown дат.
func TestMapHeroToSummary_UnknownDates(t *testing.T) {
	h := &domain.Hero{
		ID:        "hero-3",
		FirstName: "Сергей",
		LastName:  "Сидоров",
		BirthDate: domain.NewUnknownDate(),
		DeathDate: domain.NewUnknownDate(),
	}

	got := mapHeroToSummary(h)

	// Legacy Timestamps не должны быть заданы для unknown.
	if got.BirthDate != nil {
		t.Errorf("BirthDate should be nil for unknown, got %v", got.BirthDate)
	}
	if got.DeathDate != nil {
		t.Errorf("DeathDate should be nil for unknown, got %v", got.DeathDate)
	}

	// Новые info поля либо nil (если IsZero), либо с precision=UNKNOWN.
	// Зависит от реализации mapFlexibleDateToProto — он возвращает nil для IsZero.
	// Для NewUnknownDate() без DisplayText IsZero() = true → nil.
	if got.BirthDateInfo != nil {
		// Если не nil, то precision должен быть UNKNOWN.
		if got.BirthDateInfo.Precision != emhv1.DatePrecision_DATE_PRECISION_UNKNOWN {
			t.Errorf("BirthDateInfo.Precision = %v", got.BirthDateInfo.Precision)
		}
	}
}

// --- mapHeroDetailToProto ---

// TestMapHeroDetailToProto_FullAggregate проверяет полный агрегат со всеми связями.
func TestMapHeroDetailToProto_FullAggregate(t *testing.T) {
	createdAt := utcDate(2024, 1, 15)
	updatedAt := utcDate(2024, 3, 20)
	birthAnchor := utcDate(1990, 5, 15)

	hero := &domain.Hero{
		ID:        "hero-1",
		FirstName: "Иван",
		LastName:  "Петров",
		FullBio:   "Полная биография героя",
		Status:    domain.StatusPublished,
		BirthDate: domain.FlexibleDate{Precision: domain.PrecisionExact, Anchor: &birthAnchor},
		DeathDate: domain.NewUnknownDate(),
		ServiceStartDate: domain.FlexibleDate{
			Precision:   domain.PrecisionYear,
			Anchor:      ptrTime(utcDate(2010, 1, 1)),
			DisplayText: "2010",
		},
		Position:     "Командир взвода",
		CauseOfDeath: "Погиб в бою",
		Memberships:  []string{"Союз десантников"},
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	photoDate := utcDate(2020, 5, 1)
	awardDate := utcDate(2020, 6, 1)

	detail := &domain.HeroDetail{
		Hero: hero,
		Photos: []*domain.Photo{
			{ID: "p1", URL: "https://s3/1.jpg", IsMain: true, SortOrder: 0},
			{ID: "p2", URL: "https://s3/2.jpg", IsMain: false, SortOrder: 1},
		},
		Awards: []*domain.HeroAward{
			{AwardID: "a1", AwardName: "Герой России", AwardDate: &awardDate, DecreeNumber: "№1"},
		},
		Conflicts: []*domain.HeroConflict{
			{ConflictID: "c1", ConflictName: "СВО", SpecificLocation: "Донбасс"},
		},
		Locations: []*domain.HeroLocation{
			{
				LocationID: "l1",
				Type:       1,
				Location:   &domain.Location{ID: "l1", Name: "Москва"},
			},
		},
		Sources: []*domain.HeroSource{
			{ID: "s1", URL: "https://example.com", Title: "Статья"},
		},
		Relations: []*domain.HeroRelation{
			{ID: "r1", FromHeroID: "hero-1", ToHeroID: "hero-2", RelationType: "comrade"},
		},
	}

	got := mapHeroDetailToProto(detail)

	// Summary заполнен из mapHeroToSummary.
	if got.Summary == nil {
		t.Fatal("Summary is nil")
	}
	if got.Summary.Id != "hero-1" {
		t.Errorf("Summary.Id = %q", got.Summary.Id)
	}

	// Детальные поля.
	if got.FullBio != "Полная биография героя" {
		t.Errorf("FullBio = %q", got.FullBio)
	}
	if got.Position != "Командир взвода" {
		t.Errorf("Position = %q", got.Position)
	}
	if got.CauseOfDeath != "Погиб в бою" {
		t.Errorf("CauseOfDeath = %q", got.CauseOfDeath)
	}
	if got.Status != emhv1.PublicationStatus_PUBLICATION_STATUS_PUBLISHED {
		t.Errorf("Status = %v", got.Status)
	}

	// Audit.
	if got.Audit == nil {
		t.Fatal("Audit is nil")
	}

	// Memberships.
	if len(got.Memberships) != 1 || got.Memberships[0] != "Союз десантников" {
		t.Errorf("Memberships = %v", got.Memberships)
	}

	// ServiceStartDate — legacy Timestamp.
	if got.ServiceStartDate == nil {
		t.Error("ServiceStartDate is nil")
	}

	// ServiceStartDateInfo — новая гибкая дата.
	if got.ServiceStartDateInfo == nil {
		t.Fatal("ServiceStartDateInfo is nil")
	}
	if got.ServiceStartDateInfo.Precision != emhv1.DatePrecision_DATE_PRECISION_YEAR {
		t.Errorf("ServiceStartDateInfo.Precision = %v", got.ServiceStartDateInfo.Precision)
	}

	// Связи.
	if len(got.Photos) != 2 {
		t.Errorf("Photos len = %d, want 2", len(got.Photos))
	}
	if len(got.Awards) != 1 {
		t.Errorf("Awards len = %d", len(got.Awards))
	}
	if len(got.Conflicts) != 1 {
		t.Errorf("Conflicts len = %d", len(got.Conflicts))
	}
	if len(got.Locations) != 1 {
		t.Errorf("Locations len = %d", len(got.Locations))
	}
	if len(got.Sources) != 1 {
		t.Errorf("Sources len = %d", len(got.Sources))
	}
	if len(got.Relations) != 1 {
		t.Errorf("Relations len = %d", len(got.Relations))
	}

	// Проверяем, что IsMain в фото сохранился.
	if !got.Photos[0].IsMain {
		t.Error("Photos[0].IsMain should be true")
	}

	// Suppress unused variable warning.
	_ = photoDate
}

// TestMapHeroDetailToProto_EmptyRelations проверяет маппинг с пустыми связями.
func TestMapHeroDetailToProto_EmptyRelations(t *testing.T) {
	hero := &domain.Hero{
		ID:        "hero-4",
		FirstName: "Алексей",
		LastName:  "Смирнов",
		Status:    domain.StatusDraft,
		BirthDate: domain.NewUnknownDate(),
		DeathDate: domain.NewUnknownDate(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	detail := &domain.HeroDetail{Hero: hero}
	got := mapHeroDetailToProto(detail)

	// Все коллекции должны быть пустыми, но не nil (слайсы инициализированы через make).
	if got.Photos == nil {
		t.Error("Photos should be non-nil slice")
	}
	if len(got.Photos) != 0 {
		t.Errorf("Photos len = %d, want 0", len(got.Photos))
	}
	if len(got.Awards) != 0 {
		t.Errorf("Awards len = %d", len(got.Awards))
	}
	if len(got.Conflicts) != 0 {
		t.Errorf("Conflicts len = %d", len(got.Conflicts))
	}
}

// TestMapHeroDetailToProto_MainPhotoUrlFromDetail проверяет, что MainPhotoUrl
// вычисляется через HeroDetail.MainPhotoURL() (с приоритетом IsMain).
func TestMapHeroDetailToProto_MainPhotoUrlFromDetail(t *testing.T) {
	hero := &domain.Hero{
		ID:        "hero-5",
		FirstName: "Дмитрий",
		LastName:  "Кузнецов",
		Status:    domain.StatusDraft,
		BirthDate: domain.NewUnknownDate(),
		DeathDate: domain.NewUnknownDate(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	detail := &domain.HeroDetail{
		Hero: hero,
		Photos: []*domain.Photo{
			{ID: "p1", URL: "https://s3/first.jpg", IsMain: false, SortOrder: 0},
			{ID: "p2", URL: "https://s3/main.jpg", IsMain: true, SortOrder: 1},
		},
	}

	got := mapHeroDetailToProto(detail)

	// Summary.MainPhotoUrl должен быть URL главного фото (IsMain=true), а не первого.
	if got.Summary.MainPhotoUrl != "https://s3/main.jpg" {
		t.Errorf("Summary.MainPhotoUrl = %q, want %q", got.Summary.MainPhotoUrl, "https://s3/main.jpg")
	}
}
