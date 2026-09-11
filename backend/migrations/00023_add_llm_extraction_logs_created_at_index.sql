-- +goose Up
-- +goose StatementBegin
-- Индекс для ускорения очистки старых логов LLM-экстракций (M5, security.md).
-- Воркер LlmLogsCleanupWorker удаляет записи старше retention period порциями,
-- этот индекс превращает DELETE из seq scan в index scan + delete.
CREATE INDEX IF NOT EXISTS idx_llm_extraction_logs_created_at
    ON llm_extraction_logs (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_llm_extraction_logs_created_at;
-- +goose StatementEnd
