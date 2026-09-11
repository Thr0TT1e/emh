package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/storage/s3"
)

// connectDatabase создаёт и проверяет пул соединений PostgreSQL.
func connectDatabase(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MinConns = cfg.DB.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info("БД подключена", "max_conns", cfg.DB.MaxConns)
	return pool, nil
}

// connectStorage создаёт и инициализирует S3/MinIO хранилище.
// Возвращает интерфейс domain.MediaStorage, а не конкретный тип.
func connectStorage(ctx context.Context, cfg *config.Config, logger *slog.Logger) (domain.MediaStorage, error) {
	storage, err := s3.NewMediaStorage(s3.Config{
		Endpoint:       cfg.S3.Endpoint,
		PublicEndpoint: cfg.S3.PublicEndpoint,
		PublicBaseURL:  cfg.S3.PublicBaseURL,
		AccessKey:      cfg.S3.AccessKey,
		SecretKey:      cfg.S3.SecretKey,
		Bucket:         cfg.S3.Bucket,
		UseSSL:         cfg.S3.UseSSL,
		CACertPath:     cfg.S3.CACertPath,
	})
	if err != nil {
		return nil, err
	}
	if err := storage.EnsureBucket(ctx); err != nil {
		return nil, err
	}

	logger.Info("Хранилище медиа инициализировано", "bucket", cfg.S3.Bucket)
	return storage, nil
}

// newLogger создаёт структурированный JSON-логгер.
func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
