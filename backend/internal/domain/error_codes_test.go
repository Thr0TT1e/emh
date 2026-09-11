package domain

import (
	"errors"
	"fmt"
	"testing"
)

// TestCodeFromError_AllSentinels проверяет маппинг каждой ошибки на её код.
// При добавлении новой ошибки в errors.go + error_codes.go добавь кейс сюда.
func TestCodeFromError_AllSentinels(t *testing.T) {
	tests := []struct {
		err      error
		expected ErrorCode
	}{
		{ErrNotFound, ErrCodeNotFound},
		{ErrDeathBeforeBirth, ErrCodeDeathBeforeBirth},
		{ErrServiceBeforeBirth, ErrCodeServiceBeforeBirth},
		{ErrServiceAfterDeath, ErrCodeServiceAfterDeath},
		{ErrRequiredFieldEmpty, ErrCodeRequiredFieldEmpty},
		{ErrStatusUnspecified, ErrCodeStatusUnspecified},
		{ErrStatusRequired, ErrCodeStatusRequired},
		{ErrDuplicateFieldMask, ErrCodeDuplicateFieldMask},
		{ErrPhotosNotBelongToHero, ErrCodePhotosNotBelongToHero},
		{ErrSelfRelation, ErrCodeSelfRelation},
		{ErrInvalidDateFormat, ErrCodeInvalidDateFormat},
		{ErrInvalidDatePrecision, ErrCodeInvalidDatePrecision},
		{ErrExactDateRequiresAnchor, ErrCodeExactDateRequiresAnchor},
		{ErrUnknownDateMustNotHaveAnchor, ErrCodeUnknownDateMustNotHaveAnchor},
		{ErrDayMonthMustNotHaveAnchor, ErrCodeDayMonthMustNotHaveAnchor},
		{ErrHeroIDRequired, ErrCodeHeroIDRequired},
		{ErrPhotoIDsEmpty, ErrCodePhotoIDsEmpty},
		{ErrPhotosEmpty, ErrCodePhotosEmpty},
		{ErrPhotoURLRequired, ErrCodePhotoURLRequired},
	}
	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			if got := CodeFromError(tt.err); got != tt.expected {
				t.Errorf("CodeFromError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

// TestCodeFromError_Wrapped проверяет устойчивость к обёрткам fmt.Errorf.
func TestCodeFromError_Wrapped(t *testing.T) {
	wrapped := fmt.Errorf("get hero: %w", ErrNotFound)
	if got := CodeFromError(wrapped); got != ErrCodeNotFound {
		t.Errorf("CodeFromError(wrapped) = %v, want %v", got, ErrCodeNotFound)
	}

	doubleWrapped := fmt.Errorf("usecase: %w", fmt.Errorf("repo: %w", ErrDeathBeforeBirth))
	if got := CodeFromError(doubleWrapped); got != ErrCodeDeathBeforeBirth {
		t.Errorf("CodeFromError(double wrapped) = %v, want %v", got, ErrCodeDeathBeforeBirth)
	}
}

// TestCodeFromError_Join проверяет ошибки, созданные через errors.Join.
func TestCodeFromError_Join(t *testing.T) {
	joined := errors.Join(errors.New("context"), ErrRequiredFieldEmpty)
	if got := CodeFromError(joined); got != ErrCodeRequiredFieldEmpty {
		t.Errorf("CodeFromError(joined) = %v, want %v", got, ErrCodeRequiredFieldEmpty)
	}
}

// TestCodeFromError_Unknown проверяет, что неизвестная ошибка маппится на INTERNAL.
func TestCodeFromError_Unknown(t *testing.T) {
	if got := CodeFromError(errors.New("some random error")); got != ErrCodeInternal {
		t.Errorf("CodeFromError(unknown) = %v, want %v", got, ErrCodeInternal)
	}
}

// TestCodeFromError_Nil проверяет, что nil маппится на INTERNAL (default case).
func TestCodeFromError_Nil(t *testing.T) {
	if got := CodeFromError(nil); got != ErrCodeInternal {
		t.Errorf("CodeFromError(nil) = %v, want %v", got, ErrCodeInternal)
	}
}
