package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Translator interface {
	Translate(ctx context.Context, titleVI, bodyVI string) (titleEN, bodyEN string, err error)
}

type claudeTranslator struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClaudeTranslator(apiKey string) Translator {
	return &claudeTranslator{
		apiKey:     apiKey,
		baseURL:    "https://api.anthropic.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func newTranslatorWithClient(apiKey, baseURL string, client *http.Client) Translator {
	return &claudeTranslator{apiKey: apiKey, baseURL: baseURL, httpClient: client}
}

func (t *claudeTranslator) Translate(ctx context.Context, titleVI, bodyVI string) (string, string, error) {
	prompt := fmt.Sprintf(
		"Translate this Vietnamese blog post to English. Return ONLY a JSON object with two fields: \"title\" and \"body\". No other text.\n\nVietnamese title: %s\n\nVietnamese body:\n%s",
		titleVI, bodyVI,
	)

	body, _ := json.Marshal(map[string]interface{}{
		"model":      "claude-sonnet-4-6",
		"max_tokens": 4096,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("x-api-key", t.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("claude api error: status %d", resp.StatusCode)
	}

	var apiResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", "", err
	}
	if len(apiResp.Content) == 0 {
		return "", "", fmt.Errorf("empty response from claude api")
	}

	var translated struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal([]byte(apiResp.Content[0].Text), &translated); err != nil {
		return "", "", fmt.Errorf("failed to parse translation json: %w", err)
	}

	return translated.Title, translated.Body, nil
}
