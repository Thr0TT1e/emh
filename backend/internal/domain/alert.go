package domain

import "time"

// AlertType тип алерта безопасности.
type AlertType string

const (
	AlertTypeAuthBruteForce    AlertType = "auth_brute_force"
	AlertTypeRefreshTokenTheft AlertType = "refresh_token_theft"
)

// Alert доменная структура алерта безопасности.
type Alert struct {
	Type      AlertType
	Subject   string
	Body      string
	Timestamp time.Time
	Metadata  map[string]string
}
