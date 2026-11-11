package pg

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

func TestAwardCursor_RoundTrip(t *testing.T) {
	a := &domain.Award{ID: "018f2b5c-0000-7000-8000-000000000001", Name: "Орден «За заслуги» III степени", SortOrder: 42}
	raw := encodeAwardCursor(a)
	got, err := decodeAwardCursor(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != a.ID || got.Name != a.Name || got.SortOrder != a.SortOrder {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	// Курсор детерминирован: та же строка → тот же токен.
	if encodeAwardCursor(a) != raw {
		t.Fatal("cursor encoding must be deterministic")
	}
}

func TestAwardCursor_EmptyIsFirstPage(t *testing.T) {
	c, err := decodeAwardCursor("")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "" {
		t.Fatalf("empty cursor must decode to zero: %+v", c)
	}
}

func TestAwardCursor_RejectsGarbage(t *testing.T) {
	for _, bad := range []string{
		"not base64 ☃", // битый base64
		base64.URLEncoding.EncodeToString([]byte("{not json")),       // битый JSON
		base64.URLEncoding.EncodeToString([]byte(`{"s":1,"n":"x"}`)), // пустой id
	} {
		if _, err := decodeAwardCursor(bad); !errors.Is(err, domain.ErrInvalidCursor) {
			t.Errorf("garbage %q: err=%v, want ErrInvalidCursor", bad, err)
		}
	}
}

func TestAwardCursor_StopsCrossDecoding(t *testing.T) {
	// Старые id-only курсоры (hex UUID без base64url-JSON внутри) должны
	// отваливаться с ErrInvalidCursor, а не молча возвращать первую страницу.
	if _, err := decodeAwardCursor("018f2b5c-0000-7000-8000-000000000000"); !strings.Contains(err.Error(), "invalid pagination cursor") {
		t.Fatalf("legacy cursor must error, got %v", err)
	}
}
