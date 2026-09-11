package domain

import "time"

// ExtractionResult нормализованный результат извлечения данных из текста.
type ExtractionResult struct {
	Hero             *ExtractedHero
	Conflicts        []ExtractedConflict
	Awards           []ExtractedAward
	Locations        []ExtractedLocation
	SourceURLs       []string
	Warnings         []string
	DuplicateHeroIDs []string
	RawJSON          string // Сырой JSON от LLM для аудита
}

// ExtractedHero извлечённые данные героя.
type ExtractedHero struct {
	LastName         string   `json:"last_name"`
	FirstName        string   `json:"first_name"`
	MiddleName       string   `json:"middle_name,omitempty"`
	Nickname         string   `json:"nickname,omitempty"`
	Rank             string   `json:"rank,omitempty"`
	Unit             string   `json:"unit,omitempty"`
	Position         string   `json:"position,omitempty"`
	ServiceBranch    string   `json:"service_branch,omitempty"`
	BirthDate        string   `json:"birth_date,omitempty"`
	DeathDate        string   `json:"death_date,omitempty"`
	CauseOfDeath     string   `json:"cause_of_death,omitempty"`
	ServiceStartDate string   `json:"service_start_date,omitempty"`
	ShortBio         string   `json:"short_bio,omitempty"`
	FullBio          string   `json:"full_bio,omitempty"`
	Memberships      []string `json:"memberships,omitempty"`
}

// ExtractedConflict извлечённый конфликт.
type ExtractedConflict struct {
	Name                  string `json:"name"`
	SpecificLocation      string `json:"specific_location,omitempty"`
	RankAtConflict        string `json:"rank_at_conflict,omitempty"`
	SuggestedConflictType string `json:"suggested_conflict_type,omitempty"`
}

// ExtractedAward извлечённая награда.
type ExtractedAward struct {
	Name         string `json:"name"`
	AwardDate    string `json:"award_date,omitempty"`
	DecreeNumber string `json:"decree_number,omitempty"`
}

// ExtractedLocation извлечённая локация.
type ExtractedLocation struct {
	Name             string `json:"name"`
	HistoricalName   string `json:"historical_name,omitempty"`
	LocationType     string `json:"location_type,omitempty"`
	HeroLocationType string `json:"hero_location_type,omitempty"`
}

// LLMExtractionLog запись аудита LLM-экстракции.
type LLMExtractionLog struct {
	ID               string
	CreatedAt        time.Time
	Provider         string
	Model            string
	Prompt           string
	RawResponse      string
	ParsedResult     *ExtractionResult
	ProcessingTimeMs int64
	Status           string // "success", "error", "timeout"
	ErrorMessage     string
	SubmissionID     *string
}
