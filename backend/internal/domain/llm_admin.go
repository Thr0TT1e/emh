package domain

import "time"

// LLMProviderInfo информация о провайдере для админки.
type LLMProviderInfo struct {
	ID        string
	Name      string
	Type      string
	Model     string
	IsActive  bool
	Priority  int
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LLMTestResult результат тестирования провайдера.
type LLMTestResult struct {
	Success          bool
	RawResponse      string
	ProcessingTimeMs int64
	ErrorMessage     string
}
