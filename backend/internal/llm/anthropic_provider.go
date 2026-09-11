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

// AnthropicProvider провайдер для Anthropic Claude API.
type AnthropicProvider struct {
	name        string
	endpoint    string
	model       string
	apiKey      string
	timeout     time.Duration
	maxTokens   int
	temperature float64
	client      *http.Client
}

func NewAnthropicProvider(name, endpoint, model, apiKey string, timeout time.Duration, maxTokens int, temperature float64) domain.LLMProvider {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	return &AnthropicProvider{
		name:        name,
		endpoint:    endpoint,
		model:       model,
		apiKey:      apiKey,
		timeout:     timeout,
		maxTokens:   maxTokens,
		temperature: temperature,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *AnthropicProvider) Name() string {
	return p.name
}

func (p *AnthropicProvider) Type() string {
	return "anthropic"
}

func (p *AnthropicProvider) Model() string {
	return p.model
}

func (p *AnthropicProvider) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  p.maxTokens,
		"temperature": p.temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/v1/messages", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", domain.ErrLLMTimeout
		}
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("provider %s returned status %d: %s", p.name, resp.StatusCode, string(body))
	}

	var anthropicResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", domain.ErrLLMParseFailed
	}

	return anthropicResp.Content[0].Text, nil
}

func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	// У Anthropic нет простого health endpoint, делаем минимальный запрос
	_, err := p.Generate(ctx, "test")
	return err
}
