package domain

import "errors"

// ErrorCode доменный код ошибки для логирования и метрик.
// Код стабилен и не зависит от текста сообщения.
// Используется в slog как атрибут error_code для построения метрик.
type ErrorCode string

const (
	ErrCodeInternal                     ErrorCode = "INTERNAL"
	ErrCodeNotFound                     ErrorCode = "NOT_FOUND"
	ErrCodeDeathBeforeBirth             ErrorCode = "DEATH_BEFORE_BIRTH"
	ErrCodeServiceBeforeBirth           ErrorCode = "SERVICE_BEFORE_BIRTH"
	ErrCodeServiceAfterDeath            ErrorCode = "SERVICE_AFTER_DEATH"
	ErrCodeRequiredFieldEmpty           ErrorCode = "REQUIRED_FIELD_EMPTY"
	ErrCodeStatusUnspecified            ErrorCode = "STATUS_UNSPECIFIED"
	ErrCodeStatusRequired               ErrorCode = "STATUS_REQUIRED"
	ErrCodeDuplicateFieldMask           ErrorCode = "DUPLICATE_FIELD_MASK"
	ErrCodePhotosNotBelongToHero        ErrorCode = "PHOTOS_NOT_BELONG_TO_HERO"
	ErrCodeSelfRelation                 ErrorCode = "SELF_RELATION"
	ErrCodeInvalidDateFormat            ErrorCode = "INVALID_DATE_FORMAT"
	ErrCodeInvalidDatePrecision         ErrorCode = "INVALID_DATE_PRECISION"
	ErrCodeExactDateRequiresAnchor      ErrorCode = "EXACT_DATE_REQUIRES_ANCHOR"
	ErrCodeUnknownDateMustNotHaveAnchor ErrorCode = "UNKNOWN_DATE_MUST_NOT_HAVE_ANCHOR"
	ErrCodeDayMonthMustNotHaveAnchor    ErrorCode = "DAY_MONTH_MUST_NOT_HAVE_ANCHOR"
	ErrCodeHeroIDRequired               ErrorCode = "HERO_ID_REQUIRED"
	ErrCodePhotoIDsEmpty                ErrorCode = "PHOTO_IDS_EMPTY"
	ErrCodePhotosEmpty                  ErrorCode = "PHOTOS_EMPTY"
	ErrCodePhotoURLRequired             ErrorCode = "PHOTO_URL_REQUIRED"
	ErrCodeContactRateLimited           ErrorCode = "CONTACT_RATE_LIMITED"
	ErrCodeContactSMTP                  ErrorCode = "CONTACT_SMTP_ERROR"

	// Справочники (Award, Conflict, Location)
	ErrCodeEntityAssignedToHeroes ErrorCode = "ENTITY_ASSIGNED_TO_HEROES"
	ErrCodeOwnParent              ErrorCode = "OWN_PARENT"
	ErrCodeConflictDateInvalid    ErrorCode = "CONFLICT_DATE_INVALID"
	ErrCodeInvalidLatitude        ErrorCode = "INVALID_LATITUDE"
	ErrCodeInvalidLongitude       ErrorCode = "INVALID_LONGITUDE"
	ErrCodeAwardNameRequired      ErrorCode = "AWARD_NAME_REQUIRED"
	ErrCodeConflictNameRequired   ErrorCode = "CONFLICT_NAME_REQUIRED"
	ErrCodeLocationNameRequired   ErrorCode = "LOCATION_NAME_REQUIRED"

	// API Keys
	ErrCodeAPIKeyAlreadyRevoked ErrorCode = "API_KEY_ALREADY_REVOKED"
	ErrCodeAPIKeyNameRequired   ErrorCode = "API_KEY_NAME_REQUIRED"

	// Media
	ErrCodeContentTypeNotAllowed ErrorCode = "CONTENT_TYPE_NOT_ALLOWED"
	ErrCodeUnsupportedUploadType ErrorCode = "UNSUPPORTED_UPLOAD_TYPE"

	// Submission
	ErrCodeSubmitterNameRequired     ErrorCode = "SUBMITTER_NAME_REQUIRED"
	ErrCodeInvalidSubmitterEmail     ErrorCode = "INVALID_SUBMITTER_EMAIL"
	ErrCodeInvalidPayloadJSON        ErrorCode = "INVALID_PAYLOAD_JSON"
	ErrCodeSubmissionAlreadyReviewed ErrorCode = "SUBMISSION_ALREADY_REVIEWED"
	ErrCodeSubmissionDuplicate       ErrorCode = "SUBMISSION_DUPLICATE"

	ErrCodeCyclicReference ErrorCode = "CYCLIC_REFERENCE"

	// --- LLM ---
	// ErrCodeLLMUnavailable — LLM-провайдер недоступен.
	ErrCodeLLMUnavailable ErrorCode = "LLM_UNAVAILABLE"
	// ErrCodeLLMTimeout — таймаут ответа от LLM.
	ErrCodeLLMTimeout ErrorCode = "LLM_TIMEOUT"
	// ErrCodeLLMParseFailed — не удалось распарсить ответ LLM.
	ErrCodeLLMParseFailed ErrorCode = "LLM_PARSE_FAILED"
	// ErrCodeLLMDisabled — LLM-пайплайн отключён в конфигурации.
	ErrCodeLLMDisabled      ErrorCode = "LLM_DISABLED"
	ErrCodeLLMInputTooLarge ErrorCode = "LLM_INPUT_TOO_LARGE"
)

