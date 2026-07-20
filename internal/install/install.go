// Package install reconciles selected assets and integrations into an OpenCode
// configuration directory.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type fileWriteResult struct {
	changed    bool
	created    bool
	backupPath string
}

func reconcileFile(file preparedFile) (fileWriteResult, error) {
	previous, err := os.ReadFile(file.path)
	created := false
	switch {
	case err == nil:
		if file.contentMatches(previous) {
			return fileWriteResult{}, nil
		}
	case os.IsNotExist(err):
		created = true
		previous = nil
	default:
		return fileWriteResult{}, err
	}

	result := fileWriteResult{changed: true, created: created}
	if !created {
		backupPath, err := writeBackup(file.path, previous)
		if err != nil {
			return fileWriteResult{}, fmt.Errorf("writing backup: %w", err)
		}
		result.backupPath = backupPath
	}

	dir := filepath.Dir(file.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fileWriteResult{}, err
	}
	temp, err := os.CreateTemp(dir, ".angel-ai-*.tmp")
	if err != nil {
		return fileWriteResult{}, err
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	if err := temp.Chmod(file.perm); err != nil {
		cleanup()
		return fileWriteResult{}, err
	}
	if _, err := temp.Write(file.content); err != nil {
		cleanup()
		return fileWriteResult{}, err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return fileWriteResult{}, err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fileWriteResult{}, err
	}
	if err := os.Rename(tempPath, file.path); err != nil {
		_ = os.Remove(tempPath)
		return fileWriteResult{}, err
	}
	return result, nil
}

func fileResultLines(path string, result fileWriteResult) []string {
	var lines []string
	if result.backupPath != "" {
		lines = append(lines, "backup    "+result.backupPath)
	}
	switch {
	case result.created:
		lines = append(lines, "creado    "+path)
	case result.changed:
		lines = append(lines, "actualizado "+path)
	default:
		lines = append(lines, "sin cambios "+path)
	}
	return lines
}

func writeBackup(targetPath string, content []byte) (string, error) {
	pattern := "." + filepath.Base(targetPath) + ".bak-" + time.Now().Format("20060102-150405") + "-*"
	backup, err := os.CreateTemp(filepath.Dir(targetPath), pattern)
	if err != nil {
		return "", err
	}
	tempPath := backup.Name()
	backupPath := filepath.Join(filepath.Dir(targetPath), strings.TrimPrefix(filepath.Base(tempPath), "."))
	cleanup := func() {
		_ = backup.Close()
		_ = os.Remove(tempPath)
	}
	if err := backup.Chmod(0o600); err != nil {
		cleanup()
		return "", err
	}
	if _, err := backup.Write(content); err != nil {
		cleanup()
		return "", err
	}
	if err := backup.Sync(); err != nil {
		cleanup()
		return "", err
	}
	if err := backup.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Link(tempPath, backupPath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Remove(tempPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

// merge deep-merges src into dst. Objects merge recursively, plugin arrays are
// reconciled by plugin identity, and every other array is replaced because its
// order and positional meaning may be significant (for example MCP commands).
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
					if key == "plugin" {
						dst[key] = mergePluginArray(dstArr, srcArr)
					} else {
						dst[key] = srcArr
					}
					continue
				}
			}
		}
		dst[key] = value
	}
}

func mergePluginArray(existing, desired []any) []any {
	desiredLatest := make(map[string]any, len(desired))
	desiredLastIndex := make(map[string]int, len(desired))
	for index, value := range desired {
		identity := pluginIdentity(value)
		desiredLatest[identity] = value
		desiredLastIndex[identity] = index
	}

	result := make([]any, 0, len(existing)+len(desired))
	seen := make(map[string]bool, len(existing)+len(desired))
	for _, value := range existing {
		identity := pluginIdentity(value)
		if seen[identity] {
			continue
		}
		seen[identity] = true
		if replacement, ok := desiredLatest[identity]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, value)
		}
	}
	for index, value := range desired {
		identity := pluginIdentity(value)
		if !seen[identity] && desiredLastIndex[identity] == index {
			seen[identity] = true
			result = append(result, value)
		}
	}
	return result
}

func pluginIdentity(value any) string {
	text, ok := value.(string)
	if !ok {
		encoded, _ := json.Marshal(value)
		return "json:" + string(encoded)
	}
	if strings.HasPrefix(text, "file:") || strings.Contains(text, "://") || filepath.IsAbs(text) {
		return text
	}
	if strings.HasPrefix(text, "@") {
		if version := strings.Index(text[1:], "@"); version >= 0 {
			return text[:version+1]
		}
		return text
	}
	if version := strings.IndexByte(text, '@'); version >= 0 {
		return text[:version]
	}
	return text
}
