package updater

import (
	"fmt"
	"math/big"
	"net/url"
	"regexp"
)

var (
	stableVersionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	sha256Pattern        = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Manifest describes the single stable artifact published by the latest
// GitHub Release.
type Manifest struct {
	Version     string `json:"version"`
	ArtifactURL string `json:"artifact_url"`
	SHA256      string `json:"sha256"`
}

// Version is a stable vMAJOR.MINOR.PATCH release version.
type Version struct {
	parts [3]big.Int
}

// ParseStableVersion accepts only stable release versions. Prerelease and
// build metadata, omitted v prefixes, and leading zeroes are rejected.
func ParseStableVersion(value string) (Version, error) {
	matches := stableVersionPattern.FindStringSubmatch(value)
	if matches == nil {
		return Version{}, fmt.Errorf("version %q is not stable vMAJOR.MINOR.PATCH", value)
	}

	var parsed Version
	for index, part := range matches[1:] {
		if _, ok := parsed.parts[index].SetString(part, 10); !ok {
			return Version{}, fmt.Errorf("version %q contains an invalid number", value)
		}
	}
	return parsed, nil
}

// Compare reports whether version is older than (-1), equal to (0), or newer
// than other (1).
func (version Version) Compare(other Version) int {
	for index := range version.parts {
		if comparison := version.parts[index].Cmp(&other.parts[index]); comparison != 0 {
			return comparison
		}
	}
	return 0
}

// ValidateManifest validates every field used to select a release artifact.
func ValidateManifest(manifest Manifest) error {
	if _, err := ParseStableVersion(manifest.Version); err != nil {
		return fmt.Errorf("invalid manifest version: %w", err)
	}

	artifactURL, err := url.Parse(manifest.ArtifactURL)
	if err != nil || artifactURL.Scheme != "https" || artifactURL.Host == "" || artifactURL.User != nil {
		return fmt.Errorf("invalid manifest artifact_url %q: must be an absolute HTTPS URL without credentials", manifest.ArtifactURL)
	}
	if !sha256Pattern.MatchString(manifest.SHA256) {
		return fmt.Errorf("invalid manifest sha256: must be 64 lowercase hexadecimal characters")
	}
	return nil
}
