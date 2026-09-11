package domain

import (
	"errors"
	"testing"
	"time"
)

// ptrTime возвращает указатель на time (хелпер для табличных тестов).
func ptrTime(t time.Time) *time.Time { return &t }

// date создаёт дату в UTC (хелпер).
func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// --- Константы DatePrecision ---

// TestDatePrecision_Values фиксирует числовые значения точности дат.
// Они хранятся в БД как SMALLINT, поэтому изменение порядка сломает данные.
func TestDatePrecision_Values(t *testing.T) {
	tests := []struct {
		precision DatePrecision
		expected  int
	}{
		{PrecisionUnspecified, 0},
		{PrecisionExact, 1},
		{PrecisionMonth, 2},
		{PrecisionYear, 3},
		{PrecisionSeason, 4},
		{PrecisionDayMonth, 5},
		{PrecisionRange, 6},
		{PrecisionUnknown, 7},
	}
	for _, tt := range tests {
		if int(tt.precision) != tt.expected {
			t.Errorf("precision %v = %d, want %d", tt.precision, int(tt.precision), tt.expected)
		}
	}
}

// --- Конструкторы ---

// TestNewExactDate проверяет создание точной даты.
func TestNewExactDate(t *testing.T) {
	anchor := date(1994, 2, 15)
	fd := NewExactDate(anchor)

	if fd.Precision != PrecisionExact {
		t.Errorf("precision = %v, want %v", fd.Precision, PrecisionExact)
	}
	if fd.Anchor == nil {
		t.Fatal("anchor is nil, want non-nil")
	}
	if !fd.Anchor.Equal(anchor) {
		t.Errorf("anchor = %v, want %v", *fd.Anchor, anchor)
	}
	if fd.DisplayText != "" {
		t.Errorf("displayText = %q, want empty", fd.DisplayText)
	}
}

// TestNewExactDate_ZeroTime проверяет, что zero-time превращается в Unknown.
func TestNewExactDate_ZeroTime(t *testing.T) {
	fd := NewExactDate(time.Time{})

	if fd.Precision != PrecisionUnknown {
		t.Errorf("precision = %v, want %v", fd.Precision, PrecisionUnknown)
	}
	if fd.Anchor != nil {
		t.Errorf("anchor = %v, want nil", fd.Anchor)
	}
}

// TestNewUnknownDate проверяет создание неизвестной даты.
func TestNewUnknownDate(t *testing.T) {
	fd := NewUnknownDate()

	if fd.Precision != PrecisionUnknown {
		t.Errorf("precision = %v, want %v", fd.Precision, PrecisionUnknown)
	}
	if fd.Anchor != nil {
		t.Errorf("anchor = %v, want nil", fd.Anchor)
	}
	if fd.DisplayText != "" {
		t.Errorf("displayText = %q, want empty", fd.DisplayText)
	}
}

// --- IsZero ---

// TestFlexibleDate_IsZero проверяет определение фактически пустой даты.
func TestFlexibleDate_IsZero(t *testing.T) {
	anchor := date(1994, 2, 15)

	tests := []struct {
		name     string
		fd       FlexibleDate
		expected bool
	}{
		{"zero value", FlexibleDate{}, true},
		{"unknown without text", FlexibleDate{Precision: PrecisionUnknown}, true},
		{"unspecified without text", FlexibleDate{Precision: PrecisionUnspecified}, true},
		{"unknown with text", FlexibleDate{Precision: PrecisionUnknown, DisplayText: "неизвестно"}, false},
		{"unspecified with text", FlexibleDate{Precision: PrecisionUnspecified, DisplayText: "?"}, false},
		{"unknown with anchor", FlexibleDate{Precision: PrecisionUnknown, Anchor: &anchor}, false},
		{"exact with anchor", FlexibleDate{Precision: PrecisionExact, Anchor: &anchor}, false},
		{"exact without anchor", FlexibleDate{Precision: PrecisionExact}, false},
		{"day_month with text", FlexibleDate{Precision: PrecisionDayMonth, DisplayText: "28 июля"}, false},
		{"month with anchor", FlexibleDate{Precision: PrecisionMonth, Anchor: &anchor}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fd.IsZero(); got != tt.expected {
				t.Errorf("IsZero() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// --- IsValid ---

// TestFlexibleDate_IsValid проверяет внутреннюю консистентность гибкой даты.
// Покрывает все ветки функции.
func TestFlexibleDate_IsValid(t *testing.T) {
	anchor := date(1994, 2, 15)

	tests := []struct {
		name    string
		fd      FlexibleDate
		wantErr error // nil = валидна
	}{
		{"unspecified rejected", FlexibleDate{Precision: PrecisionUnspecified}, ErrInvalidDatePrecision},
		{"exact with anchor valid", FlexibleDate{Precision: PrecisionExact, Anchor: &anchor}, nil},
		{"exact without anchor invalid", FlexibleDate{Precision: PrecisionExact}, ErrExactDateRequiresAnchor},
		{"month with anchor valid", FlexibleDate{Precision: PrecisionMonth, Anchor: &anchor}, nil},
		{"month without anchor valid", FlexibleDate{Precision: PrecisionMonth}, nil},
		{"year with anchor valid", FlexibleDate{Precision: PrecisionYear, Anchor: &anchor}, nil},
		{"year without anchor valid", FlexibleDate{Precision: PrecisionYear}, nil},
		{"season with anchor valid", FlexibleDate{Precision: PrecisionSeason, Anchor: &anchor}, nil},
		{"day_month without anchor valid", FlexibleDate{Precision: PrecisionDayMonth, DisplayText: "28 июля"}, nil},
		{"day_month with anchor invalid", FlexibleDate{Precision: PrecisionDayMonth, Anchor: &anchor}, ErrDayMonthMustNotHaveAnchor},
		{"range valid (reserved)", FlexibleDate{Precision: PrecisionRange}, nil},
		{"unknown without anchor valid", FlexibleDate{Precision: PrecisionUnknown}, nil},
		{"unknown with anchor invalid", FlexibleDate{Precision: PrecisionUnknown, Anchor: &anchor}, ErrUnknownDateMustNotHaveAnchor},
		{"out-of-range high invalid", FlexibleDate{Precision: DatePrecision(99)}, ErrInvalidDatePrecision},
		{"out-of-range negative invalid", FlexibleDate{Precision: DatePrecision(-1)}, ErrInvalidDatePrecision},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fd.IsValid()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("IsValid() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("IsValid() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestFlexibleDate_IsValid_DisplayTextIndependent проверяет, что
// наличие/отсутствие DisplayText не влияет на валидацию.
func TestFlexibleDate_IsValid_DisplayTextIndependent(t *testing.T) {
	anchor := date(1994, 2, 15)

	withText := FlexibleDate{Precision: PrecisionMonth, Anchor: &anchor, DisplayText: "Февраль 1994"}
	withoutText := FlexibleDate{Precision: PrecisionMonth, Anchor: &anchor}

	if err := withText.IsValid(); err != nil {
		t.Errorf("IsValid() with display text = %v, want nil", err)
	}
	if err := withoutText.IsValid(); err != nil {
		t.Errorf("IsValid() without display text = %v, want nil", err)
	}
}
