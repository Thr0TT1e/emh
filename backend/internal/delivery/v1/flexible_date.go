package v1

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
)

// mapProtoFlexibleDate конвертирует proto FlexibleDate в domain.FlexibleDate.
//
// Возвращает nil, если входной FlexibleDate отсутствует.
func mapProtoFlexibleDate(fd *emhv1.FlexibleDate) (*domain.FlexibleDate, error) {
	if fd == nil {
		return nil, nil
	}

	d := &domain.FlexibleDate{
		Precision:   domain.DatePrecision(fd.Precision),
		DisplayText: fd.DisplayText,
	}

	if fd.AnchorDate != nil {
		t := fd.AnchorDate.AsTime()
		d.Anchor = &t
	}

	if err := d.IsValid(); err != nil {
		return nil, fmt.Errorf("map flexible date: %w", err)
	}

	return d, nil
}

// resolveCreateFlexibleDate выбирает дату для CreateHero.
//
// Приоритет:
// 1. новое поле *_info;
// 2. старое строковое поле YYYY-MM-DD.
//
// Если ничего не передано, возвращается неизвестная дата.
func resolveCreateFlexibleDate(info *emhv1.FlexibleDate, legacy string) (domain.FlexibleDate, error) {
	if info != nil {
		d, err := mapProtoFlexibleDate(info)
		if err != nil {
			return domain.FlexibleDate{}, err
		}

		if d == nil {
			return domain.NewUnknownDate(), nil
		}

		return *d, nil
	}

	return parseFlexibleDate(legacy)
}

// resolveUpdateFlexibleDate выбирает дату для UpdateHero.
//
// Приоритет:
// 1. новое поле *_info;
// 2. старое строковое поле YYYY-MM-DD.
//
// Возвращает nil, если дата не передана.
// Для UpdateHero nil означает, что поле не нужно трогать,
// если его нет в field_mask, или очистить, если оно есть в field_mask.
func resolveUpdateFlexibleDate(info *emhv1.FlexibleDate, legacy string) (*domain.FlexibleDate, error) {
	if info != nil {
		return mapProtoFlexibleDate(info)
	}

	return parseFlexibleDatePtr(legacy)
}

// mapFlexibleDateToProto конвертирует domain.FlexibleDate в proto FlexibleDate.
//
// Используется в публичных ответах:
// HeroSummary.birth_date_info,
// HeroSummary.death_date_info,
// HeroDetail.service_start_date_info.
func mapFlexibleDateToProto(d domain.FlexibleDate) *emhv1.FlexibleDate {
	if d.Precision == domain.PrecisionUnspecified {
		return nil
	}

	if d.IsZero() {
		return nil
	}

	fd := &emhv1.FlexibleDate{
		Precision:   emhv1.DatePrecision(d.Precision),
		DisplayText: d.DisplayText,
	}

	if d.Anchor != nil {
		fd.AnchorDate = timestamppb.New(*d.Anchor)
	}

	return fd
}
