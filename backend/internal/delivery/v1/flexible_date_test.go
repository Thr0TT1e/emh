package v1

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
)

// --- Хелперы ---

// ptrFloat64 возвращает указатель на float64.
func ptrFloat64(f float64) *float64 { return &f }

// ptrTime возвращает указатель на time.Time.
func ptrTime(t time.Time) *time.Time { return &t }

// utcDate создаёт дату в UTC.
func utcDate(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// --- mapFlexibleDateToProto ---

// TestMapFlexibleDateToProto_AllPrecisions проверяет маппинг domain → proto для всех precision-типов.
func TestMapFlexibleDateToProto_AllPrecisions(t *testing.T) {
	anchor := utcDate(1994, 2, 15)

	tests := []struct {
		name       string
		fd         domain.FlexibleDate
		wantNil    bool
		wantPrec   emhv1.DatePrecision
		wantAnchor bool
		wantText   string
	}{
		{
			name:    "unspecified returns nil",
			fd:      domain.FlexibleDate{Precision: domain.PrecisionUnspecified},
			wantNil: true,
		},
		{
			name:    "zero value returns nil",
			fd:      domain.FlexibleDate{},
			wantNil: true,
		},
		{
			name:       "exact with anchor",
			fd:         domain.FlexibleDate{Precision: domain.PrecisionExact, Anchor: &anchor},
			wantNil:    false,
			wantPrec:   emhv1.DatePrecision_DATE_PRECISION_EXACT,
			wantAnchor: true,
			wantText:   "",
		},
		{
			name:       "month with anchor and text",
			fd:         domain.FlexibleDate{Precision: domain.PrecisionMonth, Anchor: &anchor, DisplayText: "Февраль 1994"},
			wantNil:    false,
			wantPrec:   emhv1.DatePrecision_DATE_PRECISION_MONTH,
			wantAnchor: true,
			wantText:   "Февраль 1994",
		},
		{
			name:       "year with anchor and text",
			fd:         domain.FlexibleDate{Precision: domain.PrecisionYear, Anchor: &anchor, DisplayText: "1994"},
			wantNil:    false,
			wantPrec:   emhv1.DatePrecision_DATE_PRECISION_YEAR,
			wantAnchor: true, // ИСПРАВЛЕНО: anchor задан
			wantText:   "1994",
		},
		{
			name:       "season with anchor and text",
			fd:         domain.FlexibleDate{Precision: domain.PrecisionSeason, Anchor: &anchor, DisplayText: "Лето 1989"},
			wantNil:    false,
			wantPrec:   emhv1.DatePrecision_DATE_PRECISION_SEASON,
			wantAnchor: true, // ИСПРАВЛЕНО: anchor задан
			wantText:   "Лето 1989",
		},
		{
			name:     "day_month without anchor",
			fd:       domain.FlexibleDate{Precision: domain.PrecisionDayMonth, DisplayText: "28 июля"},
			wantNil:  false,
			wantPrec: emhv1.DatePrecision_DATE_PRECISION_DAY_MONTH,
			wantText: "28 июля",
		},
		{
			name:    "unknown without anchor or text returns nil (IsZero)",
			fd:      domain.FlexibleDate{Precision: domain.PrecisionUnknown},
			wantNil: true, // ИСПРАВЛЕНО: IsZero() = true → nil
		},
		{
			name:     "unknown with display text",
			fd:       domain.FlexibleDate{Precision: domain.PrecisionUnknown, DisplayText: "дата неизвестна"},
			wantNil:  false,
			wantPrec: emhv1.DatePrecision_DATE_PRECISION_UNKNOWN,
			wantText: "дата неизвестна",
		},
		{
			name:     "range (reserved)",
			fd:       domain.FlexibleDate{Precision: domain.PrecisionRange, DisplayText: "1994-1995"},
			wantNil:  false,
			wantPrec: emhv1.DatePrecision_DATE_PRECISION_RANGE,
			wantText: "1994-1995",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapFlexibleDateToProto(tt.fd)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if got.Precision != tt.wantPrec {
				t.Errorf("precision = %v, want %v", got.Precision, tt.wantPrec)
			}
			if got.DisplayText != tt.wantText {
				t.Errorf("displayText = %q, want %q", got.DisplayText, tt.wantText)
			}
			if tt.wantAnchor && got.AnchorDate == nil {
				t.Error("expected anchor to be set")
			}
			if !tt.wantAnchor && got.AnchorDate != nil {
				t.Errorf("expected no anchor, got %v", got.AnchorDate)
			}
		})
	}
}

