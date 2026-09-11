package domain

import "errors"

// Sentinel errors для типизированной обработки в delivery-слое.
// Delivery проверяет ошибки через errors.Is(), а не strings.Contains().
var (
	// ErrNotFound возвращается, когда запрашиваемая сущность не найдена.
	ErrNotFound = errors.New("not found")

	// ErrDeathBeforeBirth возвращается, когда дата гибели раньше даты рождения.
	ErrDeathBeforeBirth = errors.New("death date cannot be before birth date")

	// ErrServiceBeforeBirth возвращается, когда дата начала службы раньше даты рождения.
	ErrServiceBeforeBirth = errors.New("service start date cannot be before birth date")

	// ErrServiceAfterDeath возвращается, когда дата начала службы позже даты гибели.
	ErrServiceAfterDeath = errors.New("service start date cannot be after death date")

	// ErrRequiredFieldEmpty возвращается при попытке очистить обязательное поле.
	ErrRequiredFieldEmpty = errors.New("required field cannot be empty")

	// ErrStatusUnspecified возвращается при попытке установить статус UNSPECIFIED.
	ErrStatusUnspecified = errors.New("status cannot be unspecified")

	// ErrStatusRequired возвращается при попытке обновить статус без значения.
	ErrStatusRequired = errors.New("status is required when field_mask contains status")

	// ErrDuplicateFieldMask возвращается при дублировании поля в field_mask.
	ErrDuplicateFieldMask = errors.New("field_mask contains duplicate field")

	// ErrPhotosNotBelongToHero возвращается при попытке операции с чужими фото.
	ErrPhotosNotBelongToHero = errors.New("some photos do not belong to hero")

	// ErrSelfRelation возвращается при попытке создать связь героя с самим собой.
	ErrSelfRelation = errors.New("cannot create relation to itself")

	// ErrInvalidDateFormat возвращается при невалидном формате даты.
	ErrInvalidDateFormat = errors.New("invalid date format")

	// ErrInvalidDatePrecision возвращается при невалидной точности даты.
	ErrInvalidDatePrecision = errors.New("invalid date precision")

	// ErrExactDateRequiresAnchor возвращается, когда точная дата не имеет anchor.
	ErrExactDateRequiresAnchor = errors.New("exact date requires anchor")

	// ErrUnknownDateMustNotHaveAnchor возвращается, когда неизвестная дата имеет anchor.
	ErrUnknownDateMustNotHaveAnchor = errors.New("unknown date must not have anchor")

	// ErrDayMonthMustNotHaveAnchor возвращается, когда day_month дата имеет anchor.
	ErrDayMonthMustNotHaveAnchor = errors.New("day_month date must not have anchor")

	// ErrHeroIDRequired возвращается, когда ID героя не передан.
	ErrHeroIDRequired = errors.New("hero id is required")

	// ErrPhotoIDsEmpty возвращается, когда список ID фото пуст.
	ErrPhotoIDsEmpty = errors.New("photo_ids must not be empty")

	// ErrPhotosEmpty возвращается, когда список фото пуст.
	ErrPhotosEmpty = errors.New("photos must not be empty")

	// ErrPhotoURLRequired возвращается, когда URL фото не передан.
	ErrPhotoURLRequired = errors.New("photo url is required")

	// ErrContactRateLimited возвращается при превышении лимита отправки сообщений
	// обратной связи (по IP или по email).
	ErrContactRateLimited = errors.New("contact message rate limit exceeded")

	// ErrContactSMTP возвращается при ошибке отправки email-уведомления через SMTP.
	// Сообщение при этом сохранено в БД со статусом failed.
	ErrContactSMTP = errors.New("failed to send contact notification email")

	// --- Справочники (Award, Conflict, Location) ---

	// ErrEntityAssignedToHeroes возвращается при попытке удалить справочную запись,
	// привязанную к одному или нескольким героям.
	ErrEntityAssignedToHeroes = errors.New("entity is assigned to heroes")

	// ErrOwnParent возвращается при попытке установить объект собственным родителем.
	ErrOwnParent = errors.New("entity cannot be its own parent")

	// ErrConflictDateInvalid возвращается когда end_date раньше start_date.
	ErrConflictDateInvalid = errors.New("conflict end_date cannot be before start_date")

	// ErrInvalidLatitude возвращается при невалидной широте.
	ErrInvalidLatitude = errors.New("latitude must be between -90 and 90")

	// ErrInvalidLongitude возвращается при невалидной долготе.
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")

	// ErrAwardNameRequired возвращается при отсутствии названия награды.
	ErrAwardNameRequired = errors.New("award name is required")

	// ErrConflictNameRequired возвращается при отсутствии названия конфликта.
	ErrConflictNameRequired = errors.New("conflict name is required")

	// ErrLocationNameRequired возвращается при отсутствии названия локации.
	ErrLocationNameRequired = errors.New("location name is required")

	// --- API Keys ---

	// ErrAPIKeyAlreadyRevoked возвращается при попытке отозвать уже отозванный ключ.
	ErrAPIKeyAlreadyRevoked = errors.New("api key already revoked")

	// ErrAPIKeyNameRequired возвращается при отсутствии имени ключа.
	ErrAPIKeyNameRequired = errors.New("api key name is required")

	// --- Media ---

	// ErrContentTypeNotAllowed возвращается при недопустимом MIME-типе.
	ErrContentTypeNotAllowed = errors.New("content type not allowed for this upload type")

	// ErrUnsupportedUploadType возвращается при неизвестном типе загрузки.
	ErrUnsupportedUploadType = errors.New("unsupported upload type")

	// --- Submission ---

	// ErrSubmitterNameRequired возвращается при отсутствии имени отправителя.
	ErrSubmitterNameRequired = errors.New("submitter name is required")

	// ErrInvalidSubmitterEmail возвращается при невалидном email отправителя.
	ErrInvalidSubmitterEmail = errors.New("invalid submitter email")

	// ErrInvalidPayloadJSON возвращается при невалидном JSON в заявке.
	ErrInvalidPayloadJSON = errors.New("payload_json is not valid JSON")

	// ErrSubmissionAlreadyReviewed возвращается при повторной модерации заявки.
	ErrSubmissionAlreadyReviewed = errors.New("submission already reviewed")

	// ErrSubmissionDuplicate возвращается при попытке создать заявку
	// с идентичным контентом уже существующей активной заявки.
	ErrSubmissionDuplicate = errors.New("submission with identical content already exists")

	// ErrCyclicReference возвращается при попытке установить parent_id,
	// создающий транзитивный цикл в иерархии (A → B → C → A).
	ErrCyclicReference = errors.New("cyclic reference detected in hierarchy")

	ErrLLMUnavailable   = NewAppError("LLM_UNAVAILABLE", "Сервис извлечения данных временно недоступен")
	ErrLLMTimeout       = NewAppError("LLM_TIMEOUT", "Превышено время ожидания ответа от LLM")
	ErrLLMParseFailed   = NewAppError("LLM_PARSE_FAILED", "Не удалось распознать структурированные данные в тексте")
	ErrLLMDisabled      = NewAppError("LLM_DISABLED", "Функция извлечения данных отключена администратором")
	ErrLLMInputTooLarge = NewAppError("LLM_INPUT_TOO_LARGE", "размер входного текста превышает допустимый лимит")
)
