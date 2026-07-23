package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Config supplies updater dependencies. HTTP, filesystem, and process access
// are all replaceable so policy and future application tests do not touch the
// network, installed executable, or current process.
type Config struct {
	HTTP              HTTPClient
	FileSystem        FileSystem
	Process           Process
	Output            io.Writer
	ManifestURL       string
	Timeout           time.Duration
	OnUpdateAvailable func(Manifest) error
}

// Updater implements latest-release discovery, eligibility, and application.
type Updater struct {
	http              HTTPClient
	fileSystem        FileSystem
	process           Process
	output            io.Writer
	manifestURL       string
	timeout           time.Duration
	onUpdateAvailable func(Manifest) error
}

// Result records the validated latest manifest and whether it is strictly
// newer than the running stable version.
type Result struct {
	Manifest        Manifest
	UpdateAvailable bool
}

// New constructs an updater with production standard-library dependencies for
// every seam not supplied by the caller.
func New(config Config) *Updater {
	if config.HTTP == nil {
		config.HTTP = http.DefaultClient
	}
	if config.FileSystem == nil {
		config.FileSystem = osFileSystem{}
	}
	if config.Process == nil {
		config.Process = osProcess{}
	}
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.ManifestURL == "" {
		config.ManifestURL = LatestManifestURL
	}
	if config.Timeout == 0 {
		config.Timeout = RequestTimeout
	}
	return &Updater{
		http:              config.HTTP,
		fileSystem:        config.FileSystem,
		process:           config.Process,
		output:            config.Output,
		manifestURL:       config.ManifestURL,
		timeout:           config.Timeout,
		onUpdateAvailable: config.OnUpdateAvailable,
	}
}

// Check retrieves and evaluates the latest stable manifest.
func (updater *Updater) Check(ctx context.Context, currentVersion string) (Result, error) {
	current, err := ParseStableVersion(currentVersion)
	if err != nil {
		return Result{}, fmt.Errorf("invalid running version: %w", err)
	}
	manifest, err := fetchManifest(ctx, updater.http, updater.manifestURL, updater.timeout)
	if err != nil {
		return Result{}, err
	}
	latest, err := ParseStableVersion(manifest.Version)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Manifest:        manifest,
		UpdateAvailable: latest.Compare(current) > 0,
	}, nil
}

// Run performs automatic or forced update policy. Development builds bypass
// all dependencies, including HTTP construction and requests.
func (updater *Updater) Run(currentVersion string, forced bool) error {
	if hasRelaunchMarker(updater.process.Environ()) {
		return updater.completeRelaunch(currentVersion, forced)
	}
	if currentVersion == "dev" {
		if forced {
			_, err := fmt.Fprintln(updater.output, "self-update is disabled for dev builds")
			return err
		}
		return nil
	}

	result, err := updater.Check(context.Background(), currentVersion)
	if err != nil {
		return updater.warn("update discovery failed: %v", err)
	}
	if !result.UpdateAvailable {
		if forced {
			_, err := fmt.Fprintf(updater.output, "angel-ai %s is already current\n", currentVersion)
			return err
		}
		return nil
	}
	if updater.onUpdateAvailable != nil {
		if err := updater.onUpdateAvailable(result.Manifest); err != nil {
			return updater.warn("update application failed: %v", err)
		}
		return nil
	}
	if err := updater.apply(result.Manifest); err != nil {
		return updater.warn("update application failed: %v", err)
	}
	return nil
}

func (updater *Updater) warn(format string, arguments ...any) error {
	_, err := fmt.Fprintf(updater.output, "warning: "+format+"\n", arguments...)
	return err
}
