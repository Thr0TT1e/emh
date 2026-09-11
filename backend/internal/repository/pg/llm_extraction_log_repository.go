package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type llmExtractionLogRepository struct {
	pool *pgxpool.Pool
}

func NewLLMExtractionLogRepository(pool *pgxpool.Pool) repository.LLMExtractionLogRepository {
	return &llmExtractionLogRepository{pool: pool}
}

func (r *llmExtractionLogRepository) Create(ctx context.Context, log *domain.LLMExtractionLog) (string, error) {
	query := `
		INSERT INTO llm_extraction_logs (
			provider, model, prompt, raw_response, parsed_result,
			processing_time_ms, status, error_message, submission_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	var id string
	var createdAt time.Time

	parsedResultJSON, err := json.Marshal(log.ParsedResult)
	if err != nil {
		return "", fmt.Errorf("marshal parsed result: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		log.Provider,
		log.Model,
		log.Prompt,
		log.RawResponse,
		parsedResultJSON,
		log.ProcessingTimeMs,
		log.Status,
		log.ErrorMessage,
		log.SubmissionID,
	).Scan(&id, &createdAt)

	if err != nil {
		return "", fmt.Errorf("insert log: %w", err)
	}

	log.ID = id
	log.CreatedAt = createdAt
	return id, nil
}

func (r *llmExtractionLogRepository) GetByID(ctx context.Context, id string) (*domain.LLMExtractionLog, error) {
	query := `
		SELECT id, created_at, provider, model, prompt, raw_response,
		       parsed_result, processing_time_ms, status, error_message, submission_id
		FROM llm_extraction_logs
		WHERE id = $1
	`

	var log domain.LLMExtractionLog
	var parsedResultJSON []byte
	var submissionID *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.CreatedAt,
		&log.Provider,
		&log.Model,
		&log.Prompt,
		&log.RawResponse,
		&parsedResultJSON,
		&log.ProcessingTimeMs,
		&log.Status,
		&log.ErrorMessage,
		&submissionID,
	)

	if err != nil {
		return nil, fmt.Errorf("select log: %w", err)
	}

	if len(parsedResultJSON) > 0 {
		if err := json.Unmarshal(parsedResultJSON, &log.ParsedResult); err != nil {
			return nil, fmt.Errorf("unmarshal parsed result: %w", err)
		}
	}

	log.SubmissionID = submissionID
	return &log, nil
}

func (r *llmExtractionLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.LLMExtractionLog, error) {
	query := `
		SELECT id, created_at, provider, model, prompt, raw_response,
		       parsed_result, processing_time_ms, status, error_message, submission_id
		FROM llm_extraction_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.LLMExtractionLog
	for rows.Next() {
		var log domain.LLMExtractionLog
		var parsedResultJSON []byte
		var submissionID *string

		err := rows.Scan(
			&log.ID,
			&log.CreatedAt,
			&log.Provider,
			&log.Model,
			&log.Prompt,
			&log.RawResponse,
			&parsedResultJSON,
			&log.ProcessingTimeMs,
			&log.Status,
			&log.ErrorMessage,
			&submissionID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan log: %w", err)
		}

		if len(parsedResultJSON) > 0 {
			if err := json.Unmarshal(parsedResultJSON, &log.ParsedResult); err != nil {
				return nil, fmt.Errorf("unmarshal parsed result: %w", err)
			}
		}

		log.SubmissionID = submissionID
		logs = append(logs, &log)
	}

	return logs, nil
}

// DeleteOlderThan удаляет записи старше olderThan порциями по 1000 строк,
// чтобы не держать длительную транзакцию и не блокировать таблицу.
func (r *llmExtractionLogRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM llm_extraction_logs WHERE created_at < $1`
	tag, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
