package escape

import "strings"

// LIKE экранирует специальные символы PostgreSQL LIKE/ILIKE.
// Символы %, _, \ заменяются на \%, \_, \\ соответственно.
// Используется с ESCAPE '\' в SQL-запросах.
func LIKE(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return r.Replace(s)
}
