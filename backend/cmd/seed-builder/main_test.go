package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/deepteams/webp"
)

// encodeWebP — вспомогательная кодировка тестового изображения в WebP.
func encodeWebP(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.EncoderOptions{Lossless: true, Quality: 90, Method: 4}); err != nil {
		t.Fatalf("webp encode: %v", err)
	}
	return buf.Bytes()
}

func solid(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

// writeFixtures создаёт валидный каталог исходников сида.
// paths: image/ribbon — относительные пути из descriptions.json.
func writeFixtures(t *testing.T, descriptionsJSON string, images map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "descriptions.json"), []byte(descriptionsJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, data := range images {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const oneAwardDesc = `{
  "awards": [
    {
      "name": "Медаль «За отвагу»",
      "description": "тест",
      "type": "AWARD_TYPE_MEDAL",
      "jurisdiction": "AWARD_JURISDICTION_RUSSIAN_FEDERATION",
      "is_jubilee": false,
      "worn_without_bar": false,
      "sort_order": 100,
      "image": "images/img.webp",
      "ribbon": "ribbons/rib.webp"
    }
  ]
}`

func goodFixtures(t *testing.T) map[string][]byte {
	return map[string][]byte{
		"images/img.webp":  encodeWebP(t, solid(400, 400, color.White)),
		"ribbons/rib.webp": encodeWebP(t, solid(144, 48, color.White)),
	}
}

func TestBuilder_HappyPath(t *testing.T) {
	dir := writeFixtures(t, oneAwardDesc, goodFixtures(t))
	if err := run(dir); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "seed.json")); err != nil {
		t.Fatalf("seed.json not written: %v", err)
	}
}

// Побайтовый детерминизм: два прогона подряд → идентичные байты (Q13.3).
func TestBuilder_ByteDeterministic(t *testing.T) {
	dir := writeFixtures(t, oneAwardDesc, goodFixtures(t))
	if err := run(dir); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "seed.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := run(dir); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "seed.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("seed.json not byte-deterministic between runs")
	}
	// base64 одной строкой: переносов в артефакте нет (base64.StdEncoding).
	for _, line := range bytes.Split(first, []byte("\n")) {
		if bytes.Contains(line, []byte("\r")) {
			t.Fatal("CR found — base64 must have no line breaks")
		}
	}
}

func TestBuilder_RejectsAliasJurisdiction(t *testing.T) {
	// AWARD_JURISDICTION_RF — неканон (историческая ошибка пилотных данных, Q3).
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_MEDAL","jurisdiction":"AWARD_JURISDICTION_RF","image":"images/img.webp"}]}`
	dir := writeFixtures(t, desc, goodFixtures(t))
	err := run(dir)
	if err == nil {
		t.Fatal("alias jurisdiction must fail the build")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("AWARD_JURISDICTION_RF")) {
		t.Fatalf("error must name the bad value: %v", err)
	}
}

func TestBuilder_RejectsUnspecifiedType(t *testing.T) {
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_UNSPECIFIED","jurisdiction":"AWARD_JURISDICTION_USSR","image":"images/img.webp"}]}`
	dir := writeFixtures(t, desc, goodFixtures(t))
	if err := run(dir); err == nil {
		t.Fatal("UNSPECIFIED must fail the build")
	}
}

func TestBuilder_RejectsPNG(t *testing.T) {
	// Fail-fast: данные чинятся конвертацией в репозитории (Q4).
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR","image":"images/img.png"}]}`
	dir := writeFixtures(t, desc, map[string][]byte{
		"images/img.png": {0x89, 'P', 'N', 'G'},
	})
	if err := run(dir); err == nil {
		t.Fatal("non-webp file must fail the build")
	}
}

func TestBuilder_RejectsFakeWebpExtension(t *testing.T) {
	// .webp по имени, но не RIFF/WEBP по содержимому.
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR","image":"images/img.webp"}]}`
	dir := writeFixtures(t, desc, map[string][]byte{
		"images/img.webp": []byte("definitely not a webp container"),
	})
	if err := run(dir); err == nil {
		t.Fatal("bogus webp must fail the build")
	}
}

func TestBuilder_RejectsWrongDimensions(t *testing.T) {
	// Строго 400×400 / 144×48, без допуска ±10% (Q18.2).
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR","image":"images/img.webp"}]}`
	dir := writeFixtures(t, desc, map[string][]byte{
		"images/img.webp": encodeWebP(t, solid(396, 396, color.White)),
	})
	err := run(dir)
	if err == nil {
		t.Fatal("396x396 must fail — sizes are exact")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("396x396")) {
		t.Fatalf("error must report actual size: %v", err)
	}
}

func TestBuilder_RejectsWrongRibbonSize(t *testing.T) {
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR","ribbon":"ribbons/rib.webp"}]}`
	dir := writeFixtures(t, desc, map[string][]byte{
		"ribbons/rib.webp": encodeWebP(t, solid(144, 50, color.White)),
	})
	if err := run(dir); err == nil {
		t.Fatal("144x50 ribbon must fail")
	}
}

func TestBuilder_RejectsMissingFile(t *testing.T) {
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR","image":"images/absent.webp"}]}`
	dir := writeFixtures(t, desc, goodFixtures(t))
	if err := run(dir); err == nil {
		t.Fatal("missing image file must fail")
	}
}

func TestBuilder_RejectsInternalDuplicate(t *testing.T) {
	desc := `{"awards":[
	  {"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR"},
	  {"name":"x","type":"AWARD_TYPE_ORDER","jurisdiction":"AWARD_JURISDICTION_USSR"}]}`
	dir := writeFixtures(t, desc, nil)
	if err := run(dir); err == nil {
		t.Fatal("internal duplicate must fail")
	}
}

func TestBuilder_RejectsEmptyAwards(t *testing.T) {
	dir := writeFixtures(t, `{"awards":[]}`, nil)
	if err := run(dir); err == nil {
		t.Fatal("empty awards list must fail")
	}
}

func TestBuilder_OptionalImagesOmitted(t *testing.T) {
	// Награда без картинок валидна (в fallback-список).
	desc := `{"awards":[{"name":"x","type":"AWARD_TYPE_BADGE","jurisdiction":"AWARD_JURISDICTION_DEPARTMENTAL"}]}`
	dir := writeFixtures(t, desc, nil)
	if err := run(dir); err != nil {
		t.Fatalf("award without images must build: %v", err)
	}
}

func TestBuilder_Base64RoundTrip(t *testing.T) {
	dir := writeFixtures(t, oneAwardDesc, goodFixtures(t))
	path := filepath.Join(dir, "ribbons/rib.webp")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := loadImage(path, 144, 48, "test")
	if err != nil {
		t.Fatal(err)
	}
	if img.ContentType != "image/webp" {
		t.Fatalf("content_type = %q", img.ContentType)
	}
	decoded, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		t.Fatalf("base64 must decode with StdEncoding: %v", err)
	}
	if !bytes.Equal(decoded, raw) {
		t.Fatal("base64 round-trip must preserve file bytes exactly")
	}
}
