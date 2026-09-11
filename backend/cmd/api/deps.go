package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	v1 "codeberg.org/Thr0TT1e/emh/backend/internal/delivery/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/llm"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// Dependencies содержит все инициализированные зависимости приложения.
type Dependencies struct {
	// Delivery-серверы
	HeroAdminServer       *v1.HeroAdminServer
	HeroPublicServer      *v1.HeroServer
	AwardServer           *v1.AwardServer
	AwardAdminServer      *v1.AwardAdminServer
	MediaServer           *v1.MediaServer
	ConflictServer        *v1.ConflictServer
	ConflictAdminServer   *v1.ConflictAdminServer
	LocationServer        *v1.LocationServer
	LocationAdminServer   *v1.LocationAdminServer
	SubmissionServer      *v1.SubmissionServer
	SubmissionAdminServer *v1.SubmissionAdminServer
	AuthServer            *v1.AuthServer
	APIKeyAdminServer     *v1.APIKeyAdminServer
	ContactServer         *v1.ContactServer
	ExtractionServer      *v1.ExtractionServer
	LLMAdminServer        *v1.LLMAdminServer

	// Interceptor (создаётся в wireDependencies)
	AuthInterceptor *interceptor.AuthInterceptor
	alertWorker     *usecase.AlertWorker

	// Контактный rate limiter (запускается в main.go)
	ContactRateLimiter *usecase.ContactRateLimiter

	// Фоновые задачи (для остановки при shutdown)
	thumbnailWorker        *usecase.ThumbnailWorker
	authAuditWorker        *usecase.AuthAuditWorker
	authAuditCleanupWorker *usecase.AuthAuditCleanupWorker
	orphanWorker           *usecase.OrphanCleanupWorker
	llmLogCleanupWorker    *usecase.LLMLogCleanupWorker
	submissionEmailWorker  *usecase.SubmissionEmailWorker
}

