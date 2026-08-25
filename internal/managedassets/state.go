// Package managedassets records which installer selection produced the files
// under an OpenCode configuration directory.
package managedassets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"angel-ai-opencode/internal/install"
)

const (
	schemaVersion = 1
	stateFileName = ".angel-ai-state.json"
	maxStateBytes = 1 << 20
)

var errStateNotFound = errors.New("managed asset state not found")

type categorySelection struct {
	Name    string   `json:"name"`
	Sources []string `json:"sources"`
}

type selection struct {
	Categories  []categorySelection           `json:"categories"`
	Extras      map[string]bool               `json:"extras"`
	AgentModels install.AgentModelAssignments `json:"agent_models"`
}

type managedFile struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type state struct {
	SchemaVersion int           `json:"schema_version"`
	BundleDigest  string        `json:"bundle_digest"`
	Selection     selection     `json:"selection"`
	Files         []managedFile `json:"files"`
}

func statePath(configDir string) string {
	return filepath.Join(configDir, stateFileName)
}

func load(configDir string) (state, error) {
	path := statePath(configDir)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return state{}, fmt.Errorf("%w: %s", errStateNotFound, path)
	}
	if err != nil {
		return state{}, err
	}
	defer func() { _ = file.Close() }()

	limited := io.LimitReader(file, maxStateBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return state{}, err
	}
	if len(raw) > maxStateBytes {
		return state{}, fmt.Errorf("managed asset state exceeds %d bytes", maxStateBytes)
	}
	var loaded state
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loaded); err != nil {
		return state{}, fmt.Errorf("decoding managed asset state: %w", err)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return state{}, fmt.Errorf("decoding managed asset state: %w", err)
	}
	if err := validateState(loaded); err != nil {
		return state{}, err
	}
	return loaded, nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func validateState(record state) error {
	if record.SchemaVersion != schemaVersion {
		return fmt.Errorf("unsupported managed asset schema %d", record.SchemaVersion)
	}
	if !validDigest(record.BundleDigest) {
		return errors.New("managed asset state has an invalid bundle digest")
	}
	categoryNames := map[string]struct{}{}
	for _, category := range record.Selection.Categories {
		if category.Name == "" {
			return errors.New("managed asset state has an empty category name")
		}
		if _, exists := categoryNames[category.Name]; exists {
			return fmt.Errorf("managed asset state repeats category %q", category.Name)
		}
		categoryNames[category.Name] = struct{}{}
		seenSources := map[string]struct{}{}
		for _, source := range category.Sources {
			if source == "" || path.IsAbs(source) || path.Clean(source) != source || strings.HasPrefix(source, "../") {
				return fmt.Errorf("managed asset state has unsafe source %q", source)
			}
			if _, exists := seenSources[source]; exists {
				return fmt.Errorf("managed asset state repeats source %q", source)
			}
			seenSources[source] = struct{}{}
		}
	}
	seenFiles := map[string]struct{}{}
	for _, file := range record.Files {
		if !safeRelativePath(file.Path) {
			return fmt.Errorf("managed asset state has unsafe file path %q", file.Path)
		}
		if !validDigest(file.Digest) {
			return fmt.Errorf("managed asset state has an invalid digest for %q", file.Path)
		}
		if _, exists := seenFiles[file.Path]; exists {
			return fmt.Errorf("managed asset state repeats file %q", file.Path)
		}
		seenFiles[file.Path] = struct{}{}
	}
	return nil
}

func safeRelativePath(value string) bool {
	if value == "" || filepath.IsAbs(value) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	return clean == value && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func save(configDir string, record state) error {
	if err := validateState(record); err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if len(raw) > maxStateBytes {
		return fmt.Errorf("managed asset state exceeds %d bytes", maxStateBytes)
	}

	temp, err := os.CreateTemp(configDir, ".angel-ai-state-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	if err := temp.Chmod(0o600); err != nil {
		cleanup()
		return err
	}
	if _, err := temp.Write(raw); err != nil {
		cleanup()
		return err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, statePath(configDir)); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

func digestFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}

func sortedFiles(files []managedFile) []managedFile {
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}
