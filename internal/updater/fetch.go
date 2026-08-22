package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

const RequestTimeout = 2 * time.Second

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

func fetchManifest(ctx context.Context, client HTTPClient, manifestURL string, timeout time.Duration) (Manifest, error) {
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

	decoder := json.NewDecoder(response.Body)
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
