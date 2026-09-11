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

// OpenAICompatibleProvider универсальный провайдер для OpenAI-совместимых API.
// Поддерживает: OpenAI, OpenRouter, gptunnel.ru, Polza.ai, Bothub (если совместимы).
type OpenAICompatibleProvider struct {
	name        string
	endpoint    string
	model       string
	apiKey      string
	timeout     time.Duration
	maxTokens   int
	temperature float64
	client      *http.Client
}

func NewOpenAICompatibleProvider(name, endpoint, model, apiKey string, timeout time.Duration, maxTokens int, temperature float64) domain.LLMProvider {
	return &OpenAICompatibleProvider{
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

func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

func (p *OpenAICompatibleProvider) Type() string {
	return "openai_compatible"
}

func (p *OpenAICompatibleProvider) Model() string {
	return p.model
}

func (p *OpenAICompatibleProvider) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that extracts structured data from text about military heroes."},
			{"role": "user", "content": prompt},
		},
		"temperature": p.temperature,
		"max_tokens":  p.maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", domain.ErrLLMParseFailed
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func (p *OpenAICompatibleProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/models", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
