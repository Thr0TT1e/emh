package domain

import "time"

// Location представляет географический объект.
type Location struct {
	ID             string
	Name           string
	HistoricalName string
	Type           LocationType
	ParentID       *string
	Latitude       *float64
	Longitude      *float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// LocationFilter параметры фильтрации и пагинации списка локаций.
type LocationFilter struct {
	ParentID    string
	SearchQuery string
	Type        LocationType
	Cursor      string
	Limit       int
}

// LocationType тип географического объекта.
type LocationType int

const (
	LocationTypeUnspecified LocationType = iota
	LocationTypeCountry
	LocationTypeRegion
	LocationTypeCity
	LocationTypeVillage
	LocationTypeCemetery
)

// HeroLocation представляет привязку героя к локации.
type HeroLocation struct {
	HeroID     string
	LocationID string
	Location   *Location // Полный объект из JOIN
	Type       HeroLocationType
}

// HeroLocationType тип связи героя с локацией.
type HeroLocationType int

const (
	HeroLocationTypeUnspecified HeroLocationType = iota
	HeroLocationTypeBirth
	HeroLocationTypeDeath
	HeroLocationTypeBurial
	HeroLocationTypeResidence
)

// CreateLocationParams параметры создания локации.
type CreateLocationParams struct {
	Name           string
	HistoricalName string
	Type           LocationType
	ParentID       *string
	Latitude       *float64
	Longitude      *float64
}

// UpdateLocationParams параметры обновления локации.
type UpdateLocationParams struct {
	ID             string
	Name           string
	HistoricalName string
	Type           LocationType
	ParentID       *string
	Latitude       *float64
	Longitude      *float64
	FieldMask      []string
}