// AppError — ошибка с встроенным кодом.
type AppError struct {
	Code    ErrorCode
	Message string
}

func (e *AppError) Error() string { return e.Message }

// CodeFromError возвращает доменный код ошибки для логирования и метрик.
// Использует errors.Is для устойчивости к обёрткам fmt.Errorf.
func CodeFromError(err error) ErrorCode {
	// 1. Проверяем, несёт ли ошибка код сама (NewAppError)
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	// 2. Fallback: маппинг старых sentinel через errors.Is
	switch {
	// --- Существующие ---
	case errors.Is(err, ErrNotFound):
		return ErrCodeNotFound
	case errors.Is(err, ErrDeathBeforeBirth):
		return ErrCodeDeathBeforeBirth
	case errors.Is(err, ErrServiceBeforeBirth):
		return ErrCodeServiceBeforeBirth
	case errors.Is(err, ErrServiceAfterDeath):
		return ErrCodeServiceAfterDeath
	case errors.Is(err, ErrRequiredFieldEmpty):
		return ErrCodeRequiredFieldEmpty
	case errors.Is(err, ErrStatusUnspecified):
		return ErrCodeStatusUnspecified
	case errors.Is(err, ErrStatusRequired):
		return ErrCodeStatusRequired
	case errors.Is(err, ErrDuplicateFieldMask):
		return ErrCodeDuplicateFieldMask
	case errors.Is(err, ErrPhotosNotBelongToHero):
		return ErrCodePhotosNotBelongToHero
	case errors.Is(err, ErrSelfRelation):
		return ErrCodeSelfRelation
	case errors.Is(err, ErrInvalidDateFormat):
		return ErrCodeInvalidDateFormat
	case errors.Is(err, ErrInvalidDatePrecision):
		return ErrCodeInvalidDatePrecision
	case errors.Is(err, ErrExactDateRequiresAnchor):
		return ErrCodeExactDateRequiresAnchor
	case errors.Is(err, ErrUnknownDateMustNotHaveAnchor):
		return ErrCodeUnknownDateMustNotHaveAnchor
	case errors.Is(err, ErrDayMonthMustNotHaveAnchor):
		return ErrCodeDayMonthMustNotHaveAnchor
	case errors.Is(err, ErrHeroIDRequired):
		return ErrCodeHeroIDRequired
	case errors.Is(err, ErrPhotoIDsEmpty):
		return ErrCodePhotoIDsEmpty
	case errors.Is(err, ErrPhotosEmpty):
		return ErrCodePhotosEmpty
	case errors.Is(err, ErrPhotoURLRequired):
		return ErrCodePhotoURLRequired
	case errors.Is(err, ErrContactRateLimited):
		return ErrCodeContactRateLimited
	case errors.Is(err, ErrContactSMTP):
		return ErrCodeContactSMTP

	// --- Справочники ---
	case errors.Is(err, ErrEntityAssignedToHeroes):
		return ErrCodeEntityAssignedToHeroes
	case errors.Is(err, ErrOwnParent):
		return ErrCodeOwnParent
	case errors.Is(err, ErrConflictDateInvalid):
		return ErrCodeConflictDateInvalid
	case errors.Is(err, ErrInvalidLatitude):
		return ErrCodeInvalidLatitude
	case errors.Is(err, ErrInvalidLongitude):
		return ErrCodeInvalidLongitude
	case errors.Is(err, ErrAwardNameRequired):
		return ErrCodeAwardNameRequired
	case errors.Is(err, ErrConflictNameRequired):
		return ErrCodeConflictNameRequired
	case errors.Is(err, ErrLocationNameRequired):
		return ErrCodeLocationNameRequired

	// --- API Keys ---
	case errors.Is(err, ErrAPIKeyAlreadyRevoked):
		return ErrCodeAPIKeyAlreadyRevoked
	case errors.Is(err, ErrAPIKeyNameRequired):
		return ErrCodeAPIKeyNameRequired

	// --- Media ---
	case errors.Is(err, ErrContentTypeNotAllowed):
		return ErrCodeContentTypeNotAllowed
	case errors.Is(err, ErrUnsupportedUploadType):
		return ErrCodeUnsupportedUploadType

	// --- Submission ---
	case errors.Is(err, ErrSubmitterNameRequired):
		return ErrCodeSubmitterNameRequired
	case errors.Is(err, ErrInvalidSubmitterEmail):
		return ErrCodeInvalidSubmitterEmail
	case errors.Is(err, ErrInvalidPayloadJSON):
		return ErrCodeInvalidPayloadJSON
	case errors.Is(err, ErrSubmissionAlreadyReviewed):
		return ErrCodeSubmissionAlreadyReviewed
	case errors.Is(err, ErrCyclicReference):
		return ErrCodeCyclicReference
	case errors.Is(err, ErrSubmissionDuplicate):
		return ErrCodeSubmissionDuplicate

	default:
		return ErrCodeInternal
	}
}

// NewAppError создаёт ошибку с кодом.
func NewAppError(code string, message string) error {
	return &AppError{Code: ErrorCode(code), Message: message}
}
