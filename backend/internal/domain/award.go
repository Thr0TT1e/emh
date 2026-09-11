package domain

import "time"

// Award представляет награду из справочника.
type Award struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HeroAward представляет привязку награды к герою.
type HeroAward struct {
	HeroID       string
	AwardID      string
	AwardName    string // Денормализация из JOIN
	AwardDate    *time.Time
	DecreeNumber string
	// Гибкая дата награждения
	AwardDateInfo *FlexibleDate
}

// CreateAwardParams параметры создания награды.
type CreateAwardParams struct {
	Name        string
	Description string
	ImageURL    string
	SortOrder   int
}

// UpdateAwardParams параметры обновления награды.
type UpdateAwardParams struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	SortOrder   int
	FieldMask   []string
}

// AwardFilter параметры фильтрации и пагинации списка наград.
type AwardFilter struct {
	SearchQuery string
	Cursor      string
	Limit       int
}
