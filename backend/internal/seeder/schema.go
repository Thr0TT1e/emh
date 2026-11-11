// Package seeder реализует детерминированный импорт справочника наград
// из артефакта seed.json (ADR-005, секция «Сид справочника наград»).
// Используется двумя входами: авто-сид при старте бэкенда (bootstrap)
// и one-shot subcommand `seed-awards`.
package seeder

// SeedFile — десериализованный seed.json.
type SeedFile struct {
	Awards []SeedAward `json:"awards"`
}

// SeedAward — одна запись справочника в seed.json.
// Порядок полей фиксирует побайтовую детерминированность сборки (cmd/seed-builder).
type SeedAward struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Type           string     `json:"type"`
	Jurisdiction   string     `json:"jurisdiction"`
	IsJubilee      bool       `json:"is_jubilee"`
	WornWithoutBar bool       `json:"worn_without_bar"`
	SortOrder      int        `json:"sort_order"`
	Image          *SeedImage `json:"image,omitempty"`
	Ribbon         *SeedImage `json:"ribbon,omitempty"`
}

// SeedImage — встроенное изображение: content_type + base64 без префикса data:.
// Строгий контракт с сеердом: картинки подготовлены заранее
// (знак 400×400, лента 144×48, WebP) — сид загружает их «как есть», мимо воркера.
type SeedImage struct {
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64.StdEncoding
}
