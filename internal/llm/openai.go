package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatibleProvider works with OpenAI, OpenRouter, or any compatible API
type OpenAICompatibleProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewOpenAICompatibleProvider(baseURL, apiKey string) *OpenAICompatibleProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAICompatibleProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (o *OpenAICompatibleProvider) CategorizeImage(ctx context.Context, imageBytes []byte, categories []string, model string, customPrompt string) (string, error) {
	prompt := DefaultPrompt(categories, customPrompt)
	b64Image := base64.StdEncoding.EncodeToString(imageBytes)

	reqBody := chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: prompt},
					{Type: "image_url", ImageURL: &imageURL{URL: "data:image/jpeg;base64," + b64Image}},
				},
			},
		},
		MaxTokens: 50,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result chatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return strings.TrimSpace(strings.ToLower(result.Choices[0].Message.Content)), nil
}

func (o *OpenAICompatibleProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		// Many compatible endpoints don't support /models, return defaults
		return []string{"gpt-4o", "gpt-4o-mini"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{"gpt-4o", "gpt-4o-mini"}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{"gpt-4o", "gpt-4o-mini"}, nil
	}

	var result modelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return []string{"gpt-4o", "gpt-4o-mini"}, nil
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}

	if len(models) == 0 {
		return []string{"gpt-4o", "gpt-4o-mini"}, nil
	}
	return models, nil
}
