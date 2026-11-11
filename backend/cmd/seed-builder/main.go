// seed-builder собирает seed.json из исходников сида (descriptions.json +
// картинки) со строгой валидацией: артефакт коммитится в Git LFS и сидится
// «как есть», поэтому брак данных ловится на сборке, а не в проде (ADR-005).
//
// Использование: go run ./cmd/seed-builder ./seeds/awards
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"

	_ "github.com/deepteams/webp" // декодер WebP для image.DecodeConfig

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/seeder"
)

// эталонные размеры (docs/specs/awards-images-guide.md):
// знак — квадрат 400×400, лента — 144×48. Строго, без допуска:
// фронт фиксирует width/height в пикселях (CLS), ADR — «размеры фиксированные».
const (
	imageWidth, imageHeight   = 400, 400
	ribbonWidth, ribbonHeight = 144, 48
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: seed-builder <dir-with-descriptions.json>")
		os.Exit(1)
	}
	dir := os.Args[1]
	if err := run(dir); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, "descriptions.json"))
	if err != nil {
		return fmt.Errorf("read descriptions.json: %w", err)
	}
	// descriptions.json — входной формат: пути к файлам вместо base64.
	type descAward struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Type           string `json:"type"`
		Jurisdiction   string `json:"jurisdiction"`
		IsJubilee      bool   `json:"is_jubilee"`
		WornWithoutBar bool   `json:"worn_without_bar"`
		SortOrder      int    `json:"sort_order"`
		Image          string `json:"image"`
		Ribbon         string `json:"ribbon"`
	}
	var input struct {
		Awards []descAward `json:"awards"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parse descriptions.json: %w", err)
	}
	if len(input.Awards) == 0 {
		return fmt.Errorf("descriptions.json contains no awards")
	}

	out := seeder.SeedFile{Awards: make([]seeder.SeedAward, 0, len(input.Awards))}
	for i, a := range input.Awards {
		where := fmt.Sprintf("award #%d %q", i+1, a.Name)

		// Строгий канон enum — только proto-имена, без алиасов (Q3).
		if t, ok := emhv1.AwardType_value[a.Type]; !ok || t == int32(emhv1.AwardType_AWARD_TYPE_UNSPECIFIED) {
			return fmt.Errorf("%s: unknown type %q (expected AWARD_TYPE_ORDER|MEDAL|BADGE)", where, a.Type)
		}
		if j, ok := emhv1.AwardJurisdiction_value[a.Jurisdiction]; !ok || j == int32(emhv1.AwardJurisdiction_AWARD_JURISDICTION_UNSPECIFIED) {
			return fmt.Errorf("%s: unknown jurisdiction %q (expected AWARD_JURISDICTION_RUSSIAN_FEDERATION|USSR|DEPARTMENTAL)", where, a.Jurisdiction)
		}
		if a.Name == "" {
			return fmt.Errorf("%s: name is empty", where)
		}

		sa := seeder.SeedAward{
			Name:           a.Name,
			Description:    a.Description,
			Type:           a.Type,
			Jurisdiction:   a.Jurisdiction,
			IsJubilee:      a.IsJubilee,
			WornWithoutBar: a.WornWithoutBar,
			SortOrder:      a.SortOrder,
		}

		if a.Image != "" {
			img, err := loadImage(filepath.Join(dir, a.Image), imageWidth, imageHeight, where+" image")
			if err != nil {
				return err
			}
			sa.Image = img
		}
		if a.Ribbon != "" {
			rbn, err := loadImage(filepath.Join(dir, a.Ribbon), ribbonWidth, ribbonHeight, where+" ribbon")
			if err != nil {
				return err
			}
			sa.Ribbon = rbn
		}
		out.Awards = append(out.Awards, sa)
	}

	if err := validateNoInternalDuplicates(out.Awards); err != nil {
		return err
	}

	// Побайтовый детерминизм: фиксированный порядок полей структур,
	// MarshalIndent 2 пробела, порядок записей = входной файл (Q17).
	// escapeHTML=false: кириллица в name/description не должна превращаться
	// в \u00NN — артефакт остаётся человекочитаемым в diff.
	encoded, err := marshalDeterministic(out)
	if err != nil {
		return fmt.Errorf("marshal seed.json: %w", err)
	}
	outPath := filepath.Join(dir, "seed.json")
	if err := os.WriteFile(outPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write seed.json: %w", err)
	}
	fmt.Printf("seed.json written: %d awards, %s\n", len(out.Awards), humanSize(len(encoded)))
	return nil
}

// loadImage читает файл, проверяет расширение/магию/строго ожидаемый размер
// и возвращает base64 без переносов (base64.StdEncoding, Q17).
func loadImage(path string, wantW, wantH int, label string) (*seeder.SeedImage, error) {
	if filepath.Ext(path) != ".webp" {
		return nil, fmt.Errorf("%s: %s: only .webp allowed (convert data in repo, Q4)", label, filepath.Base(path))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: read: %w", label, err)
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, fmt.Errorf("%s: %s: not a RIFF/WEBP container", label, filepath.Base(path))
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s: decode %s: %w", label, filepath.Base(path), err)
	}
	if cfg.Width != wantW || cfg.Height != wantH {
		return nil, fmt.Errorf("%s: %s: got %dx%d, expected exactly %dx%d",
			label, filepath.Base(path), cfg.Width, cfg.Height, wantW, wantH)
	}
	return &seeder.SeedImage{
		ContentType: "image/webp",
		Data:        base64.StdEncoding.EncodeToString(data),
	}, nil
}

// validateNoInternalDuplicates — тот же инвариант, что проверяет seeder при импорте;
// ловим на сборке.
func validateNoInternalDuplicates(awards []seeder.SeedAward) error {
	seen := map[string]struct{}{}
	for _, a := range awards {
		key := a.Name + "\x00" + a.Type + "\x00" + a.Jurisdiction
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate award %q (type=%s, jurisdiction=%s)", a.Name, a.Type, a.Jurisdiction)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// marshalDeterministic — детерминированная сериализация seed.json:
// фиксированный порядок полей (порядок объявления структур seeder.SeedAward),
// отступ 2 пробела, экранирование HTML отключено (кириллица остаётся
// человекочитаемой в diff), завершающий перевод строки стабилен (Encoder.Encode).
func marshalDeterministic(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func humanSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f KiB", float64(n)/1024)
}
