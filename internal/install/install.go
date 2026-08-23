// Package install reconciles selected assets and integrations into an OpenCode
// configuration directory.
package install

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type pluginIdentityResolver func(any) string

var tuiPluginLiteralIdentities = map[string]string{
	"opencode-open-in-app":   "opencode-open-in-app",
	"openspec-task-progress": "opencode-openspec-task-tui",
}

type sourceToken struct {
	text    string
	literal bool
}

type fileWriteResult struct {
	changed    bool
	created    bool
	backupPath string
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

// mergeWithPluginIdentity deep-merges src into dst. Objects merge recursively, plugin arrays are
// reconciled by plugin identity, and every other array is replaced because its
// order and positional meaning may be significant (for example MCP commands).
func mergeWithPluginIdentity(dst, src map[string]any, resolveIdentity pluginIdentityResolver) {
	for key, value := range src {
		if existing, ok := dst[key]; ok {
			if dstMap, ok1 := existing.(map[string]any); ok1 {
				if srcMap, ok2 := value.(map[string]any); ok2 {
					mergeWithPluginIdentity(dstMap, srcMap, resolveIdentity)
					continue
				}
			}
			if dstArr, ok1 := existing.([]any); ok1 {
				if srcArr, ok2 := value.([]any); ok2 {
					if key == "plugin" {
						dst[key] = mergePluginArrayWithIdentity(dstArr, srcArr, resolveIdentity)
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

func mergePluginArrayWithIdentity(existing, desired []any, resolveIdentity pluginIdentityResolver) []any {
	desiredLatest := make(map[string]any, len(desired))
	desiredLastIndex := make(map[string]int, len(desired))
	for index, value := range desired {
		identity := resolveIdentity(value)
		desiredLatest[identity] = value
		desiredLastIndex[identity] = index
	}

	result := make([]any, 0, len(existing)+len(desired))
	reconciled := make(map[string]bool, len(desired))
	for _, value := range existing {
		identity := resolveIdentity(value)
		if replacement, ok := desiredLatest[identity]; ok {
			if reconciled[identity] {
				continue
			}
			reconciled[identity] = true
			result = append(result, replacement)
		} else {
			result = append(result, value)
		}
	}
	for index, value := range desired {
		identity := resolveIdentity(value)
		if !reconciled[identity] && desiredLastIndex[identity] == index {
			reconciled[identity] = true
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

func tuiPluginIdentityResolver(configDir string) pluginIdentityResolver {
	return func(value any) string {
		fallback := pluginIdentity(value)
		entry, ok := value.(string)
		if !ok {
			return fallback
		}
		bundlePath, ok := tuiPluginBundlePath(configDir, entry)
		if !ok {
			return fallback
		}
		bundle, err := os.ReadFile(bundlePath)
		if err != nil {
			return fallback
		}
		if literalID, ok := exportedPluginLiteralID(bundle); ok {
			if identity, recognized := tuiPluginLiteralIdentities[literalID]; recognized {
				return identity
			}
		}
		return fallback
	}
}

func exportedPluginLiteralID(source []byte) (string, bool) {
	tokens := tokenizePluginSource(source)
	depth := 0
	for index := 0; index+2 < len(tokens); index++ {
		switch tokens[index].text {
		case "{":
			depth++
			continue
		case "}":
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth != 0 || tokens[index].text != "export" ||
			tokens[index+1].text != "default" || tokens[index+2].text != "{" {
			continue
		}
		if literalID, ok := directLiteralIDProperty(tokens[index+3:]); ok {
			return literalID, true
		}
	}
	return "", false
}

func directLiteralIDProperty(tokens []sourceToken) (string, bool) {
	depth := 1
	propertyStart := true
	for index := 0; index < len(tokens); index++ {
		switch tokens[index].text {
		case "{":
			depth++
		case "}":
			depth--
			if depth == 0 {
				return "", false
			}
		case ",":
			if depth == 1 {
				propertyStart = true
			}
		default:
			if depth != 1 || !propertyStart {
				continue
			}
			propertyStart = false
			if tokens[index].text != "id" || index+2 >= len(tokens) ||
				tokens[index+1].text != ":" || !tokens[index+2].literal {
				continue
			}
			return tokens[index+2].text, true
		}
	}
	return "", false
}

func tokenizePluginSource(source []byte) []sourceToken {
	tokens := make([]sourceToken, 0, len(source)/4)
	for index := 0; index < len(source); {
		switch {
		case isSourceSpace(source[index]):
			index++
		case source[index] == '/' && index+1 < len(source) && source[index+1] == '/':
			index += 2
			for index < len(source) && source[index] != '\n' {
				index++
			}
		case source[index] == '/' && index+1 < len(source) && source[index+1] == '*':
			index += 2
			for index+1 < len(source) && (source[index] != '*' || source[index+1] != '/') {
				index++
			}
			if index+1 < len(source) {
				index += 2
			}
		case source[index] == '`':
			index = skipSourceString(source, index, '`')
		case source[index] == '\'' || source[index] == '"':
			quote := source[index]
			start := index + 1
			end, escaped := scanSourceString(source, index, quote)
			if !escaped && end <= len(source) && end > start {
				tokens = append(tokens, sourceToken{text: string(source[start : end-1]), literal: true})
			}
			index = end
		case isSourceIdentifierStart(source[index]):
			start := index
			index++
			for index < len(source) && isSourceIdentifierPart(source[index]) {
				index++
			}
			tokens = append(tokens, sourceToken{text: string(source[start:index])})
		default:
			tokens = append(tokens, sourceToken{text: string(source[index])})
			index++
		}
	}
	return tokens
}

func skipSourceString(source []byte, start int, quote byte) int {
	end, _ := scanSourceString(source, start, quote)
	return end
}

func scanSourceString(source []byte, start int, quote byte) (int, bool) {
	escaped := false
	for index := start + 1; index < len(source); index++ {
		if source[index] == '\\' {
			escaped = true
			index++
			continue
		}
		if source[index] == quote {
			return index + 1, escaped
		}
	}
	return len(source), escaped
}

func isSourceSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func isSourceIdentifierStart(value byte) bool {
	return value == '_' || value == '$' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func isSourceIdentifierPart(value byte) bool {
	return isSourceIdentifierStart(value) || value >= '0' && value <= '9'
}

func selectedTUIPluginIdentityResolver(configDir string, selected map[string]bool) pluginIdentityResolver {
	resolveTUIIdentity := tuiPluginIdentityResolver(configDir)
	return func(value any) string {
		identity := resolveTUIIdentity(value)
		if selected[identity] {
			return identity
		}
		return pluginIdentity(value)
	}
}

func tuiPluginBundlePath(configDir, entry string) (string, bool) {
	candidate := entry
	if strings.HasPrefix(entry, "file:") {
		parsed, err := url.Parse(entry)
		if err != nil || parsed.Scheme != "file" {
			return "", false
		}
		switch {
		case parsed.Opaque != "":
			candidate = parsed.Opaque
		case parsed.Host == "" || parsed.Host == "localhost":
			candidate = parsed.Path
		default:
			return "", false
		}
		candidate, err = url.PathUnescape(candidate)
		if err != nil {
			return "", false
		}
	} else if !filepath.IsAbs(entry) && !isRelativePluginPath(entry) {
		return "", false
	}

	candidate = filepath.FromSlash(candidate)
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(configDir, candidate)
	}
	return filepath.Clean(candidate), true
}

func isRelativePluginPath(entry string) bool {
	if entry == "" || strings.HasPrefix(entry, "@") || strings.Contains(entry, "://") {
		return false
	}
	return strings.HasPrefix(entry, "./") ||
		strings.HasPrefix(entry, "../") ||
		strings.HasPrefix(entry, `.\`) ||
		strings.HasPrefix(entry, `..\`) ||
		strings.ContainsAny(entry, `/\`)
}
