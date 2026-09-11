package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// discardTestLogger возвращает логгер, который ничего не пишет.
func discardTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- CodeFromError: полный маппинг всех sentinel errors ---

// TestCodeFromError_AllSentinels проверяет маппинг каждой из 19 ошибок на её код.
// Это smoke-тест: если мы добавим новый sentinel error и забудем его в CodeFromError,
// тест упадёт (ошибка будет маппиться на INTERNAL).
func TestCodeFromError_AllSentinels(t *testing.T) {
	tests := []struct {
		err      error
		expected domain.ErrorCode
	}{
		{domain.ErrNotFound, domain.ErrCodeNotFound},
		{domain.ErrDeathBeforeBirth, domain.ErrCodeDeathBeforeBirth},
		{domain.ErrServiceBeforeBirth, domain.ErrCodeServiceBeforeBirth},
		{domain.ErrServiceAfterDeath, domain.ErrCodeServiceAfterDeath},
		{domain.ErrRequiredFieldEmpty, domain.ErrCodeRequiredFieldEmpty},
		{domain.ErrStatusUnspecified, domain.ErrCodeStatusUnspecified},
		{domain.ErrStatusRequired, domain.ErrCodeStatusRequired},
		{domain.ErrDuplicateFieldMask, domain.ErrCodeDuplicateFieldMask},
		{domain.ErrPhotosNotBelongToHero, domain.ErrCodePhotosNotBelongToHero},
		{domain.ErrSelfRelation, domain.ErrCodeSelfRelation},
		{domain.ErrInvalidDateFormat, domain.ErrCodeInvalidDateFormat},
		{domain.ErrInvalidDatePrecision, domain.ErrCodeInvalidDatePrecision},
		{domain.ErrExactDateRequiresAnchor, domain.ErrCodeExactDateRequiresAnchor},
		{domain.ErrUnknownDateMustNotHaveAnchor, domain.ErrCodeUnknownDateMustNotHaveAnchor},
		{domain.ErrDayMonthMustNotHaveAnchor, domain.ErrCodeDayMonthMustNotHaveAnchor},
		{domain.ErrHeroIDRequired, domain.ErrCodeHeroIDRequired},
		{domain.ErrPhotoIDsEmpty, domain.ErrCodePhotoIDsEmpty},
		{domain.ErrPhotosEmpty, domain.ErrCodePhotosEmpty},
		{domain.ErrPhotoURLRequired, domain.ErrCodePhotoURLRequired},
	}
	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			if got := domain.CodeFromError(tt.err); got != tt.expected {
				t.Errorf("CodeFromError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

// TestCodeFromError_Wrapped проверяет устойчивость к обёрткам fmt.Errorf.
func TestCodeFromError_Wrapped(t *testing.T) {
	wrapped := fmt.Errorf("get hero: %w", domain.ErrNotFound)
	if got := domain.CodeFromError(wrapped); got != domain.ErrCodeNotFound {
		t.Errorf("CodeFromError(wrapped) = %v, want %v", got, domain.ErrCodeNotFound)
	}

	doubleWrapped := fmt.Errorf("usecase: %w", fmt.Errorf("repo: %w", domain.ErrDeathBeforeBirth))
	if got := domain.CodeFromError(doubleWrapped); got != domain.ErrCodeDeathBeforeBirth {
		t.Errorf("CodeFromError(double wrapped) = %v, want %v", got, domain.ErrCodeDeathBeforeBirth)
	}
}

// TestCodeFromError_Join проверяет ошибки, созданные через errors.Join.
func TestCodeFromError_Join(t *testing.T) {
	joined := errors.Join(errors.New("context"), domain.ErrRequiredFieldEmpty)
	if got := domain.CodeFromError(joined); got != domain.ErrCodeRequiredFieldEmpty {
		t.Errorf("CodeFromError(joined) = %v, want %v", got, domain.ErrCodeRequiredFieldEmpty)
	}
}

// TestCodeFromError_Unknown проверяет, что неизвестная ошибка маппится на INTERNAL.
func TestCodeFromError_Unknown(t *testing.T) {
	if got := domain.CodeFromError(errors.New("some random error")); got != domain.ErrCodeInternal {
		t.Errorf("CodeFromError(unknown) = %v, want %v", got, domain.ErrCodeInternal)
	}
}

// TestCodeFromError_Nil проверяет, что nil маппится на INTERNAL (default case).
func TestCodeFromError_Nil(t *testing.T) {
	if got := domain.CodeFromError(nil); got != domain.ErrCodeInternal {
		t.Errorf("CodeFromError(nil) = %v, want %v", got, domain.ErrCodeInternal)
	}
}

// --- connectCodeFromDomainCode ---

// TestConnectCodeFromDomainCode проверяет маппинг доменных кодов на Connect-коды.
// Логика:
// - NOT_FOUND → CodeNotFound (для 404 в API)
// - INTERNAL → CodeInternal (для 500)
// - Все остальные → CodeInvalidArgument (для 400)
func TestConnectCodeFromDomainCode(t *testing.T) {
	tests := []struct {
		code     domain.ErrorCode
		expected connect.Code
	}{
		{domain.ErrCodeNotFound, connect.CodeNotFound},
		{domain.ErrCodeInternal, connect.CodeInternal},
		// Все остальные маппятся на InvalidArgument.
		{domain.ErrCodeDeathBeforeBirth, connect.CodeInvalidArgument},
		{domain.ErrCodeServiceBeforeBirth, connect.CodeInvalidArgument},
		{domain.ErrCodeServiceAfterDeath, connect.CodeInvalidArgument},
		{domain.ErrCodeRequiredFieldEmpty, connect.CodeInvalidArgument},
		{domain.ErrCodeStatusUnspecified, connect.CodeInvalidArgument},
		{domain.ErrCodeStatusRequired, connect.CodeInvalidArgument},
		{domain.ErrCodeDuplicateFieldMask, connect.CodeInvalidArgument},
		{domain.ErrCodePhotosNotBelongToHero, connect.CodeInvalidArgument},
		{domain.ErrCodeSelfRelation, connect.CodeInvalidArgument},
		{domain.ErrCodeInvalidDateFormat, connect.CodeInvalidArgument},
		{domain.ErrCodeInvalidDatePrecision, connect.CodeInvalidArgument},
		{domain.ErrCodeExactDateRequiresAnchor, connect.CodeInvalidArgument},
		{domain.ErrCodeUnknownDateMustNotHaveAnchor, connect.CodeInvalidArgument},
		{domain.ErrCodeDayMonthMustNotHaveAnchor, connect.CodeInvalidArgument},
		{domain.ErrCodeHeroIDRequired, connect.CodeInvalidArgument},
		{domain.ErrCodePhotoIDsEmpty, connect.CodeInvalidArgument},
		{domain.ErrCodePhotosEmpty, connect.CodeInvalidArgument},
		{domain.ErrCodePhotoURLRequired, connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := connectCodeFromDomainCode(tt.code); got != tt.expected {
				t.Errorf("connectCodeFromDomainCode(%q) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

// TestConnectCodeFromDomainCode_UnknownCode проверяет, что неизвестный код
// маппится на InvalidArgument (default case).
func TestConnectCodeFromDomainCode_UnknownCode(t *testing.T) {
	if got := connectCodeFromDomainCode("SOME_UNKNOWN_CODE"); got != connect.CodeInvalidArgument {
		t.Errorf("connectCodeFromDomainCode(unknown) = %v, want %v", got, connect.CodeInvalidArgument)
	}
}

// --- localizedMessage ---

// TestLocalizedMessage_AllCodes проверяет, что для каждого кода есть русское сообщение.
// Smoke-тест: если добавлен новый код в error_codes.go без сообщения в errorMessages,
// вернётся fallback "Внутренняя ошибка сервера" и тест упадёт.
func TestLocalizedMessage_AllCodes(t *testing.T) {
	codes := []domain.ErrorCode{
		domain.ErrCodeNotFound,
		domain.ErrCodeDeathBeforeBirth,
		domain.ErrCodeServiceBeforeBirth,
		domain.ErrCodeServiceAfterDeath,
		domain.ErrCodeRequiredFieldEmpty,
		domain.ErrCodeStatusUnspecified,
		domain.ErrCodeStatusRequired,
		domain.ErrCodeDuplicateFieldMask,
		domain.ErrCodePhotosNotBelongToHero,
		domain.ErrCodeSelfRelation,
		domain.ErrCodeInvalidDateFormat,
		domain.ErrCodeInvalidDatePrecision,
		domain.ErrCodeExactDateRequiresAnchor,
		domain.ErrCodeUnknownDateMustNotHaveAnchor,
		domain.ErrCodeDayMonthMustNotHaveAnchor,
		domain.ErrCodeHeroIDRequired,
		domain.ErrCodePhotoIDsEmpty,
		domain.ErrCodePhotosEmpty,
		domain.ErrCodePhotoURLRequired,
	}

	for _, code := range codes {
		t.Run(string(code), func(t *testing.T) {
			msg := localizedMessage(code)
			if msg == "" {
				t.Errorf("localizedMessage(%q) is empty", code)
			}
			if msg == "Внутренняя ошибка сервера" {
				t.Errorf("localizedMessage(%q) returned fallback — отсутствует в errorMessages", code)
			}
			// Проверяем, что сообщение на русском (кириллица присутствует).
			hasCyrillic := false
			for _, r := range msg {
				if r >= 'А' && r <= 'я' {
					hasCyrillic = true
					break
				}
			}
			if !hasCyrillic {
				t.Errorf("localizedMessage(%q) = %q — не содержит кириллицы", code, msg)
			}
		})
	}
}

// TestLocalizedMessage_UnknownCode_ReturnsFallback проверяет fallback на неизвестный код.
func TestLocalizedMessage_UnknownCode_ReturnsFallback(t *testing.T) {
	msg := localizedMessage("SOME_UNKNOWN_CODE")
	if msg != "Внутренняя ошибка сервера" {
		t.Errorf("expected fallback message, got %q", msg)
	}
}

// --- mapDomainError ---

// TestMapDomainError_Internal проверяет, что INTERNAL ошибки маппятся на CodeInternal.
func TestMapDomainError_Internal(t *testing.T) {
	logger := discardTestLogger()
	unknownErr := errors.New("some unexpected error")
	err := mapDomainError(context.Background(), logger, unknownErr, "test action")

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
	// Сообщение должно быть "internal server error" (не локализованное).
	if connectErr.Message() != "internal server error" {
		t.Errorf("message = %q, want %q", connectErr.Message(), "internal server error")
	}
}

// TestMapDomainError_NotFound проверяет, что NOT_FOUND маппится на CodeNotFound.
func TestMapDomainError_NotFound(t *testing.T) {
	logger := discardTestLogger()
	err := mapDomainError(context.Background(), logger, domain.ErrNotFound, "test action")

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
	expectedMsg := "Запись не найдена"
	if connectErr.Message() != expectedMsg {
		t.Errorf("message = %q, want %q", connectErr.Message(), expectedMsg)
	}
}

// TestMapDomainError_BusinessRules проверяет маппинг бизнес-правил на InvalidArgument.
func TestMapDomainError_BusinessRules(t *testing.T) {
	logger := discardTestLogger()
	tests := []struct {
		err         error
		expectedMsg string
	}{
		{domain.ErrDeathBeforeBirth, "Дата гибели не может быть раньше даты рождения"},
		{domain.ErrServiceBeforeBirth, "Дата начала службы не может быть раньше даты рождения"},
		{domain.ErrServiceAfterDeath, "Дата начала службы не может быть позже даты гибели"},
		{domain.ErrRequiredFieldEmpty, "Обязательное поле не может быть пустым"},
		{domain.ErrStatusUnspecified, "Статус публикации не задан"},
		{domain.ErrDuplicateFieldMask, "Дублирование поля в field_mask"},
		{domain.ErrPhotosNotBelongToHero, "Некоторые фотографии не принадлежат данному герою"},
		{domain.ErrSelfRelation, "Нельзя создать связь героя с самим собой"},
		{domain.ErrInvalidDateFormat, "Неверный формат даты"},
		{domain.ErrInvalidDatePrecision, "Неверная точность даты"},
		{domain.ErrPhotoURLRequired, "URL фотографии обязателен"},
	}

	for _, tt := range tests {
		t.Run(string(domain.CodeFromError(tt.err)), func(t *testing.T) {
			err := mapDomainError(context.Background(), logger, tt.err, "test action")
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("expected *connect.Error, got %T", err)
			}
			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
			}
			if connectErr.Message() != tt.expectedMsg {
				t.Errorf("message = %q, want %q", connectErr.Message(), tt.expectedMsg)
			}
		})
	}
}

// TestMapDomainError_WrappedError проверяет, что обёрнутые ошибки маппятся корректно.
func TestMapDomainError_WrappedError(t *testing.T) {
	logger := discardTestLogger()
	wrappedErr := fmt.Errorf("hero repository: %w", domain.ErrNotFound)
	err := mapDomainError(context.Background(), logger, wrappedErr, "test action")

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
	if connectErr.Message() != "Запись не найдена" {
		t.Errorf("message = %q", connectErr.Message())
	}
}

// TestMapDomainError_NilError проверяет, что nil ошибка маппится на Internal.
func TestMapDomainError_NilError(t *testing.T) {
	logger := discardTestLogger()
	err := mapDomainError(context.Background(), logger, nil, "test action")

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
