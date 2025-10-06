package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/post"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
)

type searxResp struct {
	Query   string        `json:"query"`
	Results []searxResult `json:"results"`
}
type searxResult struct {
	URL       string   `json:"url"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Published *string  `json:"publishedDate"`
	Thumbnail string   `json:"thumbnail"`
	ParsedURL []string `json:"parsed_url"`
}

type ChatSearchPayload struct {
	Query   string             `json:"query"`
	Total   int                `json:"total"`
	Results []ChatSearchResult `json:"results"`
}
type ChatSearchResult struct {
	Title       string `json:"title"`
	Url         string `json:"url"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source"`
	Type        string `json:"type"`        // "pdf" | "web"
	PublishedAt string `json:"publishedAt"` // empty if unknown
	Thumbnail   string `json:"thumbnail"`   // empty if none
}

type AgentToolWebSearch interface {
	SearchExternalWeb(message string) (string, error)
}

type agentAgentToolWebSearchService struct {
	logger  *zap.Logger
	PosRepo post.PostRepositoryInterface
	env     *config.Config
}

func NewAgentToolWebSearchService(logger *zap.Logger, posRepo post.PostRepositoryInterface, env *config.Config) AgentToolWebSearch {
	return &agentAgentToolWebSearchService{
		logger:  logger,
		PosRepo: posRepo,
		env:     env,
	}
}

// ---- helper ----
func trimRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return strings.TrimSpace(s)
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:max])) + "…"
}
func guessSource(u string, parsed []string) string {
	// searx provides parsed_url = [scheme, host, path, ...]
	if len(parsed) >= 2 && parsed[1] != "" {
		return parsed[1]
	}
	uu, err := url.Parse(u)
	if err == nil && uu.Host != "" {
		return uu.Host
	}
	return ""
}
func guessType(u string) string {
	uu, err := url.Parse(u)
	if err == nil {
		ext := strings.ToLower(path.Ext(uu.Path))
		if ext == ".pdf" {
			return "pdf"
		}
	}
	return "web"
}

// ---- main ----
func (a *agentAgentToolWebSearchService) SearchExternalWeb(message string) (string, error) {
	base := a.env.Searxng_url
	if base == "" {
		return "", fmt.Errorf("searxng_url not configured")
	}
	q := url.Values{}
	q.Set("q", message)
	q.Set("format", "json")
	fullURL := base + "?" + q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("read body: %w", readErr)
	}

	a.logger.Info("SearxNG web search",
		zap.String("url", fullURL),
		zap.Int("status", resp.StatusCode),
		zap.Duration("latency", time.Since(start)),
	)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("searxng status %d: %s", resp.StatusCode, string(raw))
	}

	// Parse SearxNG JSON
	var sr searxResp
	if err := json.Unmarshal(raw, &sr); err != nil {
		// If parsing fails, return the raw response for UI debugging
		return string(raw), nil
	}

	// Normalize -> Chat UI friendly
	out := ChatSearchPayload{
		Query: strings.TrimSpace(sr.Query),
	}
	maxItems := 5
	for i, r := range sr.Results {
		if i >= maxItems {
			break
		}
		item := ChatSearchResult{
			Title:       strings.TrimSpace(r.Title),
			Url:         strings.TrimSpace(r.URL),
			Snippet:     trimRunes(strings.ReplaceAll(strings.TrimSpace(r.Content), "\n", " "), 240),
			Source:      guessSource(r.URL, r.ParsedURL),
			Type:        guessType(r.URL),
			PublishedAt: "",
			Thumbnail:   strings.TrimSpace(r.Thumbnail),
		}
		if r.Published != nil {
			item.PublishedAt = strings.TrimSpace(*r.Published)
		}
		out.Results = append(out.Results, item)
	}
	out.Total = len(out.Results)

	b, _ := json.Marshal(out) // safe: fields are already controlled
	return string(b), nil
}
