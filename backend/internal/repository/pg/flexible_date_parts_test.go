package pg

import (
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// TestFlexibleDateParts_Nil проверяет обработку nil указателя.
func TestFlexibleDateParts_Nil(t *testing.T) {
	anchor, precision, display := flexibleDateParts(nil)
	if anchor != nil {
		t.Errorf("expected nil anchor, got %v", anchor)
	}
	if precision != int(domain.PrecisionUnknown) {
		t.Errorf("expected PrecisionUnknown (%d), got %d", domain.PrecisionUnknown, precision)
	}
	if display != nil {
		t.Errorf("expected nil display, got %v", display)
	}
}

// TestFlexibleDateParts_Unspecified проверяет обработку PrecisionUnspecified.
// Это защита от нарушения CHECK constraint в БД (precision BETWEEN 1 AND 7).
func TestFlexibleDateParts_Unspecified(t *testing.T) {
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionUnspecified,
		DisplayText: "Не указано",
	}
	anchor, precision, display := flexibleDateParts(d)
	if anchor != nil {
		t.Errorf("expected nil anchor for Unspecified, got %v", anchor)
	}
	if precision != int(domain.PrecisionUnknown) {
		t.Errorf("expected PrecisionUnknown (%d), got %d", domain.PrecisionUnknown, precision)
	}
	if display != nil {
		t.Errorf("expected nil display for Unspecified, got %v", display)
	}
}

// TestFlexibleDateParts_Exact проверяет обработку точной даты.
func TestFlexibleDateParts_Exact(t *testing.T) {
	anchor := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionExact,
		Anchor:      &anchor,
		DisplayText: "15 мая 1990",
	}
	resultAnchor, precision, display := flexibleDateParts(d)
	if resultAnchor == nil {
		t.Fatal("expected non-nil anchor for Exact")
	}
	anchorPtr, ok := resultAnchor.(*time.Time)
	if !ok {
		t.Fatalf("expected *time.Time, got %T", resultAnchor)
	}
	if !anchorPtr.Equal(anchor) {
		t.Errorf("expected anchor %v, got %v", anchor, anchorPtr)
	}
	if precision != int(domain.PrecisionExact) {
		t.Errorf("expected PrecisionExact (%d), got %d", domain.PrecisionExact, precision)
	}
	if display == nil {
		t.Fatal("expected non-nil display")
	}
	// nullableString возвращает string (не *string)
	displayStr, ok := display.(string)
	if !ok {
		t.Fatalf("expected string, got %T", display)
	}
	if displayStr != "15 мая 1990" {
		t.Errorf("expected display '15 мая 1990', got %q", displayStr)
	}
}

// TestFlexibleDateParts_Year проверяет обработку даты с точностью YEAR.
func TestFlexibleDateParts_Year(t *testing.T) {
	anchor := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionYear,
		Anchor:      &anchor,
		DisplayText: "1990",
	}
	resultAnchor, precision, display := flexibleDateParts(d)
	if resultAnchor == nil {
		t.Error("expected non-nil anchor for YEAR precision")
	}
	if precision != int(domain.PrecisionYear) {
		t.Errorf("expected PrecisionYear (%d), got %d", domain.PrecisionYear, precision)
	}
	if display == nil {
		t.Fatal("expected non-nil display")
	}
	displayStr, ok := display.(string)
	if !ok {
		t.Fatalf("expected string, got %T", display)
	}
	if displayStr != "1990" {
		t.Errorf("expected display '1990', got %q", displayStr)
	}
}

// TestFlexibleDateParts_DayMonth проверяет обработку даты DAY_MONTH (без anchor).
func TestFlexibleDateParts_DayMonth(t *testing.T) {
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionDayMonth,
		DisplayText: "28 июля",
	}
	anchor, precision, display := flexibleDateParts(d)
	if anchor != nil {
		t.Errorf("expected nil anchor for DAY_MONTH, got %v", anchor)
	}
	if precision != int(domain.PrecisionDayMonth) {
		t.Errorf("expected PrecisionDayMonth (%d), got %d", domain.PrecisionDayMonth, precision)
	}
	if display == nil {
		t.Fatal("expected non-nil display")
	}
	displayStr, ok := display.(string)
	if !ok {
		t.Fatalf("expected string, got %T", display)
	}
	if displayStr != "28 июля" {
		t.Errorf("expected display '28 июля', got %q", displayStr)
	}
}

// TestFlexibleDateParts_Unknown проверяет обработку UNKNOWN (без anchor).
func TestFlexibleDateParts_Unknown(t *testing.T) {
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionUnknown,
		DisplayText: "Дата неизвестна",
	}
	anchor, precision, display := flexibleDateParts(d)
	if anchor != nil {
		t.Errorf("expected nil anchor for UNKNOWN, got %v", anchor)
	}
	if precision != int(domain.PrecisionUnknown) {
		t.Errorf("expected PrecisionUnknown (%d), got %d", domain.PrecisionUnknown, precision)
	}
	if display == nil {
		t.Fatal("expected non-nil display")
	}
	displayStr, ok := display.(string)
	if !ok {
		t.Fatalf("expected string, got %T", display)
	}
	if displayStr != "Дата неизвестна" {
		t.Errorf("expected display 'Дата неизвестна', got %q", displayStr)
	}
}

// TestFlexibleDateParts_EmptyDisplay проверяет что пустой DisplayText становится nil.
func TestFlexibleDateParts_EmptyDisplay(t *testing.T) {
	anchor := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	d := &domain.FlexibleDate{
		Precision:   domain.PrecisionYear,
		Anchor:      &anchor,
		DisplayText: "",
	}
	_, _, display := flexibleDateParts(d)
	if display != nil {
		t.Errorf("expected nil display for empty string, got %v", display)
	}
}
