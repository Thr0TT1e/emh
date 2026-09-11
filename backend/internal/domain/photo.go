package domain

import "time"

// FaceBox область интереса на фотографии (обычно лицо героя).
type FaceBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Photo представляет фотографию героя.
type Photo struct {
	ID           string
	HeroID       string
	URL          string
	ThumbnailURL string
	Description  string
	SortOrder    int
	IsMain       bool
	FaceBox      *FaceBox
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AddPhotoParams параметры добавления фото.
type AddPhotoParams struct {
	HeroID       string
	URL          string
	ThumbnailURL string
	Description  string
	SortOrder    int
	IsMain       bool
	FaceBox      *FaceBox
}

// UpdatePhotoParams параметры обновления метаданных фото.
type UpdatePhotoParams struct {
	PhotoID     string
	HeroID      string
	Description string
	FaceBox     *FaceBox
	FieldMask   []string
}
