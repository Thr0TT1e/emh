package seeder

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// --- fake AwardRepository ---

type fakeRepo struct {
	repository.AwardRepository // незащищённые методы не используются — nil-safe through panic

	existing  map[string]bool // ключ seedKey(name, type, jurisdiction)
	inserted  []domain.CreateAwardParams
	raceOnKey map[string]bool // InsertIfAbsent вернёт false (гонка с админкой)
	existsErr error
	insertErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{existing: map[string]bool{}, raceOnKey: map[string]bool{}}
}

func (f *fakeRepo) ExistsByKey(_ context.Context, name string, t domain.AwardType, j domain.AwardJurisdiction) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.existing[seedKey(name, awardTypeName(t), jurisdictionName(j))], nil
}

// awardTypeName / jurisdictionName — обратный маппинг domain→proto-имена
// (согласно Q11 ключ сида строится на proto-именах).
func awardTypeName(t domain.AwardType) string {
	switch t {
	case domain.AwardTypeOrder:
		return "AWARD_TYPE_ORDER"
	case domain.AwardTypeMedal:
		return "AWARD_TYPE_MEDAL"
	case domain.AwardTypeBadge:
		return "AWARD_TYPE_BADGE"
	}
	return "AWARD_TYPE_UNSPECIFIED"
}

func jurisdictionName(j domain.AwardJurisdiction) string {
	switch j {
	case domain.AwardJurisdictionRussianFederation:
		return "AWARD_JURISDICTION_RUSSIAN_FEDERATION"
	case domain.AwardJurisdictionUSSR:
		return "AWARD_JURISDICTION_USSR"
	case domain.AwardJurisdictionDepartmental:
		return "AWARD_JURISDICTION_DEPARTMENTAL"
	}
	return "AWARD_JURISDICTION_UNSPECIFIED"
}

func (f *fakeRepo) InsertIfAbsent(_ context.Context, p domain.CreateAwardParams) (bool, error) {
	if f.insertErr != nil {
		return false, f.insertErr
	}
	key := seedKey(p.Name, awardTypeName(p.Type), jurisdictionName(p.Jurisdiction))
	if f.raceOnKey[key] {
		return false, nil
	}
	f.inserted = append(f.inserted, p)
	return true, nil
}

func (f *fakeRepo) List(_ context.Context, _ domain.AwardFilter) ([]*domain.Award, string, int64, error) {
	return nil, "", int64(len(f.existing)), nil
}

// --- fake MediaStorage ---

type fakeStorage struct {
	domain.MediaStorage

	uploaded  map[string][]byte
	uploadErr error
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{uploaded: map[string][]byte{}}
}

func (f *fakeStorage) Upload(_ context.Context, key string, data io.Reader, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(data); err != nil {
		return err
	}
	f.uploaded[key] = buf.Bytes()
	return nil
}

func (f *fakeStorage) PublicURL(key string) string { return "https://cdn.test/" + key }

// --- helpers ---

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// validWebP — минимальный RIFF/WEBP-контейнер (проверяется только магия, не пиксели).
func validWebP() *SeedImage {
	body := []byte("RIFF" + "\x00\x00\x00\x00" + "WEBP" + "VP8 " + "\x00\x00\x00\x00")
	return &SeedImage{ContentType: "image/webp", Data: base64.StdEncoding.EncodeToString(body)}
}

func writeSeed(t *testing.T, seed SeedFile) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.json")
	// сериализация через encoding/json напрямую — тест фиксирует consumer-контракт
	data, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func baseAward(name string) SeedAward {
	return SeedAward{
		Name:         name,
		Description:  "тест",
		Type:         "AWARD_TYPE_MEDAL",
		Jurisdiction: "AWARD_JURISDICTION_RUSSIAN_FEDERATION",
		SortOrder:    1,
	}
}

// --- tests ---

