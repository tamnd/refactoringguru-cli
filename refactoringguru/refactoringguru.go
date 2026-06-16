// Package refactoringguru is the library behind the refactoringguru CLI.
package refactoringguru

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const DefaultUserAgent = "refactoringguru-cli/dev (+https://github.com/tamnd/refactoringguru-cli)"

type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://refactoring.guru",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var (
	cardRe    = regexp.MustCompile(`(?s)<a class="pattern-card ([\w-]+)" href="(/design-patterns/[\w-]+)">(.*?)</a>`)
	nameRe    = regexp.MustCompile(`<span class="pattern-name">([^<]+)<`)
	tagRe     = regexp.MustCompile(`<[^>]+>`)
)

// classify maps a pattern slug to its GoF category.
func classify(slug string) string {
	switch slug {
	case "factory-method", "abstract-factory", "builder", "prototype", "singleton":
		return "Creational"
	case "adapter", "bridge", "composite", "decorator", "facade", "flyweight", "proxy":
		return "Structural"
	default:
		return "Behavioral"
	}
}

// Patterns fetches the design-patterns catalog and returns all 22 GoF patterns.
func (c *Client) Patterns(ctx context.Context) ([]*Pattern, error) {
	body, err := c.get(ctx, c.cfg.BaseURL+"/design-patterns/catalog")
	if err != nil {
		return nil, err
	}
	html := string(body)

	var patterns []*Pattern
	rank := 0
	for _, m := range cardRe.FindAllStringSubmatch(html, -1) {
		slug := m[1]
		href := m[2]
		content := m[3]

		nm := nameRe.FindStringSubmatch(content)
		if nm == nil {
			continue
		}
		rank++
		patterns = append(patterns, &Pattern{
			Rank:     rank,
			Slug:     slug,
			Category: classify(slug),
			Title:    strings.TrimSpace(nm[1]),
			URL:      c.cfg.BaseURL + href,
		})
	}
	return patterns, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// keep tagRe from being flagged as unused if the compiler is strict
var _ = tagRe

// Stats returns aggregate statistics about the design-patterns catalog.
func (c *Client) Stats(ctx context.Context) (*Info, error) {
	patterns, err := c.Patterns(ctx)
	if err != nil {
		return nil, err
	}
	info := &Info{
		TotalPatterns: len(patterns),
		SiteURL:       c.cfg.BaseURL,
		CatalogURL:    c.cfg.BaseURL + "/design-patterns/catalog",
	}
	for _, p := range patterns {
		switch p.Category {
		case "Creational":
			info.Creational++
		case "Structural":
			info.Structural++
		default:
			info.Behavioral++
		}
	}
	return info, nil
}
