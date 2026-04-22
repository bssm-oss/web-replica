package analyzer

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	defaultMaxResponseBytes = 2 << 20
	maxRedirects            = 3
	userAgent               = "SiteforgeBot/0.1 (+for user-authorized analysis)"
)

type FetchedPage struct {
	URL         string
	ContentType string
	Body        []byte
	FetchedAt   time.Time
}

func FetchHTML(ctx context.Context, target ValidatedURL, timeout time.Duration) (FetchedPage, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if err := validateHost(ctx, host); err != nil {
				return nil, err
			}
			var d net.Dialer
			return d.DialContext(ctx, network, address)
		},
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			_, err := ValidateURL(req.Context(), req.URL.String())
			return err
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.Normalized, nil)
	if err != nil {
		return FetchedPage{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return FetchedPage{}, fmt.Errorf("fetch html: %w", err)
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, defaultMaxResponseBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		return FetchedPage{}, err
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "text/html") {
		contentType = contentType + " (warning: not text/html)"
	}
	return FetchedPage{URL: resp.Request.URL.String(), ContentType: contentType, Body: body, FetchedAt: time.Now().UTC()}, nil
}

func filenameForURL(input string) string {
	clean := path.Clean(input)
	return strings.Trim(strings.ReplaceAll(clean, "/", "-"), "-")
}