func TestSeedAwards_HappyPath(t *testing.T) {
	a := baseAward("Медаль «За отвагу»")
	a.Image = validWebP()
	a.Ribbon = validWebP()
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	repo, storage := newFakeRepo(), newFakeStorage()
	res, err := SeedAwards(context.Background(), path, repo, storage, testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Inserted != 1 || res.Skipped != 0 || res.Failed != 0 {
		t.Fatalf("result = %+v", res)
	}
	p := repo.inserted[0]
	if p.Type != domain.AwardTypeMedal || p.Jurisdiction != domain.AwardJurisdictionRussianFederation {
		t.Errorf("enum mapping: %+v", p)
	}
	if p.SortOrder != 1 || p.Description != "тест" {
		t.Errorf("scalar mapping: %+v", p)
	}
	// Детерминированный ключ объекта: awards/<images|ribbons>/<sha[:16]>.webp (Q11).
	hash := objectHash(a.Name, a.Type, a.Jurisdiction)
	for _, key := range []string{"awards/images/" + hash + ".webp", "awards/ribbons/" + hash + ".webp"} {
		if _, ok := storage.uploaded[key]; !ok {
			t.Errorf("upload missing: %s (have %v)", key, keysOf(storage.uploaded))
		}
	}
	wantURL := "https://cdn.test/awards/ribbons/" + hash + ".webp"
	if p.RibbonImageURL != wantURL {
		t.Errorf("RibbonImageURL = %q, want %q", p.RibbonImageURL, wantURL)
	}
}

func TestSeedAwards_SkipsExisting(t *testing.T) {
	a := baseAward("Существующая")
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	repo := newFakeRepo()
	repo.existing[seedKey(a.Name, a.Type, a.Jurisdiction)] = true
	res, err := SeedAwards(context.Background(), path, repo, newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped != 1 || res.Inserted != 0 || len(repo.inserted) != 0 {
		t.Fatalf("result = %+v inserted=%d", res, len(repo.inserted))
	}
}

func TestSeedAwards_RaceWithAdminCountsSkipped(t *testing.T) {
	a := baseAward("Гонка")
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	repo := newFakeRepo()
	repo.raceOnKey[seedKey(a.Name, a.Type, a.Jurisdiction)] = true
	res, err := SeedAwards(context.Background(), path, repo, newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped != 1 || res.Inserted != 0 || res.Failed != 0 {
		t.Fatalf("result = %+v", res)
	}
}

func TestSeedAwards_MinIODown_FailsRecordWithoutInsert(t *testing.T) {
	// Q14: недоступный MinIO → запись целиком не инсертится.
	a := baseAward("Без MinIO")
	a.Ribbon = validWebP()
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	repo, storage := newFakeRepo(), newFakeStorage()
	storage.uploadErr = errors.New("minio unreachable")
	res, err := SeedAwards(context.Background(), path, repo, storage, testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 || res.Inserted != 0 {
		t.Fatalf("result = %+v", res)
	}
	if len(repo.inserted) != 0 {
		t.Errorf("insert must not happen when upload fails: %+v", repo.inserted)
	}
}

func TestSeedAwards_BadBase64_FailsRecord(t *testing.T) {
	a := baseAward("Битая картинка")
	a.Image = &SeedImage{ContentType: "image/webp", Data: "%%%not-base64%%"}
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	res, err := SeedAwards(context.Background(), path, newFakeRepo(), newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 || res.Inserted != 0 {
		t.Fatalf("result = %+v", res)
	}
}

func TestSeedAwards_WrongContentType_FailsRecord(t *testing.T) {
	a := baseAward("Не webp")
	a.Image = &SeedImage{ContentType: "image/png", Data: validWebP().Data}
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	res, err := SeedAwards(context.Background(), path, newFakeRepo(), newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 {
		t.Fatalf("result = %+v", res)
	}
}

func TestSeedAwards_UnknownEnum_FailsRecordOnly(t *testing.T) {
	good := baseAward("Хорошая")
	bad := baseAward("Левый enum")
	bad.Type = "ORDER" // не proto-канон (Q3)
	path := writeSeed(t, SeedFile{Awards: []SeedAward{bad, good}})

	repo := newFakeRepo()
	res, err := SeedAwards(context.Background(), path, repo, newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 || res.Inserted != 1 {
		t.Fatalf("result = %+v", res)
	}
	if len(repo.inserted) != 1 || repo.inserted[0].Name != good.Name {
		t.Errorf("only valid record must be inserted: %+v", repo.inserted)
	}
}

func TestSeedAwards_UnspecifiedEnum_Fails(t *testing.T) {
	a := baseAward("UNSPECIFIED")
	a.Jurisdiction = "AWARD_JURISDICTION_UNSPECIFIED"
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a}})

	res, err := SeedAwards(context.Background(), path, newFakeRepo(), newFakeStorage(), testLogger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 {
		t.Fatalf("UNSPECIFIED must be rejected: %+v", res)
	}
}

func TestSeedAwards_InternalDuplicate_Fatal(t *testing.T) {
	a := baseAward("Дубль")
	b := baseAward("Дубль")
	b.Type = "AWARD_TYPE_ORDER"
	c := baseAward("Дубль") // точный дубль ключа
	path := writeSeed(t, SeedFile{Awards: []SeedAward{a, b, c}})

	_, err := SeedAwards(context.Background(), path, newFakeRepo(), newFakeStorage(), testLogger)
	if err == nil {
		t.Fatal("expected fatal error for internal duplicate")
	}
}

func TestSeedAwards_BrokenJSON_Fatal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SeedAwards(context.Background(), path, newFakeRepo(), newFakeStorage(), testLogger); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSeedAwards_MissingFile_Fatal(t *testing.T) {
	if _, err := SeedAwards(context.Background(), "/nonexistent/seed.json", newFakeRepo(), newFakeStorage(), testLogger); err == nil {
		t.Fatal("expected read error")
	}
}

func TestHasAnyAwards(t *testing.T) {
	repo := newFakeRepo()
	repo.existing[seedKey("x", "y", "z")] = true
	has, err := HasAnyAwards(context.Background(), repo)
	if err != nil || !has {
		t.Fatalf("has=%v err=%v", has, err)
	}
	empty := newFakeRepo()
	has, err = HasAnyAwards(context.Background(), empty)
	if err != nil || has {
		t.Fatalf("has=%v err=%v", has, err)
	}
}

func TestObjectHash_DeterministicAndSeparated(t *testing.T) {
	h1 := objectHash("AB", "C", "D")
	h2 := objectHash("A", "BC", "D")
	if h1 == h2 {
		t.Fatal("NUL-separator must prevent key splicing collision")
	}
	if len(h1) != 16 {
		t.Fatalf("hash len = %d, want 16", len(h1))
	}
	// Стабильность между вызовами
	if objectHash("Орден", "AWARD_TYPE_ORDER", "AWARD_JURISDICTION_USSR") !=
		objectHash("Орден", "AWARD_TYPE_ORDER", "AWARD_JURISDICTION_USSR") {
		t.Fatal("hash must be deterministic")
	}
	// Ключ не зависит от времени/состояния: зафиксированный вектор.
	if got := objectHash("a", "b", "c"); got != "8badde10c760e9b7" {
		t.Fatalf("hash vector = %q", got)
	}
}

func TestDecodeImage_RejectsNonRIFF(t *testing.T) {
	img := &SeedImage{ContentType: "image/webp", Data: base64.StdEncoding.EncodeToString([]byte("garbage garbage"))}
	if _, err := decodeImage(img); err == nil {
		t.Fatal("expected RIFF/WEBP magic error")
	}
}

// keysOf — утилита отладки сообщений.
func keysOf(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
