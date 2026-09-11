package domain

import "time"

// HeroSource источник данных о герое (ссылка на статью, архив, книгу).
type HeroSource struct {
	ID         string
	HeroID     string
	URL        string
	Title      string
	SourceType string
	Excerpt    string
	CreatedAt  time.Time
}

// AddHeroSourceParams параметры добавления источника.
type AddHeroSourceParams struct {
	HeroID     string
	URL        string
	Title      string
	SourceType string
	Excerpt    string
}
