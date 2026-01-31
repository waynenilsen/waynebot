package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	httpFetchTimeout     = 15 * time.Second
	httpFetchMaxResponse = 50 * 1024 // 50KB
)

type httpFetchArgs struct {
	URL    string            `json:"url"`
	Method string            `json:"method"`
	Header map[string]string `json:"header"`
}

// HTTPFetch returns a ToolFunc that fetches HTTP URLs with host restrictions and
// response size limits.
func HTTPFetch(cfg *SandboxConfig) ToolFunc {
	client := &http.Client{Timeout: httpFetchTimeout}

	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var args httpFetchArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid args: %w", err)
		}
		if args.URL == "" {
			return "", fmt.Errorf("url is required")
		}
		if args.Method == "" {
			args.Method = http.MethodGet
		}

		parsed, err := url.Parse(args.URL)
		if err != nil {
			return "", fmt.Errorf("invalid url: %w", err)
		}

		host := parsed.Hostname()
		if cfg.IsHostBlocked(host) {
			return "", fmt.Errorf("host %q is blocked", host)
		}

		req, err := http.NewRequestWithContext(ctx, args.Method, args.URL, nil)
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		for k, v := range args.Header {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("fetch: %w", err)
		}
		defer resp.Body.Close()

		limited := io.LimitReader(resp.Body, httpFetchMaxResponse+1)
		body, err := io.ReadAll(limited)
		if err != nil {
			return "", fmt.Errorf("read body: %w", err)
		}

		result := string(body)
		if len(body) > httpFetchMaxResponse {
			result = result[:httpFetchMaxResponse] + "\n... response truncated"
		}

		if resp.StatusCode >= 400 {
			return result, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return result, nil
	}
}
