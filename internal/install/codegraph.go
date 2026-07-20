package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const codegraphPackage = "@colbymchenry/codegraph@latest"

func applyCodegraph(assetsDir, configDir string) ([]string, error) {
	integrationDir := filepath.Join(assetsDir, "integrations", "codegraph")

	rawPatch, err := os.ReadFile(filepath.Join(integrationDir, "mcp.json"))
	if err != nil {
		return nil, fmt.Errorf("reading CodeGraph MCP config: %w", err)
	}
	var patch map[string]any
	if err := json.Unmarshal(rawPatch, &patch); err != nil {
		return nil, fmt.Errorf("parsing CodeGraph MCP config: %w", err)
	}
	guidance, err := os.ReadFile(filepath.Join(integrationDir, "AGENTS.md"))
	if err != nil {
		return nil, fmt.Errorf("reading CodeGraph agent guidance: %w", err)
	}

	installLine, err := ensureCodegraphInstalled()
	if err != nil {
		return nil, err
	}
	done := []string{installLine}

	opencodePath := filepath.Join(configDir, "opencode.json")
	lines, err := mergeJSON(opencodePath, "https://opencode.ai/config.json", []map[string]any{patch})
	done = append(done, "merge     codegraph MCP")
	done = append(done, lines...)
	if err != nil {
		return done, err
	}

	agentsPath := filepath.Join(configDir, "AGENTS.md")
	if err := upsertCodegraphGuidance(agentsPath, string(guidance)); err != nil {
		return done, err
	}
	done = append(done, "actualizado "+agentsPath)
	return done, nil
}

func ensureCodegraphInstalled() (string, error) {
	if path, err := exec.LookPath("codegraph"); err == nil {
		return "disponible " + path, nil
	}

	npm, err := exec.LookPath("npm")
	if err != nil {
		return "", fmt.Errorf("CodeGraph is selected but neither codegraph nor npm is available on PATH")
	}
	cmd := exec.Command(npm, "install", "--global", codegraphPackage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return "", fmt.Errorf("installing CodeGraph with npm: %w", err)
		}
		return "", fmt.Errorf("installing CodeGraph with npm: %w: %s", err, detail)
	}
	if _, err := exec.LookPath("codegraph"); err != nil {
		return "", fmt.Errorf("CodeGraph was installed but codegraph is not available on PATH")
	}
	return "instalado  " + codegraphPackage, nil
}

func upsertCodegraphGuidance(path, guidance string) error {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(raw)
	for _, markers := range [][2]string{
		{"<!-- codegraph-guidance -->", "<!-- /codegraph-guidance -->"},
		{"<!-- gentle-ai:codegraph-guidance -->", "<!-- /gentle-ai:codegraph-guidance -->"},
	} {
		content = removeManagedBlock(content, markers[0], markers[1])
	}
	content = strings.TrimSpace(content)
	guidance = strings.TrimSpace(guidance)
	if content != "" {
		content += "\n\n"
	}
	content += guidance + "\n"

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func removeManagedBlock(content, startMarker, endMarker string) string {
	for {
		start := strings.Index(content, startMarker)
		if start < 0 {
			return content
		}
		endOffset := strings.Index(content[start+len(startMarker):], endMarker)
		if endOffset < 0 {
			return content
		}
		end := start + len(startMarker) + endOffset + len(endMarker)
		before := strings.TrimRight(content[:start], "\n")
		after := strings.TrimLeft(content[end:], "\n")
		switch {
		case before == "":
			content = after
		case after == "":
			content = before
		default:
			content = before + "\n\n" + after
		}
	}
}
