package nutrition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrAIUnavailable = errors.New("ai unavailable")
	ErrAIBlocked     = errors.New("ai response blocked")
)

// Turn is one conversation turn sent to Gemini. Role is "user" or "model".
type Turn struct {
	Role string
	Text string
}

// AI is what the service needs; GeminiClient implements it.
type AI interface {
	Generate(ctx context.Context, system string, turns []Turn) (string, error)
}

// GeminiClient calls the Gemini API over plain HTTPS (no SDK needed).
// The API key lives ONLY on this server, never in the mobile app.
//
// Rules for Gemini 3.5 Flash-Lite (generateContent):
//   - do not send temperature / top_p / top_k (deprecated, ignored)
//   - the last turn must be a "user" turn
type GeminiClient struct {
	apiKey string
	model  string
	http   *http.Client
}

const defaultModel = "gemini-3.5-flash-lite"

var modelPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// cleanModel tolerates sloppy .env values: surrounding spaces or quotes, a
// "models/" prefix, or an empty / malformed name (falls back to the default).
func cleanModel(m string) string {
	m = strings.TrimSpace(m)
	m = strings.Trim(m, "\"'")
	m = strings.TrimPrefix(m, "models/")
	m = strings.TrimSpace(m)
	if !modelPattern.MatchString(m) {
		return defaultModel
	}
	return m
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	clean := cleanModel(model)
	if clean != model {
		log.Printf("gemini: model name %q was cleaned to %q", model, clean)
	}
	return &GeminiClient{apiKey: apiKey, model: clean, http: &http.Client{Timeout: 45 * time.Second}}
}

type gPart struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type gContent struct {
	Role  string  `json:"role,omitempty"`
	Parts []gPart `json:"parts"`
}

type gRequest struct {
	SystemInstruction *gContent  `json:"systemInstruction,omitempty"`
	Contents          []gContent `json:"contents"`
	GenerationConfig  struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

type gResponse struct {
	Candidates []struct {
		Content      gContent `json:"content"`
		FinishReason string   `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

func (g *GeminiClient) Generate(ctx context.Context, system string, turns []Turn) (string, error) {
	req := gRequest{SystemInstruction: &gContent{Parts: []gPart{{Text: system}}}}
	req.GenerationConfig.MaxOutputTokens = 900
	for _, t := range turns {
		req.Contents = append(req.Contents, gContent{Role: t.Role, Parts: []gPart{{Text: t.Text}}})
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", g.model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.http.Do(httpReq)
	if err != nil {
		log.Printf("gemini: request failed: %v", err)
		return "", ErrAIUnavailable
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 400 {
			snippet = snippet[:400]
		}
		log.Printf("gemini: status %d: %s", resp.StatusCode, snippet) // the key is never in the response
		return "", ErrAIUnavailable
	}

	var out gResponse
	if err := json.Unmarshal(body, &out); err != nil {
		log.Printf("gemini: bad JSON: %v", err)
		return "", ErrAIUnavailable
	}
	if out.PromptFeedback.BlockReason != "" {
		return "", ErrAIBlocked
	}
	if len(out.Candidates) == 0 {
		return "", ErrAIUnavailable
	}

	var sb strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		if !p.Thought {
			sb.WriteString(p.Text)
		}
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		if out.Candidates[0].FinishReason == "SAFETY" {
			return "", ErrAIBlocked
		}
		return "", ErrAIUnavailable
	}
	return text, nil
}