// Package install writes selected catalog items into the opencode config
// directory. File and directory items are copied verbatim; JSON fragments are
// deep-merged into opencode.json after saving a timestamped .bak copy.
package install

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"angel-ai-opencode/internal/catalog"
)

// Plan describes what Apply would do, one line per action.
func Plan(items []catalog.Item, configDir string) []string {
	var lines []string
	for _, item := range items {
		switch item.Kind {
		case catalog.MergeJSON:
			lines = append(lines, fmt.Sprintf("merge  %s → opencode.json", filepath.Base(item.Source)))
		case catalog.CopyDir:
			lines = append(lines, fmt.Sprintf("copiar %s/ → %s/", item.Name, filepath.Join(configDir, item.Dest)))
		default:
			lines = append(lines, fmt.Sprintf("copiar %s → %s", filepath.Base(item.Source), filepath.Join(configDir, item.Dest)))
		}
	}
	return lines
}

// Apply installs the items and returns one line per completed action.
func Apply(items []catalog.Item, configDir string) ([]string, error) {
	var done []string
	var fragments []catalog.Item
	for _, item := range items {
		if item.Kind == catalog.MergeJSON {
			fragments = append(fragments, item)
			continue
		}
		dest := filepath.Join(configDir, item.Dest)
		var err error
		if item.Kind == catalog.CopyDir {
			err = copyDir(item.Source, dest)
		} else {
			err = copyFile(item.Source, dest)
		}
		if err != nil {
			return done, fmt.Errorf("installing %s: %w", item.Name, err)
		}
		done = append(done, "instalado "+dest)
	}

	if len(fragments) > 0 {
		lines, err := mergeFragments(fragments, configDir)
		done = append(done, lines...)
		if err != nil {
			return done, err
		}
	}
	return done, nil
}

func mergeFragments(fragments []catalog.Item, configDir string) ([]string, error) {
	var patches []map[string]any
	var done []string
	for _, fragment := range fragments {
		raw, err := os.ReadFile(fragment.Source)
		if err != nil {
			return done, err
		}
		var patch map[string]any
		if err := json.Unmarshal(raw, &patch); err != nil {
			return done, fmt.Errorf("parsing fragment %s: %w", fragment.Name, err)
		}
		patches = append(patches, patch)
		done = append(done, "merge     "+filepath.Base(fragment.Source))
	}

	configPath := filepath.Join(configDir, "opencode.json")
	lines, err := mergeJSON(configPath, "https://opencode.ai/config.json", patches)
	done = append(done, lines...)
	return done, err
}

// mergeJSON deep-merges patches (applied in order) into the JSON object at
// targetPath, creating it with defaultSchema as its base if it doesn't exist
// yet. An existing file is backed up first with a timestamped .bak copy.
func mergeJSON(targetPath, defaultSchema string, patches []map[string]any) ([]string, error) {
	config := map[string]any{"$schema": defaultSchema}

	raw, err := os.ReadFile(targetPath)
	switch {
	case err == nil:
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", targetPath, err)
		}
	case !os.IsNotExist(err):
		return nil, err
	}

	for _, patch := range patches {
		merge(config, patch)
	}
	return writeJSON(targetPath, config, raw)
}

func writeJSON(targetPath string, config map[string]any, previous []byte) ([]string, error) {
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	var done []string
	if previous != nil {
		backup, err := writeBackup(targetPath, previous)
		if err != nil {
			return nil, fmt.Errorf("writing backup: %w", err)
		}
		done = append(done, "backup    "+backup)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return done, err
	}
	if err := os.WriteFile(targetPath, append(out, '\n'), 0o644); err != nil {
		return done, err
	}
	done = append(done, "escrito   "+targetPath)
	return done, nil
}

func writeBackup(targetPath string, content []byte) (string, error) {
	pattern := filepath.Base(targetPath) + ".bak-" + time.Now().Format("20060102-150405") + "-*"
	backup, err := os.CreateTemp(filepath.Dir(targetPath), pattern)
	if err != nil {
		return "", err
	}
	backupPath := backup.Name()
	cleanup := func() {
		_ = backup.Close()
		_ = os.Remove(backupPath)
	}
	if err := backup.Chmod(0o644); err != nil {
		cleanup()
		return "", err
	}
	if _, err := backup.Write(content); err != nil {
		cleanup()
		return "", err
	}
	if err := backup.Close(); err != nil {
		_ = os.Remove(backupPath)
		return "", err
	}
	return backupPath, nil
}

// merge deep-merges src into dst: maps merge recursively, arrays union
// (existing entries kept, new ones appended), scalars overwrite.
func merge(dst, src map[string]any) {
	for key, value := range src {
		if existing, ok := dst[key]; ok {
			if dstMap, ok1 := existing.(map[string]any); ok1 {
				if srcMap, ok2 := value.(map[string]any); ok2 {
					merge(dstMap, srcMap)
					continue
				}
			}
			if dstArr, ok1 := existing.([]any); ok1 {
				if srcArr, ok2 := value.([]any); ok2 {
					dst[key] = unionArray(dstArr, srcArr)
					continue
				}
			}
		}
		dst[key] = value
	}
}

func unionArray(dst, src []any) []any {
	seen := make(map[string]bool, len(dst))
	for _, v := range dst {
		seen[fmt.Sprintf("%v", v)] = true
	}
	for _, v := range src {
		if !seen[fmt.Sprintf("%v", v)] {
			dst = append(dst, v)
		}
	}
	return dst
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func copyDir(src, dest string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
