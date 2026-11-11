package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg"
	"codeberg.org/Thr0TT1e/emh/backend/internal/seeder"
	"codeberg.org/Thr0TT1e/emh/backend/pkg/env"
)

// isSeedAwardsCommand — subcommand-диспетчер: emh seed-awards <seed.json>.
func isSeedAwardsCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "seed-awards"
}

// seedPath возвращает путь к seed.json: env AWARDS_SEED_PATH или дефолт из ADR-005.
func seedPath() string {
	return env.Get("AWARDS_SEED_PATH", "/app/seeds/seed.json")
}

// runSeedAwards — one-shot импорт сида (инкремент на VPS):
//
//	podman run --rm ... emh-backend seed-awards [/app/seeds/seed.json]
//
// Выход с кодом 1, если хотя бы одна запись не импортирована или импорт
// упал фатально (битый артефакт/дубликат в файле) — one-shot-контейнер
// должен падать заметно.
func runSeedAwards(ctx context.Context, cfg *config.Config, logger *slog.Logger,
	dbPool *pgxpool.Pool, storage domain.MediaStorage,
) {
	path := seedPath()
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	repo := pg.NewAwardRepository(dbPool)
	result, err := seeder.SeedAwards(ctx, path, repo, storage, logger)
	if err != nil {
		logger.Error("Сид наград заверчился ошибкой", "path", path, "error", err)
		os.Exit(1)
	}
	logger.Info("Сид наград завершён",
		"path", path,
		"inserted", result.Inserted,
		"skipped", result.Skipped,
		"failed", result.Failed,
	)
	if result.Failed > 0 {
		os.Exit(1)
	}
}

// bootstrapAwardsSeed — авто-сид при первом запуске: справочник пуст и
// артефакт на месте. Ошибки никогда не блокируют старт сервера (ADR-005):
// API важнее наполненного каталога, админ может досидеть вручную.
// Синхронно до ListenAndServe (Q19): каталог готов к первому запросу;
// при возобновлённом после падения сиде идемпотентность даёт ON CONFLICT.
func bootstrapAwardsSeed(ctx context.Context, logger *slog.Logger,
	dbPool *pgxpool.Pool, storage domain.MediaStorage,
) {
	repo := pg.NewAwardRepository(dbPool)

	has, err := seeder.HasAnyAwards(ctx, repo)
	if err != nil {
		logger.Error("Авто-сид наград: не удалось проверить справочник, пропуск", "error", err)
		return
	}
	if has {
		return // справочник уже наполнен
	}

	path := seedPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Warn("Авто-сид наград: файл сида не найден, пропуск", "path", path)
		return
	}

	logger.Info("Авто-сид наград: справочник пуст, запускаем импорт", "path", path)
	result, err := seeder.SeedAwards(ctx, path, repo, storage, logger)
	if err != nil {
		logger.Error("Авто-сид наград не удался (сервер стартует без него)", "path", path, "error", err)
		return
	}
	logger.Info("Авто-сид наград завершён",
		"inserted", result.Inserted,
		"skipped", result.Skipped,
		"failed", result.Failed,
	)
	// Failed > 0 здесь не фатален: лог + наблюдаемость, старт не блокируем.
}