// --- mapProtoFlexibleDate ---

// TestMapProtoFlexibleDate_NilInput проверяет, что nil возвращает (nil, nil).
func TestMapProtoFlexibleDate_NilInput(t *testing.T) {
	got, err := mapProtoFlexibleDate(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// TestMapProtoFlexibleDate_AllPrecisions проверяет маппинг proto → domain.
func TestMapProtoFlexibleDate_AllPrecisions(t *testing.T) {
	anchor := utcDate(1994, 2, 15)

	tests := []struct {
		name       string
		proto      *emhv1.FlexibleDate
		wantPrec   domain.DatePrecision
		wantAnchor bool
		wantText   string
	}{
		{
			name:       "exact",
			proto:      &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_EXACT, AnchorDate: timestamppb.New(anchor)},
			wantPrec:   domain.PrecisionExact,
			wantAnchor: true,
		},
		{
			name:       "month with text",
			proto:      &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_MONTH, AnchorDate: timestamppb.New(anchor), DisplayText: "Февраль 1994"},
			wantPrec:   domain.PrecisionMonth,
			wantAnchor: true,
			wantText:   "Февраль 1994",
		},
		{
			name:       "year with anchor and text",
			proto:      &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_YEAR, AnchorDate: timestamppb.New(anchor), DisplayText: "1994"},
			wantPrec:   domain.PrecisionYear,
			wantAnchor: true, // ИСПРАВЛЕНО: anchor задан в proto
			wantText:   "1994",
		},
		{
			name:     "season with text only",
			proto:    &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_SEASON, DisplayText: "Лето 1989"},
			wantPrec: domain.PrecisionSeason,
			wantText: "Лето 1989",
		},
		{
			name:     "day_month",
			proto:    &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_DAY_MONTH, DisplayText: "28 июля"},
			wantPrec: domain.PrecisionDayMonth,
			wantText: "28 июля",
		},
		{
			name:     "unknown",
			proto:    &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_UNKNOWN},
			wantPrec: domain.PrecisionUnknown,
		},
		{
			name:     "range (reserved)",
			proto:    &emhv1.FlexibleDate{Precision: emhv1.DatePrecision_DATE_PRECISION_RANGE, DisplayText: "1994-1995"},
			wantPrec: domain.PrecisionRange,
			wantText: "1994-1995",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapProtoFlexibleDate(tt.proto)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if got.Precision != tt.wantPrec {
				t.Errorf("precision = %v, want %v", got.Precision, tt.wantPrec)
			}
			if got.DisplayText != tt.wantText {
				t.Errorf("displayText = %q, want %q", got.DisplayText, tt.wantText)
			}
			if tt.wantAnchor && got.Anchor == nil {
				t.Error("expected anchor to be set")
			}
			if !tt.wantAnchor && got.Anchor != nil {
				t.Errorf("expected no anchor, got %v", got.Anchor)
			}
		})
	}
}

// TestMapProtoFlexibleDate_InvalidPrecision проверяет ошибку при невалидной точности.
func TestMapProtoFlexibleDate_InvalidPrecision(t *testing.T) {
	proto := &emhv1.FlexibleDate{
		Precision: emhv1.DatePrecision_DATE_PRECISION_UNSPECIFIED, // 0 = unspecified, невалидно в domain
	}
	_, err := mapProtoFlexibleDate(proto)
	if err == nil {
		t.Fatal("expected error for unspecified precision")
	}
	if !errors.Is(err, domain.ErrInvalidDatePrecision) {
		t.Errorf("expected ErrInvalidDatePrecision, got %v", err)
	}
}

// TestMapProtoFlexibleDate_ExactWithoutAnchor проверяет ошибку при EXACT без anchor.
func TestMapProtoFlexibleDate_ExactWithoutAnchor(t *testing.T) {
	proto := &emhv1.FlexibleDate{
		Precision: emhv1.DatePrecision_DATE_PRECISION_EXACT,
		// AnchorDate не задан
	}
	_, err := mapProtoFlexibleDate(proto)
	if err == nil {
		t.Fatal("expected error for EXACT without anchor")
	}
	if !errors.Is(err, domain.ErrExactDateRequiresAnchor) {
		t.Errorf("expected ErrExactDateRequiresAnchor, got %v", err)
	}
}

