package domain

import "time"

// Conflict представляет военный конфликт.
type Conflict struct {
	ID               string
	Name             string
	Description      string
	Type             ConflictType
	StartDate        *time.Time
	EndDate          *time.Time
	ParentConflictID *string
	StartDateInfo    *FlexibleDate
	EndDateInfo      *FlexibleDate
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ConflictFilter параметры фильтрации и пагинации списка конфликтов.
type ConflictFilter struct {
	Type     ConflictType
	ParentID string
	Cursor   string
	Limit    int
}

// ConflictType тип конфликта.
type ConflictType int

const (
	ConflictTypeUnspecified ConflictType = iota
	ConflictTypeGlobal
	ConflictTypeLocal
	ConflictTypePeacekeeping
	ConflictTypeCounterTerrorism
	ConflictTypeSpecialOperation
)

// HeroConflict представляет участие героя в конфликте.
type HeroConflict struct {
	HeroID           string
	ConflictID       string
	ConflictName     string // Денормализация
	SpecificLocation string
	RankAtConflict   string
}

// CreateConflictParams параметры создания конфликта.
type CreateConflictParams struct {
	Name             string
	Description      string
	Type             ConflictType
	StartDate        *time.Time
	EndDate          *time.Time
	ParentConflictID *string
	StartDateInfo    *FlexibleDate
	EndDateInfo      *FlexibleDate
}

// UpdateConflictParams параметры обновления конфликта.
type UpdateConflictParams struct {
	ID               string
	Name             string
	Description      string
	Type             ConflictType
	StartDate        *time.Time
	EndDate          *time.Time
	ParentConflictID *string
	FieldMask        []string
	StartDateInfo    *FlexibleDate
	EndDateInfo      *FlexibleDate
}
