package domain

import "time"

// AuthAuditResult результат аутентификации.
type AuthAuditResult string

const (
	AuthResultSuccess AuthAuditResult = "success"
	AuthResultFailure AuthAuditResult = "failure"
)

// AuthAuditMechanism механизм аутентификации.
type AuthAuditMechanism string

const (
	AuthMechanismStaticKey AuthAuditMechanism = "static_key"
	AuthMechanismAPIKey    AuthAuditMechanism = "api_key"
	AuthMechanismJWT       AuthAuditMechanism = "jwt"
	AuthMechanismNone      AuthAuditMechanism = "none"
)

// AuthAuditEntry запись аудита аутентификации.
type AuthAuditEntry struct {
	IPAddress     string
	Mechanism     AuthAuditMechanism
	Result        AuthAuditResult
	Identity      string
	Procedure     string
	FailureReason string
	UserAgent     string
	Timestamp     time.Time
}
