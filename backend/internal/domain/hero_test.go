package domain

import "testing"

// TestPublicationStatus_Values фиксирует числовые значения статусов публикации.
// Они хранятся в БД как SMALLINT, поэтому изменение порядка сломает данные.
func TestPublicationStatus_Values(t *testing.T) {
	tests := []struct {
		status   PublicationStatus
		expected int
	}{
		{StatusUnspecified, 0},
		{StatusDraft, 1},
		{StatusPublished, 2},
		{StatusArchived, 3},
	}
	for _, tt := range tests {
		if int(tt.status) != tt.expected {
			t.Errorf("status %v = %d, want %d", tt.status, int(tt.status), tt.expected)
		}
	}
}

// TestHeroDetail_MainPhotoURL проверяет выбор главного фото для списков.
// Логика: сначала ищем IsMain, иначе первое фото, иначе пустая строка.
func TestHeroDetail_MainPhotoURL(t *testing.T) {
	tests := []struct {
		name     string
		photos   []*Photo
		expected string
	}{
		{"no photos", nil, ""},
		{"empty photos", []*Photo{}, ""},
		{"single main", []*Photo{{URL: "main.jpg", IsMain: true}}, "main.jpg"},
		{"single non-main", []*Photo{{URL: "only.jpg"}}, "only.jpg"},
		{"main not first", []*Photo{{URL: "first.jpg"}, {URL: "main.jpg", IsMain: true}, {URL: "last.jpg"}}, "main.jpg"},
		{"no main uses first", []*Photo{{URL: "first.jpg"}, {URL: "second.jpg"}}, "first.jpg"},
		{"multiple main uses first", []*Photo{{URL: "main1.jpg", IsMain: true}, {URL: "main2.jpg", IsMain: true}}, "main1.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := &HeroDetail{Hero: &Hero{}, Photos: tt.photos}
			if got := detail.MainPhotoURL(); got != tt.expected {
				t.Errorf("MainPhotoURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestHeroDetail_AwardNames проверяет сбор названий наград.
// Логика: возвращает слайс названий в порядке следования, пустой если наград нет.
func TestHeroDetail_AwardNames(t *testing.T) {
	tests := []struct {
		name     string
		awards   []*HeroAward
		expected []string
	}{
		{"no awards", nil, []string{}},
		{"empty awards", []*HeroAward{}, []string{}},
		{"single award", []*HeroAward{{AwardName: "Герой России"}}, []string{"Герой России"}},
		{"multiple awards preserve order", []*HeroAward{
			{AwardName: "Герой России"},
			{AwardName: "Орден Мужества"},
			{AwardName: "Медаль за отвагу"},
		}, []string{"Герой России", "Орден Мужества", "Медаль за отвагу"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := &HeroDetail{Hero: &Hero{}, Awards: tt.awards}
			got := detail.AwardNames()
			if len(got) != len(tt.expected) {
				t.Fatalf("AwardNames() len = %d, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("AwardNames()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
