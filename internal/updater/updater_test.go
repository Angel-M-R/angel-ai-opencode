package updater

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const validSHA256 = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type httpClientFunc func(*http.Request) (*http.Response, error)

func (function httpClientFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

func manifestResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func manifestJSON(version string) string {
	return `{"version":"` + version + `","artifact_url":"https://github.com/angelmr/angel-ai-opencode/releases/download/` + version + `/angel-ai","sha256":"` + validSHA256 + `"}`
}

func TestValidateManifestRequiresStableVersionHTTPSAndLowercaseSHA256(t *testing.T) {
	valid := Manifest{
		Version:     "v1.2.3",
		ArtifactURL: "https://github.com/angelmr/angel-ai-opencode/releases/download/v1.2.3/angel-ai",
		SHA256:      validSHA256,
	}
	if err := ValidateManifest(valid); err != nil {
		t.Fatalf("valid manifest: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{name: "malformed version", mutate: func(manifest *Manifest) { manifest.Version = "1.2.3" }},
		{name: "prerelease version", mutate: func(manifest *Manifest) { manifest.Version = "v1.2.3-beta.1" }},
		{name: "leading zero version", mutate: func(manifest *Manifest) { manifest.Version = "v1.02.3" }},
		{name: "non-HTTPS URL", mutate: func(manifest *Manifest) { manifest.ArtifactURL = "http://example.com/angel-ai" }},
		{name: "relative URL", mutate: func(manifest *Manifest) { manifest.ArtifactURL = "/angel-ai" }},
		{name: "credentialed URL", mutate: func(manifest *Manifest) { manifest.ArtifactURL = "https://user@example.com/angel-ai" }},
		{name: "uppercase SHA-256", mutate: func(manifest *Manifest) { manifest.SHA256 = strings.ToUpper(validSHA256) }},
		{name: "short SHA-256", mutate: func(manifest *Manifest) { manifest.SHA256 = validSHA256[:63] }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := valid
			test.mutate(&manifest)
			if err := ValidateManifest(manifest); err == nil {
				t.Fatal("invalid manifest was accepted")
			}
		})
	}
}

func TestCheckClassifiesEqualOlderAndNewerStableManifests(t *testing.T) {
	tests := []struct {
		name          string
		current       string
		latest        string
		wantAvailable bool
	}{
		{name: "equal", current: "v1.2.3", latest: "v1.2.3"},
		{name: "older", current: "v1.2.3", latest: "v1.2.2"},
		{name: "newer patch", current: "v1.2.3", latest: "v1.2.4", wantAvailable: true},
		{name: "newer major with large number", current: "v2.0.0", latest: "v100000000000000000000.0.0", wantAvailable: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := New(Config{HTTP: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusOK, manifestJSON(test.latest)), nil
			})})
			result, err := updater.Check(context.Background(), test.current)
			if err != nil {
				t.Fatal(err)
			}
			if result.UpdateAvailable != test.wantAvailable {
				t.Fatalf("UpdateAvailable = %v, want %v", result.UpdateAvailable, test.wantAvailable)
			}
		})
	}
}

func TestCheckRejectsManifestRetrievalFailures(t *testing.T) {
	transportError := errors.New("offline")
	tests := []struct {
		name     string
		client   HTTPClient
		wantText string
	}{
		{
			name: "transport",
			client: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return nil, transportError
			}),
			wantText: "offline",
		},
		{
			name: "status",
			client: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusServiceUnavailable, "unavailable"), nil
			}),
			wantText: "unexpected HTTP status",
		},
		{
			name: "decoding",
			client: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusOK, `{"version":`), nil
			}),
			wantText: "decoding latest manifest",
		},
		{
			name: "unknown field",
			client: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusOK, strings.TrimSuffix(manifestJSON("v1.2.4"), "}")+`,"extra":true}`), nil
			}),
			wantText: "unknown field",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := New(Config{HTTP: test.client})
			_, err := updater.Check(context.Background(), "v1.2.3")
			if err == nil || !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("error = %v, want text %q", err, test.wantText)
			}
		})
	}
}

func TestManifestRequestUsesDirectURLAndTwoSecondTimeout(t *testing.T) {
	started := time.Now()
	updater := New(Config{HTTP: httpClientFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != LatestManifestURL {
			t.Fatalf("request URL = %q", request.URL)
		}
		deadline, ok := request.Context().Deadline()
		if !ok {
			t.Fatal("manifest request has no deadline")
		}
		remaining := deadline.Sub(started)
		if remaining < RequestTimeout-100*time.Millisecond || remaining > RequestTimeout+100*time.Millisecond {
			t.Fatalf("request deadline = %v after start, want %v", remaining, RequestTimeout)
		}
		return manifestResponse(http.StatusOK, manifestJSON("v1.2.4")), nil
	})})
	if _, err := updater.Check(context.Background(), "v1.2.3"); err != nil {
		t.Fatal(err)
	}
}

func TestManifestRequestTimeoutCancelsTransport(t *testing.T) {
	updater := New(Config{
		Timeout: time.Millisecond,
		HTTP: httpClientFunc(func(request *http.Request) (*http.Response, error) {
			<-request.Context().Done()
			return nil, request.Context().Err()
		}),
	})
	_, err := updater.Check(context.Background(), "v1.2.3")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

func TestRunHandlesCurrentNewerAndDevPolicies(t *testing.T) {
	t.Run("forced current", func(t *testing.T) {
		var output bytes.Buffer
		updater := New(Config{
			Output: &output,
			HTTP: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusOK, manifestJSON("v1.2.2")), nil
			}),
		})
		if err := updater.Run("v1.2.3", true); err != nil {
			t.Fatal(err)
		}
		if output.String() != "angel-ai v1.2.3 is already current\n" {
			t.Fatalf("output = %q", output.String())
		}
	})

	t.Run("newer release", func(t *testing.T) {
		var available Manifest
		updater := New(Config{
			HTTP: httpClientFunc(func(*http.Request) (*http.Response, error) {
				return manifestResponse(http.StatusOK, manifestJSON("v1.2.4")), nil
			}),
			OnUpdateAvailable: func(manifest Manifest) error {
				available = manifest
				return nil
			},
		})
		if err := updater.Run("v1.2.3", false); err != nil {
			t.Fatal(err)
		}
		if available.Version != "v1.2.4" {
			t.Fatalf("available version = %q", available.Version)
		}
	})

	for _, forced := range []bool{false, true} {
		t.Run("dev", func(t *testing.T) {
			var output bytes.Buffer
			requests := 0
			updater := New(Config{
				Output: &output,
				HTTP: httpClientFunc(func(*http.Request) (*http.Response, error) {
					requests++
					return nil, errors.New("dev requested network")
				}),
			})
			if err := updater.Run("dev", forced); err != nil {
				t.Fatal(err)
			}
			if requests != 0 {
				t.Fatalf("network requests = %d", requests)
			}
			wantOutput := ""
			if forced {
				wantOutput = "self-update is disabled for dev builds\n"
			}
			if output.String() != wantOutput {
				t.Fatalf("output = %q, want %q", output.String(), wantOutput)
			}
		})
	}
}