// TestMapProtoFlexibleDate_UnknownWithAnchor проверяет ошибку при UNKNOWN с anchor.
func TestMapProtoFlexibleDate_UnknownWithAnchor(t *testing.T) {
	anchor := utcDate(1994, 2, 15)
	proto := &emhv1.FlexibleDate{
		Precision:  emhv1.DatePrecision_DATE_PRECISION_UNKNOWN,
		AnchorDate: timestamppb.New(anchor),
	}
	_, err := mapProtoFlexibleDate(proto)
	if err == nil {
		t.Fatal("expected error for UNKNOWN with anchor")
	}
	if !errors.Is(err, domain.ErrUnknownDateMustNotHaveAnchor) {
		t.Errorf("expected ErrUnknownDateMustNotHaveAnchor, got %v", err)
	}
}

// --- Round-trip: domain → proto → domain ---

// TestFlexibleDate_RoundTrip проверяет, что преобразование туда-обратно сохраняет данные.
func TestFlexibleDate_RoundTrip(t *testing.T) {
	anchor := utcDate(1994, 2, 15)
	originals := []domain.FlexibleDate{
		{Precision: domain.PrecisionExact, Anchor: &anchor, DisplayText: "15 февраля 1994"},
		{Precision: domain.PrecisionMonth, Anchor: &anchor, DisplayText: "Февраль 1994"},
		{Precision: domain.PrecisionYear, Anchor: &anchor, DisplayText: "1994"},
		{Precision: domain.PrecisionSeason, Anchor: &anchor, DisplayText: "Лето 1989"},
		{Precision: domain.PrecisionDayMonth, DisplayText: "28 июля"},
		{Precision: domain.PrecisionUnknown, DisplayText: "неизвестно"},
	}

	for i, orig := range originals {
		proto := mapFlexibleDateToProto(orig)
		if proto == nil {
			t.Errorf("[%d] mapFlexibleDateToProto returned nil", i)
			continue
		}
		restored, err := mapProtoFlexibleDate(proto)
		if err != nil {
			t.Errorf("[%d] mapProtoFlexibleDate error: %v", i, err)
			continue
		}
		if restored == nil {
			t.Errorf("[%d] mapProtoFlexibleDate returned nil", i)
			continue
		}
		if restored.Precision != orig.Precision {
			t.Errorf("[%d] precision: got %v, want %v", i, restored.Precision, orig.Precision)
		}
		if restored.DisplayText != orig.DisplayText {
			t.Errorf("[%d] displayText: got %q, want %q", i, restored.DisplayText, orig.DisplayText)
		}
		if orig.Anchor != nil && restored.Anchor == nil {
			t.Errorf("[%d] anchor lost in round-trip", i)
		}
	}
}

// --- resolveCreateFlexibleDate ---

// TestResolveCreateFlexibleDate_InfoHasPriority проверяет приоритет info над legacy.
func TestResolveCreateFlexibleDate_InfoHasPriority(t *testing.T) {
	info := &emhv1.FlexibleDate{
		Precision:   emhv1.DatePrecision_DATE_PRECISION_YEAR,
		AnchorDate:  timestamppb.New(utcDate(1994, 1, 1)),
		DisplayText: "1994",
	}
	legacy := "1995-06-15" // игнорируется

	got, err := resolveCreateFlexibleDate(info, legacy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Precision != domain.PrecisionYear {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionYear)
	}
	if got.DisplayText != "1994" {
		t.Errorf("displayText = %q, want %q", got.DisplayText, "1994")
	}
}

// TestResolveCreateFlexibleDate_LegacyFallback проверяет fallback на legacy при отсутствии info.
func TestResolveCreateFlexibleDate_LegacyFallback(t *testing.T) {
	got, err := resolveCreateFlexibleDate(nil, "1994-02-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Precision != domain.PrecisionExact {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionExact)
	}
	if got.Anchor == nil {
		t.Fatal("anchor is nil")
	}
	if got.Anchor.Year() != 1994 || got.Anchor.Month() != 2 || got.Anchor.Day() != 15 {
		t.Errorf("anchor = %v, want 1994-02-15", got.Anchor)
	}
}

// TestResolveCreateFlexibleDate_EmptyLegacy_ReturnsUnknown проверяет, что пустой legacy даёт Unknown.
func TestResolveCreateFlexibleDate_EmptyLegacy_ReturnsUnknown(t *testing.T) {
	got, err := resolveCreateFlexibleDate(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Precision != domain.PrecisionUnknown {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionUnknown)
	}
}

// TestResolveCreateFlexibleDate_BothNil_ReturnsUnknown проверяет, что оба nil дают Unknown.
func TestResolveCreateFlexibleDate_BothNil_ReturnsUnknown(t *testing.T) {
	got, err := resolveCreateFlexibleDate(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Precision != domain.PrecisionUnknown {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionUnknown)
	}
}

