package domain

import "time"

// LLMProviderRecord запись в таблице управления провайдерами.
type LLMProviderRecord struct {
	ID        string
	Name      string // Совпадает с name из конфига
	IsActive  bool
	Priority  int
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