// wireDependencies создаёт и связывает все зависимости приложения.
func wireDependencies(
	cfg *config.Config,
	logger *slog.Logger,
	dbPool *pgxpool.Pool,
	mediaStorage domain.MediaStorage,
) (*Dependencies, error) {
	deps := &Dependencies{}

	// ─── Репозитории ────
	heroRepo := pg.NewHeroRepository(dbPool)
	awardRepo := pg.NewAwardRepository(dbPool)
	heroAwardRepo := pg.NewHeroAwardRepository(dbPool)
	photoRepo := pg.NewPhotoRepository(dbPool)
	heroConflictRepo := pg.NewHeroConflictRepository(dbPool)
	conflictRepo := pg.NewConflictRepository(dbPool)
	heroLocationRepo := pg.NewHeroLocationRepository(dbPool)
	locationRepo := pg.NewLocationRepository(dbPool)
	submissionRepo := pg.NewSubmissionRepository(dbPool)
	heroSourceRepo := pg.NewHeroSourceRepository(dbPool)
	heroRelationRepo := pg.NewHeroRelationRepository(dbPool)
	refreshTokenRepo := pg.NewRefreshTokenRepository(dbPool)
	apiKeyRepo := pg.NewAPIKeyRepository(dbPool)
	contactRepo := pg.NewContactMessageRepository(dbPool)
	authAuditRepo := pg.NewAuthAuditRepository(dbPool)
	extractionLogRepo := pg.NewLLMExtractionLogRepository(dbPool)
	// Инициализация воркера очистки LLM логов (90 дней хранения, запуск раз в сутки)
	llmLogCleanupWorker := usecase.NewLLMLogCleanupWorker(
		extractionLogRepo,
		logger,
		90*24*time.Hour, // 90 дней
		24*time.Hour,    // интервал проверки
	)

	// ─── UseCases ───
	heroUC := usecase.NewHeroUseCase(heroRepo)
	awardUC := usecase.NewAwardUseCase(awardRepo)
	heroAwardUC := usecase.NewHeroAwardUseCase(heroAwardRepo, heroRepo, awardRepo)
	conflictUC := usecase.NewConflictUseCase(conflictRepo)
	heroConflictUC := usecase.NewHeroConflictUseCase(heroConflictRepo, heroRepo, conflictRepo)
	locationUC := usecase.NewLocationUseCase(locationRepo)
	heroLocationUC := usecase.NewHeroLocationUseCase(heroLocationRepo, heroRepo, locationRepo)
	heroSourceUC := usecase.NewHeroSourceUseCase(heroSourceRepo, heroRepo)
	heroRelationUC := usecase.NewHeroRelationUseCase(heroRelationRepo, heroRepo)
	mediaUC := usecase.NewMediaUseCase(mediaStorage, cfg.S3.PresignExpiry.Duration)
	apiKeyUC := usecase.NewAPIKeyUseCase(apiKeyRepo)

	heroQueryUC := usecase.NewHeroQueryUseCase(
		heroRepo, photoRepo, heroAwardRepo,
		heroConflictRepo, heroLocationRepo,
		heroSourceRepo, heroRelationRepo,
	)

	// Thumbnail worker
	thumbnailWorker := usecase.NewThumbnailWorker(
		photoRepo, mediaStorage, logger, cfg.Thumbnails,
	)
	photoUC := usecase.NewPhotoUseCase(photoRepo, heroRepo, mediaStorage, thumbnailWorker, logger)

	// Orphan cleanup worker
	orphanWorker := usecase.NewOrphanCleanupWorker(
		photoRepo, mediaStorage, logger,
		cfg.OrphanCleanup.Interval.Duration,
		cfg.OrphanCleanup.GracePeriod.Duration,
		cfg.OrphanCleanup.DryRun,
	)

	// Auth
	admins := make([]domain.AdminUser, 0, len(cfg.Auth.Admins))
	for _, a := range cfg.Auth.Admins {
		admins = append(admins, domain.AdminUser{
			Username:     a.Username,
			PasswordHash: a.PasswordHash,
			Role:         a.Role,
		})
	}

	// Contact / SMTP
	contactRateLimiter := usecase.NewContactRateLimiter(usecase.ContactRateLimitConfig{
		HourLimit:       cfg.Contact.RateLimitHour,
		DayLimit:        cfg.Contact.RateLimitDay,
		EntryTTL:        cfg.Contact.EntryTTL.Duration,
		CleanupInterval: cfg.Contact.CleanupInterval.Duration,
	}, logger)

	var contactSMTPSender smtp.Sender
	if cfg.Contact.SMTPEnabled {
		contactSMTPSender = smtp.NewSender(smtp.Config{
			Enabled:    true,
			Host:       cfg.Contact.SMTPHost,
			Port:       cfg.Contact.SMTPPort,
			Username:   cfg.Contact.SMTPUsername,
			Password:   cfg.Contact.SMTPPassword,
			FromName:   cfg.Contact.SMTPFromName,
			FromEmail:  cfg.Contact.SMTPFromEmail,
			To:         cfg.Contact.SMTPTo,
			Timeout:    cfg.Contact.SMTPTimeout.Duration,
			MaxRetries: cfg.Contact.SMTPMaxRetries,
			BaseDelay:  cfg.Contact.SMTPBaseDelay.Duration,
			MaxDelay:   cfg.Contact.SMTPMaxDelay.Duration,
		}, logger)
		logger.Info("contact SMTP sender инициализирован",
			"host", cfg.Contact.SMTPHost,
			"port", cfg.Contact.SMTPPort,
			"max_retries", cfg.Contact.SMTPMaxRetries,
		)
	}

	contactUC := usecase.NewContactUseCase(
		contactRepo, contactRateLimiter, contactSMTPSender,
		usecase.ContactUseCaseConfig{
			SMTPEnabled:  cfg.Contact.SMTPEnabled,
			IPPepper:     cfg.Contact.IPPepper,
			DedupeWindow: cfg.Contact.DedupeWindow.Duration,
		}, logger,
	)

	// ─── Alert senders ───
	var alertSenders []usecase.AlertSender

	// SMTP alerts (используем тот же контакт-конфиг для SMTP)
	if cfg.Contact.SMTPEnabled {
		alertSenders = append(alertSenders, usecase.NewSMTPAlertSender(contactSMTPSender))
		logger.Info("alert SMTP sender enabled")
	}

	// Telegram alerts
	if cfg.Telegram.Enabled {
		alertSenders = append(alertSenders, usecase.NewTelegramAlertSender(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			cfg.Telegram.Timeout.Duration,
		))
		logger.Info("alert Telegram sender enabled",
			"chat_id", cfg.Telegram.ChatID,
		)
	}

	// Multi-sender или nil если ничего не настроено
	var alertSender usecase.AlertSender
	if len(alertSenders) > 0 {
		alertSender = usecase.NewMultiAlertSender(logger, alertSenders...)
	} else {
		logger.Warn("no alert senders configured, alerts will only be logged")
	}

	// Alert worker
	alertWorker := usecase.NewAlertWorker(alertSender, logger, cfg.AuthAlerts.QueueSize)

	// Auth audit & alerts
	authAuditWorker := usecase.NewAuthAuditWorker(authAuditRepo, logger, cfg.AuthAudit.QueueSize)

	// Auth audit cleanup worker
	authAuditCleanupWorker := usecase.NewAuthAuditCleanupWorker(
		authAuditRepo,
		logger,
		cfg.AuthAudit.Retention.Duration,
		cfg.AuthAudit.CleanupInterval.Duration,
	)

	authAlertService := usecase.NewAuthAlertService(
		logger, cfg.AuthAlerts.Threshold, cfg.AuthAlerts.Window.Duration,
		alertWorker,
	)

	// Auth usecase (добавляем alertWorker)
	authUC := usecase.NewAuthUseCase(
		admins,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiry.Duration,
		cfg.Auth.RefreshExpiry.Duration,
		refreshTokenRepo,
		alertWorker,
		logger,
	)

	// Auth interceptor
	authInterceptor := interceptor.NewAuthInterceptor(
		cfg.Auth.JWTSecret,
		cfg.Auth.APIKeys,
		apiKeyUC,
		cfg.Auth.TrustedProxies,
		authAuditWorker,
		authAlertService,
		cfg.Contact.IPPepper,
		logger,
	)

	// Submission email worker
	var submissionEmailWorker *usecase.SubmissionEmailWorker
	if cfg.SubmissionEmails.Enabled && cfg.Contact.SMTPEnabled {
		submissionEmailWorker = usecase.NewSubmissionEmailWorker(
			contactSMTPSender,
			logger,
			cfg.SubmissionEmails,
		)
		logger.Info("submission email worker enabled",
			"queue_size", cfg.SubmissionEmails.QueueSize,
		)
	} else {
		logger.Info("submission email worker disabled")
	}

	// При создании submissionUC:
	submissionUC := usecase.NewSubmissionUseCase(
		submissionRepo,
		heroRepo,
		submissionEmailWorker,
		logger,
	)

	// ─── LLM-пайплайн ────
	var llmProvider domain.LLMProvider
	if cfg.LLM.Enabled {
		provider, err := llm.NewLLMProvider(cfg.LLM)
		if err != nil {
			return nil, err
		}
		llmProvider = provider
		logger.Info("LLM-провайдер инициализирован",
			"provider", llmProvider.Name(),
			"model", llmProvider.Model(),
		)
	} else {
		logger.Info("LLM отключён")
	}

	extractionUC := usecase.NewExtractionUseCase(
		llmProvider,
		extractionLogRepo,
		heroRepo,
		cfg.LLM.MaxInputChars,
		logger,
	)

	// ─── Delivery-серверы ────
	deps.HeroAdminServer = v1.NewHeroAdminServer(
		heroUC, heroQueryUC, heroAwardUC, heroConflictUC,
		heroLocationUC, heroSourceUC, heroRelationUC, photoUC, logger,
	)
	deps.HeroPublicServer = v1.NewHeroServer(heroUC, heroQueryUC, heroSourceUC, heroRelationUC, logger)
	deps.AwardServer = v1.NewAwardServer(awardUC, logger)
	deps.AwardAdminServer = v1.NewAwardAdminServer(awardUC, logger)
	deps.MediaServer = v1.NewMediaServer(mediaUC, logger)
	deps.ConflictServer = v1.NewConflictServer(conflictUC, logger)
	deps.ConflictAdminServer = v1.NewConflictAdminServer(conflictUC, logger)
	deps.LocationServer = v1.NewLocationServer(locationUC, logger)
	deps.LocationAdminServer = v1.NewLocationAdminServer(locationUC, logger)
	deps.SubmissionServer = v1.NewSubmissionServer(submissionUC, logger)
	deps.SubmissionAdminServer = v1.NewSubmissionAdminServer(submissionUC, logger)
	deps.AuthServer = v1.NewAuthServer(authUC, logger)
	deps.APIKeyAdminServer = v1.NewAPIKeyAdminServer(apiKeyUC, logger)
	deps.ContactServer = v1.NewContactServer(contactUC, cfg.Auth.TrustedProxies, logger)
	deps.ExtractionServer = v1.NewExtractionServer(extractionUC, logger)

	// LLM Admin
	providerRepo := pg.NewLLMProviderRepository(dbPool)
	providerManager, err := llm.NewProviderManager(cfg.LLM, providerRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("create provider manager: %w", err)
	}

	llmAdminUC := usecase.NewLLMAdminUseCase(providerRepo, providerManager, logger)
	deps.LLMAdminServer = v1.NewLLMAdminServer(llmAdminUC, logger)

	// Интерцептор и фоновые задачи
	deps.AuthInterceptor = authInterceptor
	deps.ContactRateLimiter = contactRateLimiter
	deps.thumbnailWorker = thumbnailWorker
	deps.authAuditWorker = authAuditWorker
	deps.authAuditCleanupWorker = authAuditCleanupWorker
	deps.alertWorker = alertWorker
	deps.orphanWorker = orphanWorker
	deps.llmLogCleanupWorker = llmLogCleanupWorker
	deps.submissionEmailWorker = submissionEmailWorker

	return deps, nil
}
