package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

const (
	RequestTimeout          = 2 * time.Second
	ArtifactRequestTimeout  = 2 * time.Minute
	DefaultMaxManifestBytes = int64(64 << 10)
	DefaultMaxArtifactBytes = int64(128 << 20)
)

// LatestManifestURL points directly at the manifest asset for the running
// GOOS/GOARCH pair on the latest GitHub Release and does not use the GitHub
// API. Every release publishes one manifest per supported platform so a
// self-update never downloads an artifact built for another platform.
func LatestManifestURL() string {
	return fmt.Sprintf("https://github.com/Angel-M-R/angel-ai-opencode/releases/latest/download/manifest-%s-%s.json", runtime.GOOS, runtime.GOARCH)
}

// HTTPClient is the network seam used for deterministic manifest tests.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func fetchManifest(ctx context.Context, client HTTPClient, manifestURL string, timeout time.Duration, maxBytes int64) (Manifest, error) {
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, manifestURL, nil)
	if err != nil {
		return Manifest{}, fmt.Errorf("creating manifest request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return Manifest{}, fmt.Errorf("fetching latest manifest: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Manifest{}, fmt.Errorf("fetching latest manifest: unexpected HTTP status %s", response.Status)
	}
	if response.ContentLength > maxBytes {
		return Manifest{}, fmt.Errorf("fetching latest manifest: response exceeds %d bytes", maxBytes)
	}
	body, err := readAtMost(response.Body, maxBytes)
	if err != nil {
		return Manifest{}, fmt.Errorf("fetching latest manifest: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decoding latest manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return Manifest{}, fmt.Errorf("decoding latest manifest: %w", err)
	}
	if err := ValidateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func readAtMost(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := &io.LimitedReader{R: reader, N: maxBytes + 1}
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return body, nil
}
