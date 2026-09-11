package v1

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// errorMessages содержит русские сообщения для доменных ошибок.
// Ключ — доменный код ошибки.
// При добавлении нового sentinel error в domain/errors.go
// необходимо добавить сюда соответствующее сообщение.
var errorMessages = map[domain.ErrorCode]string{
	domain.ErrCodeNotFound:                     "Запись не найдена",
	domain.ErrCodeDeathBeforeBirth:             "Дата гибели не может быть раньше даты рождения",
	domain.ErrCodeServiceBeforeBirth:           "Дата начала службы не может быть раньше даты рождения",
	domain.ErrCodeServiceAfterDeath:            "Дата начала службы не может быть позже даты гибели",
	domain.ErrCodeRequiredFieldEmpty:           "Обязательное поле не может быть пустым",
	domain.ErrCodeStatusUnspecified:            "Статус публикации не задан",
	domain.ErrCodeStatusRequired:               "Статус публикации обязателен при обновлении",
	domain.ErrCodeDuplicateFieldMask:           "Дублирование поля в field_mask",
	domain.ErrCodePhotosNotBelongToHero:        "Некоторые фотографии не принадлежат данному герою",
	domain.ErrCodeSelfRelation:                 "Нельзя создать связь героя с самим собой",
	domain.ErrCodeInvalidDateFormat:            "Неверный формат даты",
	domain.ErrCodeInvalidDatePrecision:         "Неверная точность даты",
	domain.ErrCodeExactDateRequiresAnchor:      "Точная дата требует указания даты",
	domain.ErrCodeUnknownDateMustNotHaveAnchor: "Неизвестная дата не должна содержать дату",
	domain.ErrCodeDayMonthMustNotHaveAnchor:    "Дата без года не должна содержать дату",
	domain.ErrCodeHeroIDRequired:               "ID героя обязателен",
	domain.ErrCodePhotoIDsEmpty:                "Список ID фотографий пуст",
	domain.ErrCodePhotosEmpty:                  "Список фотографий пуст",
	domain.ErrCodePhotoURLRequired:             "URL фотографии обязателен",
	domain.ErrCodeContactRateLimited:           "Слишком много сообщений. Попробуйте позже.",
	domain.ErrCodeContactSMTP:                  "Не удалось отправить уведомление. Попробуйте позже.",

	// Справочники
	domain.ErrCodeEntityAssignedToHeroes: "Запись привязана к героям и не может быть удалена",
	domain.ErrCodeOwnParent:              "Объект не может быть собственным родителем",
	domain.ErrCodeConflictDateInvalid:    "Дата окончания не может быть раньше даты начала",
	domain.ErrCodeInvalidLatitude:        "Широта должна быть в диапазоне от -90 до 90",
	domain.ErrCodeInvalidLongitude:       "Долгота должна быть в диапазоне от -180 до 180",
	domain.ErrCodeAwardNameRequired:      "Название награды обязательно",
	domain.ErrCodeConflictNameRequired:   "Название конфликта обязательно",
	domain.ErrCodeLocationNameRequired:   "Название локации обязательно",

	// API Keys
	domain.ErrCodeAPIKeyAlreadyRevoked: "Ключ уже отозван",
	domain.ErrCodeAPIKeyNameRequired:   "Имя ключа обязательно",

	// Media
	domain.ErrCodeContentTypeNotAllowed: "Недопустимый тип файла",
	domain.ErrCodeUnsupportedUploadType: "Неподдерживаемый тип загрузки",

	// Submission
	domain.ErrCodeSubmitterNameRequired:     "Имя отправителя обязательно",
	domain.ErrCodeInvalidSubmitterEmail:     "Некорректный email отправителя",
	domain.ErrCodeInvalidPayloadJSON:        "Некорректный формат данных заявки",
	domain.ErrCodeSubmissionAlreadyReviewed: "Заявка уже обработана",
	domain.ErrCodeSubmissionDuplicate:       "Заявка с таким содержанием уже существует",

	domain.ErrCodeCyclicReference: "Обнаружена циклическая ссылка в иерархии",

	// LLM
	domain.ErrCodeLLMUnavailable:   "Сервис извлечения данных временно недоступен",
	domain.ErrCodeLLMTimeout:       "Превышено время ожидания ответа от LLM",
	domain.ErrCodeLLMParseFailed:   "Не удалось распознать структурированные данные в тексте",
	domain.ErrCodeLLMDisabled:      "Функция извлечения данных отключена администратором",
	domain.ErrCodeLLMInputTooLarge: "Входной текст превышает максимально допустимый размер",
}

// localizedMessage возвращает русское сообщение для доменного кода ошибки.
func localizedMessage(code domain.ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Внутренняя ошибка сервера"
}

// connectCodeFromDomainCode маппит доменный код ошибки на Connect-код.
func connectCodeFromDomainCode(code domain.ErrorCode) connect.Code {
	switch code {
	case domain.ErrCodeNotFound:
		return connect.CodeNotFound
	case domain.ErrCodeInternal:
		return connect.CodeInternal
	case domain.ErrCodeContactRateLimited:
		return connect.CodeResourceExhausted
	case domain.ErrCodeContactSMTP:
		return connect.CodeUnavailable
	case domain.ErrCodeEntityAssignedToHeroes:
		return connect.CodeFailedPrecondition
	case domain.ErrCodeSubmissionAlreadyReviewed:
		return connect.CodeFailedPrecondition
	case domain.ErrCodeAPIKeyAlreadyRevoked:
		return connect.CodeNotFound
	case domain.ErrCodeCyclicReference:
		return connect.CodeFailedPrecondition
	case domain.ErrCodeLLMUnavailable:
		return connect.CodeUnavailable
	case domain.ErrCodeLLMTimeout:
		return connect.CodeDeadlineExceeded
	case domain.ErrCodeLLMParseFailed:
		return connect.CodeInvalidArgument
	case domain.ErrCodeLLMDisabled:
		return connect.CodeFailedPrecondition
	case domain.ErrCodeLLMInputTooLarge:
		return connect.CodeInvalidArgument
	case domain.ErrCodeSubmissionDuplicate:
		return connect.CodeAlreadyExists

	default:
		return connect.CodeInvalidArgument
	}
}

// mapDomainError маппит доменную ошибку в Connect-ошибку с локализованным сообщением.
//
// Логирование:
//   - INTERNAL логируется как Error (с оригинальной ошибкой для отладки)
//   - Остальные логируются как Warn (с кодом для метрик)
//
// Клиенту возвращается локализованное сообщение.
// Оригинальная ошибка не передаётся клиенту (безопасность).
func mapDomainError(ctx context.Context, logger *slog.Logger, err error, action string) error {
	code := domain.CodeFromError(err)

	// метрика ошибки
	layer := "domain"
	if code == domain.ErrCodeInternal {
		layer = "infra"
	}
	metrics.ErrorsTotal.WithLabelValues(string(code), layer).Inc()

	// Логируем с кодом для метрик
	if code == domain.ErrCodeInternal {
		logger.ErrorContext(ctx, action,
			"error", err,
			"error_code", string(code),
		)
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	logger.WarnContext(ctx, action,
		"error", err,
		"error_code", string(code),
	)

	connectCode := connectCodeFromDomainCode(code)
	msg := localizedMessage(code)

	return connect.NewError(connectCode, errors.New(msg))
}
