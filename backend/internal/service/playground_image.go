package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const playgroundMaxImageBytes int64 = 20 << 20

// IsAllowedPlaygroundImageURL reports whether a playground image URL is a
// public https destination. Loopback, private, link-local, and metadata
// addresses are rejected; the download client also blocks these at dial time
// and on every redirect hop.
func IsAllowedPlaygroundImageURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}
	if parsed.Host == "" || parsed.User != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "" || isBlockedImageDownloadHost(host) {
		return false
	}
	return true
}

// FetchPlaygroundImageURL downloads a previously generated playground image
// through the same SSRF-safe client used for object-storage backfill.
func FetchPlaygroundImageURL(ctx context.Context, rawURL string) ([]byte, string, error) {
	return fetchPlaygroundImageURL(ctx, rawURL, nil)
}

func playgroundImageHTTPClient() *http.Client {
	client := newImageDownloadHTTPClient(http.ProxyFromEnvironment)
	client.Timeout = 30 * time.Second
	client.CheckRedirect = playgroundImageCheckRedirect
	return client
}

func playgroundImageCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return fmt.Errorf("too many redirects")
	}
	if req.URL == nil || !IsAllowedPlaygroundImageURL(req.URL.String()) {
		return fmt.Errorf("redirect url is not allowed")
	}
	return nil
}

func fetchPlaygroundImageURL(ctx context.Context, rawURL string, client *http.Client) ([]byte, string, error) {
	if !IsAllowedPlaygroundImageURL(rawURL) {
		return nil, "", playgroundError(502, "IMAGE_DOWNLOAD_FAILED", "failed to download image")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	if client == nil {
		client = playgroundImageHTTPClient()
	} else if client.CheckRedirect == nil {
		client.CheckRedirect = playgroundImageCheckRedirect
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", playgroundError(502, "IMAGE_DOWNLOAD_FAILED", "failed to download image")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, playgroundMaxImageBytes+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > playgroundMaxImageBytes {
		return nil, "", playgroundError(502, "IMAGE_DOWNLOAD_FAILED", "image exceeds the download size limit")
	}
	return data, strings.TrimSpace(resp.Header.Get("Content-Type")), nil
}

// PlaygroundMaxImageBytes is the maximum decoded image payload accepted from
// either a remote URL or a base64 payload.
func PlaygroundMaxImageBytes() int64 {
	return playgroundMaxImageBytes
}
