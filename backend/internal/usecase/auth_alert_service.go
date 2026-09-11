package usecase

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// attemptWindow отслеживает неудачные попытки аутентификации по IP.
type attemptWindow struct {
	count       int
	lastAttempt time.Time
	alerted     bool
}

// AuthAlertService отслеживает неудачные попытки аутентификации
// и отправляет асинхронный алерт при превышении порога.
type AuthAlertService struct {
	mu          sync.Mutex
	attempts    map[string]*attemptWindow
	threshold   int
	window      time.Duration
	logger      *slog.Logger
	alertWorker *AlertWorker
}

// NewAuthAlertService создаёт сервис алертов.
func NewAuthAlertService(
	logger *slog.Logger,
	threshold int,
	window time.Duration,
	alertWorker *AlertWorker,
) *AuthAlertService {
	return &AuthAlertService{
		attempts:    make(map[string]*attemptWindow),
		threshold:   threshold,
		window:      window,
		logger:      logger,
		alertWorker: alertWorker,
	}
}

// RecordFailure фиксирует неудачную попытку аутентификации.
// При превышении порога отправляет асинхронный алерт.
func (s *AuthAlertService) RecordFailure(ip string) {
	if ip == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Ленивая очистка старых записей
	s.cleanupLocked(now)

	w, ok := s.attempts[ip]
	if !ok {
		w = &attemptWindow{}
		s.attempts[ip] = w
	}

	// Если окно прошло, сбрасываем счётчик
	if now.Sub(w.lastAttempt) > s.window {
		w.count = 0
		w.alerted = false
	}

	w.count++
	w.lastAttempt = now

	// Алерт при превышении порога (один раз за окно)
	if w.count >= s.threshold && !w.alerted {
		w.alerted = true

		// Асинхронная отправка алерта
		if s.alertWorker != nil {
			s.alertWorker.Enqueue(domain.Alert{
				Type:      domain.AlertTypeAuthBruteForce,
				Subject:   "Обнаружена попытка подбора пароля",
				Body:      "Количество неудачных попыток аутентификации превысило порог",
				Timestamp: now,
				Metadata: map[string]string{
					"ip":       ip,
					"attempts": fmt.Sprintf("%d", w.count),
					"window":   s.window.String(),
				},
			})
		}

		// Лог остаётся для аудита в stdout
		s.logger.Error("AUTH BRUTE FORCE ALERT",
			"alert_type", "auth_brute_force",
			"ip", ip,
			"attempts", w.count,
			"window", s.window.String(),
		)
	}
}

// cleanupLocked удаляет записи, у которых окно прошло.
// Вызывается под mutex.
func (s *AuthAlertService) cleanupLocked(now time.Time) {
	for ip, w := range s.attempts {
		if now.Sub(w.lastAttempt) > s.window {
			delete(s.attempts, ip)
		}
	}
}
