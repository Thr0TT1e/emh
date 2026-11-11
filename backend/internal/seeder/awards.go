package seeder

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
)

// Result — итог импорта сида.
type Result struct {
	Inserted int
	Skipped  int
	Failed   int
}

// HasAnyAwards — наполнен ли справочник (для авто-сида при старте).
// Переиспользует List с минимальной страницей: total считается COUNT(*) в репозитории.
func HasAnyAwards(ctx context.Context, repo repository.AwardRepository) (bool, error) {
	_, _, total, err := repo.List(ctx, domain.AwardFilter{Limit: 1})
	if err != nil {
		return false, fmt.Errorf("check awards count: %w", err)
	}
	return total > 0, nil
}

// SeedAwards импортирует награды из seed.json в БД и MinIO.
//
// Идемпотентность: существующие по (name, type, jurisdiction) пропускаются,
// ручные правки админки не перезаписываются (ON CONFLICT DO NOTHING).
// Ошибки отдельных записей не прерывают сид: log.Error + Failed++ (ADR-005, Q8=a).
// Битый JSON артефакта или дубликат ключа внутри файла — фатальная ошибка:
// seed.json собран builder'ом, его порча — инцидент сборки, а не штатный случай.
// MinIO недоступен (или битая картинка) → запись НЕ вставляется целиком:
// запись без ленты засорит справочник — повторный сид пропустит её по ON CONFLICT (Q14).
func SeedAwards(ctx context.Context, seedPath string,
	repo repository.AwardRepository, storage domain.MediaStorage, logger *slog.Logger,
) (*Result, error) {
	data, err := os.ReadFile(seedPath)
	if err != nil {
		return nil, fmt.Errorf("read seed file: %w", err)
	}
	var seed SeedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		return nil, fmt.Errorf("parse seed file %s: %w", seedPath, err)
	}
	if err := validateNoInternalDuplicates(seed.Awards); err != nil {
		return nil, err
	}

	result := &Result{}
	for _, a := range seed.Awards {
		p, err := mapAward(a)
		if err != nil {
			logger.Error("Сид наград: некорректная запись, пропуск", "name", a.Name, "error", err)
			result.Failed++
			continue
		}

		// Дешёвая проверка существования до загрузок в MinIO.
		exists, err := repo.ExistsByKey(ctx, p.Name, p.Type, p.Jurisdiction)
		if err != nil {
			logger.Error("Сид наград: проверка существования не удалась, пропуск", "name", a.Name, "error", err)
			result.Failed++
			continue
		}
		if exists {
			result.Skipped++
			continue
		}

		// Картинки грузятся до инсерта: частичная запись недопустима (Q14).
		if err := uploadImages(ctx, a, p, storage, logger); err != nil {
			result.Failed++ // причина залогирована внутри
			continue
		}

		inserted, err := repo.InsertIfAbsent(ctx, *p)
		if err != nil {
			logger.Error("Сид наград: вставка не удалась", "name", a.Name, "error", err)
			result.Failed++
			continue
		}
		if !inserted {
			// Race condition: запись создал админ, пока шёл сид.
			logger.Warn("Сид наград: запись уже существует (гонка с админкой)", "name", a.Name)
			result.Skipped++
			continue
		}
		result.Inserted++
	}
	return result, nil
}

// validateNoInternalDuplicates: составной ключ внутри файла обязан быть уникален —
// дубликат означает брак сборки артефакта.
func validateNoInternalDuplicates(awards []SeedAward) error {
	seen := make(map[string]struct{}, len(awards))
	for _, a := range awards {
		key := seedKey(a.Name, a.Type, a.Jurisdiction)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("seed file contains duplicate award %q (type=%s, jurisdiction=%s)",
				a.Name, a.Type, a.Jurisdiction)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// seedKey — канонический составной ключ. NUL-разделитель исключает склейку строк
// ("A|B"+"C" vs "A"+"B|C") при хешировании.
func seedKey(name, awardType, jurisdiction string) string {
	return name + "\x00" + awardType + "\x00" + jurisdiction
}

// objectHash — 16 hex-символов sha256 от ключа записи (proto-имена enum, Q11):
// детерминированно, идемпотентно, без кириллицы в ключах MinIO.
func objectHash(name, awardType, jurisdiction string) string {
	sum := sha256.Sum256([]byte(seedKey(name, awardType, jurisdiction)))
	return hex.EncodeToString(sum[:])[:16]
}

// mapAward строго валидирует enum-значения (только proto-канон, UNSPECIFIED
// неприемлем, Q3) и собирает параметры вставки.
func mapAward(a SeedAward) (*domain.CreateAwardParams, error) {
	if a.Name == "" {
		return nil, fmt.Errorf("award name is empty")
	}
	t, ok := emhv1.AwardType_value[a.Type]
	if !ok || t == int32(emhv1.AwardType_AWARD_TYPE_UNSPECIFIED) {
		return nil, fmt.Errorf("unknown award type %q", a.Type)
	}
	j, ok := emhv1.AwardJurisdiction_value[a.Jurisdiction]
	if !ok || j == int32(emhv1.AwardJurisdiction_AWARD_JURISDICTION_UNSPECIFIED) {
		return nil, fmt.Errorf("unknown award jurisdiction %q", a.Jurisdiction)
	}
	return &domain.CreateAwardParams{
		Name:           a.Name,
		Description:    a.Description,
		SortOrder:      a.SortOrder,
		Type:           domain.AwardType(t),
		Jurisdiction:   domain.AwardJurisdiction(j),
		WornWithoutBar: a.WornWithoutBar,
		IsJubilee:      a.IsJubilee,
	}, nil
}

// uploadImages декодирует base64 и грузит знак/ленту в MinIO, заполняя URL в p.
func uploadImages(ctx context.Context, a SeedAward, p *domain.CreateAwardParams,
	storage domain.MediaStorage, logger *slog.Logger,
) error {
	hash := objectHash(a.Name, a.Type, a.Jurisdiction)

	if a.Image != nil {
		data, err := decodeImage(a.Image)
		if err != nil {
			logger.Error("Сид наград: знак награды некорректен", "name", a.Name, "error", err)
			return err
		}
		key := "awards/images/" + hash + ".webp"
		if err := storage.Upload(ctx, key, bytes.NewReader(data), "image/webp"); err != nil {
			logger.Error("Сид наград: загрузка знака в MinIO не удалась", "name", a.Name, "key", key, "error", err)
			return err
		}
		p.ImageURL = storage.PublicURL(key)
	}

	if a.Ribbon != nil {
		data, err := decodeImage(a.Ribbon)
		if err != nil {
			logger.Error("Сид наград: лента некорректна", "name", a.Name, "error", err)
			return err
		}
		key := "awards/ribbons/" + hash + ".webp"
		if err := storage.Upload(ctx, key, bytes.NewReader(data), "image/webp"); err != nil {
			logger.Error("Сид наград: загрузка ленты в MinIO не удалась", "name", a.Name, "key", key, "error", err)
			return err
		}
		p.RibbonImageURL = storage.PublicURL(key)
	}
	return nil
}

// decodeImage проверяет content_type, декодирует base64 и RIFF/WEBP-магию.
// Байтовая проверка — защита от порчи артефакта после сборки; MIME/размеры
// валидированы builder'ом (строгость на сборке, Q10).
func decodeImage(img *SeedImage) ([]byte, error) {
	if img.ContentType != "image/webp" {
		return nil, fmt.Errorf("expected image/webp, got %q", img.ContentType)
	}
	data, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, fmt.Errorf("payload is not a RIFF/WEBP container")
	}
	return data, nil
}
