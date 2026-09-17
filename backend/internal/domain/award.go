package domain

import "time"

// AwardType тип награды. Значения совпадают с proto enum AwardType.
type AwardType int

const (
	AwardTypeUnspecified AwardType = iota
	AwardTypeOrder
	AwardTypeMedal
	AwardTypeBadge
)

// AwardJurisdiction государственная принадлежность награды. Значения совпадают с proto enum AwardJurisdiction.
type AwardJurisdiction int

const (
	AwardJurisdictionUnspecified AwardJurisdiction = iota
	AwardJurisdictionRussianFederation
	AwardJurisdictionUSSR
	AwardJurisdictionDepartmental
)

// Award представляет награду из справочника.
type Award struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	SortOrder   int
	// RibbonImageURL - изображение ленты для орденской планки. Пусто, если лента не загружена.
	RibbonImageURL string
	Type           AwardType
	Jurisdiction   AwardJurisdiction
	WornWithoutBar bool
	IsJubilee      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	// Денормализация из справочника наград для орденской планки
	RibbonImageURL string
	ImageURL       string
	Type           AwardType
	Jurisdiction   AwardJurisdiction
	WornWithoutBar bool
	IsJubilee      bool
}

// CreateAwardParams параметры создания награды.
type CreateAwardParams struct {
	Name           string
	Description    string
	ImageURL       string
	SortOrder      int
	RibbonImageURL string
	Type           AwardType
	Jurisdiction   AwardJurisdiction
	WornWithoutBar bool
	IsJubilee      bool
}

// UpdateAwardParams параметры обновления награды.
type UpdateAwardParams struct {
	ID             string
	Name           string
	Description    string
	ImageURL       string
	SortOrder      int
	RibbonImageURL string
	Type           AwardType
	Jurisdiction   AwardJurisdiction
	WornWithoutBar bool
	IsJubilee      bool
	FieldMask      []string
}

// AwardFilter параметры фильтрации и пагинации списка наград.
type AwardFilter struct {
	SearchQuery string
	Cursor      string
	Limit       int
}
