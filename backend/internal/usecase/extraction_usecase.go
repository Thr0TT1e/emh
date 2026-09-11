package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// ExtractionUseCase бизнес-логика извлечения данных из текста через LLM.
type ExtractionUseCase interface {
	// ExtractFromText извлекает структурированные данные героя из текста.
	ExtractFromText(ctx context.Context, text string, sourceURL string) (*domain.ExtractionResult, error)
}

type extractionUseCase struct {
	provider      domain.LLMProvider
	logRepo       repository.LLMExtractionLogRepository
	heroRepo      repository.HeroRepository
	maxInputChars int
	logger        *slog.Logger
}

// NewExtractionUseCase создаёт usecase извлечения данных.
// Если provider == nil, все вызовы возвращают ErrLLMDisabled.
func NewExtractionUseCase(
	provider domain.LLMProvider,
	logRepo repository.LLMExtractionLogRepository,
	heroRepo repository.HeroRepository,
	maxInputChars int,
	logger *slog.Logger,
) ExtractionUseCase {
	return &extractionUseCase{
		provider:      provider,
		logRepo:       logRepo,
		heroRepo:      heroRepo,
		maxInputChars: maxInputChars,
		logger:        logger,
	}
}

func (uc *extractionUseCase) ExtractFromText(ctx context.Context, text string, sourceURL string) (*domain.ExtractionResult, error) {
	if uc.provider == nil {
		return nil, domain.ErrLLMDisabled
	}

	// Проверка длины в символах (рунах), а не байтах.
	// Для кириллицы в UTF-8 один символ = 2 байта, но бизнес-правило
	// оперирует символами: 150К символов ≈ 300КБ для русского текста.
	charCount := utf8.RuneCountInString(text)
	if charCount > uc.maxInputChars {
		metrics.LLMExtractionTextTooLongTotal.Inc()
		uc.logger.Warn("LLM input text too large",
			"max_chars", uc.maxInputChars,
			"got_chars", charCount,
		)
		return nil, fmt.Errorf("%w: максимум %d символов, получено %d",
			domain.ErrLLMInputTooLarge, uc.maxInputChars, charCount)
	}

	start := time.Now()
	providerName := uc.provider.Name()
	modelName := uc.provider.Model()
	prompt := buildExtractionPrompt(text)

	// Вызываем LLM
	rawResponse, err := uc.provider.Generate(ctx, prompt)
	if err != nil {
		// метрика ошибки LLM
		status := "error"
		if errors.Is(err, context.DeadlineExceeded) {
			status = "timeout"
			metrics.LLMRequestsTotal.WithLabelValues(providerName, modelName, "timeout").Inc()
		} else {
			metrics.LLMRequestsTotal.WithLabelValues(providerName, modelName, "error").Inc()
		}
		metrics.LLMDurationSeconds.WithLabelValues(providerName, modelName).Observe(time.Since(start).Seconds())

		uc.saveLog(ctx, prompt, "", nil, time.Since(start), status, err.Error(), nil)
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Парсим JSON из ответа
	result, err := parseLLMResponse(rawResponse)
	if err != nil {
		// метрика ошибки парсинга
		metrics.LLMRequestsTotal.WithLabelValues(providerName, modelName, "error").Inc()
		metrics.LLMDurationSeconds.WithLabelValues(providerName, modelName).Observe(time.Since(start).Seconds())
		metrics.LLMParseErrorsTotal.WithLabelValues("invalid_json").Inc()

		uc.saveLog(ctx, prompt, rawResponse, nil, time.Since(start), "error", err.Error(), nil)
		return nil, domain.ErrLLMParseFailed
	}

	// метрика успешного запроса
	metrics.LLMRequestsTotal.WithLabelValues(providerName, modelName, "success").Inc()
	metrics.LLMDurationSeconds.WithLabelValues(providerName, modelName).Observe(time.Since(start).Seconds())

	// Добавляем мета-информацию
	result.RawJSON = rawResponse
	if sourceURL != "" {
		result.SourceURLs = append(result.SourceURLs, sourceURL)
	}

	// Поиск дубликатов среди существующих героев
	uc.findDuplicates(ctx, result)

	// Сохраняем лог
	uc.saveLog(ctx, prompt, rawResponse, result, time.Since(start), "success", "", nil)

	return result, nil
}

// findDuplicates ищет возможных дубликатов среди существующих героев.
// Использует полнотекстовый поиск (tsvector + GIN) по ФИО и позывному.
// Результат — список ID кандидатов для визуальной проверки админом.
func (uc *extractionUseCase) findDuplicates(ctx context.Context, result *domain.ExtractionResult) {
	if result.Hero == nil {
		return
	}

	// Собираем поисковые слова: фамилия, имя, позывной.
	// Полнотекстовый поиск найдёт совпадение независимо от порядка слов
	// и от того, в какой колонке (ФИО или позывной) находится слово.
	var words []string
	if result.Hero.LastName != "" {
		words = append(words, result.Hero.LastName)
	}
	if result.Hero.FirstName != "" {
		words = append(words, result.Hero.FirstName)
	}
	if result.Hero.Nickname != "" {
		words = append(words, result.Hero.Nickname)
	}
	if len(words) == 0 {
		return
	}

	heroes, _, _, err := uc.heroRepo.List(ctx, domain.HeroFilter{
		SearchWords: words,
		Limit:       5,
	})
	if err != nil {
		uc.logger.Warn("не удалось выполнить поиск дубликатов",
			"words", words,
			"error", err,
		)
		return
	}

	for _, h := range heroes {
		result.DuplicateHeroIDs = append(result.DuplicateHeroIDs, h.ID)
	}
}

// saveLog сохраняет запись аудита в БД (best-effort, не блокирует основной поток).
func (uc *extractionUseCase) saveLog(
	ctx context.Context,
	prompt, rawResponse string,
	result *domain.ExtractionResult,
	elapsed time.Duration,
	status, errMsg string,
	submissionID *string,
) {
	log := &domain.LLMExtractionLog{
		Provider:         uc.provider.Name(),
		Model:            uc.provider.Model(),
		Prompt:           prompt,
		RawResponse:      rawResponse,
		ParsedResult:     result,
		ProcessingTimeMs: elapsed.Milliseconds(),
		Status:           status,
		ErrorMessage:     errMsg,
		SubmissionID:     submissionID,
	}

	if _, err := uc.logRepo.Create(ctx, log); err != nil {
		uc.logger.Error("не удалось сохранить лог экстракции",
			"error", err,
			"status", status,
		)
	}
}

// ─────────────────────────────────────────────
// Промпт-инжиниринг
// ─────────────────────────────────────────────

// buildExtractionPrompt формирует полный промпт для извлечения данных.
func buildExtractionPrompt(text string) string {
	return extractionSystemPrompt + "\n\n## Текст для анализа:\n" + text
}

const extractionSystemPrompt = `Ты — эксперт по извлечению структурированных данных из текстов о военных героях России и СССР.
Твоя задача — извлечь информацию о герое из предоставленного текста и вернуть результат СТРОГО в формате JSON.

## Правила:
1. Извлекай ТОЛЬКО информацию, явно указанную в тексте. Не выдумывай и не додумывай данные.
2. Если информация отсутствует, оставь поле пустым ("") или исключи из массива.
3. Даты возвращай в формате: "YYYY-MM-DD", "YYYY-MM" или "YYYY". Если точный формат не определён, верни как в тексте.
4. Для конфликтов определи тип: "ВОВ", "Афганистан", "Чечня", "Сирия", "СВО", "Вьетнам", "Корея", "Другой".
5. Для локаций определи тип связи: "birth", "death", "burial", "residence", "service".
6. Если в тексте несколько героев, извлеки данные только ОДНОГО — наиболее подробно описанного.
7. Ответ должен содержать ТОЛЬКО валидный JSON без markdown-обёрток и пояснений.

## Формат ответа:
{
  "hero": {
    "last_name": "",
    "first_name": "",
    "middle_name": "",
    "nickname": "",
    "rank": "",
    "unit": "",
    "position": "",
    "service_branch": "",
    "birth_date": "",
    "death_date": "",
    "cause_of_death": "",
    "service_start_date": "",
    "short_bio": "",
    "full_bio": "",
    "memberships": []
  },
  "conflicts": [
    {
      "name": "",
      "specific_location": "",
      "rank_at_conflict": "",
      "suggested_conflict_type": ""
    }
  ],
  "awards": [
    {
      "name": "",
      "award_date": "",
      "decree_number": ""
    }
  ],
  "locations": [
    {
      "name": "",
      "historical_name": "",
      "location_type": "",
      "hero_location_type": ""
    }
  ],
  "warnings": []
}

## Пример:
Текст: "Иванов Пётр Сергеевич, рядовой, погиб 15 марта 2000 года в бою за высоту 776 в Чечне. Награждён Орденом Мужества посмертно."

Ответ:
{
  "hero": {
    "last_name": "Иванов",
    "first_name": "Пётр",
    "middle_name": "Сергеевич",
    "nickname": "",
    "rank": "рядовой",
    "unit": "",
    "position": "",
    "service_branch": "",
    "birth_date": "",
    "death_date": "2000-03-15",
    "cause_of_death": "погиб в бою",
    "service_start_date": "",
    "short_bio": "Рядовой Иванов Пётр Сергеевич погиб 15 марта 2000 года в бою за высоту 776 в Чечне. Награждён Орденом Мужества посмертно.",
    "full_bio": "",
    "memberships": []
  },
  "conflicts": [
    {
      "name": "Вторая чеченская война",
      "specific_location": "высота 776",
      "rank_at_conflict": "рядовой",
      "suggested_conflict_type": "Чечня"
    }
  ],
  "awards": [
    {
      "name": "Орден Мужества",
      "award_date": "",
      "decree_number": ""
    }
  ],
  "locations": [
    {
      "name": "Чечня",
      "historical_name": "",
      "location_type": "region",
      "hero_location_type": "death"
    }
  ],
  "warnings": []
}`

// ─────────────────────────────────────────────
// Парсинг ответа LLM
// ─────────────────────────────────────────────

// llmRawResponse структура для парсинга JSON от LLM.
type llmRawResponse struct {
	Hero *struct {
		LastName         string   `json:"last_name"`
		FirstName        string   `json:"first_name"`
		MiddleName       string   `json:"middle_name"`
		Nickname         string   `json:"nickname"`
		Rank             string   `json:"rank"`
		Unit             string   `json:"unit"`
		Position         string   `json:"position"`
		ServiceBranch    string   `json:"service_branch"`
		BirthDate        string   `json:"birth_date"`
		DeathDate        string   `json:"death_date"`
		CauseOfDeath     string   `json:"cause_of_death"`
		ServiceStartDate string   `json:"service_start_date"`
		ShortBio         string   `json:"short_bio"`
		FullBio          string   `json:"full_bio"`
		Memberships      []string `json:"memberships"`
	} `json:"hero"`
	Conflicts []struct {
		Name                  string `json:"name"`
		SpecificLocation      string `json:"specific_location"`
		RankAtConflict        string `json:"rank_at_conflict"`
		SuggestedConflictType string `json:"suggested_conflict_type"`
	} `json:"conflicts"`
	Awards []struct {
		Name         string `json:"name"`
		AwardDate    string `json:"award_date"`
		DecreeNumber string `json:"decree_number"`
	} `json:"awards"`
	Locations []struct {
		Name             string `json:"name"`
		HistoricalName   string `json:"historical_name"`
		LocationType     string `json:"location_type"`
		HeroLocationType string `json:"hero_location_type"`
	} `json:"locations"`
	Warnings []string `json:"warnings"`
}

// parseLLMResponse парсит JSON-ответ от LLM в ExtractionResult.
func parseLLMResponse(raw string) (*domain.ExtractionResult, error) {
	// Извлекаем JSON из возможной markdown-обёртки
	jsonStr := extractJSON(raw)

	var parsed llmRawResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal LLM response: %w", err)
	}

	result := &domain.ExtractionResult{
		Warnings: parsed.Warnings,
	}

	// Маппинг героя
	if parsed.Hero != nil {
		result.Hero = &domain.ExtractedHero{
			LastName:         parsed.Hero.LastName,
			FirstName:        parsed.Hero.FirstName,
			MiddleName:       parsed.Hero.MiddleName,
			Nickname:         parsed.Hero.Nickname,
			Rank:             parsed.Hero.Rank,
			Unit:             parsed.Hero.Unit,
			Position:         parsed.Hero.Position,
			ServiceBranch:    parsed.Hero.ServiceBranch,
			BirthDate:        parsed.Hero.BirthDate,
			DeathDate:        parsed.Hero.DeathDate,
			CauseOfDeath:     parsed.Hero.CauseOfDeath,
			ServiceStartDate: parsed.Hero.ServiceStartDate,
			ShortBio:         parsed.Hero.ShortBio,
			FullBio:          parsed.Hero.FullBio,
			Memberships:      parsed.Hero.Memberships,
		}
	}

	// Маппинг конфликтов
	for _, c := range parsed.Conflicts {
		result.Conflicts = append(result.Conflicts, domain.ExtractedConflict{
			Name:                  c.Name,
			SpecificLocation:      c.SpecificLocation,
			RankAtConflict:        c.RankAtConflict,
			SuggestedConflictType: c.SuggestedConflictType,
		})
	}

	// Маппинг наград
	for _, a := range parsed.Awards {
		result.Awards = append(result.Awards, domain.ExtractedAward{
			Name:         a.Name,
			AwardDate:    a.AwardDate,
			DecreeNumber: a.DecreeNumber,
		})
	}

	// Маппинг локаций
	for _, l := range parsed.Locations {
		result.Locations = append(result.Locations, domain.ExtractedLocation{
			Name:             l.Name,
			HistoricalName:   l.HistoricalName,
			LocationType:     l.LocationType,
			HeroLocationType: l.HeroLocationType,
		})
	}

	return result, nil
}

// extractJSON извлекает JSON из возможной markdown-обёртки.
// Обрабатывает: чистый JSON, ```json ... ```, JSON с окружающим текстом.
func extractJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)

	// Попытка 1: весь ответ — чистый JSON
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return trimmed
	}

	// Попытка 2: markdown-обёртка ```json ... ```
	if idx := strings.Index(trimmed, "```json"); idx != -1 {
		start := idx + len("```json")
		end := strings.Index(trimmed[start:], "```")
		if end != -1 {
			return strings.TrimSpace(trimmed[start : start+end])
		}
	}

	// Попытка 3: ищем первую { и последнюю }
	first := strings.Index(trimmed, "{")
	last := strings.LastIndex(trimmed, "}")
	if first != -1 && last != -1 && last > first {
		return trimmed[first : last+1]
	}

	// Fallback: возвращаем как есть (парсер вернёт ошибку)
	return trimmed
}
