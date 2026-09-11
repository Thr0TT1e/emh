package domain

// HeroRelation связь между двумя героями (отец-сын, сослуживцы и т.д.).
type HeroRelation struct {
	ID              string
	FromHeroID      string
	ToHeroID        string
	RelationType    string
	Description     string
	RelatedHeroName string // Денормализация для отображения
}

// AddHeroRelationParams параметры создания связи.
type AddHeroRelationParams struct {
	FromHeroID   string
	ToHeroID     string
	RelationType string
	Description  string
}