// TestResolveCreateFlexibleDate_InvalidLegacyFormat проверяет ошибку при невалидном legacy формате.
func TestResolveCreateFlexibleDate_InvalidLegacyFormat(t *testing.T) {
	_, err := resolveCreateFlexibleDate(nil, "15.02.1994")
	if err == nil {
		t.Fatal("expected error for invalid legacy format")
	}
	if !errors.Is(err, domain.ErrInvalidDateFormat) {
		t.Errorf("expected ErrInvalidDateFormat, got %v", err)
	}
}

// TestResolveCreateFlexibleDate_InvalidInfo проверяет ошибку при невалидном info.
func TestResolveCreateFlexibleDate_InvalidInfo(t *testing.T) {
	info := &emhv1.FlexibleDate{
		Precision: emhv1.DatePrecision_DATE_PRECISION_EXACT,
		// без AnchorDate — невалидно
	}
	_, err := resolveCreateFlexibleDate(info, "")
	if err == nil {
		t.Fatal("expected error for invalid info")
	}
	if !errors.Is(err, domain.ErrExactDateRequiresAnchor) {
		t.Errorf("expected ErrExactDateRequiresAnchor, got %v", err)
	}
}

// --- resolveUpdateFlexibleDate ---

// TestResolveUpdateFlexibleDate_InfoHasPriority проверяет приоритет info над legacy.
func TestResolveUpdateFlexibleDate_InfoHasPriority(t *testing.T) {
	info := &emhv1.FlexibleDate{
		Precision:   emhv1.DatePrecision_DATE_PRECISION_YEAR,
		AnchorDate:  timestamppb.New(utcDate(1994, 1, 1)),
		DisplayText: "1994",
	}
	legacy := "1995-06-15" // игнорируется

	got, err := resolveUpdateFlexibleDate(info, legacy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Precision != domain.PrecisionYear {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionYear)
	}
}

// TestResolveUpdateFlexibleDate_LegacyFallback проверяет fallback на legacy.
func TestResolveUpdateFlexibleDate_LegacyFallback(t *testing.T) {
	got, err := resolveUpdateFlexibleDate(nil, "1994-02-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Precision != domain.PrecisionExact {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionExact)
	}
}

// TestResolveUpdateFlexibleDate_BothEmpty_ReturnsNil проверяет, что оба пустые дают nil.
func TestResolveUpdateFlexibleDate_BothEmpty_ReturnsNil(t *testing.T) {
	got, err := resolveUpdateFlexibleDate(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// TestResolveUpdateFlexibleDate_InvalidLegacy проверяет ошибку при невалидном legacy.
func TestResolveUpdateFlexibleDate_InvalidLegacy(t *testing.T) {
	_, err := resolveUpdateFlexibleDate(nil, "invalid-date")
	if err == nil {
		t.Fatal("expected error for invalid legacy")
	}
	if !errors.Is(err, domain.ErrInvalidDateFormat) {
		t.Errorf("expected ErrInvalidDateFormat, got %v", err)
	}
}

// --- parseFlexibleDate (private helper) ---

// TestParseFlexibleDate_ValidFormats проверяет парсинг валидных форматов через resolveCreate.
// parseFlexibleDate — приватная функция, тестируем через resolveCreateFlexibleDate.
func TestParseFlexibleDate_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		year  int
		month time.Month
		day   int
	}{
		{"YYYY-MM-DD", "1994-02-15", 1994, 2, 15},
		{"RFC3339", "1994-02-15T10:30:00Z", 1994, 2, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCreateFlexibleDate(nil, tt.input)
			if err != nil {
				t.Fatalf("resolveCreate failed: %v", err)
			}
			if got.Anchor == nil {
				t.Fatal("anchor is nil")
			}
			if got.Anchor.Year() != tt.year || got.Anchor.Month() != tt.month || got.Anchor.Day() != tt.day {
				t.Errorf("date = %v, want %d-%02d-%02d", got.Anchor, tt.year, tt.month, tt.day)
			}
		})
	}
}

// TestParseFlexibleDate_EmptyString_ReturnsUnknown проверяет, что пустая строка → Unknown.
func TestParseFlexibleDate_EmptyString_ReturnsUnknown(t *testing.T) {
	got, err := resolveCreateFlexibleDate(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Precision != domain.PrecisionUnknown {
		t.Errorf("precision = %v, want %v", got.Precision, domain.PrecisionUnknown)
	}
}
