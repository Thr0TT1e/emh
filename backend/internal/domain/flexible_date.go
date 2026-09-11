package domain

import (
	"fmt"
	"time"
)

// DatePrecision определяет степень точности исторической даты.
// Значения должны совпадать с emh.v1.DatePrecision.
type DatePrecision int

const (
	// PrecisionUnspecified - точность не указана.
	// Валидная бизнес-дата не должна использовать это значение.
	PrecisionUnspecified DatePrecision = iota
	// PrecisionExact - точная дата YYYY-MM-DD.
	PrecisionExact
	// PrecisionMonth - известен только месяц и год.
	PrecisionMonth
	// PrecisionYear - известен только год.
	PrecisionYear
	// PrecisionSeason - известен сезон и год.
	PrecisionSeason
	// PrecisionDayMonth - известны день и месяц без года.
	PrecisionDayMonth
	// PrecisionRange - известен диапазон дат.
	PrecisionRange
	// PrecisionUnknown - дата неизвестна и представлена только текстом.
	PrecisionUnknown
)

// FlexibleDate представляет гибкую историческую дату.
//
// Anchor используется как машиночитаемый якорь для сортировки и фильтрации.
// DisplayText используется для пользовательского вывода.
type FlexibleDate struct {
	// Anchor - опорная дата.
	// Для точной даты это полная дата.
	// Для месяца и года рекомендуется первый день месяца или первый день года.
	// Для day_month и unknown должно быть nil.
	Anchor *time.Time

	// Precision - точность даты.
	Precision DatePrecision

	// DisplayText - человекочитаемое представление даты.
	// Например: "Февраль 1994", "1994", "28 июля".
	DisplayText string
}

// NewExactDate создает точную гибкую дату.
func NewExactDate(t time.Time) FlexibleDate {
	if t.IsZero() {
		return NewUnknownDate()
	}

	return FlexibleDate{
		Anchor:    &t,
		Precision: PrecisionExact,
	}
}

// NewUnknownDate создает неизвестную дату.
func NewUnknownDate() FlexibleDate {
	return FlexibleDate{
		Precision: PrecisionUnknown,
	}
}

// IsZero возвращает true, если дата фактически не задана.
func (d FlexibleDate) IsZero() bool {
	return d.Anchor == nil &&
		(d.Precision == PrecisionUnspecified || d.Precision == PrecisionUnknown) &&
		d.DisplayText == ""
}

// IsValid проверяет внутреннюю консистентность гибкой даты.
func (d FlexibleDate) IsValid() error {
	if d.Precision == PrecisionUnspecified {
		return ErrInvalidDatePrecision
	}

	if d.Precision < PrecisionExact || d.Precision > PrecisionUnknown {
		return fmt.Errorf("%w: %d", ErrInvalidDatePrecision, int(d.Precision))
	}

	switch d.Precision {
	case PrecisionExact:
		if d.Anchor == nil {
			return ErrExactDateRequiresAnchor
		}

	case PrecisionUnknown:
		if d.Anchor != nil {
			return ErrUnknownDateMustNotHaveAnchor
		}

	case PrecisionDayMonth:
		if d.Anchor != nil {
			return ErrDayMonthMustNotHaveAnchor
		}
	}

	return nil
}
