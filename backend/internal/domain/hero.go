package domain

import "time"

// Hero представляет полную карточку героя в бизнес-логике.
type Hero struct {
	ID               string
	FirstName        string
	LastName         string
	MiddleName       string
	ShortBio         string
	FullBio          string
	Rank             string
	BirthDate        FlexibleDate
	DeathDate        FlexibleDate
	ServiceStartDate FlexibleDate
	Status           PublicationStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Nickname         string
	Unit             string
	Position         string
	ServiceBranch    string
	CauseOfDeath     string
	Memberships      []string

	// Денормализация для списков (заполняется в List репозитория)
	MainPhotoURL     *string
	MainThumbnailURL *string
	AwardNames       []string
}

// PublicationStatus дублирует enum из proto для изоляции домена.
type PublicationStatus int

const (
	StatusUnspecified PublicationStatus = iota
	StatusDraft
	StatusPublished
	StatusArchived
)

// HeroDetail — полный агрегат героя со всеми связями (read-модель).
type HeroDetail struct {
	Hero      *Hero
	Photos    []*Photo
	Awards    []*HeroAward
	Conflicts []*HeroConflict
	Locations []*HeroLocation
	Sources   []*HeroSource
	Relations []*HeroRelation
}

// CreateHeroParams параметры создания героя (входные данные для UC).
type CreateHeroParams struct {
	FirstName        string
	LastName         string
	MiddleName       string
	ShortBio         string
	FullBio          string
	Rank             string
	BirthDate        FlexibleDate
	DeathDate        FlexibleDate
	Status           PublicationStatus
	Nickname         string
	Unit             string
	Position         string
	ServiceBranch    string
	CauseOfDeath     string
	ServiceStartDate FlexibleDate
	Memberships      []string
}

// UpdateHeroParams параметры обновления. FieldMask определяет, что менять.
type UpdateHeroParams struct {
	ID               *string
	FirstName        *string
	LastName         *string
	MiddleName       *string
	ShortBio         *string
	FullBio          *string
	Rank             *string
	BirthDate        *FlexibleDate
	DeathDate        *FlexibleDate
	Status           *PublicationStatus
	FieldMask        []string
	Nickname         *string
	Unit             *string
	Position         *string
	ServiceBranch    *string
	CauseOfDeath     *string
	ServiceStartDate *FlexibleDate
	Memberships      []string
}

// HeroFilter параметры поиска и фильтрации.
type HeroFilter struct {
	SearchQuery     string
	SearchWords     []string
	ConflictID      string
	LocationID      string
	DateFrom        *string // YYYY-MM-DD
	DateTo          *string
	Cursor          string
	Limit           int
	Status          *PublicationStatus // Явный фильтр по конкретному статусу
	IncludeArchived bool               // Если true (и Status=nil), включает архивные записи в общий список
}

// MainPhotoURL возвращает URL главного фото или пустую строку.
func (d *HeroDetail) MainPhotoURL() string {
	for _, p := range d.Photos {
		if p.IsMain {
			return p.URL
		}
	}
	if len(d.Photos) > 0 {
		return d.Photos[0].URL
	}
	return ""
}

// AwardNames возвращает названия наград для бейджей в HeroSummary.
func (d *HeroDetail) AwardNames() []string {
	names := make([]string, 0, len(d.Awards))
	for _, a := range d.Awards {
		names = append(names, a.AwardName)
	}
	return names
}
