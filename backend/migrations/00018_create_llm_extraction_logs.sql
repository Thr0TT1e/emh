-- +goose Up
CREATE TABLE llm_extraction_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    prompt TEXT NOT NULL,
    raw_response TEXT NOT NULL,
    parsed_result JSONB,
    processing_time_ms INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'error', 'timeout')),
    error_message TEXT,
    submission_id UUID REFERENCES submissions(id) ON DELETE SET NULL
);

CREATE INDEX idx_llm_extraction_logs_created_at ON llm_extraction_logs(created_at DESC);
CREATE INDEX idx_llm_extraction_logs_submission_id ON llm_extraction_logs(submission_id) WHERE submission_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS llm_extraction_logs;
