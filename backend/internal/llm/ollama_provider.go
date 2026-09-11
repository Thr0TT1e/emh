package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// OllamaProvider реализация domain.LLMProvider для локальной Ollama.
type OllamaProvider struct {
	name     string
	endpoint string
	model    string
	timeout  time.Duration
	client   *http.Client
}

// Компиляционная проверка интерфейса.
var _ domain.LLMProvider = (*OllamaProvider)(nil)

type ollamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaProvider создаёт провайдер Ollama.
func NewOllamaProvider(name, endpoint, model string, timeout time.Duration) domain.LLMProvider {
	return &OllamaProvider{
		name:     name,
		endpoint: endpoint,
		model:    model,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *OllamaProvider) Name() string  { return p.name }
func (p *OllamaProvider) Type() string  { return "ollama" }
func (p *OllamaProvider) Model() string { return p.model }

func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.1,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/generate", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", domain.ErrLLMTimeout
		}
		return "", fmt.Errorf("ollama http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

// HealthCheck проверяет доступность Ollama через /api/tags.
func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}
	return nil
}
