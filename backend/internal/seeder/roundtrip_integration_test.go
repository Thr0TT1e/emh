//go:build integration

// Round-trip сида: seed → exists → seed again → skipped (ТЗ §11).
// МинIO в CI/локальной машине недоступен (реестры закрыты), поэтому хранилище —
// фейк; формат ключей объекта и идемпотентность загрузки покрыты юнит-тестами,
// здесь проверяется связка seeder ↔ реальный pg-репозиторий с idx_awards_seed_key.
package seeder_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
	"codeberg.org/Thr0TT1e/emh/backend/internal/seeder"
)

type recordingStorage struct {
	domain.MediaStorage
	uploaded map[string][]byte
}

func (s *recordingStorage) Upload(_ context.Context, key string, data io.Reader, _ string) error {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(data); err != nil {
		return err
	}
	s.uploaded[key] = buf.Bytes()
	return nil
}

func (s *recordingStorage) PublicURL(key string) string { return "https://cdn.test/" + key }

// Минимальный, но честный RIFF/WEBP: магия + VP8L-подобное тело не декодируется —
// seeder декодирование пикселей не делает (Q10), только контент-type + магию.
var webpMagic = base64.StdEncoding.EncodeToString(
	append(append([]byte("RIFF\x00\x00\x00\x00WEBP"), "VP8 "...), "\x00\x00\x00\x00"...))

func pilotSeed(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.json")
	seed := seeder.SeedFile{Awards: []seeder.SeedAward{
		{
			Name: "Орден Тестовый", Description: "d1",
			Type: "AWARD_TYPE_ORDER", Jurisdiction: "AWARD_JURISDICTION_RUSSIAN_FEDERATION",
			SortOrder: 1, Image: &seeder.SeedImage{ContentType: "image/webp", Data: webpMagic},
		},
		{
			Name: "Медаль Тестовая", Description: "d2",
			Type: "AWARD_TYPE_MEDAL", Jurisdiction: "AWARD_JURISDICTION_USSR",
			IsJubilee: true, SortOrder: 2,
			Image:  &seeder.SeedImage{ContentType: "image/webp", Data: webpMagic},
			Ribbon: &seeder.SeedImage{ContentType: "image/webp", Data: webpMagic},
		},
	}}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(seed); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return path
}

func TestSeedAwards_RoundTrip(t *testing.T) {
	pool, cleanup := testutil.NewTestPool(t)
	defer cleanup()
	testutil.CleanupTable(t, pool, "awards")

	repo := pg.NewAwardRepository(pool)
	storage := &recordingStorage{uploaded: map[string][]byte{}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()
	path := pilotSeed(t)

	// 1. Первый прогон: обе записи вставлены.
	res, err := seeder.SeedAwards(ctx, path, repo, storage, logger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Inserted != 2 || res.Skipped != 0 || res.Failed != 0 {
		t.Fatalf("first run: %+v", res)
	}

	// 2. Справочник больше не пуст (условие авто-сида).
	has, err := seeder.HasAnyAwards(ctx, repo)
	if err != nil || !has {
		t.Fatalf("HasAnyAwards = %v, %v", has, err)
	}

	// 3. Второй прогон: всё пропущено, дублей образов нет.
	res2, err := seeder.SeedAwards(ctx, path, repo, storage, logger)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Skipped != 2 || res2.Inserted != 0 || res2.Failed != 0 {
		t.Fatalf("second run: %+v", res2)
	}
	total, _, count, err := repo.List(ctx, domain.AwardFilter{Limit: 10})
	if err != nil || count != 2 {
		t.Fatalf("awards count = %d (%v)", count, err)
	}
	_ = total

	// 4. Ручная правка админки не перезатёрта повторным сидом.
	for _, a := range total {
		if a.Name != "Медаль Тестовая" {
			continue
		}
		if _, err := repo.Update(ctx, domain.UpdateAwardParams{
			ID: a.ID, Description: "ручная правка", FieldMask: []string{"description"},
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := seeder.SeedAwards(ctx, path, repo, storage, logger); err != nil {
		t.Fatal(err)
	}
	for _, a := range total {
		if a.Name != "Медаль Тестовая" {
			continue
		}
		got, err := repo.GetByID(ctx, a.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Description != "ручная правка" {
			t.Fatalf("seed overwrote manual edit: %q", got.Description)
		}
	}

	// 5. Ключи MinIO: детерминированный sha256[:16] от составного ключа (Q11).
	sum := sha256.Sum256([]byte("Медаль Тестовая\x00AWARD_TYPE_MEDAL\x00AWARD_JURISDICTION_USSR"))
	key := hex.EncodeToString(sum[:])[:16]
	if _, ok := storage.uploaded["awards/ribbons/"+key+".webp"]; !ok {
		missing := make([]string, 0, len(storage.uploaded))
		for k := range storage.uploaded {
			missing = append(missing, k)
		}
		t.Fatalf("ribbon object missing; uploaded=%v", missing)
	}
}
